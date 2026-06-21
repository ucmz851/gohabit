package main

import (
	"database/sql"
	"sort"
	"time"

	_ "modernc.org/sqlite"
)

type Habit struct {
	ID          string                  `json:"id"`
	Name        string                  `json:"name"`
	Description string                  `json:"description"`
	GitDir      string                  `json:"git_dir"`
	IsPriority  bool                    `json:"is_priority"`
	CreatedAt   time.Time               `json:"created_at"`
	History     map[string]bool         `json:"history"`     // date string "YYYY-MM-DD" -> completed
	TimeBlocks  map[string]map[int]bool `json:"time_blocks"` // date -> hour -> focused
}

// GetStreaks calculates the current streak and the longest streak.
func (h *Habit) GetStreaks() (int, int) {
	if len(h.History) == 0 {
		return 0, 0
	}

	// Extract completed dates
	var completedDates []time.Time
	for dateStr, completed := range h.History {
		if completed {
			t, err := time.Parse("2006-01-02", dateStr)
			if err == nil {
				completedDates = append(completedDates, t)
			}
		}
	}

	if len(completedDates) == 0 {
		return 0, 0
	}

	// Sort dates chronologically
	sort.Slice(completedDates, func(i, j int) bool {
		return completedDates[i].Before(completedDates[j])
	})

	// Calculate longest streak
	longest := 0
	currentRun := 0
	var lastDate time.Time

	for i, d := range completedDates {
		if i == 0 {
			currentRun = 1
		} else {
			// Check day difference by truncating to UTC midnight to avoid DST issues
			dUTC := time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, time.UTC)
			lastUTC := time.Date(lastDate.Year(), lastDate.Month(), lastDate.Day(), 0, 0, 0, 0, time.UTC)
			dayDiff := int(dUTC.Sub(lastUTC).Hours() / 24)

			if dayDiff == 1 {
				currentRun++
			} else if dayDiff > 1 {
				if currentRun > longest {
					longest = currentRun
				}
				currentRun = 1
			}
		}
		lastDate = d
	}
	if currentRun > longest {
		longest = currentRun
	}

	// Calculate current streak
	current := 0
	today := time.Now()
	todayStr := today.Format("2006-01-02")
	yesterdayStr := today.AddDate(0, 0, -1).Format("2006-01-02")

	hasToday := h.History[todayStr]
	hasYesterday := h.History[yesterdayStr]

	if hasToday || hasYesterday {
		var startCheck time.Time
		if hasToday {
			startCheck = today
		} else {
			startCheck = today.AddDate(0, 0, -1)
		}

		for {
			checkStr := startCheck.Format("2006-01-02")
			if h.History[checkStr] {
				current++
				startCheck = startCheck.AddDate(0, 0, -1)
			} else {
				break
			}
		}
	}

	return current, longest
}

func (h *Habit) GetCompletionRate() float64 {
	if len(h.History) == 0 {
		return 0
	}

	// Calculate total days since creation
	firstDate := h.CreatedAt
	for dateStr, completed := range h.History {
		if completed {
			t, err := time.Parse("2006-01-02", dateStr)
			if err == nil && t.Before(firstDate) {
				firstDate = t
			}
		}
	}

	today := time.Now()
	todayZero := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, time.UTC)
	firstZero := time.Date(firstDate.Year(), firstDate.Month(), firstDate.Day(), 0, 0, 0, 0, time.UTC)

	days := int(todayZero.Sub(firstZero).Hours()/24) + 1
	if days <= 0 {
		days = 1
	}

	completedCount := 0
	for _, completed := range h.History {
		if completed {
			completedCount++
		}
	}

	return float64(completedCount) / float64(days) * 100
}

func (h *Habit) GetTotalCompletions() int {
	count := 0
	for _, completed := range h.History {
		if completed {
			count++
		}
	}
	return count
}

type Database struct {
	db *sql.DB
}

func NewDatabase(dbPath string) (*Database, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	// Enable foreign key support inside SQLite
	_, _ = db.Exec("PRAGMA foreign_keys = ON")

	// Create tables schema
	schema := `
	CREATE TABLE IF NOT EXISTS habits (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		description TEXT,
		created_at TEXT NOT NULL,
		git_dir TEXT,
		is_priority INTEGER DEFAULT 0
	);

	CREATE TABLE IF NOT EXISTS history (
		habit_id TEXT,
		date TEXT, -- YYYY-MM-DD
		completed INTEGER DEFAULT 1,
		PRIMARY KEY (habit_id, date),
		FOREIGN KEY (habit_id) REFERENCES habits(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS time_blocks (
		habit_id TEXT,
		date TEXT, -- YYYY-MM-DD
		hour INTEGER, -- 0-23
		PRIMARY KEY (habit_id, date, hour),
		FOREIGN KEY (habit_id) REFERENCES habits(id) ON DELETE CASCADE
	);
	`
	if _, err := db.Exec(schema); err != nil {
		_ = db.Close()
		return nil, err
	}

	// Run migration for existing databases to add git_dir and is_priority columns
	_, _ = db.Exec("ALTER TABLE habits ADD COLUMN git_dir TEXT")
	_, _ = db.Exec("ALTER TABLE habits ADD COLUMN is_priority INTEGER DEFAULT 0")

	return &Database{db: db}, nil
}

