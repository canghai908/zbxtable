package model

import "testing"

func TestDerivePeerTrafficItemKey(t *testing.T) {
	tests := []struct {
		name     string
		itemName string
		itemKey  string
		want     string
		wantErr  bool
	}{
		{
			name:     "switch outbound net key to inbound",
			itemName: "eth0-WAN",
			itemKey:  `net.if.out[ifHCOutOctets.2]`,
			want:     `net.if.in[ifHCInOctets.2]`,
		},
		{
			name:     "switch inbound net key to outbound",
			itemName: "eth0-WAN",
			itemKey:  `net.if.in[ifHCInOctets.2]`,
			want:     `net.if.out[ifHCOutOctets.2]`,
		},
		{
			name:     "fallback to name for sent item",
			itemName: "Interface eth0: Bits sent",
			itemKey:  "ifHCOutOctets.1",
			want:     "ifHCInOctets.1",
		},
		{
			name:     "fallback to name for received item",
			itemName: "Interface eth0: Bits received",
			itemKey:  "ifHCInOctets.1",
			want:     "ifHCOutOctets.1",
		},
		{
			name:     "reject unsupported key",
			itemName: "CPU usage",
			itemKey:  "system.cpu.util",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := derivePeerTrafficItemKey(tt.itemName, tt.itemKey)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got key %q", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("expected %q, got %q", tt.want, got)
			}
		})
	}
}

func TestEnsureEdgeLabel(t *testing.T) {
	var edge AEdge

	label := ensureEdgeLabel(&edge)
	if label == nil {
		t.Fatal("expected label to be initialized")
	}
	if len(edge.Labels) != 1 {
		t.Fatalf("expected 1 label, got %d", len(edge.Labels))
	}
	if edge.Labels[0].Position.Distance != "50%" {
		t.Fatalf("expected default distance, got %q", edge.Labels[0].Position.Distance)
	}
	if !edge.Labels[0].Position.Options.EnsureLegibility {
		t.Fatal("expected EnsureLegibility default to be true")
	}

	label.Attrs.Label.Text = "traffic"
	label2 := ensureEdgeLabel(&edge)
	if label2.Attrs.Label.Text != "traffic" {
		t.Fatalf("expected existing label to be preserved, got %q", label2.Attrs.Label.Text)
	}
}
