package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/jimafisk/custom_go_template/ast"
	"github.com/jimafisk/custom_go_template/parser"
	"github.com/jimafisk/custom_go_template/transformer"
)

func main() {
	// Define command-line flags
	port := flag.Int("port", 8083, "Port to run the server on")
	outputFile := flag.String("output", "", "Output file for rendered HTML (optional)")
	flag.Parse()

	// Define the directories for pages and components
	pagesDir := filepath.Join("examples", "pages")
	componentsDir := filepath.Join("examples", "components")

	// Log the directories for debugging
	log.Printf("Pages directory: %s", filepath.Join(os.Getenv("PWD"), pagesDir))
	log.Printf("Components directory: %s", filepath.Join(os.Getenv("PWD"), componentsDir))

	// If output file is specified, render to file and exit
	if *outputFile != "" {
		// Render the comprehensive example to the output file
		html := renderToString(filepath.Join(pagesDir, "comprehensive.html"), componentsDir)
		err := os.WriteFile(*outputFile, []byte(html), 0644)
		if err != nil {
			log.Fatalf("Error writing output file: %v", err)
		}
		fmt.Printf("Rendered HTML written to %s\n", *outputFile)
		return
	}

	// Create a file server for static files
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// Handle the root path
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// If the path is just "/", serve the comprehensive example
		if r.URL.Path == "/" {
			serveTemplate(w, filepath.Join(pagesDir, "comprehensive.html"), componentsDir)
			return
		}

		// Otherwise, try to serve the requested page
		pagePath := filepath.Join(pagesDir, r.URL.Path)
		if _, err := os.Stat(pagePath); os.IsNotExist(err) {
			http.NotFound(w, r)
			return
		}

		serveTemplate(w, pagePath, componentsDir)
	})

	// Start the server
	addr := fmt.Sprintf(":%d", *port)
	fmt.Printf("Server starting at http://localhost%s\n", addr)
	err := http.ListenAndServe(addr, nil)
	if err != nil {
		log.Fatal(err)
	}
}

func serveTemplate(w http.ResponseWriter, templatePath, componentsDir string) {
	// Read the template file
	content, err := os.ReadFile(templatePath)
	if err != nil {
		http.Error(w, "Error reading template file", http.StatusInternalServerError)
		return
	}

	// Parse the template
	parsedTemplate, parseErr := parser.ParseTemplate(string(content))
	if parseErr != nil {
		http.Error(w, fmt.Sprintf("Error parsing template: %v", parseErr), http.StatusInternalServerError)
		return
	}

	// Create a data scope with some example data
	dataScope := map[string]interface{}{
		"title":       "Template Example",
		"description": "This is an example of the custom Go template engine",
		"isLoggedIn":  true, // Explicitly set to true for the conditional example
		"user": map[string]interface{}{
			"name":    "John Doe",
			"email":   "john@example.com",
			"isAdmin": true,
			"role":    "admin",
			"details": map[string]interface{}{
				"email": "john@example.com",
				"phone": "123-456-7890",
			},
		},
		"products": []interface{}{
			map[string]interface{}{
				"name":  "Product 1",
				"price": 19.99,
			},
			map[string]interface{}{
				"name":  "Product 2",
				"price": 29.99,
			},
		},
		"filteredProducts": []interface{}{
			map[string]interface{}{
				"name":  "Product 1",
				"price": 19.99,
			},
			map[string]interface{}{
				"name":  "Product 2",
				"price": 29.99,
			},
		},
		"categories": []interface{}{
			map[string]interface{}{
				"name": "Category 1",
				"items": []interface{}{
					map[string]interface{}{
						"name": "Item 1",
						"tags": []string{"tag1", "tag2"},
					},
					map[string]interface{}{
						"name": "Item 2",
						"tags": []string{"tag2", "tag3"},
					},
				},
			},
			map[string]interface{}{
				"name":  "Category 2",
				"items": []interface{}{},
			},
		},
		"settings": map[string]interface{}{
			"theme":    "dark",
			"currency": "USD",
			"language": "en",
		},
		"status": "active", // Add status for the status conditional example
	}

	// Transform the template
	transformedTemplate := transformer.TransformAST(parsedTemplate, dataScope)

	// Render the template
	tmpl := template.New("page").Funcs(template.FuncMap{
		"renderNode": renderNode,
	})

	// Parse the base template
	baseTemplate := `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Custom Go Template Example</title>
    <script defer src="https://cdn.jsdelivr.net/npm/alpinejs@3.x.x/dist/cdn.min.js"></script>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif;
            line-height: 1.6;
            color: #333;
            max-width: 800px;
            margin: 0 auto;
            padding: 20px;
        }
        h1, h2, h3, h4 {
            margin-top: 1.5em;
            margin-bottom: 0.5em;
        }
        h1 { border-bottom: 2px solid #eee; padding-bottom: 0.3em; }
        h2 { border-bottom: 1px solid #eee; padding-bottom: 0.3em; }
        code {
            background-color: #f5f5f5;
            padding: 0.2em 0.4em;
            border-radius: 3px;
            font-family: SFMono-Regular, Consolas, "Liberation Mono", Menlo, monospace;
   
	
			}
        pre {
            background-color: #f5f5f5;
            padding: 1em;
            border-radius: 5px;
            overflow-x: auto;
        }
        .product { 
            border: 1px solid #ddd; 
            padding: 10px; 
            margin: 10px 0; 
            border-radius: 5px;
        }
        .user-profile {
            background-color: #f9f9f9;
            padding: 15px;
            border-radius: 5px;
            margin-bottom: 20px;
        }
        .notification {
            padding: 10px 15px;
            margin: 10px 0;
            border-radius: 4px;
        }
        .notification.success { background-color: #d4edda; color: #155724; }
        .notification.warning { background-color: #fff3cd; color: #856404; }
        .notification.error { background-color: #f8d7da; color: #721c24; }
    </style>
</head>
<body>
    {{renderNode .}}
</body>
</html>
`
	tmpl, err = tmpl.Parse(baseTemplate)
	if err != nil {
		http.Error(w, "Error parsing template", http.StatusInternalServerError)
		return
	}

	// Execute the template
	err = tmpl.Execute(w, transformedTemplate)
	if err != nil {
		http.Error(w, "Error executing template", http.StatusInternalServerError)
		return
	}
}

