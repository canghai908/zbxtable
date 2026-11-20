package model

import (
	"fmt"
	"regexp"
	"strings"
)

// ConditionNode is the versioned, shared dispatch and mute predicate tree.
type ConditionNode struct {
	Version  int             `json:"version,omitempty"`
	Op       string          `json:"op,omitempty"`
	Children []ConditionNode `json:"children,omitempty"`
	Field    string          `json:"field,omitempty"`
	Operator string          `json:"operator,omitempty"`
	Value    string          `json:"value,omitempty"`
}

type ConditionTrace struct {
	Matched  bool             `json:"matched"`
	Op       string           `json:"op,omitempty"`
	Field    string           `json:"field,omitempty"`
	Operator string           `json:"operator,omitempty"`
	Value    string           `json:"value,omitempty"`
	Actual   string           `json:"actual,omitempty"`
	Children []ConditionTrace `json:"children,omitempty"`
}

func ParseConditionTree(raw string) (*ConditionNode, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" || raw == "null" {
		return nil, nil
	}
	var root ConditionNode
	if strings.HasPrefix(raw, "[") {
		var legacy []Conditions
		if err := json.Unmarshal([]byte(raw), &legacy); err != nil {
			return nil, err
		}
		if len(legacy) == 0 {
			return nil, nil
		}
		root = ConditionNode{Version: 2, Op: "and"}
		for _, c := range legacy {
			root.Children = append(root.Children, ConditionNode{Field: c.RType, Operator: strings.TrimSpace(c.RFunc), Value: c.Rvalue})
		}
	} else {
		if err := json.Unmarshal([]byte(raw), &root); err != nil {
			return nil, err
		}
		if root.Version != 2 {
			return nil, fmt.Errorf("不支持的条件版本")
		}
		if root.Op == "" {
			return nil, fmt.Errorf("根节点必须为条件组")
		}
	}
	count := 0
	if err := validateConditionNode(root, 1, &count); err != nil {
		return nil, err
	}
	return &root, nil
}

func validateConditionNode(n ConditionNode, depth int, count *int) error {
	if n.Op != "" {
		if depth > 3 || (n.Op != "and" && n.Op != "or") || len(n.Children) == 0 || n.Field != "" || n.Operator != "" || n.Value != "" {
			return fmt.Errorf("条件组必须使用 AND/OR、非空且最多三级")
		}
		for i, c := range n.Children {
			if err := validateConditionNode(c, depth+1, count); err != nil {
				return fmt.Errorf("条件 %d: %w", i+1, err)
			}
		}
		return nil
	}
	if len(n.Children) > 0 {
		return fmt.Errorf("条件缺少分组运算符")
	}
	*count++
	if *count > 50 {
		return fmt.Errorf("最多支持 50 个条件")
	}
	switch n.Field {
	case "host", "group", "item", "key", "trigger", "severity":
	default:
		return fmt.Errorf("未知条件字段 %q", n.Field)
	}
	switch n.Operator {
	case "=", "==", "!=", "like", "not_contains":
	case "=~", "!~":
		if _, err := regexp.Compile(n.Value); err != nil {
			return fmt.Errorf("无效正则: %w", err)
		}
	default:
		return fmt.Errorf("未知条件操作符 %q", n.Operator)
	}
	return nil
}

// EvaluateConditionTree is also used for previews. It evaluates every child to
// explain the result without changing AND/OR semantics or sending notifications.
func EvaluateConditionTree(n *ConditionNode, data map[string]any) ConditionTrace {
	if n == nil {
		return ConditionTrace{Matched: false}
	}
	t := ConditionTrace{Op: n.Op, Field: n.Field, Operator: n.Operator, Value: n.Value}
	if n.Op != "" {
		t.Matched = n.Op == "and"
		for i := range n.Children {
			child := EvaluateConditionTree(&n.Children[i], data)
			t.Children = append(t.Children, child)
			if n.Op == "and" {
				t.Matched = t.Matched && child.Matched
			} else {
				t.Matched = t.Matched || child.Matched
			}
		}
		return t
	}
	actual, ok := data[n.Field].(string)
	t.Actual = actual
	if !ok {
		return t
	}
	switch n.Operator {
	case "=", "==":
		t.Matched = actual == n.Value
	case "!=":
		t.Matched = actual != n.Value
	case "like":
		t.Matched = strings.Contains(actual, n.Value)
	case "not_contains":
		t.Matched = !strings.Contains(actual, n.Value)
	case "=~", "!~":
		matched, err := regexp.MatchString(n.Value, actual)
		if err == nil {
			t.Matched = matched
			if n.Operator == "!~" {
				t.Matched = !matched
			}
		}
	}
	return t
}

func ValidateRuleConditions(raw, mType string) error {
	tree, err := ParseConditionTree(raw)
	if err != nil {
		return err
	}
	if tree == nil && mType != "2" {
		return fmt.Errorf("普通规则和屏蔽规则至少需要一个条件")
	}
	return nil
}
