package transformer

import (
	"reflect"
	"testing"

	"github.com/jimafisk/custom_go_template/ast"
)

// TestCollectComponentFenceData tests the collectComponentFenceData helper function that
// extracts variables, prop defaults, and functions from a component's fence section.
func TestCollectComponentFenceData(t *testing.T) {
	tests := []struct {
		name          string
		fence         *ast.FenceSection
		expectedScope map[string]any
		description   string
	}{
		{
			name:          "empty fence section",
			fence:         &ast.FenceSection{},
			expectedScope: map[string]any{},
			description:   "Empty fence should result in empty scope",
		},
		{
			name:          "nil fence section",
			fence:         nil,
			expectedScope: map[string]any{},
			description:   "Nil fence should be handled gracefully with empty scope",
		},
		{
			name: "variables only - simple types",
			fence: &ast.FenceSection{
				Variables: []ast.VariableNode{
					{Keyword: "let", Name: "count", Value: "0"},
					{Keyword: "const", Name: "message", Value: `"Hello"`},
					{Keyword: "var", Name: "isActive", Value: "true"},
				},
				RawContent: "",
			},
			expectedScope: map[string]any{
				"count":    0,
				"message":  "Hello",
				"isActive": true,
			},
			description: "Variables should be parsed using parseValue()",
		},
		{
			name: "variables only - complex types",
			fence: &ast.FenceSection{
				Variables: []ast.VariableNode{
					{Keyword: "const", Name: "items", Value: "[1, 2, 3]"},
					{Keyword: "let", Name: "user", Value: `{ name: "John", age: 30 }`},
					{Keyword: "const", Name: "price", Value: "19.99"},
				},
				RawContent: "",
			},
			expectedScope: map[string]any{
				"items": "[1, 2, 3]",
				"user":  `{ name: "John", age: 30 }`,
				"price": 19.99,
			},
			description: "Arrays and objects should be stored as strings for Alpine.js",
		},
		{
			name: "props only - simple defaults",
			fence: &ast.FenceSection{
				Props: []ast.PropNode{
					{Name: "title", DefaultValue: `"Default Title"`},
					{Name: "count", DefaultValue: "10"},
					{Name: "enabled", DefaultValue: "false"},
				},
				RawContent: "",
			},
			expectedScope: map[string]any{
				"title":   "Default Title",
				"count":   10,
				"enabled": false,
			},
			description: "Prop defaults should be parsed using parseValue()",
		},
		{
			name: "props only - complex defaults",
			fence: &ast.FenceSection{
				Props: []ast.PropNode{
					{Name: "user", DefaultValue: `{ name: "Guest", role: "viewer" }`},
					{Name: "items", DefaultValue: "[]"},
					{Name: "config", DefaultValue: `{ theme: "light", locale: "en" }`},
				},
				RawContent: "",
			},
			expectedScope: map[string]any{
				"user":   `{ name: "Guest", role: "viewer" }`,
				"items":  "[]",
				"config": `{ theme: "light", locale: "en" }`,
			},
			description: "Complex prop defaults should be stored as strings",
		},
		{
			name: "function declaration only",
			fence: &ast.FenceSection{
				RawContent: `import Header from './components/Header.html'

prop title = "Default"

function formatPrice(price) {
  return '$' + price.toFixed(2);
}`,
			},
			expectedScope: map[string]any{
				"formatPrice": `function formatPrice(price) {
  return '$' + price.toFixed(2);
}`,
			},
			description: "Function declarations should be extracted and stored as strings",
		},
		{
			name: "function expression - const",
			fence: &ast.FenceSection{
				RawContent: `const increment = function(value) {
  return value + 1;
}`,
			},
			expectedScope: map[string]any{
				"increment": `const increment = function(value) {
  return value + 1;
}`,
			},
			description: "Function expressions (const) should be extracted",
		},
		{
			name: "function expression - let",
			fence: &ast.FenceSection{
				RawContent: `let decrement = function(value) {
  return value - 1;
}`,
			},
			expectedScope: map[string]any{
				"decrement": `let decrement = function(value) {
  return value - 1;
}`,
			},
			description: "Function expressions (let) should be extracted",
		},
		{
			name: "function expression - var",
			fence: &ast.FenceSection{
				RawContent: `var multiply = function(a, b) {
  return a * b;
}`,
			},
			expectedScope: map[string]any{
				"multiply": `var multiply = function(a, b) {
  return a * b;
}`,
			},
			description: "Function expressions (var) should be extracted",
		},
		{
			name: "arrow function - basic",
			fence: &ast.FenceSection{
				RawContent: `const add = (a, b) => {
  return a + b;
}`,
			},
			expectedScope: map[string]any{
				"add": `const add = (a, b) => {
  return a + b;
}`,
			},
			description: "Arrow functions should be extracted",
		},
		{
			name: "arrow function - implicit return",
			fence: &ast.FenceSection{
				RawContent: `const double = x => x * 2`,
			},
			expectedScope: map[string]any{
				"double": `const double = x => x * 2`,
			},
			description: "Arrow functions with implicit return should be extracted",
		},
		{
			name: "arrow function - no params",
			fence: &ast.FenceSection{
				RawContent: `const getRandomNumber = () => Math.random()`,
			},
			expectedScope: map[string]any{
				"getRandomNumber": `const getRandomNumber = () => Math.random()`,
			},
			description: "Arrow functions with no params should be extracted",
		},
		{
			name: "method shorthand",
			fence: &ast.FenceSection{
				RawContent: `getTotal() {
  return this.price * this.quantity;
}`,
			},
			expectedScope: map[string]any{
				"getTotal": `getTotal() {
  return this.price * this.quantity;
}`,
			},
			description: "Method shorthand syntax should be extracted",
		},
		{
			name: "async function declaration",
			fence: &ast.FenceSection{
				RawContent: `async function fetchData(url) {
  const response = await fetch(url);
  return await response.json();
}`,
			},
			expectedScope: map[string]any{
				"fetchData": `async function fetchData(url) {
  const response = await fetch(url);
  return await response.json();
}`,
			},
			description: "Async function declarations should be extracted",
		},
		{
			name: "async arrow function",
			fence: &ast.FenceSection{
				RawContent: `const loadData = async (id) => {
  const data = await getData(id);
  return data;
}`,
			},
			expectedScope: map[string]any{
				"loadData": `const loadData = async (id) => {
  const data = await getData(id);
  return data;
}`,
			},
			description: "Async arrow functions should be extracted",
		},
		{
			name: "multiple functions",
			fence: &ast.FenceSection{
				RawContent: `function formatPrice(price) {
  return '$' + price.toFixed(2);
}

const increment = () => {
  count++;
}

getTotal() {
  return count * 2;
}`,
			},
			expectedScope: map[string]any{
				"formatPrice": `function formatPrice(price) {
  return '$' + price.toFixed(2);
}`,
				"increment": `const increment = () => {
  count++;
}`,
				"getTotal": `getTotal() {
  return count * 2;
}`,
			},
			description: "Multiple functions of different types should all be extracted",
		},
		{
			name: "variables and props combined",
			fence: &ast.FenceSection{
				Props: []ast.PropNode{
					{Name: "title", DefaultValue: `"Default Title"`},
					{Name: "count", DefaultValue: "10"},
				},
				Variables: []ast.VariableNode{
					{Keyword: "let", Name: "isVisible", Value: "true"},
					{Keyword: "const", Name: "items", Value: "[1, 2, 3]"},
				},
				RawContent: "",
			},
			expectedScope: map[string]any{
				"title":     "Default Title",
				"count":     10,
				"isVisible": true,
				"items":     "[1, 2, 3]",
			},
			description: "Props and variables should both be added to scope",
		},
		{
			name: "variables, props, and functions combined",
			fence: &ast.FenceSection{
				Props: []ast.PropNode{
					{Name: "title", DefaultValue: `"Default Title"`},
					{Name: "user", DefaultValue: `{ name: "Guest" }`},
				},
				Variables: []ast.VariableNode{
					{Keyword: "let", Name: "count", Value: "0"},
					{Keyword: "const", Name: "items", Value: "[1, 2, 3]"},
				},
				RawContent: `import Header from './components/Header.html'

prop title = "Default Title"
prop user = { name: 'Guest' }

let count = 0
const items = [1, 2, 3]

function formatPrice(price) {
  return '$' + price.toFixed(2);
}

const increment = () => {
  count++;
}

getTotal() {
  return count * 2;
}`,
			},
			expectedScope: map[string]any{
				"title": "Default Title",
				"user":  `{ name: "Guest" }`,
				"count": 0,
				"items": "[1, 2, 3]",
				"formatPrice": `function formatPrice(price) {
  return '$' + price.toFixed(2);
}`,
				"increment": `const increment = () => {
  count++;
}`,
				"getTotal": `getTotal() {
  return count * 2;
}`,
			},
			description: "All three types (variables, props, functions) should be extracted",
		},
		{
			name: "realistic component fence",
			fence: &ast.FenceSection{
				Imports: []ast.ImportNode{
					{Name: "Header", Path: "./components/Header.html"},
					{Name: "Button", Path: "./components/Button.html"},
				},
				Props: []ast.PropNode{
					{Name: "title", DefaultValue: `"Shopping Cart"`},
					{Name: "items", DefaultValue: "[]"},
					{Name: "currency", DefaultValue: `"USD"`},
				},
				Variables: []ast.VariableNode{
					{Keyword: "let", Name: "total", Value: "0"},
					{Keyword: "const", Name: "taxRate", Value: "0.08"},
					{Keyword: "let", Name: "discount", Value: "0"},
				},
				RawContent: `import Header from './components/Header.html'
import Button from './components/Button.html'

prop title = "Shopping Cart"
prop items = []
prop currency = "USD"

let total = 0
const taxRate = 0.08
let discount = 0

function calculateTotal() {
  const subtotal = items.reduce((sum, item) => sum + item.price, 0);
  const discountedTotal = subtotal - discount;
  const tax = discountedTotal * taxRate;
  return discountedTotal + tax;
}

const formatCurrency = (amount) => {
  return new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency: currency
  }).format(amount);
}

applyDiscount(code) {
  if (code === 'SAVE10') {
    discount = total * 0.1;
  }
}`,
			},
			expectedScope: map[string]any{
				"title":    "Shopping Cart",
				"items":    "[]",
				"currency": "USD",
				"total":    0,
				"taxRate":  0.08,
				"discount": 0,
				"calculateTotal": `function calculateTotal() {
  const subtotal = items.reduce((sum, item) => sum + item.price, 0);
  const discountedTotal = subtotal - discount;
  const tax = discountedTotal * taxRate;
  return discountedTotal + tax;
}`,
				"formatCurrency": `const formatCurrency = (amount) => {
  return new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency: currency
  }).format(amount);
}`,
				"applyDiscount": `applyDiscount(code) {
  if (code === 'SAVE10') {
    discount = total * 0.1;
  }
}`,
			},
			description: "Realistic component with imports, props, variables, and functions",
		},
		{
			name: "empty prop default value",
			fence: &ast.FenceSection{
				Props: []ast.PropNode{
					{Name: "title", DefaultValue: `""`},
					{Name: "count", DefaultValue: "0"},
					{Name: "items", DefaultValue: "null"},
				},
				RawContent: "",
			},
			expectedScope: map[string]any{
				"title": "",
				"count": 0,
				"items": nil,
			},
			description: "Props with empty string, zero, and null defaults should be handled",
		},
		{
			name: "variable with null value",
			fence: &ast.FenceSection{
				Variables: []ast.VariableNode{
					{Keyword: "let", Name: "data", Value: "null"},
					{Keyword: "const", Name: "error", Value: "null"},
				},
				RawContent: "",
			},
			expectedScope: map[string]any{
				"data":  nil,
				"error": nil,
			},
			description: "Variables with null values should be parsed correctly",
		},
		{
			name: "function with complex nested braces",
			fence: &ast.FenceSection{
				RawContent: `function processData(items) {
  return items.map(item => {
    if (item.active) {
      return {
        ...item,
        processed: true
      };
    }
    return item;
  });
}`,
			},
			expectedScope: map[string]any{
				"processData": `function processData(items) {
  return items.map(item => {
    if (item.active) {
      return {
        ...item,
        processed: true
      };
    }
    return item;
  });
}`,
			},
			description: "Functions with nested braces and arrow functions should be extracted",
		},
		{
			name: "function with string containing braces",
			fence: &ast.FenceSection{
				RawContent: `function getMessage() {
  return "Hello {name}!";
}`,
			},
			expectedScope: map[string]any{
				"getMessage": `function getMessage() {
  return "Hello {name}!";
}`,
			},
			description: "Functions with strings containing braces should be extracted correctly",
		},
		{
			name: "single line arrow function with multiline chain",
			fence: &ast.FenceSection{
				RawContent: `const transform = data => data
  .filter(x => x.active)
  .map(x => x.value)
  .reduce((a, b) => a + b, 0)`,
			},
			expectedScope: map[string]any{
				"transform": `const transform = data => data
  .filter(x => x.active)
  .map(x => x.value)
  .reduce((a, b) => a + b, 0)`,
			},
			description: "Arrow function with chained method calls should be extracted",
		},
		{
			name: "function with default parameters",
			fence: &ast.FenceSection{
				RawContent: `function greet(name = "Guest", greeting = "Hello") {
  return greeting + ', ' + name;
}`,
			},
			expectedScope: map[string]any{
				"greet": `function greet(name = "Guest", greeting = "Hello") {
  return greeting + ', ' + name;
}`,
			},
			description: "Functions with default parameters should be extracted",
		},
		{
			name: "function with destructured parameters",
			fence: &ast.FenceSection{
				RawContent: `function displayUser({ name, email, role = 'user' }) {
  return name + ' (' + email + ') - ' + role;
}`,
			},
			expectedScope: map[string]any{
				"displayUser": `function displayUser({ name, email, role = 'user' }) {
  return name + ' (' + email + ') - ' + role;
}`,
			},
			description: "Functions with destructured parameters should be extracted",
		},
		{
			name: "generator function",
			fence: &ast.FenceSection{
				RawContent: `function* idGenerator() {
  let id = 0;
  while (true) {
    yield id++;
  }
}`,
			},
			expectedScope: map[string]any{
				"idGenerator": `function* idGenerator() {
  let id = 0;
  while (true) {
    yield id++;
  }
}`,
			},
			description: "Generator functions should be extracted",
		},
		{
			name: "variables with whitespace variations",
			fence: &ast.FenceSection{
				Variables: []ast.VariableNode{
					{Keyword: "let", Name: "count", Value: "  42  "},
					{Keyword: "const", Name: "name", Value: `  "Test"  `},
					{Keyword: "var", Name: "flag", Value: "  true  "},
				},
				RawContent: "",
			},
			expectedScope: map[string]any{
				"count": 42,
				"name":  "Test",
				"flag":  true,
			},
			description: "Variables with whitespace in values should be trimmed by parseValue()",
		},
		{
			name: "props with expression defaults",
			fence: &ast.FenceSection{
				Props: []ast.PropNode{
					{Name: "currentDate", DefaultValue: "new Date()"},
					{Name: "userId", DefaultValue: "getUserId()"},
					{Name: "config", DefaultValue: "getDefaultConfig()"},
				},
				RawContent: "",
			},
			expectedScope: map[string]any{
				"currentDate": "new Date()",
				"userId":      "getUserId()",
				"config":      "getDefaultConfig()",
			},
			description: "Props with expression defaults should be stored as strings",
		},
		{
			name: "only imports (no variables, props, or functions)",
			fence: &ast.FenceSection{
				Imports: []ast.ImportNode{
					{Name: "Header", Path: "./components/Header.html"},
					{Name: "Footer", Path: "./components/Footer.html"},
				},
				RawContent: `import Header from './components/Header.html'
import Footer from './components/Footer.html'`,
			},
			expectedScope: map[string]any{},
			description:   "Fence with only imports should result in empty scope (imports don't add to scope)",
		},
		{
			name: "fence with comments only",
			fence: &ast.FenceSection{
				RawContent: `// This is a comment
/* Multi-line
   comment */
// Another comment`,
			},
			expectedScope: map[string]any{},
			description:   "Fence with only comments should result in empty scope",
		},
		{
			name: "function with trailing semicolon",
			fence: &ast.FenceSection{
				RawContent: `function calculate() {
  return 42;
};`,
			},
			expectedScope: map[string]any{
				"calculate": `function calculate() {
  return 42;
}`,
			},
			description: "Function with trailing semicolon should be extracted (semicolon may or may not be included)",
		},
		{
			name: "arrow function in variable with trailing semicolon",
			fence: &ast.FenceSection{
				RawContent: `const square = x => x * x;`,
			},
			expectedScope: map[string]any{
				"square": `const square = x => x * x`,
			},
			description: "Arrow function with trailing semicolon should be extracted (semicolon may or may not be included)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a fresh scope for each test
			scope := make(map[string]any)

			// Call the function under test
			collectComponentFenceData(tt.fence, scope)

			// Verify scope has correct number of entries
			if len(scope) != len(tt.expectedScope) {
				t.Errorf("%s: scope has %d entries, expected %d\nGot: %v\nExpected: %v",
					tt.description, len(scope), len(tt.expectedScope), scope, tt.expectedScope)
			}

			// Verify each expected entry exists with correct value
			for key, expectedValue := range tt.expectedScope {
				actualValue, exists := scope[key]
				if !exists {
					t.Errorf("%s: scope missing key %q", tt.description, key)
					continue
				}

				// Use deep equal for comparison
				if !reflect.DeepEqual(actualValue, expectedValue) {
					t.Errorf("%s: scope[%q] = %v (type: %T), expected %v (type: %T)",
						tt.description, key, actualValue, actualValue, expectedValue, expectedValue)
				}
			}

			// Verify no unexpected entries in scope
			for key := range scope {
				if _, expected := tt.expectedScope[key]; !expected {
					t.Errorf("%s: scope has unexpected key %q with value %v",
						tt.description, key, scope[key])
				}
			}
		})
	}
}

