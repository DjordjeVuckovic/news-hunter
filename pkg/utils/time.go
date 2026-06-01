package utils

import (
	"fmt"
	"time"
)

func ParseTimeRequired(v string) (time.Time, error) {
	if v == "" {
		return time.Time{}, fmt.Errorf("empty time")
	}
	return time.Parse(time.RFC3339, v)
}

func ParseTimeOptional(v string) (time.Time, error) {
	if v == "" {
		return time.Time{}, nil
	}
	return time.Parse(time.RFC3339, v)
}
