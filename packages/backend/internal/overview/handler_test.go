package overview

import (
	"slices"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestLoadGreenhouseActiveStrategies_UsesCurrentSchema(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	statements := []string{
		`CREATE TABLE climate_profiles (
			id INTEGER PRIMARY KEY,
			greenhouse_id INTEGER NOT NULL,
			name TEXT NOT NULL,
			enabled NUMERIC NOT NULL
		)`,
		`CREATE TABLE nutrient_recipes (
			id INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			status TEXT NOT NULL
		)`,
		`CREATE TABLE crop_batches (
			id INTEGER PRIMARY KEY,
			greenhouse_id INTEGER NOT NULL,
			status TEXT NOT NULL,
			active_recipe_id INTEGER
		)`,
		`INSERT INTO climate_profiles (id, greenhouse_id, name, enabled) VALUES
			(1, 4, 'Climate Enabled', 1),
			(2, 4, 'Climate Disabled', 0),
			(3, 9, 'Other Greenhouse Climate', 1)`,
		`INSERT INTO nutrient_recipes (id, name, status) VALUES
			(10, 'Recipe Active', 'ACTIVE'),
			(11, 'Recipe Draft', 'DRAFT'),
			(12, 'Recipe Other', 'ACTIVE')`,
		`INSERT INTO crop_batches (id, greenhouse_id, status, active_recipe_id) VALUES
			(100, 4, 'RUNNING', 10),
			(101, 4, 'RUNNING', 10),
			(102, 4, 'RUNNING', 11),
			(103, 4, 'PLANNED', 12),
			(104, 9, 'RUNNING', 12)`,
	}
	for _, stmt := range statements {
		if err := db.Exec(stmt).Error; err != nil {
			t.Fatalf("exec %q: %v", stmt, err)
		}
	}

	names, err := loadGreenhouseActiveStrategies(db, 4)
	if err != nil {
		t.Fatalf("load strategies: %v", err)
	}

	want := []string{"Climate Enabled", "Recipe Active"}
	if !slices.Equal(names, want) {
		t.Fatalf("expected strategies %v, got %v", want, names)
	}
}
