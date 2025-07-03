package compute

import (
	"fmt"
	"regexp"
)

type Config struct {
	ComponentName string
	Verbose       bool
}

var componentNameRegex = regexp.MustCompile(`^[a-zA-Z0-9._-]{1,100}$`)

func validateComponentName(name string) error {
	if name == "" {
		return fmt.Errorf("name cannot be empty")
	}
	if !componentNameRegex.MatchString(name) {
		return fmt.Errorf("invalid name: must be 1-100 chars, alphanumeric with .-_ only")
	}
	return nil
}