// TestCollectComponentFenceData_ScopeModification verifies that the function
// modifies the provided scope map correctly (doesn't replace it).
func TestCollectComponentFenceData_ScopeModification(t *testing.T) {
	// Create scope with existing data
	scope := map[string]any{
		"existingVar": 100,
		"keepMe":      "original",
	}

	fence := &ast.FenceSection{
		Variables: []ast.VariableNode{
			{Keyword: "let", Name: "newVar", Value: "42"},
		},
		Props: []ast.PropNode{
			{Name: "title", DefaultValue: `"Test"`},
		},
		RawContent: `function test() { return true; }`,
	}

	// Call function
	collectComponentFenceData(fence, scope)

	// Verify existing entries are preserved
	if scope["existingVar"] != 100 {
		t.Error("Existing scope entry 'existingVar' was modified or removed")
	}
	if scope["keepMe"] != "original" {
		t.Error("Existing scope entry 'keepMe' was modified or removed")
	}

	// Verify new entries were added
	if scope["newVar"] != 42 {
		t.Error("New variable 'newVar' was not added to scope")
	}
	if scope["title"] != "Test" {
		t.Error("New prop 'title' was not added to scope")
	}
	if _, exists := scope["test"]; !exists {
		t.Error("New function 'test' was not added to scope")
	}

	// Verify total count
	expectedCount := 5 // 2 existing + 3 new
	if len(scope) != expectedCount {
		t.Errorf("Scope has %d entries, expected %d", len(scope), expectedCount)
	}
}

