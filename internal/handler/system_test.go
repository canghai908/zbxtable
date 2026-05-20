package handler

import (
	"testing"
	"zbxtable/internal/model"
)

func TestSanitizeConfigsForResponseMasksCustomAPIKey(t *testing.T) {
	configs := []model.Config{
		{ConfigKey: "custom_api_key", ConfigValue: "secret-token"},
		{ConfigKey: "custom_model", ConfigValue: "gpt-compatible"},
		{ConfigKey: "encryption_key", ConfigValue: "hidden"},
	}

	got := sanitizeConfigsForResponse(configs)
	if len(got) != 2 {
		t.Fatalf("sanitized configs len = %d, want 2", len(got))
	}
	if got[0].ConfigValue != "********" {
		t.Fatalf("masked custom_api_key = %q, want ********", got[0].ConfigValue)
	}
	if got[1].ConfigValue != "gpt-compatible" {
		t.Fatalf("custom_model = %q, want %q", got[1].ConfigValue, "gpt-compatible")
	}
}
