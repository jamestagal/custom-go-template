// Package builder provides component registry and bundle generation utilities.
package builder

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// PageUsage represents the components used by a single page.
//
// Pattern: Data Transfer Object [Load: 4]
type PageUsage struct {
	// Page is the page identifier (e.g., "pages/_index", "pages/about")
	Page string

	// Components is the list of component names used on this page
	Components []string

	// Source maps each component to how it was discovered
	// Values: "content" (from JSON), "template" (from HTML), "global" (always loaded)
	Source map[string]string
}

// ComponentUsage represents aggregated usage information for a single component.
//
// Pattern: Data Transfer Object [Load: 4]
type ComponentUsage struct {
	// Name is the component name
	Name string

	// Pages lists which pages use this component
	Pages []string

	// Size is the component template size in bytes (populated during bundle generation)
	Size int

	// Category is the component category (components, content, global)
	Category string
}

// UsageAnalyzer scans content and templates to determine component usage per page.
//
// Pattern: Service Object [Load: 6]
type UsageAnalyzer struct {
	contentDir  string
	layoutsDir  string
	pageUsage   map[string]*PageUsage
	compUsage   map[string]*ComponentUsage
}

// NewUsageAnalyzer creates a new UsageAnalyzer.
//
// Pattern: Constructor [Load: 2]
func NewUsageAnalyzer(contentDir, layoutsDir string) *UsageAnalyzer {
	return &UsageAnalyzer{
		contentDir:  contentDir,
		layoutsDir:  layoutsDir,
		pageUsage:   make(map[string]*PageUsage),
		compUsage:   make(map[string]*ComponentUsage),
	}
}

// Analyze scans content and templates to build the usage map.
// Returns a map of page path to PageUsage.
//
// Pattern: Template Method [Load: 8]
func (a *UsageAnalyzer) Analyze() (map[string]*PageUsage, error) {
	// 1. Scan content JSON files for component references
	if err := a.scanContentJSON(); err != nil {
		return nil, fmt.Errorf("UsageAnalyzer.Analyze: scanning content: %w", err)
	}

	// 2. Scan template files for static component usage
	if err := a.scanTemplates(); err != nil {
		return nil, fmt.Errorf("UsageAnalyzer.Analyze: scanning templates: %w", err)
	}

	// 3. Add global components to every page
	if err := a.addGlobalComponents(); err != nil {
		return nil, fmt.Errorf("UsageAnalyzer.Analyze: adding global components: %w", err)
	}

	return a.pageUsage, nil
}

// GetComponentUsage returns aggregated usage information for all components.
// Must be called after Analyze().
//
// Pattern: Query Method [Load: 5]
func (a *UsageAnalyzer) GetComponentUsage() map[string]*ComponentUsage {
	// Build component usage from page usage
	a.compUsage = make(map[string]*ComponentUsage)

	for pagePath, pu := range a.pageUsage {
		for _, compName := range pu.Components {
			if _, exists := a.compUsage[compName]; !exists {
				a.compUsage[compName] = &ComponentUsage{
					Name:  compName,
					Pages: []string{},
				}
			}
			a.compUsage[compName].Pages = append(a.compUsage[compName].Pages, pagePath)
		}
	}

	return a.compUsage
}

// scanContentJSON scans content/*.json files for component references.
//
// Pattern: Iterator [Load: 7]
func (a *UsageAnalyzer) scanContentJSON() error {
	// Find all JSON files in content directory
	pattern := filepath.Join(a.contentDir, "**", "*.json")
	files, err := filepath.Glob(pattern)
	if err != nil {
		// Try simpler pattern if ** not supported
		files, err = a.findJSONFiles(a.contentDir)
		if err != nil {
			return fmt.Errorf("scanContentJSON: finding JSON files: %w", err)
		}
	}

	for _, file := range files {
		// Skip _defaults.json
		if strings.HasSuffix(file, "_defaults.json") {
			continue
		}

		pagePath := a.extractPagePath(file)
		components, err := a.extractComponentsFromJSON(file)
		if err != nil {
			// Log but continue - don't fail on individual file errors
			fmt.Printf("Warning: failed to parse %s: %v\n", file, err)
			continue
		}

		if len(components) > 0 {
			a.pageUsage[pagePath] = &PageUsage{
				Page:       pagePath,
				Components: components,
				Source:     make(map[string]string),
			}
			for _, comp := range components {
				a.pageUsage[pagePath].Source[comp] = "content"
			}
		}
	}

	return nil
}

// findJSONFiles recursively finds all JSON files in a directory.
//
// Pattern: Helper Function [Load: 4]
func (a *UsageAnalyzer) findJSONFiles(dir string) ([]string, error) {
	var files []string

	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(path, ".json") {
			files = append(files, path)
		}
		return nil
	})

	return files, err
}

// extractPagePath converts a content file path to a page identifier.
// Example: "content/pages/_index.json" -> "pages/_index"
//
// Pattern: Pure Function [Load: 4]
func (a *UsageAnalyzer) extractPagePath(filePath string) string {
	// Remove content directory prefix
	rel, err := filepath.Rel(a.contentDir, filePath)
	if err != nil {
		rel = filePath
	}

	// Remove .json extension
	rel = strings.TrimSuffix(rel, ".json")

	// Normalize path separators
	return filepath.ToSlash(rel)
}

