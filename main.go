package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type activeView int

const (
	viewMain activeView = iota
	viewAdd
	viewEdit
	viewConfirmDelete
)

type model struct {
	habits            []*Habit
	selectedIdx       int
	view              activeView
	db                *Database
	cfg               Config
	width             int
	height            int
	highlightedDayIdx int // 0 to 6 representing the day index in the last 7 days sparkline (6 is today)

	// Form elements
	nameInput       textinput.Model
	descInput       textinput.Model
	gitInput        textinput.Model
	isPriorityField bool
	formFocus       int // 0: name, 1: desc, 2: git_dir, 3: priority, 4: save, 5: cancel

	// Timeline
	focusTimeline   bool
	timelineHourIdx int

	// Level system
	prevLevel    int
	levelUpToast string
	toastExpiry  time.Time
}

func initialModel(db *Database, cfg Config) model {
	// Initialize inputs
	nameIn := textinput.New()
	nameIn.Placeholder = "Drink Water"
	nameIn.CharLimit = 32
	nameIn.Width = 30

	descIn := textinput.New()
	descIn.Placeholder = "8 glasses (2 Liters) of water throughout the day"
	descIn.CharLimit = 100
	descIn.Width = 30

	gitIn := textinput.New()
	gitIn.Placeholder = "/path/to/local/git/repo"
	gitIn.CharLimit = 150
	gitIn.Width = 30

	habits, err := db.LoadHabits()
	if err != nil {
		habits = []*Habit{}
	}

	m := model{
		habits:            habits,
		selectedIdx:       0,
		view:              viewMain,
		db:                db,
		cfg:               cfg,
		nameInput:         nameIn,
		descInput:         descIn,
		gitInput:          gitIn,
		isPriorityField:   false,
		focusTimeline:     false,
		timelineHourIdx:   0,
		formFocus:         0,
		highlightedDayIdx: 6, // default focus is today (last element of 7-day view)
	}
	level, _, _, _ := m.getLevelStats()
	m.prevLevel = level
	return m
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case gitUpdateMsg:
		for _, h := range m.habits {
			if h.ID == msg.HabitID {
				h.History[time.Now().Format("2006-01-02")] = msg.Completed
				break
			}
		}
		return m, nil

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		m.levelUpToast = ""
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "q":
			if m.view == viewMain && !m.focusTimeline {
				return m, tea.Quit
			}
		}
	}

	// Route updates based on current view
	switch m.view {
	case viewMain:
		m, cmd = m.updateMain(msg)
		cmds = append(cmds, cmd)
	case viewAdd, viewEdit:
		m, cmd = m.updateForm(msg)
		cmds = append(cmds, cmd)
	case viewConfirmDelete:
		m, cmd = m.updateDelete(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m model) updateMain(msg tea.Msg) (model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		keyStr := msg.String()

		if m.focusTimeline {
			switch keyStr {
			case "esc", "q", "t":
				m.focusTimeline = false
				return m, nil
			case "left", "h":
				if m.timelineHourIdx > 0 {
					m.timelineHourIdx--
				}
			case "right", "l":
				if m.timelineHourIdx < 14 {
					m.timelineHourIdx++
				}
			case " ", "enter":
				if len(m.habits) > 0 {
					h := m.habits[m.selectedIdx]
					todayStr := time.Now().Format("2006-01-02")
					hour := 8 + m.timelineHourIdx

					if h.TimeBlocks == nil {
						h.TimeBlocks = make(map[string]map[int]bool)
					}
					if h.TimeBlocks[todayStr] == nil {
						h.TimeBlocks[todayStr] = make(map[int]bool)
					}

					active := !h.TimeBlocks[todayStr][hour]
					h.TimeBlocks[todayStr][hour] = active
					_ = m.db.SaveTimeBlock(h.ID, todayStr, hour, active)

					oldLevel, _, _, _ := m.getLevelStats()

					if active {
						h.History[todayStr] = true
						_ = m.db.SaveToggle(h.ID, todayStr, true)
					} else {
						// Check if any other time block is active for today
						anyActive := false
						for _, act := range h.TimeBlocks[todayStr] {
							if act {
								anyActive = true
								break
							}
						}
						if !anyActive {
							h.History[todayStr] = false
							_ = m.db.SaveToggle(h.ID, todayStr, false)
						}
					}

					newLevel, _, _, _ := m.getLevelStats()
					if newLevel > oldLevel {
						m.levelUpToast = fmt.Sprintf("🎉 LEVEL UP! Reached Level %d! 🎉", newLevel)
						m.toastExpiry = time.Now().Add(5 * time.Second)
						triggerNotification("Level Up! 🎉", fmt.Sprintf("Congratulations! You reached Level %d!", newLevel))
					} else if newLevel < oldLevel {
						m.levelUpToast = ""
					}
				}
			}
			return m, nil
		}

		switch keyStr {
		case "up", "k":
			if m.selectedIdx > 0 {
				m.selectedIdx--
				m.highlightedDayIdx = 6 // Reset to today
			}
		case "down", "j":
			if m.selectedIdx < len(m.habits)-1 {
				m.selectedIdx++
				m.highlightedDayIdx = 6 // Reset to today
			}
		case "left", "h":
			if m.highlightedDayIdx > 0 {
				m.highlightedDayIdx--
			}
		case "right", "l":
			if m.highlightedDayIdx < 6 {
				m.highlightedDayIdx++
			}
		case " ":
			if len(m.habits) > 0 {
				h := m.habits[m.selectedIdx]
				days := getLast7Days()
				targetDayStr := days[m.highlightedDayIdx].Format("2006-01-02")

				oldLevel, _, _, _ := m.getLevelStats()

				h.History[targetDayStr] = !h.History[targetDayStr]
				_ = m.db.SaveToggle(h.ID, targetDayStr, h.History[targetDayStr])

				newLevel, _, _, _ := m.getLevelStats()
				if newLevel > oldLevel {
					m.levelUpToast = fmt.Sprintf("🎉 LEVEL UP! Reached Level %d! 🎉", newLevel)
					m.toastExpiry = time.Now().Add(5 * time.Second)
					triggerNotification("Level Up! 🎉", fmt.Sprintf("Congratulations! You reached Level %d!", newLevel))
				} else if newLevel < oldLevel {
					m.levelUpToast = ""
				}
			}
		case "t":
			if len(m.habits) > 0 {
				m.focusTimeline = true
				m.timelineHourIdx = 0
			}
		case "n":
			m.view = viewAdd
			m.nameInput.SetValue("")
			m.descInput.SetValue("")
			m.gitInput.SetValue("")
			m.isPriorityField = false
			m.formFocus = 0
			m.nameInput.Focus()
			m.descInput.Blur()
			m.gitInput.Blur()
		case "e":
			if len(m.habits) > 0 {
				m.view = viewEdit
				h := m.habits[m.selectedIdx]
				m.nameInput.SetValue(h.Name)
				m.descInput.SetValue(h.Description)
				m.gitInput.SetValue(h.GitDir)
				m.isPriorityField = h.IsPriority
				m.formFocus = 0
				m.nameInput.Focus()
				m.descInput.Blur()
				m.gitInput.Blur()
			}
		case "d", "x":
			if len(m.habits) > 0 {
				m.view = viewConfirmDelete
			}
		}
	}
	return m, nil
}

