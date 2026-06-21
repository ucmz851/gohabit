package main

import (
	"bufio"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	DBPath        string
	HeaderBgColor string
	AccentColor   string
	SuccessColor  string
	WeekStart     int // 0 = Sunday, 1 = Monday
	EveningHour   int
}

func DefaultConfig() Config {
	return Config{
		DBPath:        "habits.db",
		HeaderBgColor: "#7C3AED",
		AccentColor:   "#F59E0B",
		SuccessColor:  "#10B981",
		WeekStart:     1,  // Monday start
		EveningHour:   18, // 6 PM
	}
}

func LoadConfig(path string) (Config, error) {
	cfg := DefaultConfig()

	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Write default config template
			_ = WriteDefaultConfig(path)
			return cfg, nil
		}
		return cfg, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])

		// Strip quotes if user wrapped them
		val = strings.Trim(val, `"'`)

		switch key {
		case "db_path":
			if val != "" {
				cfg.DBPath = val
			}
		case "color_header_bg":
			if val != "" {
				cfg.HeaderBgColor = val
			}
		case "color_accent":
			if val != "" {
				cfg.AccentColor = val
			}
		case "color_success":
			if val != "" {
				cfg.SuccessColor = val
			}
		case "week_start":
			if parsed, err := strconv.Atoi(val); err == nil {
				if parsed == 0 || parsed == 1 {
					cfg.WeekStart = parsed
				}
			}
		case "evening_hour":
			if parsed, err := strconv.Atoi(val); err == nil {
				if parsed >= 0 && parsed <= 23 {
					cfg.EveningHour = parsed
				}
			}
		}
	}

	return cfg, scanner.Err()
}

func WriteDefaultConfig(path string) error {
	content := `# Habit Tracker Configuration
# You can customize the database path, colors, and calendar week start below.

# Path to the SQLite database file
db_path = habits.db

# Header background color (Hex code)
color_header_bg = #7C3AED

# Accent color for borders and highlights (Hex code)
color_accent = #F59E0B

# Success color for completed items (Hex code)
color_success = #10B981

# Start of the week for the calendar view (0 = Sunday, 1 = Monday)
week_start = 1

# Evening hour for native desktop notifications (0-23, default 18 is 6 PM)
evening_hour = 18
`
	return os.WriteFile(path, []byte(content), 0644)
}
