package parser

import (
	"strings"
	"testing"

	"github.com/jimafisk/custom_go_template/ast"
)

// TestDynamicComponentByNameParser_Basic tests basic syntax parsing
func TestDynamicComponentByNameParser_Basic(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		wantName       string
		wantPropsCount int
		wantSuccess    bool
		wantError      string
	}{
		{
			name:           "basic with name expression",
			input:          `<Component:dynamic name={component.name} />`,
			wantName:       "component.name",
			wantPropsCount: 0,
			wantSuccess:    true,
		},
		{
			name:           "name as string literal",
			input:          `<Component:dynamic name="Hero2436" />`,
			wantName:       "Hero2436",
			wantPropsCount: 0,
			wantSuccess:    true,
		},
		{
			name:           "name with single quotes",
			input:          `<Component:dynamic name='Header' />`,
			wantName:       "Header",
			wantPropsCount: 0,
			wantSuccess:    true,
		},
		{
			name:           "with whitespace",
			input:          `<Component:dynamic   name={comp.name}   />`,
			wantName:       "comp.name",
			wantPropsCount: 0,
			wantSuccess:    true,
		},
		{
			name:        "missing name attribute",
			input:       `<Component:dynamic title="test" />`,
			wantSuccess: false,
			wantError:   "name attribute",
		},
		{
			name:        "wrong tag name",
			input:       `<Component name={x} />`,
			wantSuccess: false,
		},
		{
			name:        "lowercase component:dynamic",
			input:       `<component:dynamic name={x} />`,
			wantSuccess: false,
		},
		{
			name:        "missing closing tag",
			input:       `<Component:dynamic name={x}`,
			wantSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DynamicComponentByNameParser()(tt.input)

			if result.Successful != tt.wantSuccess {
				t.Errorf("Success = %v, want %v", result.Successful, tt.wantSuccess)
				if !tt.wantSuccess && tt.wantError != "" {
					if !strings.Contains(result.Error, tt.wantError) {
						t.Errorf("Error = %q, want error containing %q", result.Error, tt.wantError)
					}
				}
				return
			}

			if !tt.wantSuccess {
				return // Test passed, error case handled correctly
			}

			node, ok := result.Value.(*ast.DynamicComponentByNameNode)
			if !ok {
				t.Fatalf("Value is not *ast.DynamicComponentByNameNode, got %T", result.Value)
			}

			if node.NameExpression != tt.wantName {
				t.Errorf("NameExpression = %q, want %q", node.NameExpression, tt.wantName)
			}

			if len(node.Props) != tt.wantPropsCount {
				t.Errorf("Props count = %d, want %d", len(node.Props), tt.wantPropsCount)
			}

			if !node.SelfClosing {
				t.Error("Expected self-closing tag")
			}
		})
	}
}