func (m model) updateForm(msg tea.Msg) (model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			m.view = viewMain
			return m, nil
		case "tab", "down":
			m.formFocus = (m.formFocus + 1) % 6
			m.updateFormFocus()
		case "shift+tab", "up":
			m.formFocus = (m.formFocus - 1 + 6) % 6
			m.updateFormFocus()
		case "enter":
			if m.formFocus == 0 {
				m.formFocus = 1
				m.updateFormFocus()
			} else if m.formFocus == 1 {
				m.formFocus = 2
				m.updateFormFocus()
			} else if m.formFocus == 2 {
				m.formFocus = 3
				m.updateFormFocus()
			} else if m.formFocus == 3 {
				m.isPriorityField = !m.isPriorityField
			} else if m.formFocus == 4 {
				m.submitForm()
				m.view = viewMain
			} else if m.formFocus == 5 {
				m.view = viewMain
			}
		case " ":
			if m.formFocus == 3 {
				m.isPriorityField = !m.isPriorityField
			}
		}
	}

	// Update active inputs
	if m.formFocus == 0 {
		m.nameInput, cmd = m.nameInput.Update(msg)
	} else if m.formFocus == 1 {
		m.descInput, cmd = m.descInput.Update(msg)
	} else if m.formFocus == 2 {
		m.gitInput, cmd = m.gitInput.Update(msg)
	}

	return m, cmd
}

func (m *model) updateFormFocus() {
	m.nameInput.Blur()
	m.descInput.Blur()
	m.gitInput.Blur()

	switch m.formFocus {
	case 0:
		m.nameInput.Focus()
	case 1:
		m.descInput.Focus()
	case 2:
		m.gitInput.Focus()
	}
}

func (m *model) submitForm() {
	name := strings.TrimSpace(m.nameInput.Value())
	desc := strings.TrimSpace(m.descInput.Value())
	gitDir := strings.TrimSpace(m.gitInput.Value())
	if name == "" {
		return // Name is required
	}

	if m.view == viewAdd {
		id := fmt.Sprintf("%d", time.Now().UnixNano())
		h := &Habit{
			ID:          id,
			Name:        name,
			Description: desc,
			GitDir:      gitDir,
			IsPriority:  m.isPriorityField,
			CreatedAt:   time.Now(),
			History:     make(map[string]bool),
			TimeBlocks:  make(map[string]map[int]bool),
		}
		m.habits = append(m.habits, h)
		m.selectedIdx = len(m.habits) - 1
		_ = m.db.AddHabit(h)
	} else if m.view == viewEdit {
		if len(m.habits) > 0 {
			h := m.habits[m.selectedIdx]
			h.Name = name
			h.Description = desc
			h.GitDir = gitDir
			h.IsPriority = m.isPriorityField
			_ = m.db.UpdateHabit(h)
		}
	}
}

