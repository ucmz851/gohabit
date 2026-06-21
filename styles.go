package main

import "github.com/charmbracelet/lipgloss"

var (
	// Colors (Modern Vibrant Reward Theme)
	navyColor       = lipgloss.Color("#7C3AED") // Purple/Indigo (reused variable name for compatibility)
	cyanColor       = lipgloss.Color("#F59E0B") // Golden accent (reused variable name for compatibility)
	greenColor      = lipgloss.Color("#10B981") // Success emerald/mint (reused variable name for compatibility)
	redColor        = lipgloss.Color("#F43F5E") // Soft Rose warning (reused variable name for compatibility)
	whiteColor      = lipgloss.Color("#F8FAFC") // Bright text
	grayColor       = lipgloss.Color("#64748B") // Muted Slate Gray
	lightGrayColor  = lipgloss.Color("#CBD5E1") // Light gray
	darkGrayColor   = lipgloss.Color("#334155") // Dark slate borders
	darkerGrayColor = lipgloss.Color("#0F172A") // Deep dark background
	slateColor      = lipgloss.Color("#475569") // Slate helper color

	// Base Styles
	mainStyle = lipgloss.NewStyle().
			Padding(1, 2)

	headerStyle = lipgloss.NewStyle().
			Foreground(whiteColor).
			Background(navyColor).
			Padding(0, 2).
			Bold(true)

	// Panel Styles (Modern Rounded Borders)
	leftPanelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(navyColor).
			Padding(1, 2)

	rightPanelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(navyColor).
			Padding(1, 2)

	// List item styles
	habitTitleStyle = lipgloss.NewStyle().
			Foreground(whiteColor).
			Bold(true)

	selectedHabitStyle = lipgloss.NewStyle().
				Foreground(cyanColor).
				Bold(true)

	selectedRowBgStyle = lipgloss.NewStyle().
				Background(darkGrayColor)

	completedStyle = lipgloss.NewStyle().
			Foreground(greenColor).
			Bold(true)

	pendingStyle = lipgloss.NewStyle().
			Foreground(grayColor)

	streakStyle = lipgloss.NewStyle().
			Foreground(cyanColor).
			Bold(true)

	// Detail styles
	detailTitleStyle = lipgloss.NewStyle().
				Foreground(navyColor).
				Bold(true).
				Underline(true)

	statCardStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(darkGrayColor).
			Padding(0, 2).
			MarginRight(1).
			Align(lipgloss.Center).
			Width(22)

	statValStyle = lipgloss.NewStyle().
			Foreground(whiteColor).
			Bold(true)

	statLabelStyle = lipgloss.NewStyle().
			Foreground(grayColor)

	// Calendar styles
	calendarDoneStyle = lipgloss.NewStyle().
				Foreground(greenColor).
				Bold(true)

	calendarMissStyle = lipgloss.NewStyle().
				Foreground(grayColor)

	calendarFutureStyle = lipgloss.NewStyle().
				Foreground(darkGrayColor).
				Faint(true)

	// Form styles
	formTitleStyle = lipgloss.NewStyle().
			Foreground(whiteColor).
			Background(navyColor).
			Padding(0, 1).
			Bold(true).
			MarginBottom(1)

	inputStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(darkGrayColor).
			Padding(0, 1)

	focusedInputStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(navyColor).
				Padding(0, 1)

	buttonStyle = lipgloss.NewStyle().
			Foreground(whiteColor).
			Background(darkGrayColor).
			Padding(0, 2).
			MarginRight(1)

	focusedButtonStyle = lipgloss.NewStyle().
				Foreground(whiteColor).
				Background(navyColor).
				Bold(true).
				Padding(0, 2).
				MarginRight(1)

	helpStyle = lipgloss.NewStyle().
			Foreground(grayColor)
)

func InitStyles(cfg Config) {
	navyColor = lipgloss.Color(cfg.HeaderBgColor)
	cyanColor = lipgloss.Color(cfg.AccentColor)
	greenColor = lipgloss.Color(cfg.SuccessColor)

	headerStyle = headerStyle.Background(navyColor)
	leftPanelStyle = leftPanelStyle.BorderForeground(navyColor)
	rightPanelStyle = rightPanelStyle.BorderForeground(navyColor)
	selectedHabitStyle = selectedHabitStyle.Foreground(cyanColor)
	completedStyle = completedStyle.Foreground(greenColor)
	streakStyle = streakStyle.Foreground(cyanColor)
	detailTitleStyle = detailTitleStyle.Foreground(navyColor)
	calendarDoneStyle = calendarDoneStyle.Foreground(greenColor)
	formTitleStyle = formTitleStyle.Background(navyColor)
	focusedInputStyle = focusedInputStyle.BorderForeground(navyColor)
	focusedButtonStyle = focusedButtonStyle.Background(navyColor)
}
