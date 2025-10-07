package parser

import (
	"strings"
	"testing"

	"github.com/jimafisk/custom_go_template/ast"
)

// TestStyleParser_SingleStyleBlock tests parsing of a single <style> block
func TestStyleParser_SingleStyleBlock(t *testing.T) {
	template := `<style>
  .header { background-color: #f8f9fa; }
  .brand svg { height: 32px; }
</style>`

	result, err := ParseTemplate(template)
	if err != nil {
		t.Fatalf("ParseTemplate failed: %v", err)
	}

	if len(result.RootNodes) != 1 {
		t.Fatalf("Expected 1 root node, got %d", len(result.RootNodes))
	}

	styleSection, ok := result.RootNodes[0].(*ast.StyleSection)
	if !ok {
		t.Fatalf("Expected *ast.StyleSection, got %T", result.RootNodes[0])
	}

	expectedContent := `.header { background-color: #f8f9fa; }
  .brand svg { height: 32px; }`

	if strings.TrimSpace(styleSection.Content) != strings.TrimSpace(expectedContent) {
		t.Errorf("Style content mismatch.\nExpected:\n%s\nGot:\n%s",
			expectedContent, styleSection.Content)
	}
}

// TestStyleParser_MultipleStyleBlocks tests parsing multiple <style> blocks in one component
func TestStyleParser_MultipleStyleBlocks(t *testing.T) {
	template := `<style>
  .header { color: red; }
</style>

<header>Content</header>

<style>
  .footer { color: blue; }
</style>`

	result, err := ParseTemplate(template)
	if err != nil {
		t.Fatalf("ParseTemplate failed: %v", err)
	}

	// Count StyleSection nodes
	styleCount := 0
	for _, node := range result.RootNodes {
		if _, ok := node.(*ast.StyleSection); ok {
			styleCount++
		}
	}

	if styleCount != 2 {
		t.Errorf("Expected 2 StyleSection nodes, got %d", styleCount)
	}

	// Verify first style
	firstStyle, ok := result.RootNodes[0].(*ast.StyleSection)
	if !ok {
		t.Fatalf("First node should be *ast.StyleSection, got %T", result.RootNodes[0])
	}

	if !strings.Contains(firstStyle.Content, ".header") {
		t.Errorf("First style should contain .header, got: %s", firstStyle.Content)
	}

	// Verify second style
	var secondStyle *ast.StyleSection
	for _, node := range result.RootNodes {
		if style, ok := node.(*ast.StyleSection); ok {
			if strings.Contains(style.Content, ".footer") {
				secondStyle = style
				break
			}
		}
	}

	if secondStyle == nil {
		t.Errorf("Second StyleSection with .footer not found")
	}
}

// TestStyleParser_EmptyStyleBlock tests parsing empty <style> blocks
func TestStyleParser_EmptyStyleBlock(t *testing.T) {
	template := `<style></style>`

	result, err := ParseTemplate(template)
	if err != nil {
		t.Fatalf("ParseTemplate failed: %v", err)
	}

	if len(result.RootNodes) != 1 {
		t.Fatalf("Expected 1 root node, got %d", len(result.RootNodes))
	}

	styleSection, ok := result.RootNodes[0].(*ast.StyleSection)
	if !ok {
		t.Fatalf("Expected *ast.StyleSection, got %T", result.RootNodes[0])
	}

	if strings.TrimSpace(styleSection.Content) != "" {
		t.Errorf("Expected empty content, got: %q", styleSection.Content)
	}
}

// TestStyleParser_StyleWithWhitespace tests style block with only whitespace
func TestStyleParser_StyleWithWhitespace(t *testing.T) {
	template := `<style>


</style>`

	result, err := ParseTemplate(template)
	if err != nil {
		t.Fatalf("ParseTemplate failed: %v", err)
	}

	if len(result.RootNodes) != 1 {
		t.Fatalf("Expected 1 root node, got %d", len(result.RootNodes))
	}

	styleSection, ok := result.RootNodes[0].(*ast.StyleSection)
	if !ok {
		t.Fatalf("Expected *ast.StyleSection, got %T", result.RootNodes[0])
	}

	// Content may have whitespace, which is fine
	if styleSection.Content == "" {
		// This is acceptable - whitespace was trimmed
	} else if strings.TrimSpace(styleSection.Content) != "" {
		t.Errorf("Expected only whitespace content, got: %q", styleSection.Content)
	}
}

