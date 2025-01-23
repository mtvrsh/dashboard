package main

import (
	"testing"
	"time"
)

func TestPrettyPrintDuration(t *testing.T) {
	testData := []struct {
		duration time.Duration
		expected string
	}{
		{time.Hour*1000 + time.Microsecond, "41d 16h"},
		{time.Hour * 72, "3d"},
		{time.Hour * 90, "3d 18h"},
		{time.Hour*50 + time.Minute*50 + time.Second*50, "2d 2h 50m"},
		{time.Second, ""},
		{-time.Hour, "-1h"},
	}

	for _, tt := range testData {
		got := prettyPrintDuration(tt.duration)
		if got != tt.expected {
			t.Errorf("durationToString(%v) got: %q, expected: %q", tt.duration, got, tt.expected)
		}
	}
}
