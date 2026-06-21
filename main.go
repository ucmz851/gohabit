package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"flag"

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
	nameInput textinput.Model
	descInput textinput.Model
	formFocus int // 0: name, 1: desc, 2: save, 3: cancel
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

	habits, err := db.LoadHabits()
	if err != nil {
		habits = []*Habit{}
	}

	return model{
		habits:            habits,
		selectedIdx:       0,
		view:              viewMain,
		db:                db,
		cfg:               cfg,
		nameInput:         nameIn,
		descInput:         descIn,
		formFocus:         0,
		highlightedDayIdx: 6, // default focus is today (last element of 7-day view)
	}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "q":
			if m.view == viewMain {
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
		switch msg.String() {
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

				h.History[targetDayStr] = !h.History[targetDayStr]
				_ = m.db.SaveToggle(h.ID, targetDayStr, h.History[targetDayStr])
			}
		case "n":
			m.view = viewAdd
			m.nameInput.SetValue("")
			m.descInput.SetValue("")
			m.formFocus = 0
			m.nameInput.Focus()
			m.descInput.Blur()
		case "e":
			if len(m.habits) > 0 {
				m.view = viewEdit
				h := m.habits[m.selectedIdx]
				m.nameInput.SetValue(h.Name)
				m.descInput.SetValue(h.Description)
				m.formFocus = 0
				m.nameInput.Focus()
				m.descInput.Blur()
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
			m.formFocus = (m.formFocus + 1) % 4
			m.updateFormFocus()
		case "shift+tab", "up":
			m.formFocus = (m.formFocus - 1 + 4) % 4
			m.updateFormFocus()
		case "enter":
			if m.formFocus == 0 {
				m.formFocus = 1
				m.updateFormFocus()
			} else if m.formFocus == 1 {
				m.formFocus = 2
				m.updateFormFocus()
			} else if m.formFocus == 2 {
				m.submitForm()
				m.view = viewMain
			} else if m.formFocus == 3 {
				m.view = viewMain
			}
		}
	}

	// Update active inputs
	if m.formFocus == 0 {
		m.nameInput, cmd = m.nameInput.Update(msg)
	} else if m.formFocus == 1 {
		m.descInput, cmd = m.descInput.Update(msg)
	}

	return m, cmd
}

func (m *model) updateFormFocus() {
	m.nameInput.Blur()
	m.descInput.Blur()

	switch m.formFocus {
	case 0:
		m.nameInput.Focus()
	case 1:
		m.descInput.Focus()
	}
}

