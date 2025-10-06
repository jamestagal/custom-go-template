package transformer

import (
	"strings"
	"testing"

	"github.com/jimafisk/custom_go_template/ast"
)

func TestExtractGetters(t *testing.T) {
	// Test case: Extract getter methods from fence content
	fenceContent := `
prop user = { name: "John" }

function formatDate(dateString) {
  return new Date(dateString).toLocaleDateString();
}

get roleBadge() {
  return getRoleBadge(this.user.role);
}

get formattedJoinDate() {
  return formatDate(this.user.joinDate);
}
`

	fence := &ast.FenceSection{
		RawContent: fenceContent,
		Props: []ast.PropNode{
			{Name: "user", DefaultValue: `{ name: "John" }`},
		},
		Variables: []ast.VariableNode{},
	}

	scope := make(map[string]any)
	collectComponentFenceData(fence, scope)

	// Check that getters were extracted
	if _, exists := scope["roleBadge"]; !exists {
		t.Error("Expected 'roleBadge' getter to be extracted")
	}

	if _, exists := scope["formattedJoinDate"]; !exists {
		t.Error("Expected 'formattedJoinDate' getter to be extracted")
	}

	// Check that regular function was also extracted
	if _, exists := scope["formatDate"]; !exists {
		t.Error("Expected 'formatDate' function to be extracted")
	}

	// Verify the getter definition contains 'get' keyword
	roleBadge, ok := scope["roleBadge"].(string)
	if !ok {
		t.Fatalf("Expected roleBadge to be a string, got %T", scope["roleBadge"])
	}

	if !strings.HasPrefix(strings.TrimSpace(roleBadge), "get ") {
		t.Errorf("Expected getter to start with 'get ', got: %s", roleBadge)
	}

	if !strings.Contains(roleBadge, "return getRoleBadge(this.user.role)") {
		t.Errorf("Expected getter body to contain 'return getRoleBadge(this.user.role)', got: %s", roleBadge)
	}
}

func TestExtractSetters(t *testing.T) {
	// Test case: Extract setter methods from fence content
	fenceContent := `
set userName(value) {
  this.user.name = value;
}
`

	fence := &ast.FenceSection{
		RawContent: fenceContent,
		Props:      []ast.PropNode{},
		Variables:  []ast.VariableNode{},
	}

	scope := make(map[string]any)
	collectComponentFenceData(fence, scope)

	// Check that setter was extracted
	if _, exists := scope["userName"]; !exists {
		t.Error("Expected 'userName' setter to be extracted")
	}

	// Verify the setter definition contains 'set' keyword
	userName, ok := scope["userName"].(string)
	if !ok {
		t.Fatalf("Expected userName to be a string, got %T", scope["userName"])
	}

	if !strings.HasPrefix(strings.TrimSpace(userName), "set ") {
		t.Errorf("Expected setter to start with 'set ', got: %s", userName)
	}
}
