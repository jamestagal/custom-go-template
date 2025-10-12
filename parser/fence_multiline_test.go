package parser

import (
	"testing"
)

func TestParseFenceContent_MultiLineArray(t *testing.T) {
	input := `prop links = [
  { label: "Home", url: "/" },
  { label: "About", url: "/about" }
]`

	fence := parseFenceContent(input)

	if len(fence.Props) != 1 {
		t.Errorf("Expected 1 prop, got %d", len(fence.Props))
		return
	}

	if fence.Props[0].Name != "links" {
		t.Errorf("Expected prop name 'links', got '%s'", fence.Props[0].Name)
	}

	expectedValue := `[
  { label: "Home", url: "/" },
  { label: "About", url: "/about" }
]`

	if fence.Props[0].DefaultValue != expectedValue {
		t.Errorf("Expected full array value, got:\n%s\n\nExpected:\n%s",
			fence.Props[0].DefaultValue, expectedValue)
	}
}

func TestParseFenceContent_MultiLineObject(t *testing.T) {
	input := `prop config = {
  theme: "dark",
  options: {
    nested: true
  }
}`

	fence := parseFenceContent(input)

	if len(fence.Props) != 1 {
		t.Errorf("Expected 1 prop, got %d", len(fence.Props))
		return
	}

	if fence.Props[0].Name != "config" {
		t.Errorf("Expected prop name 'config', got '%s'", fence.Props[0].Name)
	}

	expectedValue := `{
  theme: "dark",
  options: {
    nested: true
  }
}`

	if fence.Props[0].DefaultValue != expectedValue {
		t.Errorf("Expected full object value, got:\n%s\n\nExpected:\n%s",
			fence.Props[0].DefaultValue, expectedValue)
	}
}

func TestParseFenceContent_FooterLinks(t *testing.T) {
	// Real-world test case from Footer.html
	input := `prop links = [
  { label: "Home", url: "/" },
  { label: "About", url: "/about" },
  { label: "Products", url: "/products" },
  { label: "Contact", url: "/contact" },
  { label: "Terms of Service", url: "/terms" },
  { label: "Privacy Policy", url: "/privacy" }
]`

	fence := parseFenceContent(input)

	if len(fence.Props) != 1 {
		t.Errorf("Expected 1 prop, got %d", len(fence.Props))
		return
	}

	if fence.Props[0].Name != "links" {
		t.Errorf("Expected prop name 'links', got '%s'", fence.Props[0].Name)
	}

	// Check that the value is the complete array, not just "["
	if fence.Props[0].DefaultValue == "[" {
		t.Error("Bug detected: prop value was truncated to first line only")
	}

	// The value should contain all the links
	value := fence.Props[0].DefaultValue
	if !contains(value, "Home") || !contains(value, "Privacy Policy") {
		t.Errorf("Multi-line array incomplete: %s", value)
	}
}

func TestParseFenceContent_MixedProps(t *testing.T) {
	input := `prop companyName = "Custom Template Co."
prop year = new Date().getFullYear()
prop links = [
  { label: "Home", url: "/" },
  { label: "About", url: "/about" }
]
prop count = 42`

	fence := parseFenceContent(input)

	if len(fence.Props) != 4 {
		t.Errorf("Expected 4 props, got %d", len(fence.Props))
		return
	}

	// Check each prop
	tests := []struct {
		name  string
		value string
	}{
		{"companyName", `"Custom Template Co."`},
		{"year", "new Date().getFullYear()"},
		{"links", "["},  // Should start with [
		{"count", "42"},
	}

	for i, tt := range tests {
		if i >= len(fence.Props) {
			t.Errorf("Missing prop at index %d", i)
			continue
		}

		if fence.Props[i].Name != tt.name {
			t.Errorf("Prop %d: expected name '%s', got '%s'", i, tt.name, fence.Props[i].Name)
		}

		// For the links prop, verify it's multi-line
		if tt.name == "links" {
			if fence.Props[i].DefaultValue == "[" {
				t.Errorf("Bug: links prop was truncated to single line")
			}
			if !contains(fence.Props[i].DefaultValue, "Home") {
				t.Errorf("links prop incomplete: %s", fence.Props[i].DefaultValue)
			}
		} else {
			// For single-line props, check exact value
			if fence.Props[i].DefaultValue != tt.value {
				t.Errorf("Prop %s: expected value '%s', got '%s'",
					tt.name, tt.value, fence.Props[i].DefaultValue)
			}
		}
	}
}

func TestParseFenceContent_FunctionExpression(t *testing.T) {
	input := `prop year = new Date().getFullYear()`

	fence := parseFenceContent(input)

	if len(fence.Props) != 1 {
		t.Errorf("Expected 1 prop, got %d", len(fence.Props))
		return
	}

	expected := "new Date().getFullYear()"
	if fence.Props[0].DefaultValue != expected {
		t.Errorf("Expected function expression '%s', got '%s'",
			expected, fence.Props[0].DefaultValue)
	}
}

