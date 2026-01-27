package transformer

import (
	"testing"

	"github.com/jimafisk/custom_go_template/ast"
)

// TestGetterPropertyInClassAttribute tests class="profile-role {roleBadge.bg} {roleBadge.text}"
// where roleBadge could be either a getter definition (string) or an actual map
func TestGetterPropertyInClassAttribute(t *testing.T) {
	attrs := []ast.Attribute{
		{
			Name:  "class",
			Value: "profile-role {roleBadge.bg} {roleBadge.text}",
		},
	}

	t.Run("roleBadge_is_getter_definition", func(t *testing.T) {
		// When roleBadge is a getter definition (string), property access should fail
		// and result in a runtime binding
		dataScope := map[string]any{
			"roleBadge": "get roleBadge() { return this.getRoleBadge(this.user.role); }",
			"user":      map[string]any{"role": "admin"},
		}
		transformed := transformAttributeExpressions(attrs, dataScope)

		t.Logf("Number of attributes: %d", len(transformed))
		for _, attr := range transformed {
			t.Logf("Attr: Name=%q, Value=%q, Dynamic=%v", attr.Name, attr.Value, attr.Dynamic)
		}

		if len(transformed) != 1 {
			t.Errorf("Expected 1 attribute, got %d", len(transformed))
		}

		// Should be a dynamic binding since getter can't be resolved at build-time
		if !transformed[0].Dynamic {
			t.Errorf("Expected Dynamic=true for getter property access")
		}
	})

	t.Run("roleBadge_is_actual_map", func(t *testing.T) {
		// When roleBadge is an actual map, property access should succeed
		// and result in a fully interpolated static class (or partial)
		dataScope := map[string]any{
			"roleBadge": map[string]any{
				"bg":   "bg-red-100",
				"text": "text-red-800",
			},
			"user": map[string]any{"role": "admin"},
		}
		transformed := transformAttributeExpressions(attrs, dataScope)

		t.Logf("Number of attributes: %d", len(transformed))
		for _, attr := range transformed {
			t.Logf("Attr: Name=%q, Value=%q, Dynamic=%v", attr.Name, attr.Value, attr.Dynamic)
		}

		if len(transformed) != 1 {
			t.Errorf("Expected 1 attribute, got %d", len(transformed))
		}
	})
}