// TestDynamicComponentByNameParser_WithSpread tests spread operator parsing
func TestDynamicComponentByNameParser_WithSpread(t *testing.T) {
	tests := []struct {
		name              string
		input             string
		wantName          string
		wantSpreadCount   int
		wantSpreadExprs   []string
		wantRegularProps  int
		wantSuccess       bool
	}{
		{
			name:            "single spread",
			input:           `<Component:dynamic name={component.name} {...component.fields} />`,
			wantName:        "component.name",
			wantSpreadCount: 1,
			wantSpreadExprs: []string{"component.fields"},
			wantRegularProps: 0,
			wantSuccess:     true,
		},
		{
			name:            "multiple spreads",
			input:           `<Component:dynamic name={comp.name} {...defaults} {...overrides} />`,
			wantName:        "comp.name",
			wantSpreadCount: 2,
			wantSpreadExprs: []string{"defaults", "overrides"},
			wantRegularProps: 0,
			wantSuccess:     true,
		},
		{
			name:            "spread with nested property",
			input:           `<Component:dynamic name={x} {...component.fields.main} />`,
			wantName:        "x",
			wantSpreadCount: 1,
			wantSpreadExprs: []string{"component.fields.main"},
			wantRegularProps: 0,
			wantSuccess:     true,
		},
		{
			name:            "spread with complex expression",
			input:           `<Component:dynamic name="Hero" {...$store.settings.theme} />`,
			wantName:        "Hero",
			wantSpreadCount: 1,
			wantSpreadExprs: []string{"$store.settings.theme"},
			wantRegularProps: 0,
			wantSuccess:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DynamicComponentByNameParser()(tt.input)

			if result.Successful != tt.wantSuccess {
				t.Errorf("Success = %v, want %v. Error: %s", result.Successful, tt.wantSuccess, result.Error)
				return
			}

			node, ok := result.Value.(*ast.DynamicComponentByNameNode)
			if !ok {
				t.Fatalf("Value is not *ast.DynamicComponentByNameNode, got %T", result.Value)
			}

			if node.NameExpression != tt.wantName {
				t.Errorf("NameExpression = %q, want %q", node.NameExpression, tt.wantName)
			}

			if len(node.SpreadProps) != tt.wantSpreadCount {
				t.Errorf("SpreadProps count = %d, want %d", len(node.SpreadProps), tt.wantSpreadCount)
			}

			for i, wantExpr := range tt.wantSpreadExprs {
				if i >= len(node.SpreadProps) {
					t.Errorf("Missing spread prop at index %d", i)
					continue
				}
				if node.SpreadProps[i] != wantExpr {
					t.Errorf("SpreadProps[%d] = %q, want %q", i, node.SpreadProps[i], wantExpr)
				}
			}

			if len(node.Props) != tt.wantRegularProps {
				t.Errorf("Props count = %d, want %d", len(node.Props), tt.wantRegularProps)
			}
		})
	}
}

// TestDynamicComponentByNameParser_MixedProps tests mixed spread and regular props
func TestDynamicComponentByNameParser_MixedProps(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		wantName        string
		wantSpreadCount int
		wantPropsCount  int
		wantPropsNames  []string
		wantSuccess     bool
	}{
		{
			name:            "spread then regular",
			input:           `<Component:dynamic name={component.name} {...component.fields} theme="dark" />`,
			wantName:        "component.name",
			wantSpreadCount: 1,
			wantPropsCount:  1,
			wantPropsNames:  []string{"theme"},
			wantSuccess:     true,
		},
		{
			name:            "regular then spread",
			input:           `<Component:dynamic name="Hero" title="Hello" {...overrides} />`,
			wantName:        "Hero",
			wantSpreadCount: 1,
			wantPropsCount:  1,
			wantPropsNames:  []string{"title"},
			wantSuccess:     true,
		},
		{
			name:            "multiple spreads and regular",
			input:           `<Component:dynamic name={comp.name} {...defaults} title="Test" {...overrides} debug={true} />`,
			wantName:        "comp.name",
			wantSpreadCount: 2,
			wantPropsCount:  2,
			wantPropsNames:  []string{"title", "debug"},
			wantSuccess:     true,
		},
		{
			name:            "all prop types",
			input:           `<Component:dynamic name={x} {...comp.fields} allContent={allContent} content={content} />`,
			wantName:        "x",
			wantSpreadCount: 1,
			wantPropsCount:  2,
			wantPropsNames:  []string{"allContent", "content"},
			wantSuccess:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DynamicComponentByNameParser()(tt.input)

			if result.Successful != tt.wantSuccess {
				t.Errorf("Success = %v, want %v. Error: %s", result.Successful, tt.wantSuccess, result.Error)
				return
			}

			node, ok := result.Value.(*ast.DynamicComponentByNameNode)
			if !ok {
				t.Fatalf("Value is not *ast.DynamicComponentByNameNode, got %T", result.Value)
			}

			if node.NameExpression != tt.wantName {
				t.Errorf("NameExpression = %q, want %q", node.NameExpression, tt.wantName)
			}

			if len(node.SpreadProps) != tt.wantSpreadCount {
				t.Errorf("SpreadProps count = %d, want %d", len(node.SpreadProps), tt.wantSpreadCount)
			}

			if len(node.Props) != tt.wantPropsCount {
				t.Errorf("Props count = %d, want %d", len(node.Props), tt.wantPropsCount)
			}

			for i, wantName := range tt.wantPropsNames {
				if i >= len(node.Props) {
					t.Errorf("Missing prop at index %d", i)
					continue
				}
				if node.Props[i].Name != wantName {
					t.Errorf("Props[%d].Name = %q, want %q", i, node.Props[i].Name, wantName)
				}
			}
		})
	}
}