func TestParseFenceContent_NestedArraysAndObjects(t *testing.T) {
	input := `prop data = {
  items: [
    { id: 1, tags: ["a", "b"] },
    { id: 2, tags: ["c", "d"] }
  ],
  meta: { count: 2 }
}`

	fence := parseFenceContent(input)

	if len(fence.Props) != 1 {
		t.Errorf("Expected 1 prop, got %d", len(fence.Props))
		return
	}

	value := fence.Props[0].DefaultValue

	// Verify it contains all the nested content
	requiredStrings := []string{"items", "tags", "meta", "count"}
	for _, str := range requiredStrings {
		if !contains(value, str) {
			t.Errorf("Multi-line prop missing '%s': %s", str, value)
		}
	}
}

func TestParseFenceContent_Variables(t *testing.T) {
	input := `let items = [
  { id: 1 },
  { id: 2 }
]
const config = {
  enabled: true
}`

	fence := parseFenceContent(input)

	if len(fence.Variables) != 2 {
		t.Errorf("Expected 2 variables, got %d", len(fence.Variables))
		return
	}

	// Check items variable
	if fence.Variables[0].Name != "items" {
		t.Errorf("Expected variable name 'items', got '%s'", fence.Variables[0].Name)
	}
	if fence.Variables[0].Keyword != "let" {
		t.Errorf("Expected keyword 'let', got '%s'", fence.Variables[0].Keyword)
	}
	if !contains(fence.Variables[0].Value, "id") {
		t.Errorf("Variable value incomplete: %s", fence.Variables[0].Value)
	}

	// Check config variable
	if fence.Variables[1].Name != "config" {
		t.Errorf("Expected variable name 'config', got '%s'", fence.Variables[1].Name)
	}
	if fence.Variables[1].Keyword != "const" {
		t.Errorf("Expected keyword 'const', got '%s'", fence.Variables[1].Keyword)
	}
}

func TestParseFenceContent_NavItemsConditional(t *testing.T) {
	// Real-world test case from Header.html - complex ternary with arrays
	input := `const navItems = isLoggedIn ? [
  { label: "Home", url: "/" },
  { label: "Products", url: "/products" },
  { label: "Categories", url: "/categories" },
  user.role === "admin" ? { label: "Admin Panel", url: "/admin" } : null,
  { label: "Account", url: "/account" }
].filter(Boolean) : [
  { label: "Home", url: "/" },
  { label: "Products", url: "/products" },
  { label: "Login", url: "/login" },
  { label: "Register", url: "/register" }
]`

	fence := parseFenceContent(input)

	if len(fence.Variables) != 1 {
		t.Errorf("Expected 1 variable, got %d", len(fence.Variables))
		return
	}

	if fence.Variables[0].Name != "navItems" {
		t.Errorf("Expected variable name 'navItems', got '%s'", fence.Variables[0].Name)
	}

	if fence.Variables[0].Keyword != "const" {
		t.Errorf("Expected keyword 'const', got '%s'", fence.Variables[0].Keyword)
	}

	value := fence.Variables[0].Value

	// Check that it was NOT truncated to just "isLoggedIn ? ["
	if value == "isLoggedIn ? [" {
		t.Fatal("Bug detected: value was truncated to first line only!")
	}

	// Check that it contains key parts of the conditional
	requiredStrings := []string{
		"isLoggedIn",
		"filter(Boolean)",
		"Login",
		"Register",
		"Admin Panel",
		"Categories",
	}

	for _, str := range requiredStrings {
		if !contains(value, str) {
			t.Errorf("Missing '%s' in navItems value", str)
		}
	}

	// Should be multi-line and substantial
	if len(value) < 200 {
		t.Errorf("Value seems too short for the full ternary, only %d chars: %s", len(value), value)
	}

	t.Logf("Successfully parsed navItems with %d characters", len(value))
}

// ===== STORE PARSING TESTS =====

func TestParseFenceContent_SingleLineStore(t *testing.T) {
	input := `store counter = { count: 0, increment() { this.count++; } }`

	fence := parseFenceContent(input)

	if len(fence.Stores) != 1 {
		t.Errorf("Expected 1 store, got %d", len(fence.Stores))
		return
	}

	storeDef, exists := fence.Stores["counter"]
	if !exists {
		t.Errorf("Store 'counter' not found in stores map")
		return
	}

	expectedValue := `{ count: 0, increment() { this.count++; } }`
	if storeDef != expectedValue {
		t.Errorf("Expected store definition:\n%s\n\nGot:\n%s", expectedValue, storeDef)
	}
}

