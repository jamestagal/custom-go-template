package tests

import (
	"testing"

	"github.com/jimafisk/custom_go_template/ast"
	"github.com/jimafisk/custom_go_template/renderer"
)

// TestSimpleFlatJSONInjection tests injecting simple flat JSON into exported props
// Cognitive Load: 5
func TestSimpleFlatJSONInjection(t *testing.T) {
	fence := &ast.FenceSection{
		ExportedProps: []string{"title", "description"},
		Props:         []ast.PropNode{},
		Variables:     []ast.VariableNode{},
	}

	contentData := map[string]interface{}{
		"title":       "My Blog Post",
		"description": "This is a test blog post",
	}

	result, err := renderer.InjectContentProps(fence, contentData)
	if err != nil {
		t.Fatalf("InjectContentProps failed: %v", err)
	}

	// Check that exported props were injected
	if len(result.Variables) != 2 {
		t.Errorf("Expected 2 variables after injection, got %d", len(result.Variables))
	}

	// Verify values were injected correctly (strconv.Quote uses double quotes)
	foundTitle := false
	foundDescription := false
	for _, v := range result.Variables {
		if v.Name == "title" && v.Value == `"My Blog Post"` {
			foundTitle = true
		}
		if v.Name == "description" && v.Value == `"This is a test blog post"` {
			foundDescription = true
		}
	}

	if !foundTitle {
		t.Errorf("Expected 'title' variable to be injected with correct value")
		// Debug: print actual values
		for _, v := range result.Variables {
			if v.Name == "title" {
				t.Logf("Actual title value: %q", v.Value)
			}
		}
	}
	if !foundDescription {
		t.Error("Expected 'description' variable to be injected with correct value")
	}
}

// TestPlentiComponentsArrayInjection tests Plenti components array format
// Cognitive Load: 8
func TestPlentiComponentsArrayInjection(t *testing.T) {
	fence := &ast.FenceSection{
		ExportedProps: []string{"title", "subtitle"},
		Props:         []ast.PropNode{},
		Variables:     []ast.VariableNode{},
	}

	// Plenti format with components array
	contentData := map[string]interface{}{
		"components": []interface{}{
			map[string]interface{}{
				"name": "page_header",
				"fields": map[string]interface{}{
					"title":    "Welcome to My Site",
					"subtitle": "A great place to learn",
				},
			},
		},
	}

	// Need to extract component fields first
	componentFields, ok := contentData["components"].([]interface{})
	if !ok || len(componentFields) == 0 {
		t.Fatal("Test setup error: invalid components array")
	}

	firstComponent := componentFields[0].(map[string]interface{})
	fields := firstComponent["fields"].(map[string]interface{})

	result, err := renderer.InjectContentProps(fence, fields)
	if err != nil {
		t.Fatalf("InjectContentProps failed: %v", err)
	}

	// Check that exported props were injected
	if len(result.Variables) != 2 {
		t.Errorf("Expected 2 variables after injection, got %d", len(result.Variables))
	}

	// Verify values (using double quotes)
	foundTitle := false
	foundSubtitle := false
	for _, v := range result.Variables {
		if v.Name == "title" && v.Value == `"Welcome to My Site"` {
			foundTitle = true
		}
		if v.Name == "subtitle" && v.Value == `"A great place to learn"` {
			foundSubtitle = true
		}
	}

	if !foundTitle {
		t.Error("Expected 'title' variable to be injected")
	}
	if !foundSubtitle {
		t.Error("Expected 'subtitle' variable to be injected")
	}
}

// TestMixedExportLetAndRegularProps tests both exported and regular props coexisting
// Cognitive Load: 7
func TestMixedExportLetAndRegularProps(t *testing.T) {
	fence := &ast.FenceSection{
		ExportedProps: []string{"title", "author"},
		Props: []ast.PropNode{
			{Name: "readTime", DefaultValue: "5 min"},
			{Name: "tags", DefaultValue: "['go', 'testing']"},
		},
		Variables: []ast.VariableNode{},
	}

	contentData := map[string]interface{}{
		"title":  "Understanding Go",
		"author": "Jane Doe",
	}

	result, err := renderer.InjectContentProps(fence, contentData)
	if err != nil {
		t.Fatalf("InjectContentProps failed: %v", err)
	}

	// Check that exported props were injected as variables
	if len(result.Variables) != 2 {
		t.Errorf("Expected 2 variables (exported props), got %d", len(result.Variables))
	}

	// Check that regular props remain unchanged
	if len(result.Props) != 2 {
		t.Errorf("Expected 2 props to remain, got %d", len(result.Props))
	}

	// Verify prop values weren't modified
	for _, p := range result.Props {
		if p.Name == "readTime" && p.DefaultValue != "5 min" {
			t.Errorf("Expected readTime prop to remain '5 min', got '%s'", p.DefaultValue)
		}
	}
}