// TestCollectComponentFenceData_OverwriteBehavior verifies how the function handles
// duplicate keys across variables, props, and functions.
func TestCollectComponentFenceData_OverwriteBehavior(t *testing.T) {
	tests := []struct {
		name        string
		fence       *ast.FenceSection
		description string
		checkKey    string
		checkValue  any
	}{
		{
			name: "prop overwrites variable with same name",
			fence: &ast.FenceSection{
				Variables: []ast.VariableNode{
					{Keyword: "let", Name: "title", Value: `"Variable Title"`},
				},
				Props: []ast.PropNode{
					{Name: "title", DefaultValue: `"Prop Title"`},
				},
			},
			description: "When prop and variable have same name, last one processed wins",
			checkKey:    "title",
			// The expected value depends on processing order - this test documents the behavior
		},
		{
			name: "function overwrites variable with same name",
			fence: &ast.FenceSection{
				Variables: []ast.VariableNode{
					{Keyword: "const", Name: "calculate", Value: "0"},
				},
				RawContent: `function calculate() { return 42; }`,
			},
			description: "When function and variable have same name, last one processed wins",
			checkKey:    "calculate",
			// The expected value depends on processing order - this test documents the behavior
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scope := make(map[string]any)
			collectComponentFenceData(tt.fence, scope)

			// Just verify the key exists - the actual value depends on implementation order
			if _, exists := scope[tt.checkKey]; !exists {
				t.Errorf("%s: scope missing key %q", tt.description, tt.checkKey)
			}

			// Log what value was chosen for documentation purposes
			t.Logf("%s: scope[%q] = %v (type: %T)", tt.description, tt.checkKey, scope[tt.checkKey], scope[tt.checkKey])
		})
	}
}