func TestParseFenceContent_MultiLineStore(t *testing.T) {
	input := `store auth = {
  isLoggedIn: false,
  user: null,
  login() {
    this.isLoggedIn = true;
  },
  logout() {
    this.isLoggedIn = false;
  }
}`

	fence := parseFenceContent(input)

	if len(fence.Stores) != 1 {
		t.Errorf("Expected 1 store, got %d", len(fence.Stores))
		return
	}

	storeDef, exists := fence.Stores["auth"]
	if !exists {
		t.Errorf("Store 'auth' not found in stores map")
		return
	}

	// Check that it's the complete multi-line definition
	expectedValue := `{
  isLoggedIn: false,
  user: null,
  login() {
    this.isLoggedIn = true;
  },
  logout() {
    this.isLoggedIn = false;
  }
}`

	if storeDef != expectedValue {
		t.Errorf("Expected store definition:\n%s\n\nGot:\n%s", expectedValue, storeDef)
	}

	// Verify key components
	requiredStrings := []string{"isLoggedIn", "user", "login()", "logout()"}
	for _, str := range requiredStrings {
		if !contains(storeDef, str) {
			t.Errorf("Store definition missing '%s'", str)
		}
	}
}

func TestParseFenceContent_MultipleStores(t *testing.T) {
	input := `store auth = {
  isLoggedIn: false,
  user: null
}

store cart = {
  items: [],
  total: 0,
  addItem(item) {
    this.items.push(item);
    this.total += item.price;
  }
}

store theme = {
  mode: "light",
  toggle() {
    this.mode = this.mode === "light" ? "dark" : "light";
  }
}`

	fence := parseFenceContent(input)

	if len(fence.Stores) != 3 {
		t.Errorf("Expected 3 stores, got %d", len(fence.Stores))
		return
	}

	// Check auth store
	authDef, exists := fence.Stores["auth"]
	if !exists {
		t.Errorf("Store 'auth' not found")
	} else {
		if !contains(authDef, "isLoggedIn") || !contains(authDef, "user") {
			t.Errorf("auth store incomplete: %s", authDef)
		}
	}

	// Check cart store
	cartDef, exists := fence.Stores["cart"]
	if !exists {
		t.Errorf("Store 'cart' not found")
	} else {
		if !contains(cartDef, "items") || !contains(cartDef, "addItem") {
			t.Errorf("cart store incomplete: %s", cartDef)
		}
	}

	// Check theme store
	themeDef, exists := fence.Stores["theme"]
	if !exists {
		t.Errorf("Store 'theme' not found")
	} else {
		if !contains(themeDef, "mode") || !contains(themeDef, "toggle") {
			t.Errorf("theme store incomplete: %s", themeDef)
		}
	}
}

func TestParseFenceContent_StoreWithNestedObjects(t *testing.T) {
	input := `store user = {
  profile: {
    name: "John",
    email: "john@example.com",
    settings: {
      theme: "dark",
      notifications: true
    }
  },
  updateProfile(data) {
    this.profile = { ...this.profile, ...data };
  }
}`

	fence := parseFenceContent(input)

	if len(fence.Stores) != 1 {
		t.Errorf("Expected 1 store, got %d", len(fence.Stores))
		return
	}

	storeDef, exists := fence.Stores["user"]
	if !exists {
		t.Errorf("Store 'user' not found")
		return
	}

	// Verify all nested content is present
	requiredStrings := []string{
		"profile",
		"name",
		"email",
		"settings",
		"theme",
		"notifications",
		"updateProfile",
	}

	for _, str := range requiredStrings {
		if !contains(storeDef, str) {
			t.Errorf("Store definition missing '%s': %s", str, storeDef)
		}
	}
}

func TestParseFenceContent_MixedPropsVarsAndStores(t *testing.T) {
	input := `import Header from "./components/Header.html"

prop title = "My App"
prop count = 42

let localVar = "test"

store auth = {
  isLoggedIn: false,
  user: null
}

store cart = {
  items: [],
  total: 0
}

const config = {
  debug: true
}`

	fence := parseFenceContent(input)

	// Verify imports
	if len(fence.Imports) != 1 {
		t.Errorf("Expected 1 import, got %d", len(fence.Imports))
	} else if fence.Imports[0].Name != "Header" {
		t.Errorf("Expected import 'Header', got '%s'", fence.Imports[0].Name)
	}

	// Verify props
	if len(fence.Props) != 2 {
		t.Errorf("Expected 2 props, got %d", len(fence.Props))
	} else {
		if fence.Props[0].Name != "title" {
			t.Errorf("Expected prop 'title', got '%s'", fence.Props[0].Name)
		}
		if fence.Props[1].Name != "count" {
			t.Errorf("Expected prop 'count', got '%s'", fence.Props[1].Name)
		}
	}

	// Verify variables
	if len(fence.Variables) != 2 {
		t.Errorf("Expected 2 variables, got %d", len(fence.Variables))
	} else {
		if fence.Variables[0].Name != "localVar" {
			t.Errorf("Expected variable 'localVar', got '%s'", fence.Variables[0].Name)
		}
		if fence.Variables[1].Name != "config" {
			t.Errorf("Expected variable 'config', got '%s'", fence.Variables[1].Name)
		}
	}

	// Verify stores
	if len(fence.Stores) != 2 {
		t.Errorf("Expected 2 stores, got %d", len(fence.Stores))
		return
	}

	if _, exists := fence.Stores["auth"]; !exists {
		t.Errorf("Store 'auth' not found")
	}

	if _, exists := fence.Stores["cart"]; !exists {
		t.Errorf("Store 'cart' not found")
	}
}

