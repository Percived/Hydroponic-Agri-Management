package db

import (
	"reflect"
	"testing"
	"time"

	gormlogger "gorm.io/gorm/logger"
)

func TestNewGormLoggerDisablesSlowSQL(t *testing.T) {
	logger := newGormLogger()

	value := reflect.ValueOf(logger)
	if value.Kind() != reflect.Ptr {
		t.Fatalf("expected pointer logger, got %T", logger)
	}

	configField := value.Elem().FieldByName("Config")
	if !configField.IsValid() {
		t.Fatalf("expected Config field on logger, got %T", logger)
	}

	slowThreshold := configField.FieldByName("SlowThreshold")
	if got := time.Duration(slowThreshold.Int()); got != 0 {
		t.Fatalf("expected slow threshold 0, got %s", got)
	}

	logLevel := configField.FieldByName("LogLevel")
	if got := gormlogger.LogLevel(logLevel.Int()); got != gormlogger.Warn {
		t.Fatalf("expected warn log level, got %v", got)
	}
}