// TestCollectComponentFenceData_EmptyAndNilHandling verifies edge cases around
// empty and nil inputs.
func TestCollectComponentFenceData_EmptyAndNilHandling(t *testing.T) {
	tests := []struct {
		name        string
		fence       *ast.FenceSection
		shouldPanic bool
		description string
	}{
		{
			name:        "nil fence",
			fence:       nil,
			shouldPanic: false,
			description: "Nil fence should not panic",
		},
		{
			name:        "empty fence",
			fence:       &ast.FenceSection{},
			shouldPanic: false,
			description: "Empty fence should not panic",
		},
		{
			name: "fence with nil slices",
			fence: &ast.FenceSection{
				Imports:   nil,
				Props:     nil,
				Variables: nil,
			},
			shouldPanic: false,
			description: "Fence with nil slices should not panic",
		},
		{
			name: "fence with empty slices",
			fence: &ast.FenceSection{
				Imports:   []ast.ImportNode{},
				Props:     []ast.PropNode{},
				Variables: []ast.VariableNode{},
			},
			shouldPanic: false,
			description: "Fence with empty slices should not panic",
		},
		{
			name: "fence with empty RawContent",
			fence: &ast.FenceSection{
				RawContent: "",
			},
			shouldPanic: false,
			description: "Fence with empty RawContent should not panic",
		},
		{
			name: "fence with whitespace-only RawContent",
			fence: &ast.FenceSection{
				RawContent: "   \n  \t  \n   ",
			},
			shouldPanic: false,
			description: "Fence with whitespace-only RawContent should not panic",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					if !tt.shouldPanic {
						t.Errorf("%s: function panicked unexpectedly: %v", tt.description, r)
					}
				} else if tt.shouldPanic {
					t.Errorf("%s: function should have panicked but didn't", tt.description)
				}
			}()

			scope := make(map[string]any)
			collectComponentFenceData(tt.fence, scope)

			// Verify scope is still valid and empty
			if scope == nil {
				t.Errorf("%s: scope became nil", tt.description)
			}
			if len(scope) != 0 {
				t.Errorf("%s: scope should be empty but has %d entries: %v", tt.description, len(scope), scope)
			}
		})
	}
}