func (m model) updateDelete(msg tea.Msg) (model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "y", "Y", "enter":
			if len(m.habits) > 0 {
				h := m.habits[m.selectedIdx]
				_ = m.db.DeleteHabit(h.ID)

				m.habits = append(m.habits[:m.selectedIdx], m.habits[m.selectedIdx+1:]...)
				if m.selectedIdx >= len(m.habits) && m.selectedIdx > 0 {
					m.selectedIdx--
				}
			}
			m.view = viewMain
		case "n", "N", "esc":
			m.view = viewMain
		}
	}
	return m, nil
}

func (m model) getLevelStats() (int, int, int, float64) {
	totalCompletions := 0
	totalStreaks := 0
	for _, h := range m.habits {
		totalCompletions += h.GetTotalCompletions()
		currStreak, _ := h.GetStreaks()
		totalStreaks += currStreak
	}

	// 10 XP per completion, 5 XP per day of active streak
	totalXP := (totalCompletions * 10) + (totalStreaks * 5)
	
	// Level formula: Level = (XP / 100) + 1
	level := (totalXP / 100) + 1
	xpInLevel := totalXP % 100
	nextLevelXP := 100
	percent := float64(xpInLevel) / float64(nextLevelXP)
	
	return level, xpInLevel, nextLevelXP, percent
}

func renderProgressBar(percent float64, width int) string {
	filledWidth := int(percent * float64(width))
	if filledWidth < 0 {
		filledWidth = 0
	}
	if filledWidth > width {
		filledWidth = width
	}
	
	filledChar := "█"
	emptyChar := "░"
	
	filled := strings.Repeat(filledChar, filledWidth)
	empty := strings.Repeat(emptyChar, width-filledWidth)
	
	return lipgloss.NewStyle().Foreground(greenColor).Render(filled) + 
		lipgloss.NewStyle().Foreground(darkGrayColor).Render(empty)
}

func (m model) renderWelcomeView() string {
	welcomeTitle := lipgloss.NewStyle().
		Foreground(cyanColor).
		Bold(true).
		Underline(true).
		Render("✨ WELCOME TO GOHABIT ✨")

	welcomeSubtitle := lipgloss.NewStyle().
		Foreground(whiteColor).
		Render("Your journey to building better habits starts today!")

	rewardSection := lipgloss.NewStyle().
		Foreground(lightGrayColor).
		Render("Track daily habits • Build streaks • Earn XP • Level Up!")

	actionButton := focusedButtonStyle.Render("[ Press 'n' to create your first habit! ]")

	cardContent := lipgloss.JoinVertical(lipgloss.Center,
		welcomeTitle,
		"\n",
		welcomeSubtitle,
		"\n",
		rewardSection,
		"\n\n",
		actionButton,
	)

	cardWidth := 68
	if m.width < cardWidth {
		cardWidth = m.width - 4
		if cardWidth < 40 {
			cardWidth = 40
		}
	}

	card := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(navyColor).
		Padding(2, 5).
		Width(cardWidth).
		Align(lipgloss.Center).
		Render(cardContent)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, card)
}

func (m model) View() string {
	if m.width == 0 || m.height == 0 {
		return "Initializing habit tracker..."
	}

	if len(m.habits) == 0 && m.view == viewMain {
		return m.renderWelcomeView()
	}

	var content string

	switch m.view {
	case viewMain:
		content = m.renderMainView()
	case viewAdd:
		content = m.renderFormView("=== CREATE NEW HABIT ===")
	case viewEdit:
		content = m.renderFormView("=== EDIT HABIT ===")
	case viewConfirmDelete:
		content = m.renderDeleteConfirmation()
	}

	level, xp, nextXP, xpPercent := m.getLevelStats()
	xpBar := renderProgressBar(xpPercent, 12)
	
	levelStr := fmt.Sprintf(" ⭐ Level %d  [%s]  %d/%d XP ", level, xpBar, xp, nextXP)
	levelStr = lipgloss.NewStyle().Foreground(cyanColor).Bold(true).Render(levelStr)

	logoBadge := lipgloss.NewStyle().
		Foreground(whiteColor).
		Background(navyColor).
		Padding(0, 1).
		Bold(true).
		Render(" GOHABIT ")

	var completedToday, totalHabits int
	todayStr := time.Now().Format("2006-01-02")
	for _, h := range m.habits {
		totalHabits++
		if h.History[todayStr] {
			completedToday++
		}
	}
	
	statsStr := fmt.Sprintf(" Today: %d/%d ", completedToday, totalHabits)
	if totalHabits > 0 {
		statsStr += fmt.Sprintf("(%.0f%%)", float64(completedToday)/float64(totalHabits)*100)
	}
	statsStrBadge := lipgloss.NewStyle().
		Foreground(whiteColor).
		Background(darkGrayColor).
		Padding(0, 1).
		Bold(true).
		Render(statsStr)

	fullHeader := lipgloss.JoinHorizontal(lipgloss.Center, logoBadge, "  ", levelStr, "  ", statsStrBadge)
	
	var headerLayout string
	if m.levelUpToast != "" {
		toastMsg := lipgloss.NewStyle().
			Foreground(cyanColor).
			Background(navyColor).
			Bold(true).
			Padding(0, 2).
			Render(m.levelUpToast)
		headerLayout = lipgloss.JoinVertical(lipgloss.Center, fullHeader, "\n", toastMsg)
	} else {
		headerLayout = fullHeader
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		"\n",
		headerLayout,
		"\n",
		content,
		"\n",
		m.renderHelp(),
	)
}