func renderToString(templatePath, componentsDir string) string {
	// Read the template file
	content, err := os.ReadFile(templatePath)
	if err != nil {
		log.Fatalf("Error reading template file: %v", err)
		return ""
	}

	// Parse the template
	parsedTemplate, parseErr := parser.ParseTemplate(string(content))
	if parseErr != nil {
		log.Fatalf("Error parsing template: %v", parseErr)
		return ""
	}

	// Create a data scope with example data
	dataScope := map[string]interface{}{
		"title":       "Template Example",
		"description": "This is an example of the custom Go template engine",
		"isLoggedIn":  true, // Explicitly set to true for the conditional example
		"user": map[string]interface{}{
			"name":    "John Doe",
			"email":   "john@example.com",
			"isAdmin": true,
			"role":    "admin",
			"details": map[string]interface{}{
				"email": "john@example.com",
				"phone": "123-456-7890",
			},
		},
		"products": []interface{}{
			map[string]interface{}{
				"name":  "Product 1",
				"price": 19.99,
			},
			map[string]interface{}{
				"name":  "Product 2",
				"price": 29.99,
			},
		},
		"filteredProducts": []interface{}{
			map[string]interface{}{
				"name":  "Product 1",
				"price": 19.99,
			},
			map[string]interface{}{
				"name":  "Product 2",
				"price": 29.99,
			},
		},
		"categories": []interface{}{
			map[string]interface{}{
				"name": "Category 1",
				"items": []interface{}{
					map[string]interface{}{
						"name": "Item 1",
						"tags": []string{"tag1", "tag2"},
					},
					map[string]interface{}{
						"name": "Item 2",
						"tags": []string{"tag2", "tag3"},
					},
				},
			},
			map[string]interface{}{
				"name":  "Category 2",
				"items": []interface{}{},
			},
		},
		"settings": map[string]interface{}{
			"theme":    "dark",
			"currency": "USD",
			"language": "en",
		},
		"status": "active", // Add status for the status conditional example
	}

	// Transform the template
	transformedTemplate := transformer.TransformAST(parsedTemplate, dataScope)

	// Render the template to a string
	var sb strings.Builder
	sb.WriteString("<!DOCTYPE html>\n<html>\n<head>\n")
	sb.WriteString("    <meta charset=\"UTF-8\">\n")
	sb.WriteString("    <meta name=\"viewport\" content=\"width=device-width, initial-scale=1.0\">\n")
	sb.WriteString("    <title>Custom Go Template Example</title>\n")
	sb.WriteString("    <script defer src=\"https://cdn.jsdelivr.net/npm/alpinejs@3.x.x/dist/cdn.min.js\"></script>\n")
	sb.WriteString("    <style>\n")
	sb.WriteString("        body {\n")
	sb.WriteString("            font-family: -apple-system, BlinkMacSystemFont, \"Segoe UI\", Roboto, Helvetica, Arial, sans-serif;\n")
	sb.WriteString("            line-height: 1.6;\n")
	sb.WriteString("            color: #333;\n")
	sb.WriteString("            max-width: 800px;\n")
	sb.WriteString("            margin: 0 auto;\n")
	sb.WriteString("            padding: 20px;\n")
	sb.WriteString("        }\n")
	sb.WriteString("    </style>\n")
	sb.WriteString("</head>\n<body>\n")

	// Render the transformed template
	for _, node := range transformedTemplate.RootNodes {
		renderNodeToHTML(&sb, node)
	}

	sb.WriteString("\n</body>\n</html>")
	return sb.String()
}

// renderNode renders an AST node to HTML
func renderNode(node *ast.Template) string {
	if node == nil {
		return ""
	}

	// Render the root nodes
	var sb strings.Builder
	for _, n := range node.RootNodes {
		renderNodeToHTML(&sb, n)
	}

	return sb.String()
}

// renderNodeToHTML renders a node to HTML
func renderNodeToHTML(sb *strings.Builder, node ast.Node) {
	if node == nil {
		return
	}

	switch n := node.(type) {
	case *ast.Element:
		// Render the opening tag
		sb.WriteString("<" + n.TagName)
		for _, attr := range n.Attributes {
			sb.WriteString(" " + attr.Name)
			if attr.Value != "" {
				sb.WriteString("=\"" + attr.Value + "\"")
			}
		}
		if n.SelfClosing {
			sb.WriteString(" />")
			return
		}
		sb.WriteString(">")

		// Render the children
		for _, child := range n.Children {
			renderNodeToHTML(sb, child)
		}

		// Render the closing tag
		sb.WriteString("</" + n.TagName + ">")

	case *ast.TextNode:
		sb.WriteString(n.Content)

	case *ast.ExpressionNode:
		sb.WriteString("{{ " + n.Expression + " }}")

	case *ast.ComponentNode:
		sb.WriteString("<component name=\"" + n.Name + "\"")
		for _, prop := range n.Props {
			sb.WriteString(" " + prop.Name + "=\"" + prop.Value + "\"")
		}
		sb.WriteString(" />")

	default:
		sb.WriteString(fmt.Sprintf("<!-- Unknown node type: %T -->", node))
	}
}