// TestStyleParser_CompleteComponent tests style parsing in a complete component
func TestStyleParser_CompleteComponent(t *testing.T) {
	template := `---
prop title = "Default"
---

<style>
  .component { padding: 1rem; }
</style>

<div class="component">{title}</div>`

	result, err := ParseTemplate(template)
	if err != nil {
		t.Fatalf("ParseTemplate failed: %v", err)
	}

	// Find the StyleSection node
	var styleSection *ast.StyleSection
	for _, node := range result.RootNodes {
		if style, ok := node.(*ast.StyleSection); ok {
			styleSection = style
			break
		}
	}

	if styleSection == nil {
		t.Fatalf("No StyleSection found in parsed template")
	}

	if !strings.Contains(styleSection.Content, ".component") {
		t.Errorf("Style content should contain .component, got: %s", styleSection.Content)
	}
}

// TestStyleParser_RealWorldHeaderSimple tests parsing the real HeaderSimple.html component
func TestStyleParser_RealWorldHeaderSimple(t *testing.T) {
	template := `---
---

<style>
  .header {
    background-color: #f8f9fa;
    padding: 1rem 0;
    border-bottom: 1px solid #e9ecef;
    margin-bottom: 2rem;
  }

  .header-container {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 0 2rem;
  }

  .brand {
    display: flex;
    align-items: center;
    text-decoration: none;
  }

  .brand svg {
    height: 32px;
    width: auto;
  }

  .nav {
    display: flex;
    gap: 1.5rem;
  }

  .nav-item {
    color: #495057;
    text-decoration: none;
  }

  .nav-item:hover {
    color: #228be6;
  }
</style>

<header class="header">
  <div class="header-container">
    <a href="/" class="brand">
      <svg>...</svg>
    </a>
    <nav class="nav">
      <a href="/" class="nav-item">Home</a>
    </nav>
  </div>
</header>`

	result, err := ParseTemplate(template)
	if err != nil {
		t.Fatalf("ParseTemplate failed: %v", err)
	}

	// Find the StyleSection node
	var styleSection *ast.StyleSection
	for _, node := range result.RootNodes {
		if style, ok := node.(*ast.StyleSection); ok {
			styleSection = style
			break
		}
	}

	if styleSection == nil {
		t.Fatalf("No StyleSection found in parsed template")
	}

	// Verify key CSS classes are present
	expectedClasses := []string{".header", ".header-container", ".brand", ".nav", ".nav-item"}
	for _, class := range expectedClasses {
		if !strings.Contains(styleSection.Content, class) {
			t.Errorf("Style content should contain %s", class)
		}
	}

	// Verify it's a substantial style block
	if len(styleSection.Content) < 100 {
		t.Errorf("Expected substantial style content, got only %d characters", len(styleSection.Content))
	}
}

// TestStyleParser_MissingClosingTag tests error handling for missing </style>
func TestStyleParser_MissingClosingTag(t *testing.T) {
	template := `<style>
  .header { color: red; }
`
	// Note: Missing </style> tag

	result, err := ParseTemplate(template)

	// The parser should either:
	// 1. Return an error, OR
	// 2. Successfully parse with no StyleSection (if it gracefully handles missing tag)

	if err != nil {
		// Error is acceptable for malformed input
		t.Logf("ParseTemplate returned error (acceptable): %v", err)
		return
	}

	// If no error, verify no StyleSection was parsed
	for _, node := range result.RootNodes {
		if _, ok := node.(*ast.StyleSection); ok {
			t.Errorf("Should not parse StyleSection without closing tag")
		}
	}
}