func (m model) renderMainView() string {
	if len(m.habits) == 0 {
		return m.renderWelcomeView()
	}

	// Render left panel (habit list)
	var listItems []string
	for idx, h := range m.habits {
		var item string
		isSelected := idx == m.selectedIdx

		todayStr := time.Now().Format("2006-01-02")
		completedToday := h.History[todayStr]

		checkbox := "[ ]"
		if completedToday {
			checkbox = completedStyle.Render("[✓]")
		} else {
			checkbox = pendingStyle.Render("[ ]")
		}

		title := h.Name
		if h.IsPriority {
			title = title + " (!)"
		}
		if isSelected {
			title = selectedHabitStyle.Render(title)
		} else {
			if h.IsPriority {
				title = lipgloss.NewStyle().Foreground(redColor).Bold(true).Render(title)
			} else {
				title = habitTitleStyle.Render(title)
			}
		}

		currentStreak, _ := h.GetStreaks()
		streakStr := ""
		if currentStreak > 0 {
			streakStr = streakStyle.Render(fmt.Sprintf(" [Streak: %d]", currentStreak))
		}

		selector := "  "
		if isSelected {
			selector = lipgloss.NewStyle().Foreground(cyanColor).Render("* ")
		}

		firstRow := fmt.Sprintf("%s%s %s%s", selector, checkbox, title, streakStr)
		historyLine := "    " + renderHistorySparkline(h, isSelected, m.highlightedDayIdx)

		item = firstRow + "\n" + historyLine

		listItems = append(listItems, item)
	}

	selectedHabit := m.habits[m.selectedIdx]

	// Render right panel (selected habit details)
	detailTitle := detailTitleStyle.Render(selectedHabit.Name)
	detailDesc := selectedHabit.Description
	if detailDesc == "" {
		detailDesc = "No description provided."
	}
	detailDesc = lipgloss.NewStyle().Foreground(lightGrayColor).Render(detailDesc)

	currStreak, longStreak := selectedHabit.GetStreaks()
	totalDone := selectedHabit.GetTotalCompletions()
	rate := selectedHabit.GetCompletionRate()

	card1 := statCardStyle.Render(fmt.Sprintf("🔥 %s\n%s", statValStyle.Render(fmt.Sprintf("%d days", currStreak)), statLabelStyle.Render("Current Streak")))
	card2 := statCardStyle.Render(fmt.Sprintf("🏆 %s\n%s", statValStyle.Render(fmt.Sprintf("%d days", longStreak)), statLabelStyle.Render("Longest Streak")))
	card3 := statCardStyle.Render(fmt.Sprintf("✨ %s\n%s", statValStyle.Render(fmt.Sprintf("%d times", totalDone)), statLabelStyle.Render("Total Completed")))
	card4 := statCardStyle.Render(fmt.Sprintf("📈 %s\n%s", statValStyle.Render(fmt.Sprintf("%.0f%%", rate)), statLabelStyle.Render("Success Rate")))

	statsGrid := lipgloss.JoinVertical(lipgloss.Left,
		lipgloss.JoinHorizontal(lipgloss.Top, card1, card2),
		"\n",
		lipgloss.JoinHorizontal(lipgloss.Top, card3, card4),
	)

	// Hover/selected day description info
	days := getLast7Days()
	hlDay := days[m.highlightedDayIdx]
	hlDateStr := hlDay.Format("2006-01-02")
	hlStatus := "Pending"
	if selectedHabit.History[hlDateStr] {
		hlStatus = "Completed"
	}
	hoverInfo := fmt.Sprintf("📅 Selected Day: %s (%s)", hlDay.Format("Monday, Jan 2"), hlStatus)
	hoverInfoRendered := lipgloss.NewStyle().Foreground(cyanColor).Bold(true).Render(hoverInfo)

	calendarTitle := lipgloss.NewStyle().Foreground(whiteColor).Bold(true).Render("=== 30-DAY CALENDAR GRID ===")
	calendarView := renderCalendar(selectedHabit, m.cfg.WeekStart)

	rightContent := lipgloss.JoinVertical(lipgloss.Left,
		detailTitle,
		"\n",
		detailDesc,
		"\n\n",
		statsGrid,
		"\n\n",
		hoverInfoRendered,
		"\n",
		calendarTitle,
		calendarView,
	)

	if m.width < 96 {
		panelWidth := m.width - 6
		if panelWidth < 45 {
			panelWidth = 45
		}

		listHeight := len(m.habits)*2 + 1
		if listHeight > 10 {
			listHeight = 10
		}
		if listHeight < 4 {
			listHeight = 4
		}

		leftListRendered := leftPanelStyle.Width(panelWidth).Height(listHeight).Render(
			strings.Join(listItems, "\n\n"),
		)
		timelineRendered := m.renderTimeline(selectedHabit, panelWidth)
		leftPanel := lipgloss.JoinVertical(lipgloss.Left, leftListRendered, "\n", timelineRendered)

		rightPanel := rightPanelStyle.Width(panelWidth).Height(18).Render(rightContent)
		return lipgloss.JoinVertical(lipgloss.Left, leftPanel, "\n", rightPanel)
	} else {
		leftWidth := m.width/2 - 3
		if leftWidth < 45 {
			leftWidth = 45
		}
		rightWidth := m.width - leftWidth - 6
		if rightWidth < 45 {
			rightWidth = 45
		}

		listHeight := m.height - 19
		if listHeight < 5 {
			listHeight = 5
		}

		leftListRendered := leftPanelStyle.Width(leftWidth).Height(listHeight).Render(
			strings.Join(listItems, "\n\n"),
		)
		timelineRendered := m.renderTimeline(selectedHabit, leftWidth)
		leftPanel := lipgloss.JoinVertical(lipgloss.Left, leftListRendered, "\n", timelineRendered)

		rightPanel := rightPanelStyle.Width(rightWidth).Height(m.height - 10).Render(rightContent)
		return lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, rightPanel)
	}
}

