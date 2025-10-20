package builder

import (
	"fmt"
	"strings"
	"testing"
)

func TestSpreadOperatorDebug(t *testing.T) {
	input := "animals = [newAnimal, ...animals]; newAnimal = ''"
	expected := "props.animals = [props.newAnimal, ...props.animals]; props.newAnimal = ''"
	
	skipList := make(map[string]bool)
	result := prefixIdentifiersInExpression(input, skipList)
	
	fmt.Println("Input:   ", input)
	fmt.Println("Expected:", expected)
	fmt.Println("Got:     ", result)
	fmt.Println()
	
	// Find the difference
	for i := 0; i < len(expected) && i < len(result); i++ {
		if expected[i] != result[i] {
			fmt.Printf("First difference at position %d:\n", i)
			fmt.Printf("Expected: '%s' (byte %d)\n", string(expected[i]), expected[i])
			fmt.Printf("Got:      '%s' (byte %d)\n", string(result[i]), result[i])
			fmt.Printf("Context expected: ...%s...\n", expected[max(0,i-10):min(len(expected),i+20)])
			fmt.Printf("Context got:      ...%s...\n", result[max(0,i-10):min(len(result),i+20)])
			break
		}
	}
	
	if result != expected {
		t.Errorf("Mismatch")
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Manual step-through for debugging
func TestManualDebug(t *testing.T) {
	expr := "];"
	var result strings.Builder
	var currentToken strings.Builder
	combinedSkip := make(map[string]bool)
	
	for i := 0; i < len(expr); i++ {
		ch := expr[i]
		fmt.Printf("i=%d ch='%c' currentToken='%s'\n", i, ch, currentToken.String())
		
		if isOperatorOrDelimiter(ch) {
			if currentToken.Len() > 0 {
				token := currentToken.String()
				processed := processToken(token, combinedSkip)
				fmt.Printf("  Processing token '%s' -> '%s'\n", token, processed)
				result.WriteString(processed)
				currentToken.Reset()
			}
			result.WriteByte(ch)
			fmt.Printf("  Wrote delimiter '%c', result so far: '%s'\n", ch, result.String())
		} else {
			currentToken.WriteByte(ch)
			fmt.Printf("  Accumulated to token: '%s'\n", currentToken.String())
		}
	}
	
	fmt.Println("Final result:", result.String())
}
