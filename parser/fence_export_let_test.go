package parser

import (
	"testing"
)

// TestParseFenceContent_ExportLetSingle tests single export let declaration
func TestParseFenceContent_ExportLetSingle(t *testing.T) {
	input := `export let title`

	fence := parseFenceContent(input)

	if len(fence.ExportedProps) != 1 {
		t.Errorf("Expected 1 exported prop, got %d", len(fence.ExportedProps))
		return
	}

	if fence.ExportedProps[0] != "title" {
		t.Errorf("Expected exported prop 'title', got '%s'", fence.ExportedProps[0])
	}
}

// TestParseFenceContent_ExportLetMultiple tests comma-separated export let
func TestParseFenceContent_ExportLetMultiple(t *testing.T) {
	input := `export let title, description, author`

	fence := parseFenceContent(input)

	if len(fence.ExportedProps) != 3 {
		t.Errorf("Expected 3 exported props, got %d", len(fence.ExportedProps))
		return
	}

	expectedProps := []string{"title", "description", "author"}
	for i, expected := range expectedProps {
		if fence.ExportedProps[i] != expected {
			t.Errorf("Expected exported prop[%d] '%s', got '%s'", i, expected, fence.ExportedProps[i])
		}
	}
}

// TestParseFenceContent_ExportLetWithWhitespace tests whitespace handling
func TestParseFenceContent_ExportLetWithWhitespace(t *testing.T) {
	input := `export  let  title  ,  description  ,  author`

	fence := parseFenceContent(input)

	if len(fence.ExportedProps) != 3 {
		t.Errorf("Expected 3 exported props, got %d", len(fence.ExportedProps))
		return
	}

	expectedProps := []string{"title", "description", "author"}
	for i, expected := range expectedProps {
		if fence.ExportedProps[i] != expected {
			t.Errorf("Expected exported prop[%d] '%s', got '%s'", i, expected, fence.ExportedProps[i])
		}
	}
}

// TestParseFenceContent_ExportLetMixedWithProps tests export let alongside prop declarations
func TestParseFenceContent_ExportLetMixedWithProps(t *testing.T) {
	input := `export let title, description
prop age = 25
prop name = "Default"`

	fence := parseFenceContent(input)

	// Check exported props
	if len(fence.ExportedProps) != 2 {
		t.Errorf("Expected 2 exported props, got %d", len(fence.ExportedProps))
	}

	expectedExports := []string{"title", "description"}
	for i, expected := range expectedExports {
		if i < len(fence.ExportedProps) && fence.ExportedProps[i] != expected {
			t.Errorf("Expected exported prop[%d] '%s', got '%s'", i, expected, fence.ExportedProps[i])
		}
	}

	// Check regular props
	if len(fence.Props) != 2 {
		t.Errorf("Expected 2 props, got %d", len(fence.Props))
		return
	}

	if fence.Props[0].Name != "age" {
		t.Errorf("Expected prop 'age', got '%s'", fence.Props[0].Name)
	}

	if fence.Props[1].Name != "name" {
		t.Errorf("Expected prop 'name', got '%s'", fence.Props[1].Name)
	}
}

// TestParseFenceContent_ExportLetEmpty tests empty export let
func TestParseFenceContent_ExportLetEmpty(t *testing.T) {
	input := `export let`

	fence := parseFenceContent(input)

	// Empty export let should result in no exported props
	if len(fence.ExportedProps) != 0 {
		t.Errorf("Expected 0 exported props for empty 'export let', got %d", len(fence.ExportedProps))
	}
}

// TestParseFenceContent_ExportLetTrailingComma tests trailing comma handling
func TestParseFenceContent_ExportLetTrailingComma(t *testing.T) {
	input := `export let title, description,`

	fence := parseFenceContent(input)

	// Should ignore trailing comma
	if len(fence.ExportedProps) != 2 {
		t.Errorf("Expected 2 exported props (ignoring trailing comma), got %d", len(fence.ExportedProps))
		return
	}

	expectedProps := []string{"title", "description"}
	for i, expected := range expectedProps {
		if fence.ExportedProps[i] != expected {
			t.Errorf("Expected exported prop[%d] '%s', got '%s'", i, expected, fence.ExportedProps[i])
		}
	}
}