func (m model) renderFormView(title string) string {
	formTitle := formTitleStyle.Render(title)

	var nameView, descView, gitView, priorityView string
	if m.formFocus == 0 {
		nameView = focusedInputStyle.Render(m.nameInput.View())
	} else {
		nameView = inputStyle.Render(m.nameInput.View())
	}

	if m.formFocus == 1 {
		descView = focusedInputStyle.Render(m.descInput.View())
	} else {
		descView = inputStyle.Render(m.descInput.View())
	}

	if m.formFocus == 2 {
		gitView = focusedInputStyle.Render(m.gitInput.View())
	} else {
		gitView = inputStyle.Render(m.gitInput.View())
	}

	priorityText := "[ ] No (Press Space/Enter to toggle)"
	if m.isPriorityField {
		priorityText = "[X] Yes (Press Space/Enter to toggle)"
	}
	if m.formFocus == 3 {
		priorityView = focusedInputStyle.Render(priorityText)
	} else {
		priorityView = inputStyle.Render(priorityText)
	}

	var saveBtn, cancelBtn string
	if m.formFocus == 4 {
		saveBtn = focusedButtonStyle.Render("[ Save ]")
	} else {
		saveBtn = buttonStyle.Render("[ Save ]")
	}

	if m.formFocus == 5 {
		cancelBtn = focusedButtonStyle.Render("[ Cancel ]")
	} else {
		cancelBtn = buttonStyle.Render("[ Cancel ]")
	}

	buttons := lipgloss.JoinHorizontal(lipgloss.Top, saveBtn, "  ", cancelBtn)

	formContent := lipgloss.JoinVertical(lipgloss.Left,
		formTitle,
		"\n",
		lipgloss.NewStyle().Foreground(whiteColor).Bold(true).Render("Habit Name:"),
		nameView,
		"\n",
		lipgloss.NewStyle().Foreground(whiteColor).Bold(true).Render("Description:"),
		descView,
		"\n",
		lipgloss.NewStyle().Foreground(whiteColor).Bold(true).Render("Git Directory (Optional):"),
		gitView,
		"\n",
		lipgloss.NewStyle().Foreground(whiteColor).Bold(true).Render("Priority:"),
		priorityView,
		"\n\n",
		buttons,
	)

	formWidth := 48
	if m.width < formWidth {
		formWidth = m.width - 4
		if formWidth < 30 {
			formWidth = 30
		}
	}

	card := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(navyColor).
		Padding(2, 4).
		Width(formWidth).
		Render(formContent)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, card)
}

func (m model) renderDeleteConfirmation() string {
	selectedHabit := m.habits[m.selectedIdx]

	title := lipgloss.NewStyle().
		Foreground(redColor).
		Bold(true).
		Render("=== DELETE HABIT ===")

	msg := fmt.Sprintf("Are you sure you want to delete '%s'?\nThis action cannot be undone.", selectedHabit.Name)
	msgView := lipgloss.NewStyle().Foreground(whiteColor).Render(msg)

	yesBtn := focusedButtonStyle.Render("y - Yes, Delete")
	noBtn := buttonStyle.Render("n - Cancel")

	buttons := lipgloss.JoinHorizontal(lipgloss.Top, yesBtn, "  ", noBtn)

	content := lipgloss.JoinVertical(lipgloss.Left,
		title,
		"\n",
		msgView,
		"\n\n",
		buttons,
	)

	deleteWidth := 44
	if m.width < deleteWidth {
		deleteWidth = m.width - 4
		if deleteWidth < 30 {
			deleteWidth = 30
		}
	}

	deleteCard := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(redColor).
		Padding(2, 4).
		Width(deleteWidth).
		Render(content)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, deleteCard)
}

