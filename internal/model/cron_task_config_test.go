package model

import "testing"

func TestTaskConfigKeysAreRecognized(t *testing.T) {
	defs := getScheduledTaskDefinitions()
	if len(defs) == 0 {
		t.Fatal("expected scheduled task definitions")
	}

	for _, def := range defs {
		if !IsTaskConfigKey(def.EnabledKey) {
			t.Fatalf("enabled key %s should be recognized as task config", def.EnabledKey)
		}
		if !IsTaskEnabledConfigKey(def.EnabledKey) {
			t.Fatalf("enabled key %s should be recognized as task enabled config", def.EnabledKey)
		}
		if !IsTaskConfigKey(def.CronKey) {
			t.Fatalf("cron key %s should be recognized as task config", def.CronKey)
		}
		if !IsTaskCronConfigKey(def.CronKey) {
			t.Fatalf("cron key %s should be recognized as task cron config", def.CronKey)
		}
	}
}

func TestValidateTaskConfigValue(t *testing.T) {
	if err := ValidateTaskConfigValue("top_sync_enabled", "1"); err != nil {
		t.Fatalf("expected valid enabled value, got %v", err)
	}
	if err := ValidateTaskConfigValue("top_sync_enabled", "2"); err == nil {
		t.Fatal("expected invalid enabled value to fail")
	}
	if err := ValidateTaskConfigValue("top_sync_cron", "0 */5 * * * *"); err != nil {
		t.Fatalf("expected valid cron value, got %v", err)
	}
	if err := ValidateTaskConfigValue("top_sync_cron", "invalid cron"); err == nil {
		t.Fatal("expected invalid cron value to fail")
	}
}

func TestGetTaskConfigsIncludesPairs(t *testing.T) {
	taskConfigs := getTaskConfigs()
	defs := getScheduledTaskDefinitions()

	if len(taskConfigs) != len(defs)*2 {
		t.Fatalf("expected %d task configs, got %d", len(defs)*2, len(taskConfigs))
	}
}
