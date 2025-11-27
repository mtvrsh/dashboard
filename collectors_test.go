package main

import (
	"testing"
	"time"
)

func TestPrettyPrintDuration(t *testing.T) {
	cases := []struct {
		duration time.Duration
		expected string
	}{
		{0, "0m"},
		{-time.Second, "0m"},
		{time.Second * 59, "0m"},
		{-time.Hour, "-1h"},
		{-(time.Hour*1 + time.Minute*30 + time.Second), "-1h 30m"},
		{time.Minute * 30, "30m"},
		{time.Minute*1440 + time.Minute*1, "1d 1m"},
		{time.Hour*1 + time.Minute*30, "1h 30m"},
		{time.Hour * 26, "1d 2h"},
		{time.Hour*1000 + time.Microsecond, "41d 16h"},
		{time.Hour * 72, "3d"},
		{time.Hour*50 + time.Minute*50 + time.Second*50, "2d 2h 50m"},
	}

	for _, c := range cases {
		got := prettyPrintDuration(c.duration)
		if got != c.expected {
			t.Errorf("prettyPrintDuration(%v) expected: %q, got: %q", c.duration, c.expected, got)
		}
	}
}