func (m model) renderHelp() string {
	var keys []string
	switch m.view {
	case viewMain:
		if m.focusTimeline {
			keys = []string{
				"←/→ or h/l: navigate hours",
				"space/enter: toggle hour block",
				"t/esc/q: exit timeline",
			}
		} else {
			keys = []string{
				"↑/↓ or k/j: select habit",
				"←/→ or h/l: navigate sparkline days",
				"space: toggle highlighted day",
				"t: toggle timeline",
				"n: add new",
				"e: edit selected",
				"d/x: delete selected",
				"q: quit",
			}
		}
	case viewAdd, viewEdit:
		keys = []string{
			"tab/shift+tab: move focus",
			"enter: submit/select/toggle",
			"esc: cancel",
		}
	case viewConfirmDelete:
		keys = []string{
			"y: delete",
			"n/esc: cancel",
		}
	}

	return helpStyle.Render("Controls: " + strings.Join(keys, "  |  "))
}

func getLast7Days() []time.Time {
	days := make([]time.Time, 7)
	now := time.Now()
	for i := 0; i < 7; i++ {
		days[6-i] = now.AddDate(0, 0, -i)
	}
	return days
}

func renderHistorySparkline(h *Habit, isSelectedHabit bool, highlightedIdx int) string {
	days := getLast7Days()
	var parts []string
	
	for idx, d := range days {
		dateStr := d.Format("2006-01-02")
		completed := h.History[dateStr]

		icon := "○"
		if completed {
			icon = "●"
		}

		var style lipgloss.Style
		if completed {
			style = completedStyle
		} else {
			style = pendingStyle
		}

		if isSelectedHabit && idx == highlightedIdx {
			parts = append(parts, lipgloss.NewStyle().Foreground(cyanColor).Bold(true).Render(fmt.Sprintf("[%s]", icon)))
		} else {
			parts = append(parts, style.Render(icon))
		}
	}
	return strings.Join(parts, "  ")
}

func getCalendarGrid(weekStart int) [][]time.Time {
	now := time.Now()
	wd := now.Weekday()

	var daysToStart int
	if weekStart == 1 { // Monday start
		daysToStart = int(wd) - 1
		if wd == time.Sunday {
			daysToStart = 6
		}
	} else { // Sunday start
		daysToStart = int(wd)
	}

	// Start date is the start day of 4 weeks ago (5 weeks total)
	startDate := now.AddDate(0, 0, -daysToStart-28)

	grid := make([][]time.Time, 5)
	for w := 0; w < 5; w++ {
		week := make([]time.Time, 7)
		for d := 0; d < 7; d++ {
			week[d] = startDate.AddDate(0, 0, w*7+d)
		}
		grid[w] = week
	}
	return grid
}

func renderCalendar(h *Habit, weekStart int) string {
	grid := getCalendarGrid(weekStart)
	now := time.Now()
	todayStr := now.Format("2006-01-02")

	var sb strings.Builder
	if weekStart == 1 {
		sb.WriteString("  Mo  Tu  We  Th  Fr  Sa  Su\n")
	} else {
		sb.WriteString("  Su  Mo  Tu  We  Th  Fr  Sa\n")
	}

	for w, week := range grid {
		sb.WriteString("  ")
		for _, d := range week {
			dateStr := d.Format("2006-01-02")

			if dateStr > todayStr {
				// Future day
				sb.WriteString(calendarFutureStyle.Render("·") + "   ")
			} else {
				completed := h.History[dateStr]
				if completed {
					sb.WriteString(calendarDoneStyle.Render("■") + "   ")
				} else {
					sb.WriteString(calendarMissStyle.Render("□") + "   ")
				}
			}
		}
		if w < 4 {
			sb.WriteString("\n")
		}
	}
	return sb.String()
}

const version = "0.0.1"

type gitUpdateMsg struct {
	HabitID   string
	Completed bool
}

func expandPath(path string) string {
	if strings.HasPrefix(path, "~") {
		home, err := os.UserHomeDir()
		if err == nil {
			if path == "~" {
				return home
			} else if strings.HasPrefix(path, "~/") {
				return filepath.Join(home, path[2:])
			}
		}
	}
	return path
}

func getLatestGitCommit(dir string) (string, error) {
	if dir == "" {
		return "", nil
	}
	dir = expandPath(dir)
	todayStr := time.Now().Format("2006-01-02")
	cmd := exec.Command("git", "log", "-n", "1", "--since="+todayStr+" 00:00:00", "--format=%H")
	cmd.Dir = dir
	output, err := cmd.Output()
	if err != nil {
		return "", nil // Ignore error (e.g. not a git repo) and assume no commits
	}
	return strings.TrimSpace(string(output)), nil
}

func triggerNotification(title, msg string) {
	// Try notify-send (Linux)
	_, err := exec.LookPath("notify-send")
	if err == nil {
		_ = exec.Command("notify-send", title, msg).Run()
		return
	}

	// Try osascript (macOS)
	_, err = exec.LookPath("osascript")
	if err == nil {
		script := fmt.Sprintf(`display notification "%s" with title "%s"`, msg, title)
		_ = exec.Command("osascript", "-e", script).Run()
		return
	}
}

