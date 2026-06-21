package main

import "github.com/charmbracelet/lipgloss"

var (
	// Colors (Classic DOS/UNIX terminal theme)
	navyColor       = lipgloss.Color("#005F87") // Navy Blue header
	cyanColor       = lipgloss.Color("#00AFDF") // Cyan accent
	greenColor      = lipgloss.Color("#00AF5F") // Success Green
	redColor        = lipgloss.Color("#DF0000") // Red warning
	whiteColor      = lipgloss.Color("#FFFFFF") // White text
	grayColor       = lipgloss.Color("#8A8A8A") // Muted gray
	lightGrayColor  = lipgloss.Color("#D7D7D7") // Light gray
	darkGrayColor   = lipgloss.Color("#303030") // Dark gray borders
	darkerGrayColor = lipgloss.Color("#121212") // Dark background

	// Base Styles
	mainStyle = lipgloss.NewStyle().
			Padding(1, 2)

	headerStyle = lipgloss.NewStyle().
			Foreground(whiteColor).
			Background(navyColor).
			Padding(0, 2).
			Bold(true)

	// Panel Styles (Classic Double Line Borders)
	leftPanelStyle = lipgloss.NewStyle().
			Border(lipgloss.DoubleBorder()).
			BorderForeground(cyanColor).
			Padding(1, 2)

	rightPanelStyle = lipgloss.NewStyle().
			Border(lipgloss.DoubleBorder()).
			BorderForeground(cyanColor).
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
				Foreground(cyanColor).
				Bold(true).
				Underline(true)

	statCardStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
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
			Border(lipgloss.NormalBorder()).
			BorderForeground(darkGrayColor).
			Padding(0, 1)

	focusedInputStyle = lipgloss.NewStyle().
				Border(lipgloss.NormalBorder()).
				BorderForeground(cyanColor).
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
	leftPanelStyle = leftPanelStyle.BorderForeground(cyanColor)
	rightPanelStyle = rightPanelStyle.BorderForeground(cyanColor)
	selectedHabitStyle = selectedHabitStyle.Foreground(cyanColor)
	completedStyle = completedStyle.Foreground(greenColor)
	streakStyle = streakStyle.Foreground(cyanColor)
	detailTitleStyle = detailTitleStyle.Foreground(cyanColor)
	calendarDoneStyle = calendarDoneStyle.Foreground(greenColor)
	formTitleStyle = formTitleStyle.Background(navyColor)
	focusedInputStyle = focusedInputStyle.BorderForeground(cyanColor)
	focusedButtonStyle = focusedButtonStyle.Background(navyColor)
}
