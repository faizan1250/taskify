package config

import (
	"testing"

	"github.com/go-playground/validator/v10"
)

func TestCronApplyDefaultsPartial(t *testing.T) {
	c := &CronConfig{BatchSize: 50}
	c.applyDefaults(DefaultCronConfig())

	if c.BatchSize != 50 {
		t.Fatalf("explicit BatchSize overwritten: got %d, want 50", c.BatchSize)
	}
	def := DefaultCronConfig()
	if c.ArchiveDaysThreshold != def.ArchiveDaysThreshold {
		t.Errorf("ArchiveDaysThreshold: got %d, want default %d", c.ArchiveDaysThreshold, def.ArchiveDaysThreshold)
	}
	if c.ReminderHours != def.ReminderHours {
		t.Errorf("ReminderHours: got %d, want default %d", c.ReminderHours, def.ReminderHours)
	}
	if c.MaxTodosPerUserNotification != def.MaxTodosPerUserNotification {
		t.Errorf("MaxTodosPerUserNotification: got %d, want default %d", c.MaxTodosPerUserNotification, def.MaxTodosPerUserNotification)
	}

	if err := validator.New().Struct(c); err != nil {
		t.Fatalf("defaulted CronConfig failed validation: %v", err)
	}
}

func TestCronApplyDefaultsAllZero(t *testing.T) {
	c := &CronConfig{}
	c.applyDefaults(DefaultCronConfig())

	if *c != *DefaultCronConfig() {
		t.Fatalf("all-zero CronConfig not fully defaulted: got %+v", c)
	}
}

func TestCronNegativeSurvivesForValidation(t *testing.T) {
	c := &CronConfig{ArchiveDaysThreshold: -5}
	c.applyDefaults(DefaultCronConfig())

	if c.ArchiveDaysThreshold != -5 {
		t.Fatalf("negative value masked by default: got %d", c.ArchiveDaysThreshold)
	}
	if err := validator.New().Struct(c); err == nil {
		t.Fatal("expected validation error for negative ArchiveDaysThreshold, got nil")
	}
}