// extractComponentsFromJSON parses a JSON file and extracts component names.
// Looks for the "components" array with "name" fields.
//
// Pattern: Parser [Load: 6]
func (a *UsageAnalyzer) extractComponentsFromJSON(filePath string) ([]string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("extractComponentsFromJSON: reading file: %w", err)
	}

	var content map[string]any
	if err := json.Unmarshal(data, &content); err != nil {
		return nil, fmt.Errorf("extractComponentsFromJSON: parsing JSON: %w", err)
	}

	// Look for "components" array
	componentsRaw, ok := content["components"]
	if !ok {
		return nil, nil // No components array - not an error
	}

	components, ok := componentsRaw.([]any)
	if !ok {
		return nil, nil // Not an array - not an error
	}

	var names []string
	for _, comp := range components {
		compMap, ok := comp.(map[string]any)
		if !ok {
			continue
		}

		name, ok := compMap["name"].(string)
		if !ok {
			continue
		}

		names = append(names, name)
	}

	return names, nil
}

// scanTemplates scans template files for static component references.
// Looks for <ComponentName /> patterns in HTML templates.
//
// Pattern: Scanner [Load: 6]
func (a *UsageAnalyzer) scanTemplates() error {
	// Component reference pattern: <ComponentName or <Component:dynamic
	componentPattern := regexp.MustCompile(`<([A-Z][a-zA-Z0-9_]*|Component:dynamic)[\s/>]`)

	// Scan content templates for each page
	for pagePath, pu := range a.pageUsage {
		// Determine template path based on page type
		templatePath := a.getTemplatePath(pagePath)
		if templatePath == "" {
			continue
		}

		data, err := os.ReadFile(templatePath)
		if err != nil {
			continue // Template doesn't exist - skip
		}

		matches := componentPattern.FindAllStringSubmatch(string(data), -1)
		for _, match := range matches {
			if len(match) < 2 {
				continue
			}
			compName := match[1]

			// Skip Component:dynamic - these are runtime resolved
			if compName == "Component:dynamic" {
				continue
			}

			// Add if not already present
			if _, exists := pu.Source[compName]; !exists {
				pu.Components = append(pu.Components, compName)
				pu.Source[compName] = "template"
			}
		}
	}

	return nil
}

// getTemplatePath returns the template file path for a given page path.
//
// Pattern: Lookup [Load: 3]
func (a *UsageAnalyzer) getTemplatePath(pagePath string) string {
	// For pages/*, use layouts/content/pages.html
	if strings.HasPrefix(pagePath, "pages/") {
		return filepath.Join(a.layoutsDir, "content", "pages.html")
	}

	// For news/*, use layouts/content/news.html
	if strings.HasPrefix(pagePath, "news/") {
		return filepath.Join(a.layoutsDir, "content", "news.html")
	}

	// Default: use page-specific template
	return filepath.Join(a.layoutsDir, "content", pagePath+".html")
}

// addGlobalComponents adds components from layouts/global/* to every page.
//
// Pattern: Enricher [Load: 5]
func (a *UsageAnalyzer) addGlobalComponents() error {
	globalDir := filepath.Join(a.layoutsDir, "global")

	// Check if global directory exists
	if _, err := os.Stat(globalDir); os.IsNotExist(err) {
		return nil // No global directory - not an error
	}

	// Find all HTML files in global directory
	files, err := filepath.Glob(filepath.Join(globalDir, "*.html"))
	if err != nil {
		return fmt.Errorf("addGlobalComponents: listing global files: %w", err)
	}

	// Extract component names from filenames
	var globalComponents []string
	for _, file := range files {
		name := strings.TrimSuffix(filepath.Base(file), ".html")
		globalComponents = append(globalComponents, name)
	}

	// Add global components to all pages
	for _, pu := range a.pageUsage {
		for _, comp := range globalComponents {
			if _, exists := pu.Source[comp]; !exists {
				pu.Components = append(pu.Components, comp)
				pu.Source[comp] = "global"
			}
		}
	}

	return nil
}

// GetCommonComponents returns components used on 2+ pages (threshold-based extraction).
//
// Pattern: Filter [Load: 4]
func (a *UsageAnalyzer) GetCommonComponents(threshold int) []string {
	compUsage := a.GetComponentUsage()

	var common []string
	for name, usage := range compUsage {
		if len(usage.Pages) >= threshold {
			common = append(common, name)
		}
	}

	// Sort for deterministic output
	sort.Strings(common)
	return common
}

// GetPageComponents returns components unique to a specific page.
// Excludes components that are in the common chunk.
//
// Pattern: Filter [Load: 4]
func (a *UsageAnalyzer) GetPageComponents(pagePath string, commonComponents []string) []string {
	pu, exists := a.pageUsage[pagePath]
	if !exists {
		return nil
	}

	// Create lookup set for common components
	commonSet := make(map[string]bool)
	for _, c := range commonComponents {
		commonSet[c] = true
	}

	// Filter out common components
	var unique []string
	for _, comp := range pu.Components {
		if !commonSet[comp] {
			unique = append(unique, comp)
		}
	}

	// Sort for deterministic output
	sort.Strings(unique)
	return unique
}

// GetAllPages returns all page paths that were analyzed.
//
// Pattern: Accessor [Load: 2]
func (a *UsageAnalyzer) GetAllPages() []string {
	pages := make([]string, 0, len(a.pageUsage))
	for page := range a.pageUsage {
		pages = append(pages, page)
	}
	sort.Strings(pages)
	return pages
}

// UsesCommonChunk returns true if the page uses any components from the common chunk.
//
// Pattern: Predicate [Load: 3]
func (a *UsageAnalyzer) UsesCommonChunk(pagePath string, commonComponents []string) bool {
	pu, exists := a.pageUsage[pagePath]
	if !exists {
		return false
	}

	commonSet := make(map[string]bool)
	for _, c := range commonComponents {
		commonSet[c] = true
	}

	for _, comp := range pu.Components {
		if commonSet[comp] {
			return true
		}
	}

	return false
}
