package model

import "testing"

func TestNormalizeHostTagFilters(t *testing.T) {
	filters := []HostTagFilter{
		{Tag: " env ", Value: " prod "},
		{Tag: "", Value: "ignored"},
		{Tag: "app", Value: ""},
	}

	got := normalizeHostTagFilters(filters)
	if len(got) != 2 {
		t.Fatalf("expected 2 filters, got %d", len(got))
	}
	if got[0].Tag != "env" || got[0].Value != "prod" {
		t.Fatalf("unexpected first filter: %+v", got[0])
	}
	if got[1].Tag != "app" || got[1].Value != "" {
		t.Fatalf("unexpected second filter: %+v", got[1])
	}
}

func TestHostMatchesTagFilters(t *testing.T) {
	hostTags := []HostTag{
		{Tag: "env", Value: "prod"},
		{Tag: "app", Value: "web"},
		{Tag: "team", Value: "ops"},
	}

	tests := []struct {
		name    string
		filters []HostTagFilter
		match   bool
	}{
		{
			name: "single exact match",
			filters: []HostTagFilter{
				{Tag: "env", Value: "prod"},
			},
			match: true,
		},
		{
			name: "multiple and match",
			filters: []HostTagFilter{
				{Tag: "env", Value: "prod"},
				{Tag: "app", Value: "web"},
			},
			match: true,
		},
		{
			name: "tag only match",
			filters: []HostTagFilter{
				{Tag: "team", Value: ""},
			},
			match: true,
		},
		{
			name: "value mismatch",
			filters: []HostTagFilter{
				{Tag: "env", Value: "stage"},
			},
			match: false,
		},
		{
			name: "missing tag",
			filters: []HostTagFilter{
				{Tag: "region", Value: "cn"},
			},
			match: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := hostMatchesTagFilters(hostTags, tt.filters); got != tt.match {
				t.Fatalf("expected %v, got %v", tt.match, got)
			}
		})
	}
}