func (d *Database) Close() error {
	return d.db.Close()
}

func (d *Database) LoadHabits() ([]*Habit, error) {
	rows, err := d.db.Query("SELECT id, name, description, created_at, git_dir, is_priority FROM habits ORDER BY created_at ASC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var habits []*Habit
	for rows.Next() {
		h := &Habit{
			History:    make(map[string]bool),
			TimeBlocks: make(map[string]map[int]bool),
		}
		var createdAtStr string
		var gitDir sql.NullString
		var isPriorityVal int
		err := rows.Scan(&h.ID, &h.Name, &h.Description, &createdAtStr, &gitDir, &isPriorityVal)
		if err != nil {
			return nil, err
		}

		h.CreatedAt, _ = time.Parse(time.RFC3339, createdAtStr)
		if h.CreatedAt.IsZero() {
			h.CreatedAt = time.Now()
		}

		if gitDir.Valid {
			h.GitDir = gitDir.String
		}
		h.IsPriority = (isPriorityVal == 1)

		habits = append(habits, h)
	}

	// Load completion history for each habit
	for _, h := range habits {
		historyRows, err := d.db.Query("SELECT date, completed FROM history WHERE habit_id = ?", h.ID)
		if err != nil {
			return nil, err
		}

		for historyRows.Next() {
			var dateStr string
			var completed int
			if err := historyRows.Scan(&dateStr, &completed); err == nil {
				h.History[dateStr] = (completed == 1)
			}
		}
		historyRows.Close()

		// Load time blocks
		tbRows, err := d.db.Query("SELECT date, hour FROM time_blocks WHERE habit_id = ?", h.ID)
		if err == nil {
			for tbRows.Next() {
				var dateStr string
				var hr int
				if err := tbRows.Scan(&dateStr, &hr); err == nil {
					if h.TimeBlocks[dateStr] == nil {
						h.TimeBlocks[dateStr] = make(map[int]bool)
					}
					h.TimeBlocks[dateStr][hr] = true
				}
			}
			tbRows.Close()
		}
	}

	return habits, nil
}

func (d *Database) AddHabit(h *Habit) error {
	priorityVal := 0
	if h.IsPriority {
		priorityVal = 1
	}
	_, err := d.db.Exec("INSERT INTO habits (id, name, description, created_at, git_dir, is_priority) VALUES (?, ?, ?, ?, ?, ?)",
		h.ID, h.Name, h.Description, h.CreatedAt.Format(time.RFC3339), h.GitDir, priorityVal)
	return err
}

func (d *Database) UpdateHabit(h *Habit) error {
	priorityVal := 0
	if h.IsPriority {
		priorityVal = 1
	}
	_, err := d.db.Exec("UPDATE habits SET name = ?, description = ?, git_dir = ?, is_priority = ? WHERE id = ?",
		h.Name, h.Description, h.GitDir, priorityVal, h.ID)
	return err
}

func (d *Database) DeleteHabit(id string) error {
	_, err := d.db.Exec("DELETE FROM habits WHERE id = ?", id)
	return err
}

func (d *Database) SaveToggle(habitID string, dateStr string, completed bool) error {
	if completed {
		_, err := d.db.Exec("INSERT OR REPLACE INTO history (habit_id, date, completed) VALUES (?, ?, 1)",
			habitID, dateStr)
		return err
	} else {
		_, err := d.db.Exec("DELETE FROM history WHERE habit_id = ? AND date = ?",
			habitID, dateStr)
		return err
	}
}

func (d *Database) SaveTimeBlock(habitID string, dateStr string, hour int, active bool) error {
	if active {
		_, err := d.db.Exec("INSERT OR REPLACE INTO time_blocks (habit_id, date, hour) VALUES (?, ?, ?)",
			habitID, dateStr, hour)
		return err
	} else {
		_, err := d.db.Exec("DELETE FROM time_blocks WHERE habit_id = ? AND date = ? AND hour = ?",
			habitID, dateStr, hour)
		return err
	}
}
