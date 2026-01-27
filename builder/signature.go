package builder

import (
	"path/filepath"
	"strings"
)

// GenerateSignature converts a layout file path to a Plenti-compatible signature.
//
// Plenti's signature format encodes the semantic hierarchy of layouts:
//   - components/ = dynamically loadable UI components
//   - content/ = page type templates (one per content type)
//   - global/ = site-wide wrappers (always loaded)
//   - scripts/ = utilities (stores, helpers)
//
// Examples:
//   layouts/components/Hero2436.html → layouts_components_Hero2436_html
//   layouts/content/blog.html → layouts_content_blog_html
//   layouts/global/nav.html → layouts_global_nav_html
//   layouts/content/_index.html → layouts_content__index_html
//
// Pattern: Pure Function [Load: 3]
func GenerateSignature(filePath string) string {
	// Normalize path separators (handle Windows paths)
	normalized := filepath.ToSlash(filePath)

	// Remove leading ./ if present
	normalized = strings.TrimPrefix(normalized, "./")

	// Replace / and . with _
	signature := strings.ReplaceAll(normalized, "/", "_")
	signature = strings.ReplaceAll(signature, ".", "_")

	return signature
}

// SignatureInfo represents the parsed components of a Plenti signature.
type SignatureInfo struct {
	Valid     bool
	Category  string // "components", "content", "global", "scripts"
	Name      string // "Hero2436", "blog", "nav", "_index"
	Extension string // "html", "svelte"
}

// ParseSignature extracts components from a Plenti-compatible signature.
//
// Examples:
//   layouts_components_Hero2436_html → {Category: "components", Name: "Hero2436", Extension: "html"}
//   layouts_content__index_html → {Category: "content", Name: "_index", Extension: "html"}
//   layouts_global_nav_html → {Category: "global", Name: "nav", Extension: "html"}
//
// Pattern: Parser Function [Load: 5]
func ParseSignature(signature string) SignatureInfo {
	parts := strings.Split(signature, "_")

	// Must have at least 4 parts: layouts, category, name, extension
	if len(parts) < 4 || parts[0] != "layouts" {
		return SignatureInfo{Valid: false}
	}

	category := parts[1]

	// Validate category
	if !isValidCategory(category) {
		return SignatureInfo{Valid: false}
	}

	// Extension is always the last part
	extension := parts[len(parts)-1]

	// Name is everything between category and extension
	// Handle names with underscores (e.g., "_index" becomes "__index" in signature)
	nameParts := parts[2 : len(parts)-1]
	name := strings.Join(nameParts, "_")

	return SignatureInfo{
		Valid:     true,
		Category:  category,
		Name:      name,
		Extension: extension,
	}
}

// isValidCategory checks if a category is one of the valid Plenti categories.
func isValidCategory(category string) bool {
	switch category {
	case "components", "content", "global", "scripts":
		return true
	default:
		return false
	}
}

// IsComponent returns true if the signature is a dynamically loadable component.
func (s SignatureInfo) IsComponent() bool {
	return s.Valid && s.Category == "components"
}

// IsGlobal returns true if the signature is a global layout (always loaded).
func (s SignatureInfo) IsGlobal() bool {
	return s.Valid && s.Category == "global"
}

// IsContentTemplate returns true if the signature is a content type template.
func (s SignatureInfo) IsContentTemplate() bool {
	return s.Valid && s.Category == "content"
}

// IsScript returns true if the signature is a utility script.
func (s SignatureInfo) IsScript() bool {
	return s.Valid && s.Category == "scripts"
}

// ToFilePath reconstructs the original file path from a signature.
//
// Example:
//   layouts_components_Hero2436_html → layouts/components/Hero2436.html
//
// Pattern: Inverse Function [Load: 4]
func (s SignatureInfo) ToFilePath() string {
	if !s.Valid {
		return ""
	}
	return "layouts/" + s.Category + "/" + s.Name + "." + s.Extension
}

// ShortNameToSignature converts a short component name to a full Plenti signature.
// This is the lookup pattern used by Plenti's allLayouts registry.
//
// Example:
//   ShortNameToSignature("Hero2436", "html") → "layouts_components_Hero2436_html"
//
// This mirrors Plenti's runtime lookup:
//   allLayouts["layouts_components_" + name + "_svelte"]
//
// Pattern: Builder Function [Load: 3]
func ShortNameToSignature(name string, extension string) string {
	return "layouts_components_" + name + "_" + extension
}

// SignatureToShortName extracts just the component name from a full signature.
//
// Example:
//   SignatureToShortName("layouts_components_Hero2436_html") → "Hero2436"
//
// Pattern: Extractor Function [Load: 3]
func SignatureToShortName(signature string) string {
	info := ParseSignature(signature)
	if !info.Valid {
		return ""
	}
	return info.Name
}