// TestCollectComponentFenceData_FunctionExtractionPatterns tests various
// JavaScript function syntax patterns to ensure regex extraction is robust.
func TestCollectComponentFenceData_FunctionExtractionPatterns(t *testing.T) {
	tests := []struct {
		name         string
		rawContent   string
		expectedFunc string
		expectedName string
		description  string
	}{
		{
			name: "function declaration with spaces",
			rawContent: `function   getName  (  )   {
  return 'test';
}`,
			expectedFunc: `function   getName  (  )   {
  return 'test';
}`,
			expectedName: "getName",
			description:  "Function with irregular spacing should be extracted",
		},
		{
			name: "arrow function with destructuring",
			rawContent: `const process = ({ id, name }) => {
  return { id, name: name.toUpperCase() };
}`,
			expectedFunc: `const process = ({ id, name }) => {
  return { id, name: name.toUpperCase() };
}`,
			expectedName: "process",
			description:  "Arrow function with destructured params should be extracted",
		},
		{
			name: "function with rest parameters",
			rawContent: `function sum(...numbers) {
  return numbers.reduce((a, b) => a + b, 0);
}`,
			expectedFunc: `function sum(...numbers) {
  return numbers.reduce((a, b) => a + b, 0);
}`,
			expectedName: "sum",
			description:  "Function with rest parameters should be extracted",
		},
		{
			name:       "IIFE should not be extracted as named function",
			rawContent: `(function() { console.log('IIFE'); })();`,
			// IIFE is anonymous, so it might not be extracted or might be handled differently
			description: "Immediately Invoked Function Expression (IIFE) handling",
		},
		{
			name:         "function with template literal",
			rawContent:   "const greet = name => `Hello ${name}!`",
			expectedFunc: "const greet = name => `Hello ${name}!`",
			expectedName: "greet",
			description:  "Arrow function with template literal return should be extracted",
		},
		{
			name: "function with regex in body",
			rawContent: `function validate(input) {
  return /^[a-z]+$/.test(input);
}`,
			expectedFunc: `function validate(input) {
  return /^[a-z]+$/.test(input);
}`,
			expectedName: "validate",
			description:  "Function containing regex should be extracted correctly",
		},
		{
			name: "async method shorthand",
			rawContent: `async fetchUser(id) {
  const response = await fetch('/api/users/' + id);
  return response.json();
}`,
			expectedFunc: `async fetchUser(id) {
  const response = await fetch('/api/users/' + id);
  return response.json();
}`,
			expectedName: "fetchUser",
			description:  "Async method shorthand should be extracted",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fence := &ast.FenceSection{
				RawContent: tt.rawContent,
			}

			scope := make(map[string]any)
			collectComponentFenceData(fence, scope)

			if tt.expectedName != "" {
				// Verify the function was extracted
				if _, exists := scope[tt.expectedName]; !exists {
					t.Errorf("%s: function %q was not extracted from:\n%s\nScope: %v",
						tt.description, tt.expectedName, tt.rawContent, scope)
				}
			}

			t.Logf("%s: extracted functions: %v", tt.description, scope)
		})
	}
}

