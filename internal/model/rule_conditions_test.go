package model

import "testing"

func TestConditionTrees(t *testing.T) {
	data := map[string]any{"host": "prod-01", "trigger": "CPU high", "severity": "High"}
	cases := []struct {
		name, raw string
		want      bool
	}{
		{"legacy equality alias", `[{"r_type":"host","r_func":"==","r_value":"prod-01"}]`, true},
		{"legacy AND", `[{"r_type":"host","r_func":"like","r_value":"prod-"},{"r_type":"severity","r_func":"=","r_value":"High"}]`, true},
		{"legacy AND mismatch", `[{"r_type":"host","r_func":"=","r_value":"other"},{"r_type":"severity","r_func":"=","r_value":"High"}]`, false},
		{"nested OR", `{"version":2,"op":"and","children":[{"field":"host","operator":"like","value":"prod-"},{"op":"or","children":[{"field":"trigger","operator":"like","value":"Memory"},{"field":"trigger","operator":"like","value":"CPU"}]}]}`, true},
		{"OR mismatch", `{"version":2,"op":"or","children":[{"field":"host","operator":"=","value":"other"},{"field":"severity","operator":"=","value":"Low"}]}`, false},
		{"literal escapes", `{"version":2,"op":"and","children":[{"field":"host","operator":"!=","value":"a'\\b"}]}`, true},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			tree, err := ParseConditionTree(tt.raw)
			if err != nil {
				t.Fatal(err)
			}
			trace := EvaluateConditionTree(tree, data)
			if trace.Matched != tt.want {
				t.Fatalf("trace: %+v", trace)
			}
			if got := matchRule(&Rule{MType: "1", Conditions: tt.raw}, data, GetCELMatcher()); got != tt.want {
				t.Fatalf("production differs: %v", got)
			}
		})
	}
}

func TestConditionValidation(t *testing.T) {
	for _, raw := range []string{
		`{"version":3,"op":"and","children":[]}`,
		`{"version":2,"op":"or","children":[]}`,
		`{"version":2,"op":"and","children":[{"field":"bad","operator":"=","value":"x"}]}`,
		`[{"r_type":"host","r_func":"=~","r_value":"["}]`,
		`{"version":2,"op":"and","children":[{"field":"host","operator":"bad","value":"x"}]}`,
		`{"version":2,"op":"and","children":[{"op":"and","children":[{"op":"or","children":[{"op":"and","children":[{"field":"host","operator":"=","value":"x"}]}]}]}]}`,
	} {
		if _, err := ParseConditionTree(raw); err == nil {
			t.Errorf("accepted %s", raw)
		}
	}
	if err := ValidateRuleConditions("[]", "1"); err == nil {
		t.Fatal("accepted empty custom")
	}
	if err := ValidateRuleConditions("", "2"); err != nil {
		t.Fatal(err)
	}
	tree, _ := ParseConditionTree(`[{"r_type":"host","r_func":"!=","r_value":"x"}]`)
	if EvaluateConditionTree(tree, map[string]any{}).Matched {
		t.Fatal("missing field matched negative")
	}
}

func TestConditionLimitsAndMuteParity(t *testing.T) {
	root := ConditionNode{Version: 2, Op: "and"}
	for i := 0; i < 51; i++ {
		root.Children = append(root.Children, ConditionNode{Field: "host", Operator: "=", Value: "prod"})
	}
	raw, _ := json.Marshal(root)
	if _, err := ParseConditionTree(string(raw)); err == nil {
		t.Fatal("accepted 51 leaves")
	}
	root.Children = root.Children[:50]
	raw, _ = json.Marshal(root)
	if _, err := ParseConditionTree(string(raw)); err != nil {
		t.Fatal(err)
	}
	rule := Rule{MType: "3", Conditions: `{"version":2,"op":"or","children":[{"field":"host","operator":"=","value":"prod"},{"field":"trigger","operator":"like","value":"CPU"}]}`}
	if !MeetEventConditions(&Event{Host: "other", Message: "CPU high"}, &rule) {
		t.Fatal("mute did not use OR")
	}
	if !MeetConditions(&Alarm{Host: "other", Message: "CPU high"}, &rule) {
		t.Fatal("dispatch differs from mute")
	}
	root = ConditionNode{Version: 2, Op: "and", Children: []ConditionNode{{Op: "or", Children: []ConditionNode{{Op: "and", Children: []ConditionNode{{Field: "host", Operator: "=", Value: "prod"}}}}}}}
	raw, _ = json.Marshal(root)
	if _, err := ParseConditionTree(string(raw)); err != nil {
		t.Fatal("third group level rejected", err)
	}
}

func TestConditionNegativeOperatorsAndTrace(t *testing.T) {
	raw := `{"version":2,"op":"and","children":[{"field":"host","operator":"not_contains","value":"dev"},{"field":"trigger","operator":"!~","value":"Memory.*"}]}`
	tree, err := ParseConditionTree(raw)
	if err != nil {
		t.Fatal(err)
	}
	result := EvaluateConditionTree(tree, map[string]any{"host": "prod", "trigger": "CPU high"})
	if !result.Matched || len(result.Children) != 2 || result.Children[0].Actual != "prod" {
		t.Fatalf("bad explanation: %+v", result)
	}
	if EvaluateConditionTree(tree, map[string]any{"host": "dev", "trigger": "CPU high"}).Matched {
		t.Fatal("negative contains failed")
	}
	if EvaluateConditionTree(tree, map[string]any{"host": "prod", "trigger": "Memory high"}).Matched {
		t.Fatal("negative regex failed")
	}
}