func (m *model) submitForm() {
	name := strings.TrimSpace(m.nameInput.Value())
	desc := strings.TrimSpace(m.descInput.Value())
	if name == "" {
		return // Name is required
	}

	if m.view == viewAdd {
		id := fmt.Sprintf("%d", time.Now().UnixNano())
		h := &Habit{
			ID:          id,
			Name:        name,
			Description: desc,
			CreatedAt:   time.Now(),
			History:     make(map[string]bool),
		}
		m.habits = append(m.habits, h)
		m.selectedIdx = len(m.habits) - 1
		_ = m.db.AddHabit(h)
	} else if m.view == viewEdit {
		if len(m.habits) > 0 {
			h := m.habits[m.selectedIdx]
			h.Name = name
			h.Description = desc
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

func (m model) View() string {
	if m.width == 0 || m.height == 0 {
		return "Initializing habit tracker..."
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

	// Header banner
	header := headerStyle.Render("  GOHABIT  ")

	// Summary Stats for Today
	var completedToday, totalHabits int
	todayStr := time.Now().Format("2006-01-02")
	for _, h := range m.habits {
		totalHabits++
		if h.History[todayStr] {
			completedToday++
		}
	}

	statsStr := fmt.Sprintf("  Today: %d/%d completed", completedToday, totalHabits)
	if totalHabits > 0 {
		percent := float64(completedToday) / float64(totalHabits) * 100
		statsStr += fmt.Sprintf(" (%.1f%%)", percent)
	}
	statsStr = lipgloss.NewStyle().Foreground(cyanColor).Bold(true).Render(statsStr)

	fullHeader := lipgloss.JoinHorizontal(lipgloss.Center, header, statsStr)

	return lipgloss.JoinVertical(lipgloss.Left,
		"\n",
		fullHeader,
		"\n",
		content,
		"\n",
		m.renderHelp(),
	)
}

func (m model) renderMainView() string {
	if len(m.habits) == 0 {
		return lipgloss.NewStyle().
			Foreground(grayColor).
			Padding(4).
			Width(m.width - 4).
			Align(lipgloss.Center).
			Render("No habits tracked yet.\n\nPress 'n' to add your first habit!")
	}

	// Adjust widths dynamically
	leftWidth := m.width/2 - 3
	if leftWidth < 45 {
		leftWidth = 45
	}
	rightWidth := m.width - leftWidth - 6
	if rightWidth < 45 {
		rightWidth = 45
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
			checkbox = completedStyle.Render("[X]")
		} else {
			checkbox = pendingStyle.Render("[ ]")
		}

		title := h.Name
		if isSelected {
			title = selectedHabitStyle.Render(h.Name)
		} else {
			title = habitTitleStyle.Render(h.Name)
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

		if isSelected {
			// Highlights the entire row block using padding and lipgloss bg (optional styling)
		}

		listItems = append(listItems, item)
	}

	leftPanel := leftPanelStyle.Width(leftWidth).Height(m.height - 10).Render(
		strings.Join(listItems, "\n\n"),
	)

	// Render right panel (selected habit details)
	selectedHabit := m.habits[m.selectedIdx]

	detailTitle := detailTitleStyle.Render(selectedHabit.Name)
	detailDesc := selectedHabit.Description
	if detailDesc == "" {
		detailDesc = "No description provided."
	}
	detailDesc = lipgloss.NewStyle().Foreground(lightGrayColor).Render(detailDesc)

	currStreak, longStreak := selectedHabit.GetStreaks()
	totalDone := selectedHabit.GetTotalCompletions()
	rate := selectedHabit.GetCompletionRate()

	card1 := statCardStyle.Render(fmt.Sprintf("%s\n%s", statValStyle.Render(fmt.Sprintf("%d days", currStreak)), statLabelStyle.Render("Current Streak")))
	card2 := statCardStyle.Render(fmt.Sprintf("%s\n%s", statValStyle.Render(fmt.Sprintf("%d days", longStreak)), statLabelStyle.Render("Longest Streak")))
	card3 := statCardStyle.Render(fmt.Sprintf("%s\n%s", statValStyle.Render(fmt.Sprintf("%d times", totalDone)), statLabelStyle.Render("Total Completed")))
	card4 := statCardStyle.Render(fmt.Sprintf("%s\n%s", statValStyle.Render(fmt.Sprintf("%.1f%%", rate)), statLabelStyle.Render("Success Rate")))

	statsGrid := lipgloss.JoinVertical(lipgloss.Left,
		lipgloss.JoinHorizontal(lipgloss.Top, card1, card2),
		"\n",
		lipgloss.JoinHorizontal(lipgloss.Top, card3, card4),
	)

	calendarTitle := lipgloss.NewStyle().Foreground(whiteColor).Bold(true).Render("=== 30-DAY CALENDAR GRID ===")
	calendarView := renderCalendar(selectedHabit, m.cfg.WeekStart)

	rightContent := lipgloss.JoinVertical(lipgloss.Left,
		detailTitle,
		"\n",
		detailDesc,
		"\n\n",
		statsGrid,
		"\n\n",
		calendarTitle,
		calendarView,
	)

	rightPanel := rightPanelStyle.Width(rightWidth).Height(m.height - 10).Render(
		rightContent,
	)

	return lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, rightPanel)
}

func (m model) renderFormView(title string) string {
	formTitle := formTitleStyle.Render(title)

	var nameView, descView string
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

	var saveBtn, cancelBtn string
	if m.formFocus == 2 {
		saveBtn = focusedButtonStyle.Render("[ Save ]")
	} else {
		saveBtn = buttonStyle.Render("[ Save ]")
	}

	if m.formFocus == 3 {
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
		"\n\n",
		buttons,
	)

	return lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(cyanColor).
		Padding(2, 4).
		MarginLeft((m.width - 45) / 2).
		MarginTop((m.height - 15) / 2).
		Width(40).
		Render(formContent)
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

	return lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(redColor).
		Padding(2, 4).
		MarginLeft((m.width - 45) / 2).
		MarginTop((m.height - 12) / 2).
		Width(40).
		Render(content)
}

func (m model) renderHelp() string {
	var keys []string
	switch m.view {
	case viewMain:
		keys = []string{
			"↑/↓ or k/j: select habit",
			"←/→ or h/l: navigate sparkline days",
			"space: toggle highlighted day",
			"n: add new",
			"e: edit selected",
			"d/x: delete selected",
			"q: quit",
		}
	case viewAdd, viewEdit:
		keys = []string{
			"tab/shift+tab: move focus",
			"enter: submit/select",
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

		icon := "-"
		if completed {
			icon = "X"
		}

		var style lipgloss.Style
		if completed {
			style = completedStyle
		} else {
			style = pendingStyle
		}

		dayLetter := d.Format("Mon")[:2]

		var itemStr string
		if isSelectedHabit && idx == highlightedIdx {
			itemStr = fmt.Sprintf("[%s:%s]", dayLetter, style.Render(icon))
		} else {
			itemStr = fmt.Sprintf(" %s:%s ", dayLetter, style.Render(icon))
		}
		parts = append(parts, itemStr)
	}
	return strings.Join(parts, "")
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
				sb.WriteString(calendarFutureStyle.Render(".") + "   ")
			} else {
				completed := h.History[dateStr]
				if completed {
					sb.WriteString(calendarDoneStyle.Render("X") + "   ")
				} else {
					sb.WriteString(calendarMissStyle.Render("-") + "   ")
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

func printHelp() {
	fmt.Println(`Gohabit - A clean retro terminal habit tracker

Usage:
  gohabit                       Launch the interactive TUI dashboard
  gohabit <command> [args]      Run in CLI mode to perform quick actions

Commands:
  add "<name>" ["<desc>"]       Add a new habit with optional description
  check "<name>"                Toggle today's completion for a habit
  list                          List all habits, streaks, and today's status
  help                          Show this help documentation

Flags:
  -v, --version                 Print version information and exit
  -h, --help                    Print this help menu and exit

Configuration:
  Settings are loaded from 'config.ini' in the current directory.
  Completions are saved to the database specified in 'config.ini'.`)
}

func handleCLI(db *Database, args []string) {
	subcommand := strings.ToLower(args[0])

	switch subcommand {
	case "help":
		printHelp()

	case "add":
		if len(args) < 2 {
			fmt.Println("Error: Missing habit name.\nUsage: gohabit add \"<name>\" [\"<description>\"]")
			os.Exit(1)
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
			os.Exit(1)
		}
		fmt.Printf("Habit '%s' added successfully.\n", name)

	case "check":
		if len(args) < 2 {
			fmt.Println("Error: Missing habit name.\nUsage: gohabit check \"<name>\"")
			os.Exit(1)
		}
		nameQuery := args[1]
		habits, err := db.LoadHabits()
		if err != nil {
			fmt.Printf("Error loading habits: %v\n", err)
			os.Exit(1)
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
			os.Exit(1)
		}

		todayStr := time.Now().Format("2006-01-02")
		newStatus := !target.History[todayStr]

		err = db.SaveToggle(target.ID, todayStr, newStatus)
		if err != nil {
			fmt.Printf("Error saving toggle: %v\n", err)
			os.Exit(1)
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
			os.Exit(1)
		}

		if len(habits) == 0 {
			fmt.Println("No habits tracked yet. Run 'gohabit add \"<name>\"' to add one.")
			return
		}

		fmt.Printf("%-25s %-15s %-15s\n", "Habit Name", "Today's Status", "Streak")
		fmt.Println(strings.Repeat("-", 60))

		todayStr := time.Now().Format("2006-01-02")
		for _, h := range habits {
			status := "[ ] Pending"
			if h.History[todayStr] {
				status = "[X] Completed"
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
		os.Exit(1)
	}
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
	defer db.Close()

	// Check if there are CLI arguments
	args := flag.Args()
	if len(args) > 0 {
		handleCLI(db, args)
		return
	}

	// Initialize styles based on config theme colors
	InitStyles(cfg)

	p := tea.NewProgram(initialModel(db, cfg), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
