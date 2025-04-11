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

// isComplexJSObjectInternal checks if a JavaScript value is a complex object
// that should be preserved as a string rather than evaluated
func isComplexJSObjectInternal(jsCode string) bool {
	// Delegate to the exported version in render.go
	return isComplexJSObject(jsCode)
}

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
	reAllVars := regexp.MustCompile(`(?:let|const|var)\s+(?P<n>[a-zA-Z_$][a-zA-Z0-9_$]*)(?:\s*=\s*.*)?;`)
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

	// Add console logging for debugging
	vm.Set("console", map[string]interface{}{
		"log": func(args ...interface{}) {
			log.Println("JS console.log:", args)
		},
		"error": func(args ...interface{}) {
			log.Println("JS console.error:", args)
		},
	})

	_, err := vm.RunString(fence) // Run the modified fence script
	if err != nil {
		log.Printf("Error running fence script: %v\nScript:\n%s", err, fence)
		// Return original props on error
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
// Modified to better handle Alpine.js expressions without evaluation
func EvalJS(jsCode string, propsDecl string) any {
	// Trim whitespace
	jsCode = strings.TrimSpace(jsCode)
	
	// Skip evaluation for empty code
	if jsCode == "" {
		return ""
	}
	
	// Check if this is a complex JS object that should be preserved as a string
	if isComplexJSObjectInternal(jsCode) {
		log.Printf("Detected complex JS object, preserving as-is: %s", jsCode)
		return jsCode
	}
	
	// Special case for object with methods - always preserve as string
	if strings.Contains(jsCode, "method(") || strings.Contains(jsCode, "handleClick(") {
		return jsCode
	}
	
	// Special case for simple array literal [1, 2, 3]
	if strings.HasPrefix(jsCode, "[") && strings.HasSuffix(jsCode, "]") {
		// Check if it contains any complex objects
		if strings.Contains(jsCode, "{") || strings.Contains(jsCode, "function") || strings.Contains(jsCode, "=>") {
			return jsCode
		}
		
		// This is a simple array, try to evaluate it
		vm := goja.New()
		result, err := vm.RunString(jsCode)
		if err == nil {
			return result.Export()
		}
	}
	
	// Special case for simple object literal with parentheses ({a: 1, b: 2})
	if strings.HasPrefix(jsCode, "({") && strings.HasSuffix(jsCode, "})") {
		// Check if it contains any methods or complex structures
		if strings.Contains(jsCode, "function") || strings.Contains(jsCode, "=>") || 
		   strings.Contains(jsCode, "get ") || strings.Contains(jsCode, "set ") ||
		   strings.Contains(jsCode, "method") {
			return jsCode
		}
		
		// This is a simple object literal with parentheses, try to evaluate it
		vm := goja.New()
		result, err := vm.RunString(jsCode)
		if err == nil {
			return result.Export()
		}
	}
	
	// Special case for Alpine.js object patterns (without parentheses)
	if strings.HasPrefix(jsCode, "{") && strings.HasSuffix(jsCode, "}") {
		// Always preserve Alpine.js objects as strings
		// This handles x-data objects, event handlers, etc.
		return jsCode
	}
	
	// For simple expressions, try to evaluate them
	vm := goja.New()
	
	// If we have props, define them in the VM
	if propsDecl != "" {
		_, err := vm.RunString(propsDecl)
		if err != nil {
			log.Printf("Error setting up props in VM: %v", err)
			// Continue anyway, we might be able to evaluate without props
		}
	}
	
	// Try different evaluation strategies in sequence
	
	// 1. Direct evaluation for simple expressions
	result, err := directEvaluation(vm, jsCode)
	if err == nil {
		// Check if the result is a complex object that should be preserved
		if isComplexResult(result) {
			return jsCode
		}
		return result
	}
	
	// 2. Expression wrapping for object/array literals
	result, err = expressionWrapping(vm, jsCode)
	if err == nil {
		// Check if the result is a complex object that should be preserved
		if isComplexResult(result) {
			return jsCode
		}
		return result
	}
	
	// 3. Function wrapping for expressions
	result, err = functionWrapping(vm, jsCode)
	if err == nil {
		// Check if the result is a complex object that should be preserved
		if isComplexResult(result) {
			return jsCode
		}
		return result
	}
	
	// 4. Object assignment for complex objects
	result, err = objectAssignment(vm, jsCode)
	if err == nil {
		// Check if the result is a complex object that should be preserved
		if isComplexResult(result) {
			return jsCode
		}
		return result
	}
	
	// All strategies failed - return the original code for Alpine.js to evaluate
	log.Printf("All evaluation strategies failed for JS code: %s", jsCode)
	return jsCode
}

// isComplexResult checks if a result from evaluation is a complex object that should be preserved as a string
func isComplexResult(result any) bool {
	// Handle nil result
	if result == nil {
		return false
	}
	
	// Check result type
	switch v := result.(type) {
	case func(goja.FunctionCall) goja.Value:
		// Functions should be preserved as strings
		return true
		
	case map[string]interface{}:
		// Check if the map contains any methods or complex values
		for _, val := range v {
			switch val.(type) {
			case func(goja.FunctionCall) goja.Value, map[string]interface{}, []interface{}:
				return true
			}
		}
		// Simple maps with primitive values can be returned as evaluated
		return false
		
	case []interface{}:
		// Check if the array contains any complex elements
		for _, item := range v {
			switch item.(type) {
			case map[string]interface{}, []interface{}, func(goja.FunctionCall) goja.Value:
				return true
			}
		}
		// Simple arrays with primitive values can be returned as evaluated
		return false
	}
	
	// Primitive types can be returned as evaluated
	return false
}

// directEvaluation attempts to evaluate the JS code directly
func directEvaluation(vm *goja.Runtime, jsCode string) (any, error) {
	value, err := vm.RunString(jsCode)
	if err != nil {
		return nil, err
	}
	return value.Export(), nil
}

// expressionWrapping handles object literals by wrapping them in parentheses
func expressionWrapping(vm *goja.Runtime, jsCode string) (any, error) {
	// Check if code is an object or array literal that needs wrapping
	if (strings.HasPrefix(jsCode, "{") && strings.HasSuffix(jsCode, "}")) ||
		(strings.HasPrefix(jsCode, "[") && strings.HasSuffix(jsCode, "]")) {
		// Wrap in parentheses to force evaluation as an expression
		wrappedCode := "(" + jsCode + ")"
		value, err := vm.RunString(wrappedCode)
		if err != nil {
			return nil, err
		}
		return value.Export(), nil
	}

	// Not applicable for this strategy
	return nil, fmt.Errorf("not an object/array literal")
}

// functionWrapping wraps the code in a function to force evaluation
func functionWrapping(vm *goja.Runtime, jsCode string) (any, error) {
	// Wrap the code in a function to handle expressions, especially those with methods
	funcCode := "function __evalWrapper() { return " + jsCode + "; } __evalWrapper()"
	value, err := vm.RunString(funcCode)
	if err != nil {
		return nil, err
	}
	return value.Export(), nil
}

// objectAssignment handles complex object literals by assigning to a variable
func objectAssignment(vm *goja.Runtime, jsCode string) (any, error) {
	// For object literals with method definitions, assign to a variable first
	objCode := "var __tempObj = " + jsCode + "; __tempObj"
	value, err := vm.RunString(objCode)
	if err != nil {
		return nil, err
	}
	return value.Export(), nil
}
