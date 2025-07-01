package main

import "fmt"

func validateComponentName(name string) error {
	if len(name) == 0 {
		return fmt.Errorf("name cannot be empty")
	}

	if len(name) > 100 {
		return fmt.Errorf("name too long: maximum 100 characters")
	}

	for _, char := range name {
		if !isValidChar(char) {
			return fmt.Errorf("invalid character '%c' in name", char)
		}
	}

	return nil
}

func isValidChar(r rune) bool {
	return (r >= 'a' && r <= 'z') ||
		(r >= 'A' && r <= 'Z') ||
		(r >= '0' && r <= '9') ||
		r == '-' || r == '_' || r == '.'
}
