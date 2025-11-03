package utils

import (
	"fmt"
	"strconv"
)

// ParseInt converts a string token into an integer and wraps parsing errors.
func ParseInt(value string) (int, error) {
	n, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("parse int: %w", err)
	}
	return n, nil
}
