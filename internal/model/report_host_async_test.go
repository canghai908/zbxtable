package model

import "testing"

func TestPrepareHostReportExecutionConfig_UsesPersistedResolvedConfig(t *testing.T) {
	report := &Report{
		ID:              123,
		ReportType:      "host",
		HostIds:         `[{"selection_mode":"host_group","zid":"1","group_ids":["10"],"reference_host_id":"1001","reference_item_ids":["2001"]}]`,
		ResolvedHostIds: `[{"zid":"1","host_id":"3001","item_ids":["4001","4002"]}]`,
		ItemIds:         `["4001","4002"]`,
	}

	if err := PrepareHostReportExecutionConfig(report); err != nil {
		t.Fatalf("PrepareHostReportExecutionConfig returned error: %v", err)
	}

	expectedHostIDs := `[{"zid":"1","host_id":"3001","item_ids":["4001","4002"]}]`
	if report.HostIds != expectedHostIDs {
		t.Fatalf("expected host_ids to switch to resolved payload, got %s", report.HostIds)
	}

	expectedItemIDs := `["4001","4002"]`
	if report.ItemIds != expectedItemIDs {
		t.Fatalf("expected item_ids to remain resolved payload, got %s", report.ItemIds)
	}
}

func TestPrepareHostReportExecutionConfig_SkipsNonBulkConfig(t *testing.T) {
	report := &Report{
		ReportType: "host",
		HostIds:    `[{"zid":"1","host_id":"3001","item_ids":["4001"]}]`,
		ItemIds:    `["4001"]`,
	}

	if err := PrepareHostReportExecutionConfig(report); err != nil {
		t.Fatalf("PrepareHostReportExecutionConfig returned error: %v", err)
	}

	if report.HostIds != `[{"zid":"1","host_id":"3001","item_ids":["4001"]}]` {
		t.Fatalf("expected non-bulk host_ids to stay unchanged, got %s", report.HostIds)
	}
}