// TestDynamicComponentByNameParser_MultipleSpread tests multiple spread operators
func TestDynamicComponentByNameParser_MultipleSpread(t *testing.T) {
	input := `<Component:dynamic name={component.name} {...defaults} {...component.fields} {...overrides} />`

	result := DynamicComponentByNameParser()(input)

	if !result.Successful {
		t.Fatalf("Parse failed: %s", result.Error)
	}

	node, ok := result.Value.(*ast.DynamicComponentByNameNode)
	if !ok {
		t.Fatalf("Value is not *ast.DynamicComponentByNameNode, got %T", result.Value)
	}

	if node.NameExpression != "component.name" {
		t.Errorf("NameExpression = %q, want %q", node.NameExpression, "component.name")
	}

	if len(node.SpreadProps) != 3 {
		t.Fatalf("SpreadProps count = %d, want 3", len(node.SpreadProps))
	}

	wantSpreads := []string{"defaults", "component.fields", "overrides"}
	for i, want := range wantSpreads {
		if node.SpreadProps[i] != want {
			t.Errorf("SpreadProps[%d] = %q, want %q", i, node.SpreadProps[i], want)
		}
	}
}

// TestDynamicComponentByNameParser_RegularPropsOnly tests regular props without spread
func TestDynamicComponentByNameParser_RegularPropsOnly(t *testing.T) {
	input := `<Component:dynamic name="Header" title="Hello" subtitle={pageTitle} />`

	result := DynamicComponentByNameParser()(input)

	if !result.Successful {
		t.Fatalf("Parse failed: %s", result.Error)
	}

	node, ok := result.Value.(*ast.DynamicComponentByNameNode)
	if !ok {
		t.Fatalf("Value is not *ast.DynamicComponentByNameNode, got %T", result.Value)
	}

	if node.NameExpression != "Header" {
		t.Errorf("NameExpression = %q, want %q", node.NameExpression, "Header")
	}

	if len(node.SpreadProps) != 0 {
		t.Errorf("SpreadProps count = %d, want 0", len(node.SpreadProps))
	}

	if len(node.Props) != 2 {
		t.Fatalf("Props count = %d, want 2", len(node.Props))
	}

	// Check title prop
	if node.Props[0].Name != "title" {
		t.Errorf("Props[0].Name = %q, want %q", node.Props[0].Name, "title")
	}
	if node.Props[0].Value != "Hello" {
		t.Errorf("Props[0].Value = %q, want %q", node.Props[0].Value, "Hello")
	}
	if node.Props[0].IsDynamic {
		t.Error("Props[0].IsDynamic should be false for static string")
	}

	// Check subtitle prop
	if node.Props[1].Name != "subtitle" {
		t.Errorf("Props[1].Name = %q, want %q", node.Props[1].Name, "subtitle")
	}
	if node.Props[1].Value != "pageTitle" {
		t.Errorf("Props[1].Value = %q, want %q", node.Props[1].Value, "pageTitle")
	}
	if !node.Props[1].IsDynamic {
		t.Error("Props[1].IsDynamic should be true for expression")
	}
}