// TestMissingPropsWithDefaults tests missing exported props that have default values
// Cognitive Load: 6
func TestMissingPropsWithDefaults(t *testing.T) {
	fence := &ast.FenceSection{
		ExportedProps: []string{"title", "description"},
		Props: []ast.PropNode{
			{Name: "title", DefaultValue: "Default Title"},
			{Name: "description", DefaultValue: "Default Description"},
		},
		Variables: []ast.VariableNode{},
	}

	// Content data is missing both exported props
	contentData := map[string]interface{}{}

	result, err := renderer.InjectContentProps(fence, contentData)
	if err != nil {
		t.Fatalf("Expected no error when defaults exist, got: %v", err)
	}

	// Should use default values from Props
	if len(result.Variables) != 2 {
		t.Errorf("Expected 2 variables with defaults, got %d", len(result.Variables))
	}

	// Verify default values were used
	foundTitleDefault := false
	foundDescDefault := false
	for _, v := range result.Variables {
		if v.Name == "title" && v.Value == "Default Title" {
			foundTitleDefault = true
		}
		if v.Name == "description" && v.Value == "Default Description" {
			foundDescDefault = true
		}
	}

	if !foundTitleDefault {
		t.Error("Expected 'title' to use default value")
	}
	if !foundDescDefault {
		t.Error("Expected 'description' to use default value")
	}
}

// TestMissingPropsWithoutDefaults tests error when exported prop missing and no default
// Cognitive Load: 5
func TestMissingPropsWithoutDefaults(t *testing.T) {
	fence := &ast.FenceSection{
		ExportedProps: []string{"title", "author"},
		Props:         []ast.PropNode{},
		Variables:     []ast.VariableNode{},
	}

	// Content data is missing 'author'
	contentData := map[string]interface{}{
		"title": "My Post",
	}

	_, err := renderer.InjectContentProps(fence, contentData)
	if err == nil {
		t.Fatal("Expected error for missing prop without default, got nil")
	}

	// Error message should mention the missing prop
	if err.Error() == "" {
		t.Error("Expected non-empty error message")
	}
}

// TestEmptyJSONUsesDefaults tests empty JSON uses all default values
// Cognitive Load: 5
func TestEmptyJSONUsesDefaults(t *testing.T) {
	fence := &ast.FenceSection{
		ExportedProps: []string{"title", "description", "author"},
		Props: []ast.PropNode{
			{Name: "title", DefaultValue: "Untitled"},
			{Name: "description", DefaultValue: "No description"},
			{Name: "author", DefaultValue: "Anonymous"},
		},
		Variables: []ast.VariableNode{},
	}

	// Empty content data
	contentData := map[string]interface{}{}

	result, err := renderer.InjectContentProps(fence, contentData)
	if err != nil {
		t.Fatalf("Expected no error with all defaults, got: %v", err)
	}

	// All exported props should use defaults
	if len(result.Variables) != 3 {
		t.Errorf("Expected 3 variables with defaults, got %d", len(result.Variables))
	}
}

// TestExportedPropsOverrideDefaults tests that JSON values override prop defaults
// Cognitive Load: 6
func TestExportedPropsOverrideDefaults(t *testing.T) {
	fence := &ast.FenceSection{
		ExportedProps: []string{"title", "description"},
		Props: []ast.PropNode{
			{Name: "title", DefaultValue: "Default Title"},
			{Name: "description", DefaultValue: "Default Description"},
		},
		Variables: []ast.VariableNode{},
	}

	// Content data provides new values
	contentData := map[string]interface{}{
		"title":       "Actual Title",
		"description": "Actual Description",
	}

	result, err := renderer.InjectContentProps(fence, contentData)
	if err != nil {
		t.Fatalf("InjectContentProps failed: %v", err)
	}

	// Verify JSON values were used, not defaults (double quotes from strconv.Quote)
	foundActualTitle := false
	foundActualDesc := false
	for _, v := range result.Variables {
		if v.Name == "title" && v.Value == `"Actual Title"` {
			foundActualTitle = true
		}
		if v.Name == "description" && v.Value == `"Actual Description"` {
			foundActualDesc = true
		}
	}

	if !foundActualTitle {
		t.Error("Expected 'title' to override default with JSON value")
	}
	if !foundActualDesc {
		t.Error("Expected 'description' to override default with JSON value")
	}
}

