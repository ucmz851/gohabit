# 🌟 Gohabit

A beautiful, sleek, and gamified Terminal User Interface (TUI) for tracking daily habits, built with **Go**, **Bubble Tea**, and **Lipgloss**.

Level up your life, earn XP, maintain streaks, and build consistent habits inside your favorite shell!

---

## ✨ Features

* 🎮 **RPG-style Gamification**:
  * **Level & XP Engine**: Earn +10 XP for completing habits, plus daily streak bonus XP.
  * **XP Progress Bar**: A visual indicator right in the app header: `⭐ Level 3 [██████░░░░░░] 40/100 XP`.
  * **Level Up Celebrations**: In-app animations and desktop notifications celebrate your growth!
* 🧼 **Friendly Clean Slate**: Fresh installations greet you with a gorgeous centered Welcome Card, encouraging you to jump right into habit creation.
* 🎨 **Vibrant Modern Aesthetics**:
  * Styled with a premium dark theme: **Indigo Purple** borders, **Golden** streaks, and **Emerald Green** checkmarks.
  * Sleek rounded panels replace blocky retro borders.
  * GitHub-style calendar grids (`■`/`□`/`·`) show your monthly habit achievements.
* 📅 **Interactive Retroactive Checks**: Highlight days on the 7-day sparkline history using `←`/`→` keys and retroactively toggle completions.
* ⏰ **Collapsible Hourly Time-Blocking**: Organize your day into time slots. The timeline stays collapsed until you decide to plan.
* 🐚 **Git Repo Integration**: Link habits to local git repos. The background worker checks commits and auto-completes habits when you commit code (with loop protection so you can still override it).
* 💾 **Reliable SQLite Storage**: Your logs are saved in a local SQLite file (`habits.db`), now protected with Write-Ahead Logging (WAL) and busy timeout retries to prevent locking.

---

## ⌨️ Controls & Keybindings

### Main View
* `↑` / `↓` or `k` / `j` : Navigate and select habits
* `←` / `→` or `h` / `l` : Navigate between days on the 7-day sparkline (indicated by `[●]`)
* `Space` : Toggle completion for the highlighted day or selected habit
* `t` : Toggle/open hourly time-blocking timeline
* `n` : Create a new habit
* `e` : Edit the selected habit name, description, priority, or Git repo
* `d` / `x` : Delete the selected habit
* `q` / `Ctrl+C` : Quit the application

### Form Input (Add / Edit Habit)
* `Tab` / `Shift+Tab` or `↓` / `↑` : Move focus between fields (Name, Description, Git Path, Priority, Save/Cancel buttons)
* `Space` / `Enter` : Toggle Priority checkbox or trigger button action
* `Esc` : Cancel form and return to the main dashboard

---

## 💻 Command Line Interface (CLI) Mode

Gohabit supports quick shell commands for scripting, aliases, or cron jobs.

### List tracked habits
```bash
gohabit list
```

### Toggle completion for today
```bash
gohabit check "Drink Water"
```

### Add a new habit
```bash
gohabit add "Workout" "Go to the gym for 45 minutes"
```

### Display help or version
```bash
gohabit --help
gohabit --version
```

---

## ⚙️ Customization (`config.ini`)

Gohabit generates a template configuration file (`config.ini`) on first run:

```ini
# Path to the SQLite database file
db_path = habits.db

# Theme Colors (Hex codes)
color_header_bg = #7C3AED   # Indigo
color_accent = #F59E0B      # Gold
color_success = #10B981     # Emerald

# Calendar start day (0 = Sunday, 1 = Monday)
week_start = 1

# Evening notification trigger hour (0-23)
evening_hour = 18
```

---

## 🚀 Installation & Running

### Prerequisites
* Go (version 1.18 or higher) installed on your system.
* Your local `$GOBIN` (e.g. `~/go/bin`) or `~/.local/bin` added to your system's `$PATH`.

### Option 1: One-Line Script Installer (Recommended)
Automatically builds from source if Go is present, otherwise downloads the pre-compiled binary for your architecture and updates your shell profile:
```bash
curl -sSfL https://raw.githubusercontent.com/ucmz851/gohabit/main/get-gohabit.sh | sh
```

### Option 2: Go Installer
Installs directly into your `$GOBIN`:
```bash
go install github.com/ucmz851/gohabit@latest
```

### Option 3: Build via Makefile
```bash
git clone https://github.com/ucmz851/gohabit.git
cd gohabit
make install
```
