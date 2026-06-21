# ⚡ Gohabit

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



## 💻 Command Line Interface (CLI) Mode

Gohabit can be run directly from your terminal using command-line arguments to perform quick checks or add new habits. This allows easy integration with aliases, cron jobs, or shell scripts.

### List tracked habits
Display a clean table of all habits, their completion status for today, and current streaks:
```bash
gohabit list
```

### Check/toggle today's completion
Toggle today's check-in status for a habit. The matching supports case-insensitive substrings (e.g. `gohabit check water` will toggle `Drink Water`):
```bash
gohabit check "<habit-name>"
```

### Add a new habit
Create a new habit with an optional description directly:
```bash
gohabit add "<habit-name>" ["<optional-description>"]
```

### Display help or version
Show the command reference table or check the installed release:
```bash
gohabit -h         # or gohabit --help
gohabit -v         # or gohabit --version
```

---

## ⚙️ Manual Configuration (`config.ini`)

Upon its first execution, the application automatically generates a template `config.ini` configuration file in the current directory if one does not already exist.

This file allows you to customize the following settings:
- **`db_path`**: Absolute or relative path to the SQLite database file (default: `habits.db`).
- **`color_header_bg`**: Hex color code for the top header background (default: `#005F87`).
- **`color_accent`**: Hex color code for panel borders and selected items (default: `#00AFDF`).
- **`color_success`**: Hex color code for completed checkboxes and days (default: `#00AF5F`).
- **`week_start`**: Choose whether the calendar grid starts on Sunday (`0`) or Monday (`1`) (default: `1`).

---

## 🚀 Installation & Running

### Prerequisites
- Go (version 1.18 or higher) installed on your system.
- Your local `$GOBIN` (e.g. `~/go/bin`) or `~/.local/bin` added to your system's `$PATH` environment variable.

### Option 1: Quick One-Line Installer (Recommended)
You can download and configure `gohabit` automatically by running the following command in your terminal. 

This script checks for Go on your system to build from source, otherwise it auto-downloads the pre-compiled binary matching your OS/architecture from the latest GitHub Release and updates your shell profile path settings:

```bash
curl -sSfL https://raw.githubusercontent.com/ucmz851/gohabit/main/get-gohabit.sh | sh
```

---

### Option 2: Native Go Installer
If you are a Go developer, you can install `gohabit` directly into your `$GOBIN` path using Go's package manager:
```bash
go install github.com/ucmz851/gohabit@latest
```
Once installed, you can launch the TUI from anywhere in your shell:
```bash
gohabit
```

### Option 3: Build & Install via Makefile
Clone the repository and run the installation script using `make`. This automatically:
1. Compiles the `gohabit` binary locally.
2. Creates the target directory `~/.local/bin` if it does not exist.
3. Copies the binary to that folder.
4. Detects your active shell (e.g. Bash or Zsh) and automatically configures your PATH profile (e.g. `~/.bashrc`, `~/.zshrc`) if `~/.local/bin` is not already in your system's `$PATH`.

```bash
# Clone the repository
git clone https://github.com/ucmz851/gohabit.git
cd gohabit

# Build & install automatically
make install
```
After executing, restart your terminal or source your profile to start using `gohabit` globally.

### Option 4: Manual Compilation
If you prefer to compile and run the binary locally:
```bash
# Build the binary
go build -o gohabit

# Run the binary locally
./gohabit
```