// TestNoExportedPropsStillWorks tests backward compatibility without export let
// Cognitive Load: 4
func TestNoExportedPropsStillWorks(t *testing.T) {
	fence := &ast.FenceSection{
		ExportedProps: []string{}, // No exported props
		Props: []ast.PropNode{
			{Name: "title", DefaultValue: "Default"},
		},
		Variables: []ast.VariableNode{
			{Keyword: "let", Name: "counter", Value: "0"},
		},
	}

	contentData := map[string]interface{}{
		"title": "Should be ignored",
	}

	result, err := renderer.InjectContentProps(fence, contentData)
	if err != nil {
		t.Fatalf("InjectContentProps should not fail without exported props: %v", err)
	}

	// Props and variables should remain unchanged
	if len(result.Props) != 1 {
		t.Errorf("Expected 1 prop to remain, got %d", len(result.Props))
	}

	if len(result.Variables) != 1 {
		t.Errorf("Expected 1 variable to remain, got %d", len(result.Variables))
	}
}

// TestNumericValueInjection tests injecting numeric values from JSON
// Cognitive Load: 5
func TestNumericValueInjection(t *testing.T) {
	fence := &ast.FenceSection{
		ExportedProps: []string{"count", "price"},
		Props:         []ast.PropNode{},
		Variables:     []ast.VariableNode{},
	}

	contentData := map[string]interface{}{
		"count": float64(42),
		"price": float64(99.99),
	}

	result, err := renderer.InjectContentProps(fence, contentData)
	if err != nil {
		t.Fatalf("InjectContentProps failed: %v", err)
	}

	// Verify numeric values were injected
	foundCount := false
	foundPrice := false
	for _, v := range result.Variables {
		if v.Name == "count" && v.Value == "42" {
			foundCount = true
		}
		if v.Name == "price" && v.Value == "99.99" {
			foundPrice = true
		}
	}

	if !foundCount {
		t.Error("Expected 'count' numeric value to be injected")
	}
	if !foundPrice {
		t.Error("Expected 'price' numeric value to be injected")
	}
}

// TestBooleanValueInjection tests injecting boolean values from JSON
// Cognitive Load: 5
func TestBooleanValueInjection(t *testing.T) {
	fence := &ast.FenceSection{
		ExportedProps: []string{"isPublished", "featured"},
		Props:         []ast.PropNode{},
		Variables:     []ast.VariableNode{},
	}

	contentData := map[string]interface{}{
		"isPublished": true,
		"featured":    false,
	}

	result, err := renderer.InjectContentProps(fence, contentData)
	if err != nil {
		t.Fatalf("InjectContentProps failed: %v", err)
	}

	// Verify boolean values were injected
	foundPublished := false
	foundFeatured := false
	for _, v := range result.Variables {
		if v.Name == "isPublished" && v.Value == "true" {
			foundPublished = true
		}
		if v.Name == "featured" && v.Value == "false" {
			foundFeatured = true
		}
	}

	if !foundPublished {
		t.Error("Expected 'isPublished' boolean to be injected")
	}
	if !foundFeatured {
		t.Error("Expected 'featured' boolean to be injected")
	}
}

// TestPartialContentInjection tests injecting only some exported props
// Cognitive Load: 6
func TestPartialContentInjection(t *testing.T) {
	fence := &ast.FenceSection{
		ExportedProps: []string{"title", "author", "date"},
		Props: []ast.PropNode{
			{Name: "author", DefaultValue: "Anonymous"},
			{Name: "date", DefaultValue: "2024-01-01"},
		},
		Variables: []ast.VariableNode{},
	}

	// Content provides only title
	contentData := map[string]interface{}{
		"title": "My Article",
	}

	result, err := renderer.InjectContentProps(fence, contentData)
	if err != nil {
		t.Fatalf("Expected no error with partial content and defaults: %v", err)
	}

	// Should have 3 variables: 1 from content, 2 from defaults
	if len(result.Variables) != 3 {
		t.Errorf("Expected 3 variables (1 from content, 2 from defaults), got %d", len(result.Variables))
	}

	// Verify the injected value
	foundTitle := false
	foundAuthorDefault := false
	foundDateDefault := false
	for _, v := range result.Variables {
		if v.Name == "title" && v.Value == `"My Article"` {
			foundTitle = true
		}
		if v.Name == "author" && v.Value == "Anonymous" {
			foundAuthorDefault = true
		}
		if v.Name == "date" && v.Value == "2024-01-01" {
			foundDateDefault = true
		}
	}

	if !foundTitle {
		t.Error("Expected 'title' from content to be injected")
	}
	if !foundAuthorDefault {
		t.Error("Expected 'author' to use default value")
	}
	if !foundDateDefault {
		t.Error("Expected 'date' to use default value")
	}
}
