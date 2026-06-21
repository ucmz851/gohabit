# ⚡ Gobit

A beautiful, sleek, and highly functional Terminal User Interface (TUI) for tracking daily habits, built using **Go**, **Bubble Tea**, and **Lipgloss**.

It features a clean split-panel design showing your habits, progress sparklines, and a beautiful detailed view containing streak statistics and a 30-day contribution grid (similar to GitHub's commit history graph).

---

## ✨ Features

- ⚡ **Split-Panel Layout**: 
  - **Left Panel**: A list of habits with daily check status, current streak indicators (`🔥`), and a interactive 7-day sparkline history.
  - **Right Panel**: A dashboard for the selected habit showing detailed stats and a 30-day visual grid.
- 📅 **Interactive Retroactive Checking**: Navigate backward/forward through the last 7 days using arrow keys and toggle completions for past days.
- 📊 **Stat Cards Dashboard**:
  - **Current Streak**: Number of consecutive days completed up to today/yesterday.
  - **Longest Streak**: The maximum streak length achieved in history.
  - **Total Done**: Absolute count of completions.
  - **Success Rate**: Overall percentage of days completed since the habit's creation.
- 🎨 **Github-like Calendar Grid**: A 30-day grid visualizing completed, missed, and future days.
- 💾 **Automatic Persistence**: Habits and completions are saved to a local SQLite database (`habits.db`) in the current folder.
- ✏️ **Full CRUD**: Add, edit, or delete habits directly from within the TUI.
- 🎨 **Classic Terminal Aesthetics**: Styled with a traditional DOS/UNIX terminal palette (navy headers, cyan highlights, double-line panel borders) using Lipgloss.

---

## ⌨️ Controls & Keybindings

### Main View
- `↑` / `↓` or `k` / `j`: Select habit
- `←` / `→` or `h` / `l`: Navigate between days on the 7-day sparkline (indicated by `[Day:●]`)
- `Space`: Toggle completion for the currently highlighted day
- `n`: Add a new habit
- `e`: Edit the selected habit name/description
- `d` or `x`: Delete the selected habit
- `q` or `Ctrl+C`: Quit the application

### Form / Inputs (Add/Edit Habit)
- `Tab` / `Shift+Tab` or `↓` / `↑`: Move focus between Name, Description, and Save/Cancel buttons
- `Enter`: Submit form or trigger button action
- `Esc`: Cancel and return to the main dashboard

### Delete Confirmation
- `y` or `Enter`: Confirm delete
- `n` or `Esc`: Cancel and keep the habit



## ⚙️ Manual Configuration (`config.ini`)

Upon its first execution, the application automatically generates a template `config.ini` configuration file in the current directory if one does not already exist.

This file allows you to customize the following settings:
- **`db_path`**: Absolute or relative path to the SQLite database file (default: `habits.db`).
- **`color_header_bg`**: Hex color code for the top header background (default: `#005F87`).
- **`color_accent`**: Hex color code for panel borders and selected items (default: `#00AFDF`).
- **`color_success`**: Hex color code for completed checkboxes and days (default: `#00AF5F`).
- **`week_start`**: Choose whether the calendar grid starts on Sunday (`0`) or Monday (`1`) (default: `1`).

---

## 🚀 How to Run

### Prerequisites
- Go (version 1.18 or higher) installed on your system.

### Option 1: Quick Run
You can run the application directly using:
```bash
go run .
```

### Option 2: Build and Run Executable
Compile the binary first:
```bash
go build -o gobit
```
Then run the compiled binary:
```bash
./gobit
```
