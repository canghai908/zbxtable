package model

import "testing"

func TestHostMatchesKeyword(t *testing.T) {
	host := Hosts{
		Name:         "prod-linux-01",
		Host:         "prod-linux-01.internal",
		Interfaces:   "10.0.0.8",
		ResourceID:   "ASSET-001",
		SerialNo:     "SN-7788",
		Model:        "Dell R760",
		Location:     "Shanghai-A1",
		Department:   "SRE",
		InstanceName: "zabbix-prod",
		OS:           "Ubuntu 22.04",
	}

	cases := []struct {
		name    string
		keyword string
		match   bool
	}{
		{name: "host name", keyword: "linux-01", match: true},
		{name: "ip", keyword: "10.0.0", match: true},
		{name: "asset id", keyword: "asset-001", match: true},
		{name: "location", keyword: "shanghai", match: true},
		{name: "department", keyword: "sre", match: true},
		{name: "instance", keyword: "zabbix-prod", match: true},
		{name: "unknown", keyword: "does-not-exist", match: false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := hostMatchesKeyword(host, tc.keyword)
			if got != tc.match {
				t.Fatalf("hostMatchesKeyword(%q) = %v, want %v", tc.keyword, got, tc.match)
			}
		})
	}
}

func TestLimitHosts(t *testing.T) {
	hosts := []Hosts{{HostID: "1"}, {HostID: "2"}, {HostID: "3"}}

	limited := limitHosts(hosts, 2)
	if len(limited) != 2 {
		t.Fatalf("len(limitHosts(...)) = %d, want 2", len(limited))
	}

	unlimited := limitHosts(hosts, 0)
	if len(unlimited) != 3 {
		t.Fatalf("len(limitHosts(..., 0)) = %d, want 3", len(unlimited))
	}
}

func TestSearchHostsInOverviewData(t *testing.T) {
	overview := map[string][]Hosts{
		"VM_LIN": {
			{
				HostID:       "1",
				TypeCode:     "VM_LIN",
				Name:         "prod-linux-01",
				Interfaces:   "10.0.0.8",
				ResourceID:   "ASSET-001",
				Location:     "Shanghai-A1",
				Department:   "SRE",
				InstanceName: "zabbix-prod",
			},
		},
		"HW_NET": {
			{
				HostID:       "2",
				TypeCode:     "HW_NET",
				Name:         "switch-01",
				Interfaces:   "172.16.1.10",
				ResourceID:   "NET-001",
				Location:     "Beijing-B2",
				Department:   "Network",
				InstanceName: "zabbix-net",
			},
		},
	}

	locationMatches := searchHostsInOverviewData(overview, "shanghai", 10)
	if len(locationMatches) != 1 || locationMatches[0].HostID != "1" {
		t.Fatalf("search by location returned %+v, want host 1", locationMatches)
	}

	ipMatches := searchHostsInOverviewData(overview, "172.16.1", 10)
	if len(ipMatches) != 1 || ipMatches[0].HostID != "2" {
		t.Fatalf("search by ip returned %+v, want host 2", ipMatches)
	}

	limitedMatches := searchHostsInOverviewData(overview, "zabbix", 1)
	if len(limitedMatches) != 1 {
		t.Fatalf("expected 1 limited match, got %d", len(limitedMatches))
	}
}

func TestNormalizeHostAvailabilityUsesActiveWhenUnknown(t *testing.T) {
	tests := []struct {
		name            string
		available       string
		activeAvailable string
		want            string
	}{
		{name: "active available replaces unknown", available: "0", activeAvailable: "1", want: "1"},
		{name: "active unavailable replaces unknown", available: "0", activeAvailable: "2", want: "2"},
		{name: "passive unavailable wins", available: "2", activeAvailable: "1", want: "2"},
		{name: "empty active keeps current", available: "0", activeAvailable: "", want: "0"},
		{name: "empty current uses active", available: "", activeAvailable: "1", want: "1"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := normalizeHostAvailability(tc.available, tc.activeAvailable)
			if got != tc.want {
				t.Fatalf("normalizeHostAvailability(%q, %q) = %q, want %q", tc.available, tc.activeAvailable, got, tc.want)
			}
		})
	}
}

func TestBuildHostFromListHostUsesActiveAvailabilityForActiveAgent(t *testing.T) {
	host := buildHostFromListHost(&APIInstance{ZID: 1, Name: "zabbix", IsV54OrLater: true}, ListHost{
		Hostid:          "10001",
		Host:            "linux-active",
		Name:            "Linux active",
		ActiveAvailable: "1",
		Interfaces: []HostListInterface{
			{IP: "10.0.0.1", Available: "0"},
		},
	}, "VM_LIN")

	if host.Available != "1" {
		t.Fatalf("expected active availability to mark host available, got %q", host.Available)
	}
}