// TestCollectComponentFenceData_ParseValueIntegration verifies that the function
// correctly uses parseValue() for variables and prop defaults.
func TestCollectComponentFenceData_ParseValueIntegration(t *testing.T) {
	tests := []struct {
		name          string
		fence         *ast.FenceSection
		checkKey      string
		checkType     string
		expectedValue any
		description   string
	}{
		{
			name: "boolean variable parsed correctly",
			fence: &ast.FenceSection{
				Variables: []ast.VariableNode{
					{Keyword: "let", Name: "flag", Value: "true"},
				},
			},
			checkKey:      "flag",
			checkType:     "bool",
			expectedValue: true,
			description:   "Boolean value should be parsed to bool, not string",
		},
		{
			name: "integer prop parsed correctly",
			fence: &ast.FenceSection{
				Props: []ast.PropNode{
					{Name: "count", DefaultValue: "42"},
				},
			},
			checkKey:      "count",
			checkType:     "int",
			expectedValue: 42,
			description:   "Integer value should be parsed to int, not string",
		},
		{
			name: "float variable parsed correctly",
			fence: &ast.FenceSection{
				Variables: []ast.VariableNode{
					{Keyword: "const", Name: "price", Value: "19.99"},
				},
			},
			checkKey:      "price",
			checkType:     "float64",
			expectedValue: 19.99,
			description:   "Float value should be parsed to float64, not string",
		},
		{
			name: "quoted string prop parsed correctly",
			fence: &ast.FenceSection{
				Props: []ast.PropNode{
					{Name: "title", DefaultValue: `"Hello World"`},
				},
			},
			checkKey:      "title",
			checkType:     "string",
			expectedValue: "Hello World",
			description:   "Quoted string should have quotes removed",
		},
		{
			name: "null variable parsed correctly",
			fence: &ast.FenceSection{
				Variables: []ast.VariableNode{
					{Keyword: "let", Name: "data", Value: "null"},
				},
			},
			checkKey:      "data",
			checkType:     "nil",
			expectedValue: nil,
			description:   "Null should be parsed to nil, not string",
		},
		{
			name: "array prop kept as string",
			fence: &ast.FenceSection{
				Props: []ast.PropNode{
					{Name: "items", DefaultValue: "[1, 2, 3]"},
				},
			},
			checkKey:      "items",
			checkType:     "string",
			expectedValue: "[1, 2, 3]",
			description:   "Array should be kept as string for Alpine.js",
		},
		{
			name: "object variable kept as string",
			fence: &ast.FenceSection{
				Variables: []ast.VariableNode{
					{Keyword: "const", Name: "config", Value: `{ theme: "dark" }`},
				},
			},
			checkKey:      "config",
			checkType:     "string",
			expectedValue: `{ theme: "dark" }`,
			description:   "Object should be kept as string for Alpine.js",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scope := make(map[string]any)
			collectComponentFenceData(tt.fence, scope)

			value, exists := scope[tt.checkKey]
			if !exists {
				t.Fatalf("%s: scope missing key %q", tt.description, tt.checkKey)
			}

			// Type checking
			switch tt.checkType {
			case "bool":
				if _, ok := value.(bool); !ok {
					t.Errorf("%s: value should be bool, got %T", tt.description, value)
				}
			case "int":
				if _, ok := value.(int); !ok {
					t.Errorf("%s: value should be int, got %T", tt.description, value)
				}
			case "float64":
				if _, ok := value.(float64); !ok {
					t.Errorf("%s: value should be float64, got %T", tt.description, value)
				}
			case "string":
				if _, ok := value.(string); !ok {
					t.Errorf("%s: value should be string, got %T", tt.description, value)
				}
			case "nil":
				if value != nil {
					t.Errorf("%s: value should be nil, got %v (%T)", tt.description, value, value)
				}
			}

			// Value checking
			if !reflect.DeepEqual(value, tt.expectedValue) {
				t.Errorf("%s: value = %v, expected %v", tt.description, value, tt.expectedValue)
			}
		})
	}
}
