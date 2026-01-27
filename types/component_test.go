package types

import (
	"testing"
)

// TestGenerateSignature tests the signature generation with the edge case matrix from the spec.
func TestGenerateSignature(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		notes    string
	}{
		{
			name:     "standard component",
			input:    "layouts/components/Hero2436.html",
			expected: "layouts_components_Hero2436_html",
			notes:    "Standard component path",
		},
		{
			name:     "leading underscore in name",
			input:    "layouts/content/_index.html",
			expected: "layouts_content__index_html",
			notes:    "Leading underscore becomes double underscore",
		},
		{
			name:     "underscore in name",
			input:    "layouts/components/Hero_2436.html",
			expected: "layouts_components_Hero_2436_html",
			notes:    "Underscore in name preserved",
		},
		{
			name:     "nested path",
			input:    "layouts/components/forms/Input.html",
			expected: "layouts_components_forms_Input_html",
			notes:    "Nested directory becomes underscore",
		},
		{
			name:     "leading dot-slash",
			input:    "./layouts/global/nav.html",
			expected: "layouts_global_nav_html",
			notes:    "Leading ./ is stripped",
		},
		{
			name:     "content template",
			input:    "layouts/content/blog.html",
			expected: "layouts_content_blog_html",
			notes:    "Content template path",
		},
		{
			name:     "global layout",
			input:    "layouts/global/header.html",
			expected: "layouts_global_header_html",
			notes:    "Global layout path",
		},
		{
			name:     "scripts layout",
			input:    "layouts/scripts/helpers.js",
			expected: "layouts_scripts_helpers_js",
			notes:    "Scripts layout path with .js extension",
		},
		{
			name:     "svelte extension",
			input:    "layouts/components/Card.svelte",
			expected: "layouts_components_Card_svelte",
			notes:    "Svelte file extension",
		},
		{
			name:     "deeply nested path",
			input:    "layouts/components/ui/forms/inputs/TextInput.html",
			expected: "layouts_components_ui_forms_inputs_TextInput_html",
			notes:    "Deeply nested directory structure",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GenerateSignature(tt.input)
			if got != tt.expected {
				t.Errorf("GenerateSignature(%q) = %q, want %q (%s)",
					tt.input, got, tt.expected, tt.notes)
			}
		})
	}
}

// TestParseSignature tests parsing signatures back to their components.
func TestParseSignature(t *testing.T) {
	tests := []struct {
		name      string
		signature string
		expected  SignatureInfo
	}{
		{
			name:      "standard component",
			signature: "layouts_components_Hero2436_html",
			expected: SignatureInfo{
				Valid:     true,
				Category:  "components",
				Name:      "Hero2436",
				Extension: "html",
			},
		},
		{
			name:      "leading underscore in name",
			signature: "layouts_content__index_html",
			expected: SignatureInfo{
				Valid:     true,
				Category:  "content",
				Name:      "_index",
				Extension: "html",
			},
		},
		{
			name:      "underscore in name",
			signature: "layouts_components_Hero_2436_html",
			expected: SignatureInfo{
				Valid:     true,
				Category:  "components",
				Name:      "Hero_2436",
				Extension: "html",
			},
		},
		{
			name:      "nested path (forms_Input)",
			signature: "layouts_components_forms_Input_html",
			expected: SignatureInfo{
				Valid:     true,
				Category:  "components",
				Name:      "forms_Input",
				Extension: "html",
			},
		},
		{
			name:      "global layout",
			signature: "layouts_global_nav_html",
			expected: SignatureInfo{
				Valid:     true,
				Category:  "global",
				Name:      "nav",
				Extension: "html",
			},
		},
		{
			name:      "scripts layout",
			signature: "layouts_scripts_helpers_js",
			expected: SignatureInfo{
				Valid:     true,
				Category:  "scripts",
				Name:      "helpers",
				Extension: "js",
			},
		},
		{
			name:      "invalid - missing layouts prefix",
			signature: "components_Hero2436_html",
			expected: SignatureInfo{
				Valid: false,
			},
		},
		{
			name:      "invalid - unknown category",
			signature: "layouts_unknown_Hero2436_html",
			expected: SignatureInfo{
				Valid: false,
			},
		},
		{
			name:      "invalid - too few parts",
			signature: "layouts_components_html",
			expected: SignatureInfo{
				Valid: false,
			},
		},
		{
			name:      "invalid - empty string",
			signature: "",
			expected: SignatureInfo{
				Valid: false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseSignature(tt.signature)
			if got.Valid != tt.expected.Valid {
				t.Errorf("ParseSignature(%q).Valid = %v, want %v",
					tt.signature, got.Valid, tt.expected.Valid)
			}
			if got.Valid {
				if got.Category != tt.expected.Category {
					t.Errorf("ParseSignature(%q).Category = %q, want %q",
						tt.signature, got.Category, tt.expected.Category)
				}
				if got.Name != tt.expected.Name {
					t.Errorf("ParseSignature(%q).Name = %q, want %q",
						tt.signature, got.Name, tt.expected.Name)
				}
				if got.Extension != tt.expected.Extension {
					t.Errorf("ParseSignature(%q).Extension = %q, want %q",
						tt.signature, got.Extension, tt.expected.Extension)
				}
			}
		})
	}
}

