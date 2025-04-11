package renderer

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"

	// Assuming utils and scoping packages are in the same module path
	"github.com/jimafisk/custom_go_template/scoping"
	"github.com/jimafisk/custom_go_template/utils"
	// Need access to Render function, indicates potential circular dependency or need for restructuring.
	// For now, assume Render is accessible or will be passed in/refactored.
)

// getCompArgs parses component arguments string into props and data maps.
func getCompArgs(comp_args []string, parentProps map[string]any) (map[string]any, map[string]any) {
	comp_props := make(map[string]any)
	comp_data := make(map[string]any)                             // For x-data generation
	reArg := regexp.MustCompile(`^({?)([a-zA-Z0-9_]+)(}?)$`)      // For shorthand {prop}
	reArgEq := regexp.MustCompile(`^([a-zA-Z0-9_]+)\s*=\s*(.*)$`) // For prop={expr} or prop="static"

	for _, comp_arg := range comp_args {
		comp_arg = strings.TrimSpace(comp_arg)
		if comp_arg == "" {
			continue
		}

		if matches := reArg.FindStringSubmatch(comp_arg); len(matches) == 4 && matches[1] == "{" && matches[3] == "}" {
			// Shorthand {prop}
			prop_name := matches[2]
			if val, ok := parentProps[prop_name]; ok {
				comp_props[prop_name] = val
				comp_data[prop_name] = prop_name // Reference parent prop in x-data
			} else {
				log.Printf("Warning: Shorthand prop '%s' used but not found in parent props.", prop_name)
			}
		} else if matches := reArgEq.FindStringSubmatch(comp_arg); len(matches) == 3 {
			// prop=value format
			prop_name := matches[1]
			prop_value_str := strings.TrimSpace(matches[2])

			if strings.HasPrefix(prop_value_str, "{") && strings.HasSuffix(prop_value_str, "}") {
				// Dynamic value: prop={expression}
				expression := strings.Trim(prop_value_str, "{}")
				// Generate parent props declarations to pass to EvalJS
				parentPropsDecl := utils.DeclProps(parentProps)
				prop_value := EvalJS(expression, parentPropsDecl) // Pass declarations string
				comp_props[prop_name] = prop_value
				comp_data[prop_name] = expression // Use expression for x-data getter
			} else {
				// Static value: prop="string" or prop='string' or prop=number etc.
				var staticValue interface{} // Declare only once
				// Use utils.AnyToJSValue logic for parsing static values might be better
				// For now, keep simplified parsing:
				// var staticValue interface{} // Remove redeclaration
				if (strings.HasPrefix(prop_value_str, `"`) && strings.HasSuffix(prop_value_str, `"`)) ||
					(strings.HasPrefix(prop_value_str, `'`) && strings.HasSuffix(prop_value_str, `'`)) {
					// Use Unquote for proper handling of escaped quotes within the string
					unquotedValue, err := strconv.Unquote(prop_value_str)
					if err == nil {
						staticValue = unquotedValue
					} else {
						log.Printf("Warning: Failed to unquote static string prop '%s': %v", prop_name, err)
						staticValue = prop_value_str // Fallback to raw string
					}
				} else {
					// Attempt to parse as number or bool before defaulting to string
					if val, err := strconv.ParseFloat(prop_value_str, 64); err == nil {
						staticValue = val
					} else if val, err := strconv.ParseBool(prop_value_str); err == nil {
						staticValue = val
					} else {
						staticValue = prop_value_str // Default to string
					}
				}
				comp_props[prop_name] = staticValue
				comp_data[prop_name] = utils.AnyToJSValue(staticValue) // Use utils.AnyToJSValue for x-data
			}
		} else {
			log.Printf("Warning: Invalid component argument format: %s", comp_arg)
		}
	}
	return comp_props, comp_data
}