func monitorGitCommits(p *tea.Program, db *Database, eveningHour int) {
	ticker := time.NewTicker(15 * time.Second)
	var lastNotificationDate string
	lastProcessedCommits := make(map[string]string)

	for {
		habits, err := db.LoadHabits()
		if err == nil {
			todayStr := time.Now().Format("2006-01-02")

			// 1. Check Git dirs
			for _, h := range habits {
				if h.GitDir != "" {
					commitHash, err := getLatestGitCommit(h.GitDir)
					if err == nil && commitHash != "" {
						lastCommit := lastProcessedCommits[h.ID]
						if commitHash != lastCommit {
							if !h.History[todayStr] {
								_ = db.SaveToggle(h.ID, todayStr, true)
								p.Send(gitUpdateMsg{
									HabitID:   h.ID,
									Completed: true,
								})
							}
							lastProcessedCommits[h.ID] = commitHash
						}
					}
				}
			}

			// 2. Check evening hour notifications
			now := time.Now()
			if now.Hour() >= eveningHour && lastNotificationDate != todayStr {
				pendingPriorityCount := 0
				for _, h := range habits {
					if h.IsPriority && !h.History[todayStr] {
						pendingPriorityCount++
					}
				}
				if pendingPriorityCount > 0 {
					var msg string
					if pendingPriorityCount == 1 {
						msg = "You have 1 pending high-priority habit!"
					} else {
						msg = fmt.Sprintf("You have %d pending high-priority habits!", pendingPriorityCount)
					}
					triggerNotification("Gohabit Reminder", msg)
					lastNotificationDate = todayStr
				}
			}
		}

		select {
		case <-ticker.C:
		}
	}
}

func printStatusWarning(db *Database) {
	habits, err := db.LoadHabits()
	if err != nil {
		return
	}

	todayStr := time.Now().Format("2006-01-02")
	pendingCount := 0
	for _, h := range habits {
		if h.IsPriority && !h.History[todayStr] {
			pendingCount++
		}
	}

	if pendingCount > 0 {
		if pendingCount == 1 {
			fmt.Println("\033[1;31m[Gohabit] You have 1 pending high-priority habit today!\033[0m")
		} else {
			fmt.Printf("\033[1;31m[Gohabit] You have %d pending high-priority habits today!\033[0m\n", pendingCount)
		}
	}
}

func (m model) renderTimeline(h *Habit, width int) string {
	if h == nil {
		return ""
	}
	todayStr := time.Now().Format("2006-01-02")

	hasAnyBlocks := false
	if h.TimeBlocks != nil && h.TimeBlocks[todayStr] != nil {
		for _, active := range h.TimeBlocks[todayStr] {
			if active {
				hasAnyBlocks = true
				break
			}
		}
	}

	if !m.focusTimeline && !hasAnyBlocks {
		return lipgloss.NewStyle().
			Foreground(grayColor).
			Width(width).
			Align(lipgloss.Center).
			Render("[ Press 't' to open Hourly Time-Blocking ]")
	}

	var headerCells []string
	var valCells []string

	for i := 0; i < 15; i++ {
		hour := 8 + i
		hourStr := fmt.Sprintf("%02d", hour)

		active := false
		if h.TimeBlocks != nil && h.TimeBlocks[todayStr] != nil {
			active = h.TimeBlocks[todayStr][hour]
		}

		isFocused := m.focusTimeline && i == m.timelineHourIdx

		var hStr, bStr string
		if isFocused {
			hStr = lipgloss.NewStyle().Foreground(cyanColor).Bold(true).Underline(true).Render(hourStr)
			if active {
				bStr = lipgloss.NewStyle().Foreground(greenColor).Background(darkGrayColor).Bold(true).Render("[■]")
			} else {
				bStr = lipgloss.NewStyle().Foreground(cyanColor).Background(darkGrayColor).Bold(true).Render("[□]")
			}
		} else {
			hStr = lipgloss.NewStyle().Foreground(grayColor).Render(hourStr)
			if active {
				bStr = completedStyle.Render("[■]")
			} else {
				bStr = pendingStyle.Render("[□]")
			}
		}
		headerCells = append(headerCells, hStr)
		valCells = append(valCells, bStr)
	}

	var borderStyle lipgloss.Style
	if m.focusTimeline {
		borderStyle = leftPanelStyle.Copy().BorderForeground(navyColor)
	} else {
		borderStyle = leftPanelStyle.Copy().BorderForeground(darkGrayColor)
	}
	borderStyle = borderStyle.Width(width)

	timelineTitle := "⏰ TIME-BLOCKS (Today)"
	if m.focusTimeline {
		timelineTitle = "⏰ TIME-BLOCKS (Today) [Active - space to toggle]"
	}
	titleRendered := lipgloss.NewStyle().Foreground(whiteColor).Bold(true).Render(timelineTitle)

	headerRow := strings.Join(headerCells, " ")
	valRow := strings.Join(valCells, "")

	content := fmt.Sprintf("%s\n\n %s\n%s",
		titleRendered,
		headerRow,
		valRow,
	)

	return borderStyle.Render(content)
}