// TestStyleParser_StyleInRootNodes verifies StyleSection is added to Template.RootNodes
func TestStyleParser_StyleInRootNodes(t *testing.T) {
	template := `<style>
  .test { color: green; }
</style>

<div>Content</div>`

	result, err := ParseTemplate(template)
	if err != nil {
		t.Fatalf("ParseTemplate failed: %v", err)
	}

	// Verify we have at least 2 nodes (style + element)
	if len(result.RootNodes) < 2 {
		t.Fatalf("Expected at least 2 root nodes, got %d", len(result.RootNodes))
	}

	// Verify first node is StyleSection
	_, ok := result.RootNodes[0].(*ast.StyleSection)
	if !ok {
		t.Errorf("First root node should be *ast.StyleSection, got %T", result.RootNodes[0])
	}

	// Verify second node is Element
	_, ok = result.RootNodes[1].(*ast.Element)
	if !ok {
		t.Errorf("Second root node should be *ast.Element, got %T", result.RootNodes[1])
	}
}

// TestStyleParser_NodeType verifies StyleSection implements Node interface correctly
func TestStyleParser_NodeType(t *testing.T) {
	template := `<style>.test{}</style>`

	result, err := ParseTemplate(template)
	if err != nil {
		t.Fatalf("ParseTemplate failed: %v", err)
	}

	styleSection, ok := result.RootNodes[0].(*ast.StyleSection)
	if !ok {
		t.Fatalf("Expected *ast.StyleSection, got %T", result.RootNodes[0])
	}

	nodeType := styleSection.NodeType()
	if nodeType != "StyleSection" {
		t.Errorf("Expected NodeType() to return 'StyleSection', got %q", nodeType)
	}
}

// TestStyleParser_ComplexCSS tests parsing of complex CSS with various features
func TestStyleParser_ComplexCSS(t *testing.T) {
	template := `<style>
  /* Comment */
  @media (max-width: 768px) {
    .responsive { width: 100%; }
  }

  .complex-selector > .child:hover {
    background: linear-gradient(to bottom, #fff, #000);
    transform: translateX(10px);
  }

  @keyframes slide {
    from { left: 0; }
    to { left: 100%; }
  }
</style>`

	result, err := ParseTemplate(template)
	if err != nil {
		t.Fatalf("ParseTemplate failed: %v", err)
	}

	styleSection, ok := result.RootNodes[0].(*ast.StyleSection)
	if !ok {
		t.Fatalf("Expected *ast.StyleSection, got %T", result.RootNodes[0])
	}

	// Verify complex CSS features are preserved
	expectedFeatures := []string{"@media", "@keyframes", "linear-gradient", "transform", ":hover"}
	for _, feature := range expectedFeatures {
		if !strings.Contains(styleSection.Content, feature) {
			t.Errorf("Style content should contain %q", feature)
		}
	}
}

// TestStyleParser_WithAttributes tests parsing <style> tag with attributes
func TestStyleParser_WithAttributes(t *testing.T) {
	template := `<style scoped>
  .test { color: red; }
</style>`

	result, err := ParseTemplate(template)
	if err != nil {
		t.Fatalf("ParseTemplate failed: %v", err)
	}

	if len(result.RootNodes) != 1 {
		t.Fatalf("Expected 1 root node, got %d", len(result.RootNodes))
	}

	styleSection, ok := result.RootNodes[0].(*ast.StyleSection)
	if !ok {
		t.Fatalf("Expected *ast.StyleSection, got %T", result.RootNodes[0])
	}

	if !strings.Contains(styleSection.Content, ".test") {
		t.Errorf("Style content should contain .test, got: %s", styleSection.Content)
	}
}

// TestStyleParser_WithMultipleAttributes tests parsing <style> tag with multiple attributes
func TestStyleParser_WithMultipleAttributes(t *testing.T) {
	template := `<style type="text/css" scoped>
  .component { padding: 1rem; }
</style>`

	result, err := ParseTemplate(template)
	if err != nil {
		t.Fatalf("ParseTemplate failed: %v", err)
	}

	if len(result.RootNodes) != 1 {
		t.Fatalf("Expected 1 root node, got %d", len(result.RootNodes))
	}

	styleSection, ok := result.RootNodes[0].(*ast.StyleSection)
	if !ok {
		t.Fatalf("Expected *ast.StyleSection, got %T", result.RootNodes[0])
	}

	if !strings.Contains(styleSection.Content, ".component") {
		t.Errorf("Style content should contain .component, got: %s", styleSection.Content)
	}
}
