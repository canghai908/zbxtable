package model

import "testing"

func TestBuildIndexInfoFromCounts(t *testing.T) {
	assetTypes := []AssetType{
		{Name: "Linux", TypeCode: "VM_LIN", Icon: "desktop"},
		{Name: "Windows", TypeCode: "VM_WIN", Icon: "windows"},
		{Name: "网络设备", TypeCode: "HW_NET", Icon: "cluster"},
		{Name: "物理服务器", TypeCode: "HW_SRV", Icon: "database"},
		{Name: "存储设备", TypeCode: "HW_STO", Icon: "hdd"},
	}
	counts := map[string]int64{
		"VM_LIN": 12,
		"VM_WIN": 8,
		"HW_NET": 5,
		"HW_SRV": 3,
		"HW_STO": 7,
	}

	info := buildIndexInfoFromCounts(assetTypes, counts)

	if info.TotalCount != 35 {
		t.Fatalf("expected total_count=35, got %d", info.TotalCount)
	}
	if info.LinCount != 12 || info.WinCount != 8 || info.NetCount != 5 || info.SrvCount != 3 {
		t.Fatalf("unexpected legacy counts: %+v", info)
	}
	if len(info.AssetTypeCounts) != len(assetTypes) {
		t.Fatalf("expected %d asset type counts, got %d", len(assetTypes), len(info.AssetTypeCounts))
	}
	if info.AssetTypeCounts[4].TypeCode != "HW_STO" || info.AssetTypeCounts[4].Count != 7 {
		t.Fatalf("expected custom asset type to be preserved, got %+v", info.AssetTypeCounts[4])
	}
	if info.HostCounts["HW_STO"] != 7 {
		t.Fatalf("expected host_counts to contain custom asset type, got %v", info.HostCounts["HW_STO"])
	}
}

func TestMigrateAssetTypeNamesTargetNames(t *testing.T) {
	defaults := []AssetType{
		{Name: "Linux", TypeCode: "VM_LIN"},
		{Name: "Windows", TypeCode: "VM_WIN"},
	}

	if defaults[0].Name != "Linux" || defaults[1].Name != "Windows" {
		t.Fatalf("expected default asset type names to be normalized, got %+v", defaults)
	}
}