// TestDynamicComponentByNameParser_NonSelfClosing tests non-self-closing tags
func TestDynamicComponentByNameParser_NonSelfClosing(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantSuccess bool
		wantError   string
	}{
		{
			name:        "non-self-closing with children",
			input:       `<Component:dynamic name={x}>children</Component:dynamic>`,
			wantSuccess: false,
			wantError:   "self-closing",
		},
		{
			name:        "non-self-closing empty",
			input:       `<Component:dynamic name={x}></Component:dynamic>`,
			wantSuccess: false,
			wantError:   "self-closing",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DynamicComponentByNameParser()(tt.input)

			if result.Successful != tt.wantSuccess {
				t.Errorf("Success = %v, want %v", result.Successful, tt.wantSuccess)
			}

			if !tt.wantSuccess && tt.wantError != "" {
				if !strings.Contains(result.Error, tt.wantError) {
					t.Errorf("Error = %q, want error containing %q", result.Error, tt.wantError)
				}
			}
		})
	}
}

// TestDynamicComponentByNameParser_ErrorCases tests various error conditions
func TestDynamicComponentByNameParser_ErrorCases(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantError string
	}{
		{
			name:      "invalid spread syntax - missing closing brace",
			input:     `<Component:dynamic name={x} {...props />`,
			wantError: "spread",
		},
		{
			name:      "malformed name expression",
			input:     `<Component:dynamic name={x />`,
			wantError: "brace",
		},
		{
			name:      "empty name attribute",
			input:     `<Component:dynamic name="" />`,
			wantError: "name",
		},
		{
			name:      "name attribute without value",
			input:     `<Component:dynamic name />`,
			wantError: "name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DynamicComponentByNameParser()(tt.input)

			if result.Successful {
				t.Error("Expected parse to fail, but it succeeded")
			}

			if !strings.Contains(result.Error, tt.wantError) {
				t.Errorf("Error = %q, want error containing %q", result.Error, tt.wantError)
			}
		})
	}
}

// TestDynamicComponentByNameParser_Whitespace tests whitespace handling
func TestDynamicComponentByNameParser_Whitespace(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantOk  bool
	}{
		{
			name:   "extra whitespace around attributes",
			input:  `<Component:dynamic   name={x}   {...props}   theme="dark"   />`,
			wantOk: true,
		},
		{
			name:   "newlines in tag",
			input:  "<Component:dynamic\n  name={x}\n  {...props}\n/>",
			wantOk: true,
		},
		{
			name:   "tabs in tag",
			input:  "<Component:dynamic\tname={x}\t{...props}\t/>",
			wantOk: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DynamicComponentByNameParser()(tt.input)

			if result.Successful != tt.wantOk {
				t.Errorf("Success = %v, want %v. Error: %s", result.Successful, tt.wantOk, result.Error)
			}
		})
	}
}

// TestDynamicComponentByNameParser_OrderPreservation tests that prop order is preserved
func TestDynamicComponentByNameParser_OrderPreservation(t *testing.T) {
	input := `<Component:dynamic name={x} {...a} prop1="1" {...b} prop2="2" {...c} prop3="3" />`

	result := DynamicComponentByNameParser()(input)

	if !result.Successful {
		t.Fatalf("Parse failed: %s", result.Error)
	}

	node, ok := result.Value.(*ast.DynamicComponentByNameNode)
	if !ok {
		t.Fatalf("Value is not *ast.DynamicComponentByNameNode, got %T", result.Value)
	}

	// Check spread props order
	wantSpreads := []string{"a", "b", "c"}
	if len(node.SpreadProps) != len(wantSpreads) {
		t.Fatalf("SpreadProps count = %d, want %d", len(node.SpreadProps), len(wantSpreads))
	}
	for i, want := range wantSpreads {
		if node.SpreadProps[i] != want {
			t.Errorf("SpreadProps[%d] = %q, want %q", i, node.SpreadProps[i], want)
		}
	}

	// Check regular props order
	wantPropNames := []string{"prop1", "prop2", "prop3"}
	if len(node.Props) != len(wantPropNames) {
		t.Fatalf("Props count = %d, want %d", len(node.Props), len(wantPropNames))
	}
	for i, want := range wantPropNames {
		if node.Props[i].Name != want {
			t.Errorf("Props[%d].Name = %q, want %q", i, node.Props[i].Name, want)
		}
	}
}
