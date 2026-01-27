// Package types provides shared types used across the template engine.
// This package serves as the canonical source for component-related types
// to ensure consistency between the transformer, builder, and server packages.
package types

import (
	"path/filepath"
	"strings"

	"github.com/jimafisk/custom_go_template/ast"
)

// ComponentTemplate represents a unified component with metadata and parsed content.
// This struct is the canonical representation used by:
//   - transformer: Component registration and lookup
//   - builder: Registry generation for runtime components
//   - server: Component discovery and registration
//
// Pattern: Canonical Type [Load: 5]
type ComponentTemplate struct {
	// Name is the short name of the component (e.g., "Hero2436", "nav", "_index")
	// This is the primary identifier used in templates: <Hero2436 />
	Name string

	// Signature is the Plenti-compatible signature for registry lookup
	// Format: layouts_{category}_{name}_{extension}
	// Example: "layouts_components_Hero2436_html"
	Signature string

	// FilePath is the original file path relative to the project root
	// Example: "layouts/components/Hero2436.html"
	FilePath string

	// Category indicates the component type for organizational purposes
	// Values: "components", "content", "global", "scripts"
	Category string

	// AST is the parsed and optionally transformed template
	// This may be nil if the component hasn't been parsed yet
	AST *ast.Template

	// Props contains the names of props declared in the component's fence section
	// These are extracted from prop declarations and export let statements
	Props []string
}

// NewComponentTemplate creates a new ComponentTemplate with auto-generated metadata.
// It extracts the name, signature, and category from the file path.
//
// Parameters:
//   - filePath: Relative path to the component file (e.g., "layouts/components/Hero2436.html")
//   - template: Parsed AST of the component (can be nil for deferred parsing)
//   - props: List of prop names declared in the component
//
// Pattern: Constructor Function [Load: 5]
func NewComponentTemplate(filePath string, template *ast.Template, props []string) *ComponentTemplate {
	// Normalize path for consistent handling
	normalized := normalizeFilePath(filePath)

	return &ComponentTemplate{
		Name:      ExtractNameFromPath(normalized),
		Signature: GenerateSignature(normalized),
		FilePath:  normalized,
		Category:  CategoryFromPath(normalized),
		AST:       template,
		Props:     props,
	}
}

// normalizeFilePath normalizes a file path for consistent handling.
// It converts to forward slashes and removes leading ./
//
// Pattern: Pure Function [Load: 2]
func normalizeFilePath(filePath string) string {
	// Convert to forward slashes (handle Windows paths)
	normalized := filepath.ToSlash(filePath)

	// Remove leading ./ if present
	normalized = strings.TrimPrefix(normalized, "./")

	return normalized
}

// ExtractNameFromPath extracts the component name from a file path.
// The name is the filename without extension.
//
// Examples:
//   - "layouts/components/Hero2436.html" -> "Hero2436"
//   - "layouts/content/_index.html" -> "_index"
//   - "layouts/global/nav.html" -> "nav"
//   - "layouts/components/forms/Input.html" -> "Input"
//
// Pattern: Extractor Function [Load: 4]
func ExtractNameFromPath(filePath string) string {
	// Normalize path
	normalized := normalizeFilePath(filePath)

	// Get the base filename
	base := filepath.Base(normalized)

	// Remove the extension
	ext := filepath.Ext(base)
	name := strings.TrimSuffix(base, ext)

	return name
}

// CategoryFromPath extracts the category from a file path.
// The category is determined by the directory structure.
//
// Examples:
//   - "layouts/components/Hero2436.html" -> "components"
//   - "layouts/content/pages.html" -> "content"
//   - "layouts/global/nav.html" -> "global"
//   - "layouts/scripts/helpers.js" -> "scripts"
//   - "layouts/components/forms/Input.html" -> "components"
//
// Pattern: Classifier Function [Load: 5]
func CategoryFromPath(filePath string) string {
	// Normalize path
	normalized := normalizeFilePath(filePath)

	// Split path into parts
	parts := strings.Split(normalized, "/")

	// Expected structure: layouts/{category}/...
	// Find the category (second element after "layouts")
	for i, part := range parts {
		if part == "layouts" && i+1 < len(parts) {
			category := parts[i+1]
			// Validate category
			if isValidCategory(category) {
				return category
			}
		}
	}

	// Default to "components" if category cannot be determined
	return "components"
}

// GenerateSignature converts a layout file path to a Plenti-compatible signature.
//
// Plenti's signature format encodes the semantic hierarchy of layouts:
//   - components/ = dynamically loadable UI components
//   - content/ = page type templates (one per content type)
//   - global/ = site-wide wrappers (always loaded)
//   - scripts/ = utilities (stores, helpers)
//
// Examples:
//   - layouts/components/Hero2436.html -> layouts_components_Hero2436_html
//   - layouts/content/blog.html -> layouts_content_blog_html
//   - layouts/global/nav.html -> layouts_global_nav_html
//   - layouts/content/_index.html -> layouts_content__index_html
//   - ./layouts/components/Hero2436.html -> layouts_components_Hero2436_html
//
// Pattern: Pure Function [Load: 3]
func GenerateSignature(filePath string) string {
	// Normalize path
	normalized := normalizeFilePath(filePath)

	// Replace / and . with _
	signature := strings.ReplaceAll(normalized, "/", "_")
	signature = strings.ReplaceAll(signature, ".", "_")

	return signature
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

// SignatureInfo represents the parsed components of a Plenti signature.
// This is used for reverse-engineering paths from signatures.
type SignatureInfo struct {
	// Valid indicates whether the signature was successfully parsed
	Valid bool

	// Category is the component category: "components", "content", "global", "scripts"
	Category string

	// Name is the component name (may contain underscores)
	Name string

	// Extension is the file extension without the dot: "html", "svelte", "js"
	Extension string
}

// ParseSignature extracts components from a Plenti-compatible signature.
//
// Examples:
//   - layouts_components_Hero2436_html -> {Category: "components", Name: "Hero2436", Extension: "html"}
//   - layouts_content__index_html -> {Category: "content", Name: "_index", Extension: "html"}
//   - layouts_global_nav_html -> {Category: "global", Name: "nav", Extension: "html"}
//   - layouts_components_Hero_2436_html -> {Category: "components", Name: "Hero_2436", Extension: "html"}
//   - layouts_components_forms_Input_html -> {Category: "components", Name: "forms_Input", Extension: "html"}
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
// Note: For nested paths (e.g., forms/Input), this returns a flattened path
// since the original directory structure cannot be fully reconstructed.
//
// Example:
//   - layouts_components_Hero2436_html -> layouts/components/Hero2436.html
//   - layouts_components_forms_Input_html -> layouts/components/forms_Input.html
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
//   - ShortNameToSignature("Hero2436", "html") -> "layouts_components_Hero2436_html"
//
// This mirrors Plenti's runtime lookup:
//   - allLayouts["layouts_components_" + name + "_svelte"]
//
// Pattern: Builder Function [Load: 3]
func ShortNameToSignature(name string, extension string) string {
	return "layouts_components_" + name + "_" + extension
}

// SignatureToShortName extracts just the component name from a full signature.
//
// Example:
//   - SignatureToShortName("layouts_components_Hero2436_html") -> "Hero2436"
//   - SignatureToShortName("layouts_components_forms_Input_html") -> "forms_Input"
//
// Pattern: Extractor Function [Load: 3]
func SignatureToShortName(signature string) string {
	info := ParseSignature(signature)
	if !info.Valid {
		return ""
	}
	return info.Name
}