// RenderComponents handles rendering static and dynamic components within the markup.
// TODO: Refactor this to work with TemplateParts or an AST instead of regex on markup string.
// It currently relies on string replacement and recursive calls to a Render function (which needs to be accessible).
func RenderComponents(markup, script, style string, props map[string]any, components []Component, renderFunc func(string, map[string]any) (string, string, string)) (string, string, string) {

	// Handle staticly imported components
	for _, component := range components {
		reComponent := regexp.MustCompile(fmt.Sprintf(`<%s(\s+[^>]*)?/>`, component.Name))
		for {
			match := reComponent.FindStringSubmatchIndex(markup)
			if match == nil {
				break
			}
			argsStr := ""
			if match[2] != -1 {
				argsStr = strings.TrimSpace(markup[match[2]:match[3]])
			}
			// Improved Arg Splitting
			var comp_args []string
			reArgs := regexp.MustCompile(`({[^}]*}|"[^"]*"|'[^']*'|[^\s=]+)\s*=\s*({[^}]*}|"[^"]*"|'[^']*'|[^\s}]+)|({[^}]+})|[^\s]+`)
			matches := reArgs.FindAllString(argsStr, -1)
			currentArg := ""
			for _, m := range matches {
				if strings.Contains(m, "=") && !strings.HasPrefix(m, "{") {
					if currentArg != "" {
						comp_args = append(comp_args, currentArg)
					}
					comp_args = append(comp_args, m)
					currentArg = ""
				} else if strings.HasPrefix(m, "{") && strings.HasSuffix(m, "}") {
					if currentArg != "" && strings.Contains(currentArg, "=") {
						currentArg += m
						comp_args = append(comp_args, currentArg)
						currentArg = ""
					} else {
						if currentArg != "" {
							comp_args = append(comp_args, currentArg)
						}
						comp_args = append(comp_args, m)
						currentArg = ""
					}
				} else {
					if currentArg != "" {
						currentArg += " " + m
					} else {
						currentArg = m
					}
				}
			}
			if currentArg != "" {
				comp_args = append(comp_args, currentArg)
			}

			comp_props, comp_data := getCompArgs(comp_args, props)
			// Recursive call using the passed renderFunc
			comp_markup, comp_script, comp_style := renderFunc(component.Path, comp_props)

			// Scoping should ideally happen after all rendering is complete.
			// Applying it here within the loop is problematic for nested scopes.
			comp_markup, comp_scopedElements := scoping.ScopeHTMLComp(comp_markup, comp_props, comp_data)
			comp_style, _ = scoping.ScopeCSS(comp_style, comp_scopedElements) // Use scoping package
			comp_script = scoping.ScopeJS(comp_script, comp_scopedElements)   // Use scoping package

			markup = markup[:match[0]] + comp_markup + markup[match[1]:]
			script += comp_script
			style += comp_style
		}
	}

	// Handle dynamic components <= ... />
	reDynamicComponent := regexp.MustCompile(`<=(".*?"|'.*?'|{.*?})\s*(.*?)?\s*/>`)
	for {
		match := reDynamicComponent.FindStringSubmatchIndex(markup)
		if match == nil {
			break
		}
		compPathExpr := markup[match[2]:match[3]]
		argsStr := ""
		if match[4] != -1 {
			argsStr = strings.TrimSpace(markup[match[4]:match[5]])
		}
		// Generate props declarations to pass to EvalJS for path evaluation
		propsDecl := utils.DeclProps(props)
		comp_path_any := EvalJS(strings.Trim(compPathExpr, "{}"), propsDecl) // Pass declarations string
		comp_path, ok := comp_path_any.(string)
		if !ok {
			log.Printf("Warning: Dynamic component path expression did not evaluate to a string: %s", compPathExpr)
			// Remove the tag and continue
			markup = markup[:match[0]] + markup[match[1]:]
			continue
		}
		comp_path = strings.Trim(comp_path, "\"'`") // Trim quotes if path was a string literal

		// Use improved arg splitting here too
		var comp_args []string
		reArgs := regexp.MustCompile(`({[^}]*}|"[^"]*"|'[^']*'|[^\s=]+)\s*=\s*({[^}]*}|"[^"]*"|'[^']*'|[^\s}]+)|({[^}]+})|[^\s]+`)
		matches := reArgs.FindAllString(argsStr, -1)
		currentArg := ""
		for _, m := range matches {
			if strings.Contains(m, "=") && !strings.HasPrefix(m, "{") {
				if currentArg != "" {
					comp_args = append(comp_args, currentArg)
				}
				comp_args = append(comp_args, m)
				currentArg = ""
			} else if strings.HasPrefix(m, "{") && strings.HasSuffix(m, "}") {
				if currentArg != "" && strings.Contains(currentArg, "=") {
					currentArg += m
					comp_args = append(comp_args, currentArg)
					currentArg = ""
				} else {
					if currentArg != "" {
						comp_args = append(comp_args, currentArg)
					}
					comp_args = append(comp_args, m)
					currentArg = ""
				}
			} else {
				if currentArg != "" {
					currentArg += " " + m
				} else {
					currentArg = m
				}
			}
		}
		if currentArg != "" {
			comp_args = append(comp_args, currentArg)
		}

		comp_props, comp_data := getCompArgs(comp_args, props)
		// Recursive call using the passed renderFunc
		comp_markup, comp_script, comp_style := renderFunc(comp_path, comp_props)

		comp_markup, comp_scopedElements := scoping.ScopeHTMLComp(comp_markup, comp_props, comp_data)
		comp_style, _ = scoping.ScopeCSS(comp_style, comp_scopedElements) // Use scoping package
		comp_script = scoping.ScopeJS(comp_script, comp_scopedElements)   // Use scoping package

		markup = markup[:match[0]] + comp_markup + markup[match[1]:]
		script += comp_script
		style += comp_style
	}
	return markup, script, style
}