func TestParseFenceContent_StoreWithArrays(t *testing.T) {
	input := `store notifications = {
  items: [
    { id: 1, message: "Welcome", read: false },
    { id: 2, message: "New message", read: false }
  ],
  unreadCount() {
    return this.items.filter(n => !n.read).length;
  },
  markAsRead(id) {
    const notification = this.items.find(n => n.id === id);
    if (notification) {
      notification.read = true;
    }
  }
}`

	fence := parseFenceContent(input)

	if len(fence.Stores) != 1 {
		t.Errorf("Expected 1 store, got %d", len(fence.Stores))
		return
	}

	storeDef, exists := fence.Stores["notifications"]
	if !exists {
		t.Errorf("Store 'notifications' not found")
		return
	}

	// Verify array and methods are present
	requiredStrings := []string{
		"items",
		"message",
		"Welcome",
		"unreadCount",
		"markAsRead",
		"filter",
		"find",
	}

	for _, str := range requiredStrings {
		if !contains(storeDef, str) {
			t.Errorf("Store definition missing '%s'", str)
		}
	}
}

func TestParseFenceContent_StoreWithComplexMethods(t *testing.T) {
	input := `store dataStore = {
  data: [],
  async fetchData() {
    const response = await fetch('/api/data');
    this.data = await response.json();
  },
  filterBy(predicate) {
    return this.data.filter(predicate);
  },
  sortBy(key, order = 'asc') {
    return [...this.data].sort((a, b) => {
      if (order === 'asc') {
        return a[key] > b[key] ? 1 : -1;
      }
      return a[key] < b[key] ? 1 : -1;
    });
  }
}`

	fence := parseFenceContent(input)

	if len(fence.Stores) != 1 {
		t.Errorf("Expected 1 store, got %d", len(fence.Stores))
		return
	}

	storeDef, exists := fence.Stores["dataStore"]
	if !exists {
		t.Errorf("Store 'dataStore' not found")
		return
	}

	// Verify complex method content
	requiredStrings := []string{
		"async fetchData",
		"await fetch",
		"filterBy",
		"predicate",
		"sortBy",
		"order = 'asc'",
		".sort(",
	}

	for _, str := range requiredStrings {
		if !contains(storeDef, str) {
			t.Errorf("Store definition missing '%s': %s", str, storeDef)
		}
	}
}

func TestParseFenceContent_StoreNameValidation(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		storeName   string
		shouldExist bool
	}{
		{
			name:        "simple name",
			input:       `store auth = { loggedIn: false }`,
			storeName:   "auth",
			shouldExist: true,
		},
		{
			name:        "underscore in name",
			input:       `store user_profile = { name: "" }`,
			storeName:   "user_profile",
			shouldExist: true,
		},
		{
			name:        "camelCase name",
			input:       `store userAuth = { token: null }`,
			storeName:   "userAuth",
			shouldExist: true,
		},
		{
			name:        "PascalCase name",
			input:       `store UserAuth = { token: null }`,
			storeName:   "UserAuth",
			shouldExist: true,
		},
		{
			name:        "name with numbers",
			input:       `store auth2 = { loggedIn: false }`,
			storeName:   "auth2",
			shouldExist: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fence := parseFenceContent(tt.input)

			if tt.shouldExist {
				if len(fence.Stores) != 1 {
					t.Errorf("Expected 1 store, got %d", len(fence.Stores))
					return
				}

				if _, exists := fence.Stores[tt.storeName]; !exists {
					t.Errorf("Expected store '%s' not found", tt.storeName)
				}
			}
		})
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 &&
		(s == substr || len(s) > len(substr) &&
		(s[0:len(substr)] == substr ||
		 s[len(s)-len(substr):] == substr ||
		 containsInner(s, substr)))
}

func containsInner(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
