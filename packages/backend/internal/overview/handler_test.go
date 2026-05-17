package overview

import (
	"slices"
	"testing"
	"time"

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

func TestLoadGreenhouseLastCollectedAt_PicksLatestTelemetryAcrossMetrics(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	statements := []string{
		`CREATE TABLE sensor_devices (
			id INTEGER PRIMARY KEY,
			greenhouse_id INTEGER NOT NULL
		)`,
		`CREATE TABLE sensor_channels (
			id INTEGER PRIMARY KEY,
			sensor_device_id INTEGER NOT NULL
		)`,
		`CREATE TABLE telemetry_records (
			id INTEGER PRIMARY KEY,
			sensor_channel_id INTEGER NOT NULL,
			metric_code TEXT NOT NULL,
			value REAL NOT NULL,
			collected_at DATETIME NOT NULL
		)`,
		`INSERT INTO sensor_devices (id, greenhouse_id) VALUES
			(1, 7),
			(2, 8)`,
		`INSERT INTO sensor_channels (id, sensor_device_id) VALUES
			(11, 1),
			(12, 1),
			(21, 2)`,
		`INSERT INTO telemetry_records (id, sensor_channel_id, metric_code, value, collected_at) VALUES
			(100, 11, 'TEMP', 24.5, '2026-05-16 08:00:00'),
			(101, 12, 'EC', 1.8, '2026-05-16 08:05:00'),
			(102, 11, 'HUMIDITY', 65.0, '2026-05-16 08:10:00'),
			(103, 21, 'TEMP', 21.0, '2026-05-16 08:20:00')`,
	}
	for _, stmt := range statements {
		if err := db.Exec(stmt).Error; err != nil {
			t.Fatalf("exec %q: %v", stmt, err)
		}
	}

	got, err := loadGreenhouseLastCollectedAt(db, 7)
	if err != nil {
		t.Fatalf("load last collected at: %v", err)
	}

	want := time.Date(2026, 5, 16, 8, 10, 0, 0, time.UTC)
	if got == nil {
		t.Fatalf("expected timestamp %s, got nil", want.Format(time.RFC3339))
	}
	if !got.Equal(want) {
		t.Fatalf("expected %s, got %s", want.Format(time.RFC3339), got.Format(time.RFC3339))
	}
}
