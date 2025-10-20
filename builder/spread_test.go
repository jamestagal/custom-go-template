package builder

import (
	"testing"
)

func TestSpreadOperatorBugFix(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "spread in array literal",
			input:    "[newAnimal, ...animals]",
			expected: "[props.newAnimal, ...props.animals]",
		},
		{
			name:     "assignment with spread",
			input:    "animals = [newAnimal, ...animals]",
			expected: "props.animals = [props.newAnimal, ...props.animals]",
		},
		{
			name:     "full expression from jim-test.html",
			input:    "animals = [newAnimal, ...animals]; newAnimal = ''",
			expected: "props.animals = [props.newAnimal, ...props.animals]; props.newAnimal = ''",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			skipList := make(map[string]bool)
			result := prefixIdentifiersInExpression(tt.input, skipList)
			
			if result != tt.expected {
				t.Errorf("\nInput:    %s\nExpected: %s\nGot:      %s", tt.input, tt.expected, result)
			}
		})
	}
}
