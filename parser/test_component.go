package parser

import (
	"log"
)

// TestComponentParser is a simple function to test the component parser
func TestComponentParser() {
	// Test cases
	testCases := []string{
		"<Button />",
		"<Button label=\"Click me\" />",
		"<Button label=\"Click me\" onClick={handleClick} />",
		"<AdminPanel user={currentUser} />",
		"<UserProfile user={currentUser} />",
	}
	
	for _, tc := range testCases {
		log.Printf("=== Testing Component Parser on: %s ===", tc)
		result := ComponentParser()(tc)
		if result.Successful {
			log.Printf("SUCCESS: Parsed component: %+v", result.Value)
		} else {
			log.Printf("FAILED: %s", result.Error)
		}
		log.Printf("Remaining: %s", result.Remaining)
		log.Printf("=== End Test ===\n")
	}
}
