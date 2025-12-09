package github

import "fmt"

// parseInt is a helper function to parse string to int.
func parseInt(s string) int {
	var result int
	_, _ = fmt.Sscanf(s, "%d", &result)
	return result
}
