package utils

import (
	"fmt"
	"log"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

// formatArray formats a Go slice/array into a JS array literal string.
func formatArray(value any) string {
	val := reflect.ValueOf(value)
	if val.Kind() != reflect.Slice && val.Kind() != reflect.Array {
		return "[]" // Return empty array for non-slice/array types
	}
	var elements []string
	for i := 0; i < val.Len(); i++ {
		elem := val.Index(i).Interface()
		elements = append(elements, AnyToJSValue(elem)) // Recursively format each element
	}
	return "[" + strings.Join(elements, ", ") + "]"
}

// formatObject formats a Go map into a JS object literal string.
func formatObject(value any) string {
	val := reflect.ValueOf(value)
	if val.Kind() != reflect.Map {
		return "{}" // Return empty object for non-maps
	}
	var pairs []string
	mapKeys := val.MapKeys()
	// Consider sorting keys for deterministic output if needed:
	// sort.Slice(mapKeys, func(i, j int) bool { return mapKeys[i].String() < mapKeys[j].String() })
	for _, key := range mapKeys {
		// Ensure map key is a string for JS object keys
		keyStr := fmt.Sprintf("%v", key.Interface())

		// DEBUG: Log if key contains "this."
		if strings.Contains(keyStr, "this.") {
			log.Printf("[DEBUG formatObject] WARNING: Map key contains 'this.': %q", keyStr)
			log.Printf("[DEBUG formatObject] Stack trace follows...")
		}

		// Quote the key if it's not a simple identifier
		quotedKey := keyStr
		if !regexp.MustCompile(`^[a-zA-Z_$][a-zA-Z0-9_$]*$`).MatchString(keyStr) {
			quotedKey = strconv.Quote(keyStr)
		}
		mapValue := val.MapIndex(key)
		pairs = append(pairs, quotedKey+": "+AnyToJSValue(mapValue.Interface())) // Use JSValue formatting recursively
	}
	return "{" + strings.Join(pairs, ", ") + "}"
}

// formatJSElement formats a basic Go type into its JS literal representation.
func formatJSElement(value any) string {
	switch v := value.(type) {
	case string:
		return strconv.Quote(v) // JS strings need quotes
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%d", v) // Numbers are fine
	case float32, float64:
		// Use %g for potentially more concise output, or %f for standard decimal
		return fmt.Sprintf("%g", v)
	case bool:
		return strconv.FormatBool(v) // true/false are fine
	case nil:
		return "null" // null is fine
	default:
		// Fallback for unknown types: convert to string and quote it.
		// This might not be suitable for complex structs.
		return strconv.Quote(fmt.Sprintf("%v", v))
	}
}

// AnyToJSValue converts an arbitrary Go value into its JavaScript literal string representation.
func AnyToJSValue(value any) string {
	if value == nil {
		return "null"
	}
	val := reflect.ValueOf(value)
	switch val.Kind() {
	case reflect.Array, reflect.Slice:
		return formatArray(value)
	case reflect.Map:
		return formatObject(value)
	default:
		return formatJSElement(value)
	}
}

// IsBoolAndTrue checks if a value is specifically the boolean true.
func IsBoolAndTrue(value any) bool {
	if b, ok := value.(bool); ok && b {
		return true
	}
	// Note: This does not implement JavaScript-like truthiness.
	return false
}

// AnyToSlice attempts to convert an interface{} value to a slice of interface{}.
// Handles basic slice/array types. Returns nil if conversion is not possible.
func AnyToSlice(value any) []any {
	if value == nil {
		return nil
	}
	val := reflect.ValueOf(value)
	kind := val.Kind()

	if kind == reflect.Slice || kind == reflect.Array {
		length := val.Len()
		result := make([]any, length)
		for i := 0; i < length; i++ {
			result[i] = val.Index(i).Interface()
		}
		return result
	}

	// Could potentially handle other iterable types here if needed
	log.Printf("Warning: AnyToSlice could not convert value of type %T", value)
	return nil // Return nil if not a slice or array
}

// MakeGetter creates a JS object literal string for Alpine x-data,
// where keys are prop names and values are JS expressions/literals
// representing how to get the prop value (often just the prop name itself).
func MakeGetter(comp_data map[string]any) string {
	log.Printf("[DEBUG MakeGetter] Called with %d keys", len(comp_data))
	for k := range comp_data {
		if strings.Contains(k, "this.") {
			log.Printf("[DEBUG MakeGetter] WARNING: Key contains 'this.': %q", k)
		}
		log.Printf("[DEBUG MakeGetter]   Key: %q", k)
	}

	var comp_data_str string
	for name, expr := range comp_data {
		// CRITICAL FIX: Use AnyToJSValue to properly format Go values as JavaScript
		// This prevents Go's default map format (map[...]) from appearing in JS
		formattedExpr := AnyToJSValue(expr)
		comp_data_str += fmt.Sprintf("get %s() { return %s },", name, formattedExpr)
	}
	return "{" + strings.TrimSuffix(comp_data_str, ",") + "}"
}

// DeclProps generates JS variable declarations (let name = value;) from a props map.
func DeclProps(props map[string]any) string {
	var builder strings.Builder
	for name, value := range props {
		// Use AnyToJSValue to ensure value is correctly formatted as a JS literal
		builder.WriteString(fmt.Sprintf("let %s = %s;\n", name, AnyToJSValue(value)))
	}
	return builder.String()
}
