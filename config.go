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
}

func DefaultConfig() Config {
	return Config{
		DBPath:        "habits.db",
		HeaderBgColor: "#005F87",
		AccentColor:   "#00AFDF",
		SuccessColor:  "#00AF5F",
		WeekStart:     1, // Monday start
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
color_header_bg = #005F87

# Accent color for borders and highlights (Hex code)
color_accent = #00AFDF

# Success color for completed items (Hex code)
color_success = #00AF5F

# Start of the week for the calendar view (0 = Sunday, 1 = Monday)
week_start = 1
`
	return os.WriteFile(path, []byte(content), 0644)
}
