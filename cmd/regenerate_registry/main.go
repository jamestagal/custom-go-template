package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/jimafisk/custom_go_template/ast"
	"github.com/jimafisk/custom_go_template/builder"
	"github.com/jimafisk/custom_go_template/parser"
	"github.com/jimafisk/custom_go_template/transformer"
	"github.com/jimafisk/custom_go_template/utils"
)

func main() {
	log.Println("Regenerating component registry...")

	// Register stores first
	storeRegistry := registerStores()
	log.Printf("Registered %d store(s)", len(storeRegistry))

	// Collect all components
	var components []builder.ComponentTemplate

	// Register from components/ directory
	components = append(components, loadComponentsFromDir("examples/components", "../components/", storeRegistry)...)

	// Register from global/ directory
	components = append(components, loadComponentsFromDir("layouts/global", "../global/", storeRegistry)...)

	// Register from layouts/components/ directory - NEW!
	// This is the primary directory for components in the Plenti-style architecture
	components = append(components, loadComponentsFromDir("layouts/components", "../components/", storeRegistry)...)

	// Register from layouts/content/ directory
	components = append(components, loadComponentsFromDir("layouts/content", "../content/", storeRegistry)...)

	log.Printf("Loaded %d components total", len(components))

	// Generate registry
	registryJS := builder.GenerateComponentRegistry(components)

	// Write to file
	outputDir := "static/js"
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		log.Fatalf("Failed to create directory: %v", err)
	}

	outputPath := filepath.Join(outputDir, "component-registry.js")
	if err := os.WriteFile(outputPath, []byte(registryJS), 0644); err != nil {
		log.Fatalf("Failed to write file: %v", err)
	}

	log.Printf("Component registry generated successfully: %s (%d components)", outputPath, len(components))
}

func loadComponentsFromDir(dir string, pathPrefix string, storeRegistry map[string]string) []builder.ComponentTemplate {
	var components []builder.ComponentTemplate

	files, err := os.ReadDir(dir)
	if err != nil {
		log.Printf("Warning: Failed to read directory %s: %v", dir, err)
		return components
	}

	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".html") {
			componentPath := fmt.Sprintf("%s/%s", dir, file.Name())

			// Read component file
			componentContent, err := os.ReadFile(componentPath)
			if err != nil {
				log.Printf("Error reading component %s: %v", componentPath, err)
				continue
			}

			// Parse the component template
			componentAST, err := parser.ParseTemplate(string(componentContent))
			if err != nil {
				log.Printf("Error parsing component %s: %v", componentPath, err)
				continue
			}

			// Re-parse fence if component has store imports
			for i, node := range componentAST.RootNodes {
				if fence, ok := node.(*ast.FenceSection); ok {
					if strings.Contains(fence.RawContent, "import store from") {
						fenceWithStores := parser.ParseFenceContentWithStores(fence.RawContent, storeRegistry)
						componentAST.RootNodes[i] = fenceWithStores
					}
					break
				}
			}

			// Transform the component
			componentProps := extractComponentProps(componentAST)
			transformedAST := transformer.TransformAST(componentAST, componentProps)

			// CRITICAL FIX: Register component TWICE to match server behavior
			// 1. With full path prefix (for import resolution)
			compNameWithPath := fmt.Sprintf("%s%s", pathPrefix, file.Name())
			components = append(components, builder.ComponentTemplate{
				Name: compNameWithPath,
				AST:  transformedAST,
			})
			log.Printf("Loaded component: %s from %s", compNameWithPath, componentPath)

			// 2. With just the base name (without .html) for JSON lookups
			// This matches what the runtime loader expects when component.name = "jim_test_greeting"
			baseName := strings.TrimSuffix(file.Name(), ".html")
			components = append(components, builder.ComponentTemplate{
				Name: baseName,
				AST:  transformedAST,
			})
			log.Printf("Also registered as: %s (for JSON component.name lookups)", baseName)
		}
	}

	return components
}

func extractComponentProps(template *ast.Template) map[string]interface{} {
	props := make(map[string]interface{})

	for _, node := range template.RootNodes {
		if fence, ok := node.(*ast.FenceSection); ok {
			// Extract props from fence (Props is []PropNode)
			for _, prop := range fence.Props {
				props[prop.Name] = prop.DefaultValue
			}
			break
		}
	}

	return props
}

// registerStores loads all store files from the stores/ directory.
// This is a wrapper around utils.RegisterStores() for consistency.
func registerStores() map[string]string {
	return utils.RegisterStores()
}
