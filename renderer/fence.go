package renderer

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/dop251/goja"
	// Assuming utils package is in the same module path
	"github.com/jimafisk/custom_go_template/utils"
)

// Component struct definition might need to be shared or moved to a common place if used by multiple packages.
// For now, let's redefine it here or assume it will be moved later.
type Component struct {
	Name string
	Path string
}

// GetComponents extracts component imports from fence content and returns the cleaned fence.
func GetComponents(fence string) (string, []Component) {
	components := []Component{}
	cleanedFence := ""
	reImport := regexp.MustCompile(`import\s+([A-Za-z_][A-Za-z_0-9]*)\s+from\s*"([^"]+)";`)
	lines := strings.Split(fence, "\n")
	for _, line := range lines {
		match := reImport.FindStringSubmatch(line)
		if len(match) > 1 {
			compName := match[1]
			compPath := match[2]
			// TODO: Resolve relative paths based on the current file's path
			components = append(components, Component{
				Name: compName,
				Path: compPath, // Path might need resolving
			})
		} else {
			cleanedFence += line + "\n" // Keep non-import lines
		}
	}
	return strings.TrimSpace(cleanedFence), components
}

// SetProps replaces 'prop' declarations in the fence with 'let' assignments using provided prop values.
func SetProps(fence string, props map[string]any) string {
	// Replace declared props with their passed values
	for name, value := range props {
		// Regex to find 'prop name;' or 'prop name = defaultValue;'
		reProp := regexp.MustCompile(fmt.Sprintf(`prop\s+(%s)(\s*=\s*(.*?))?;`, regexp.QuoteMeta(name)))
		// Replace with 'let name = value;' using the actual passed value
		fence = reProp.ReplaceAllString(fence, "let "+name+" = "+utils.AnyToJSValue(value)+";") // Use utils.AnyToJSValue
	}
	// Convert any remaining 'prop name = defaultValue;' or 'prop name;' to 'let name = defaultValue;' or 'let name;'
	rePropDefaults := regexp.MustCompile(`prop\s+(.*?);`)
	fence = rePropDefaults.ReplaceAllString(fence, "let $1;")

	return fence
}

// GetAllVars extracts all variable names declared with let, const, or var in the fence.
func GetAllVars(fence string) []string {
	allVars := []string{}
	// Regex to find variable declarations (let, const, var)
	// It captures the variable name, ignoring potential assignments for this function's purpose.
	reAllVars := regexp.MustCompile(`(?:let|const|var)\s+(?P<name>[a-zA-Z_$][a-zA-Z0-9_$]*)(?:\s*=\s*.*)?;`)
	nameIndex := reAllVars.SubexpIndex("name")
	matches := reAllVars.FindAllStringSubmatch(fence, -1)
	for _, currentVar := range matches {
		if nameIndex >= 0 && nameIndex < len(currentVar) {
			allVars = append(allVars, currentVar[nameIndex])
		}
	}
	return allVars
}

// EvaluateProps runs the fence script in Goja and updates the props map with evaluated variable values.
func EvaluateProps(fence string, allVars []string, props map[string]any) map[string]any {
	vm := goja.New()
	// Prepend prop declarations (already processed by SetProps)
	// No, SetProps modified the fence string itself.

	_, err := vm.RunString(fence) // Run the modified fence script
	if err != nil {
		log.Printf("Error running fence script: %v\nScript:\n%s", err, fence)
		// Return original props on error? Or an empty map?
		return props
	}

	// Re-evaluate all declared variables (props and computed ones)
	evaluatedProps := make(map[string]any)
	for _, name := range allVars {
		evaluated_value := vm.Get(name).Export()
		// Handle nil (JS undefined) appropriately
		if evaluated_value == nil {
			// Maybe represent as Go nil or a specific marker?
			evaluated_value = nil // Or keep as nil
		}
		evaluatedProps[name] = evaluated_value
	}

	// Ensure original props passed in are preserved if not overwritten by fence logic
	// (This might not be desired - fence logic should potentially override)
	// Let's assume fence logic takes precedence. The evaluatedProps map contains the final state.

	return evaluatedProps
}

// EvalJS evaluates a snippet of JS code within a given variable context.
func EvalJS(jsCode string, propsDecl string) any {
	// props_decl := declProps(props) // Removed: propsDecl is now passed in
	vm := goja.New()
	// Consider adding console logging from Goja to Go logs
	// vm.Set("console", map[string]interface{}{"log": log.Println, "error": log.Println})
	// Use the passed-in propsDecl string
	goja_value, err := vm.RunString(propsDecl + jsCode)
	if err != nil {
		log.Printf("Error evaluating JS '%s' with context: %v", jsCode, err) // Improved error log
		return ""                                                            // Return empty string on error, or perhaps nil?
	}
	exportedVal := goja_value.Export()
	if exportedVal == nil {
		return "" // Or nil?
	}
	return exportedVal
}