// TestParseFenceContent_MultipleExportLetStatements tests multiple export let lines
func TestParseFenceContent_MultipleExportLetStatements(t *testing.T) {
	input := `export let title, description
export let author
export let publishDate`

	fence := parseFenceContent(input)

	// Should combine all export let statements
	if len(fence.ExportedProps) != 4 {
		t.Errorf("Expected 4 exported props from multiple statements, got %d", len(fence.ExportedProps))
		return
	}

	expectedProps := []string{"title", "description", "author", "publishDate"}
	for i, expected := range expectedProps {
		if fence.ExportedProps[i] != expected {
			t.Errorf("Expected exported prop[%d] '%s', got '%s'", i, expected, fence.ExportedProps[i])
		}
	}
}

// TestParseFenceContent_ExportLetWithSemicolon tests export let with semicolon
func TestParseFenceContent_ExportLetWithSemicolon(t *testing.T) {
	input := `export let title, description;`

	fence := parseFenceContent(input)

	if len(fence.ExportedProps) != 2 {
		t.Errorf("Expected 2 exported props, got %d", len(fence.ExportedProps))
		return
	}

	expectedProps := []string{"title", "description"}
	for i, expected := range expectedProps {
		if fence.ExportedProps[i] != expected {
			t.Errorf("Expected exported prop[%d] '%s', got '%s'", i, expected, fence.ExportedProps[i])
		}
	}
}

// TestParseFenceContent_ExportLetRealWorld tests real-world blog post scenario
func TestParseFenceContent_ExportLetRealWorld(t *testing.T) {
	input := `export let title, description, author, publishDate
prop readTime = "5 min"
let formattedDate = new Date(publishDate).toLocaleDateString()`

	fence := parseFenceContent(input)

	// Check exported props
	if len(fence.ExportedProps) != 4 {
		t.Errorf("Expected 4 exported props, got %d", len(fence.ExportedProps))
	}

	expectedExports := []string{"title", "description", "author", "publishDate"}
	for i, expected := range expectedExports {
		if i < len(fence.ExportedProps) && fence.ExportedProps[i] != expected {
			t.Errorf("Expected exported prop[%d] '%s', got '%s'", i, expected, fence.ExportedProps[i])
		}
	}

	// Check regular props
	if len(fence.Props) != 1 {
		t.Errorf("Expected 1 prop, got %d", len(fence.Props))
	}

	// Check variables
	if len(fence.Variables) != 1 {
		t.Errorf("Expected 1 variable, got %d", len(fence.Variables))
	}
}

// TestParseFenceContent_NoExportLet tests backward compatibility when no export let
func TestParseFenceContent_NoExportLet(t *testing.T) {
	input := `prop title = "Default Title"
prop description = "Default description"
let counter = 0`

	fence := parseFenceContent(input)

	// Should have no exported props
	if len(fence.ExportedProps) != 0 {
		t.Errorf("Expected 0 exported props when no export let present, got %d", len(fence.ExportedProps))
	}

	// But should still have props and variables
	if len(fence.Props) != 2 {
		t.Errorf("Expected 2 props, got %d", len(fence.Props))
	}

	if len(fence.Variables) != 1 {
		t.Errorf("Expected 1 variable, got %d", len(fence.Variables))
	}
}

// TestParseFenceContent_ExportLetSingleWord tests single-word variable names
func TestParseFenceContent_ExportLetSingleWord(t *testing.T) {
	input := `export let x, y, z`

	fence := parseFenceContent(input)

	if len(fence.ExportedProps) != 3 {
		t.Errorf("Expected 3 exported props, got %d", len(fence.ExportedProps))
		return
	}

	expectedProps := []string{"x", "y", "z"}
	for i, expected := range expectedProps {
		if fence.ExportedProps[i] != expected {
			t.Errorf("Expected exported prop[%d] '%s', got '%s'", i, expected, fence.ExportedProps[i])
		}
	}
}

// TestParseFenceContent_ExportLetCamelCase tests camelCase variable names
func TestParseFenceContent_ExportLetCamelCase(t *testing.T) {
	input := `export let firstName, lastName, emailAddress`

	fence := parseFenceContent(input)

	if len(fence.ExportedProps) != 3 {
		t.Errorf("Expected 3 exported props, got %d", len(fence.ExportedProps))
		return
	}

	expectedProps := []string{"firstName", "lastName", "emailAddress"}
	for i, expected := range expectedProps {
		if fence.ExportedProps[i] != expected {
			t.Errorf("Expected exported prop[%d] '%s', got '%s'", i, expected, fence.ExportedProps[i])
		}
	}
}
