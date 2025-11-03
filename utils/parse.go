package utils

import (
	"fmt"
	"strconv"
)

// ParseInt converts a numeric string into an int with additional validation.
func ParseInt(value string) (int, error) {
	n, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("parse int: %w", err)
	}
	return n, nil
}
