package utils

import (
	"fmt"
	"time"
)

const fabricTime = "2006-01-02 15:04:05 -0700"

func NormalizeLease(input string) (string, error) {
	if input == "" {
		return "", nil
	}
	if t, err := time.Parse(time.RFC3339, input); err == nil {
		return t.Format(fabricTime), nil
	}
	if _, err := time.Parse(fabricTime, input); err == nil {
		return input, nil
	}
	return "", fmt.Errorf("invalid lease_end_time format: %s", input)
}