// TestRoundTrip tests generate → parse → ToFilePath round-trip consistency.
func TestRoundTrip(t *testing.T) {
	tests := []struct {
		name         string
		inputPath    string
		expectedPath string // May differ from input for nested paths
		notes        string
	}{
		{
			name:         "standard component",
			inputPath:    "layouts/components/Hero2436.html",
			expectedPath: "layouts/components/Hero2436.html",
			notes:        "Round-trip preserves standard paths",
		},
		{
			name:         "leading underscore",
			inputPath:    "layouts/content/_index.html",
			expectedPath: "layouts/content/_index.html",
			notes:        "Round-trip preserves leading underscore",
		},
		{
			name:         "underscore in name",
			inputPath:    "layouts/components/Hero_2436.html",
			expectedPath: "layouts/components/Hero_2436.html",
			notes:        "Round-trip preserves underscore in name",
		},
		{
			name:         "nested path becomes flattened",
			inputPath:    "layouts/components/forms/Input.html",
			expectedPath: "layouts/components/forms_Input.html",
			notes:        "Nested paths are flattened (underscore replaces slash)",
		},
		{
			name:         "global layout",
			inputPath:    "layouts/global/nav.html",
			expectedPath: "layouts/global/nav.html",
			notes:        "Round-trip preserves global paths",
		},
		{
			name:         "with leading dot-slash",
			inputPath:    "./layouts/components/Card.html",
			expectedPath: "layouts/components/Card.html",
			notes:        "Leading ./ is normalized away",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Generate signature from path
			signature := GenerateSignature(tt.inputPath)

			// Parse the signature
			info := ParseSignature(signature)
			if !info.Valid {
				t.Fatalf("ParseSignature(%q) returned invalid for input %q",
					signature, tt.inputPath)
			}

			// Reconstruct path
			reconstructed := info.ToFilePath()
			if reconstructed != tt.expectedPath {
				t.Errorf("Round-trip: %q → %q → %q, want %q (%s)",
					tt.inputPath, signature, reconstructed, tt.expectedPath, tt.notes)
			}
		})
	}
}

// TestExtractNameFromPath tests component name extraction from file paths.
func TestExtractNameFromPath(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "standard component",
			input:    "layouts/components/Hero2436.html",
			expected: "Hero2436",
		},
		{
			name:     "leading underscore",
			input:    "layouts/content/_index.html",
			expected: "_index",
		},
		{
			name:     "underscore in name",
			input:    "layouts/components/Hero_2436.html",
			expected: "Hero_2436",
		},
		{
			name:     "nested path",
			input:    "layouts/components/forms/Input.html",
			expected: "Input",
		},
		{
			name:     "with leading dot-slash",
			input:    "./layouts/global/nav.html",
			expected: "nav",
		},
		{
			name:     "svelte extension",
			input:    "layouts/components/Card.svelte",
			expected: "Card",
		},
		{
			name:     "js extension",
			input:    "layouts/scripts/helpers.js",
			expected: "helpers",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractNameFromPath(tt.input)
			if got != tt.expected {
				t.Errorf("ExtractNameFromPath(%q) = %q, want %q",
					tt.input, got, tt.expected)
			}
		})
	}
}

// TestCategoryFromPath tests category extraction from file paths.
func TestCategoryFromPath(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "components category",
			input:    "layouts/components/Hero2436.html",
			expected: "components",
		},
		{
			name:     "content category",
			input:    "layouts/content/blog.html",
			expected: "content",
		},
		{
			name:     "global category",
			input:    "layouts/global/nav.html",
			expected: "global",
		},
		{
			name:     "scripts category",
			input:    "layouts/scripts/helpers.js",
			expected: "scripts",
		},
		{
			name:     "nested components path",
			input:    "layouts/components/forms/Input.html",
			expected: "components",
		},
		{
			name:     "with leading dot-slash",
			input:    "./layouts/content/_index.html",
			expected: "content",
		},
		{
			name:     "invalid path - defaults to components",
			input:    "some/other/path.html",
			expected: "components",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CategoryFromPath(tt.input)
			if got != tt.expected {
				t.Errorf("CategoryFromPath(%q) = %q, want %q",
					tt.input, got, tt.expected)
			}
		})
	}
}

// TestShortNameToSignature tests short name to full signature conversion.
func TestShortNameToSignature(t *testing.T) {
	tests := []struct {
		name      string
		shortName string
		extension string
		expected  string
	}{
		{
			name:      "html extension",
			shortName: "Hero2436",
			extension: "html",
			expected:  "layouts_components_Hero2436_html",
		},
		{
			name:      "svelte extension",
			shortName: "Card",
			extension: "svelte",
			expected:  "layouts_components_Card_svelte",
		},
		{
			name:      "name with underscore",
			shortName: "Hero_2436",
			extension: "html",
			expected:  "layouts_components_Hero_2436_html",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ShortNameToSignature(tt.shortName, tt.extension)
			if got != tt.expected {
				t.Errorf("ShortNameToSignature(%q, %q) = %q, want %q",
					tt.shortName, tt.extension, got, tt.expected)
			}
		})
	}
}