func printHelp() {
	fmt.Println(`Gohabit - A clean retro terminal habit tracker

Usage:
  gohabit                       Launch the interactive TUI dashboard
  gohabit <command> [args]      Run in CLI mode to perform quick actions

Commands:
  add "<name>" ["<desc>"]       Add a new habit with optional description
  check "<name>"                Toggle today's completion for a habit
  list                          List all habits, streaks, and today's status
  track status                  Show warning if priority habits are pending
  help                          Show this help documentation

Flags:
  -v, --version                 Print version information and exit
  -h, --help                    Print this help menu and exit

Configuration:
  Settings are loaded from 'config.ini' in the current directory.
  Completions are saved to the database specified in 'config.ini'.`)
}

func handleCLI(db *Database, args []string) int {
	subcommand := strings.ToLower(args[0])

	switch subcommand {
	case "help":
		printHelp()

	case "track":
		if len(args) < 2 || args[1] != "status" {
			fmt.Println("Usage: gohabit track status")
			return 1
		}
		printStatusWarning(db)

	case "add":
		if len(args) < 2 {
			fmt.Println("Error: Missing habit name.\nUsage: gohabit add \"<name>\" [\"<description>\"]")
			return 1
		}
		name := args[1]
		desc := ""
		if len(args) >= 3 {
			desc = args[2]
		}

		id := fmt.Sprintf("%d", time.Now().UnixNano())
		h := &Habit{
			ID:          id,
			Name:        name,
			Description: desc,
			CreatedAt:   time.Now(),
		}

		err := db.AddHabit(h)
		if err != nil {
			fmt.Printf("Error adding habit: %v\n", err)
			return 1
		}
		fmt.Printf("Habit '%s' added successfully.\n", name)

	case "check":
		if len(args) < 2 {
			fmt.Println("Error: Missing habit name.\nUsage: gohabit check \"<name>\"")
			return 1
		}
		nameQuery := args[1]
		habits, err := db.LoadHabits()
		if err != nil {
			fmt.Printf("Error loading habits: %v\n", err)
			return 1
		}

		var target *Habit
		for _, h := range habits {
			if strings.EqualFold(h.Name, nameQuery) || strings.Contains(strings.ToLower(h.Name), strings.ToLower(nameQuery)) {
				target = h
				break
			}
		}

		if target == nil {
			fmt.Printf("Error: No habit found matching '%s'.\n", nameQuery)
			return 1
		}

		todayStr := time.Now().Format("2006-01-02")
		newStatus := !target.History[todayStr]

		err = db.SaveToggle(target.ID, todayStr, newStatus)
		if err != nil {
			fmt.Printf("Error saving toggle: %v\n", err)
			return 1
		}

		if newStatus {
			fmt.Printf("Marked habit '%s' as COMPLETED for today!\n", target.Name)
		} else {
			fmt.Printf("Marked habit '%s' as PENDING for today.\n", target.Name)
		}

	case "list":
		habits, err := db.LoadHabits()
		if err != nil {
			fmt.Printf("Error loading habits: %v\n", err)
			return 1
		}

		if len(habits) == 0 {
			fmt.Println("No habits tracked yet. Run 'gohabit add \"<name>\"' to add one.")
			return 0
		}

		fmt.Printf("%-25s %-15s %-15s\n", "Habit Name", "Today's Status", "Streak")
		fmt.Println(strings.Repeat("-", 60))

		todayStr := time.Now().Format("2006-01-02")
		for _, h := range habits {
			status := "[ ] Pending"
			if h.History[todayStr] {
				status = "[✓] Completed"
			}
			currStreak, _ := h.GetStreaks()
			streakStr := fmt.Sprintf("%d days", currStreak)
			if currStreak == 0 {
				streakStr = "0 days"
			}
			fmt.Printf("%-25s %-15s %-15s\n", h.Name, status, streakStr)
		}

	default:
		fmt.Printf("Error: Unknown command '%s'\n\n", args[0])
		printHelp()
		return 1
	}
	return 0
}

func main() {
	// Set custom usage
	flag.Usage = printHelp

	// Parse version flags
	versionFlag := flag.Bool("v", false, "print version and exit")
	flag.BoolVar(versionFlag, "version", false, "print version and exit")
	flag.Parse()

	if *versionFlag {
		fmt.Printf("gohabit version %s\n", version)
		return
	}

	// Load configuration
	cfg, err := LoadConfig("config.ini")
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Initialize database from config database path
	db, err := NewDatabase(cfg.DBPath)
	if err != nil {
		fmt.Printf("Error initializing database: %v\n", err)
		os.Exit(1)
	}

	// Check if there are CLI arguments
	args := flag.Args()
	if len(args) > 0 {
		code := handleCLI(db, args)
		db.Close()
		os.Exit(code)
	}

	defer db.Close()

	// Initialize styles based on config theme colors
	InitStyles(cfg)

	p := tea.NewProgram(initialModel(db, cfg), tea.WithAltScreen())
	go monitorGitCommits(p, db, cfg.EveningHour)
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