// TestSignatureToShortName tests extracting short name from full signature.
func TestSignatureToShortName(t *testing.T) {
	tests := []struct {
		name      string
		signature string
		expected  string
	}{
		{
			name:      "standard signature",
			signature: "layouts_components_Hero2436_html",
			expected:  "Hero2436",
		},
		{
			name:      "name with underscore",
			signature: "layouts_components_Hero_2436_html",
			expected:  "Hero_2436",
		},
		{
			name:      "nested path signature",
			signature: "layouts_components_forms_Input_html",
			expected:  "forms_Input",
		},
		{
			name:      "invalid signature",
			signature: "invalid_signature",
			expected:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SignatureToShortName(tt.signature)
			if got != tt.expected {
				t.Errorf("SignatureToShortName(%q) = %q, want %q",
					tt.signature, got, tt.expected)
			}
		})
	}
}

// TestSignatureInfoMethods tests the category checking methods on SignatureInfo.
func TestSignatureInfoMethods(t *testing.T) {
	tests := []struct {
		name             string
		signature        string
		isComponent      bool
		isGlobal         bool
		isContentTemplate bool
		isScript         bool
	}{
		{
			name:             "component",
			signature:        "layouts_components_Hero2436_html",
			isComponent:      true,
			isGlobal:         false,
			isContentTemplate: false,
			isScript:         false,
		},
		{
			name:             "global",
			signature:        "layouts_global_nav_html",
			isComponent:      false,
			isGlobal:         true,
			isContentTemplate: false,
			isScript:         false,
		},
		{
			name:             "content",
			signature:        "layouts_content_blog_html",
			isComponent:      false,
			isGlobal:         false,
			isContentTemplate: true,
			isScript:         false,
		},
		{
			name:             "scripts",
			signature:        "layouts_scripts_helpers_js",
			isComponent:      false,
			isGlobal:         false,
			isContentTemplate: false,
			isScript:         true,
		},
		{
			name:             "invalid",
			signature:        "invalid",
			isComponent:      false,
			isGlobal:         false,
			isContentTemplate: false,
			isScript:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := ParseSignature(tt.signature)

			if got := info.IsComponent(); got != tt.isComponent {
				t.Errorf("IsComponent() = %v, want %v", got, tt.isComponent)
			}
			if got := info.IsGlobal(); got != tt.isGlobal {
				t.Errorf("IsGlobal() = %v, want %v", got, tt.isGlobal)
			}
			if got := info.IsContentTemplate(); got != tt.isContentTemplate {
				t.Errorf("IsContentTemplate() = %v, want %v", got, tt.isContentTemplate)
			}
			if got := info.IsScript(); got != tt.isScript {
				t.Errorf("IsScript() = %v, want %v", got, tt.isScript)
			}
		})
	}
}

// TestNewComponentTemplate tests the constructor function.
func TestNewComponentTemplate(t *testing.T) {
	tests := []struct {
		name              string
		filePath          string
		props             []string
		expectedName      string
		expectedSignature string
		expectedCategory  string
	}{
		{
			name:              "standard component",
			filePath:          "layouts/components/Hero2436.html",
			props:             []string{"title", "description"},
			expectedName:      "Hero2436",
			expectedSignature: "layouts_components_Hero2436_html",
			expectedCategory:  "components",
		},
		{
			name:              "content template",
			filePath:          "layouts/content/_index.html",
			props:             []string{},
			expectedName:      "_index",
			expectedSignature: "layouts_content__index_html",
			expectedCategory:  "content",
		},
		{
			name:              "nested path",
			filePath:          "layouts/components/forms/Input.html",
			props:             []string{"value", "placeholder"},
			expectedName:      "Input",
			expectedSignature: "layouts_components_forms_Input_html",
			expectedCategory:  "components",
		},
		{
			name:              "with leading dot-slash",
			filePath:          "./layouts/global/nav.html",
			props:             nil,
			expectedName:      "nav",
			expectedSignature: "layouts_global_nav_html",
			expectedCategory:  "global",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ct := NewComponentTemplate(tt.filePath, nil, tt.props)

			if ct.Name != tt.expectedName {
				t.Errorf("Name = %q, want %q", ct.Name, tt.expectedName)
			}
			if ct.Signature != tt.expectedSignature {
				t.Errorf("Signature = %q, want %q", ct.Signature, tt.expectedSignature)
			}
			if ct.Category != tt.expectedCategory {
				t.Errorf("Category = %q, want %q", ct.Category, tt.expectedCategory)
			}
			if len(ct.Props) != len(tt.props) {
				t.Errorf("Props length = %d, want %d", len(ct.Props), len(tt.props))
			}
		})
	}
}
