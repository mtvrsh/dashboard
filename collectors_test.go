package main

import (
	"os/user"
	"reflect"
	"testing"
	"time"
)

func Test_prettyPrintDuration(t *testing.T) {
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

func Test_getMountpointUsers(t *testing.T) {
	currentUser, _ := user.Current()

	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		paths   []string
		want    map[string][]string
		wantErr bool
	}{
		{
			// this case will fail if other user than user executing this test will
			// access those 2 mountpoints
			name:  "2 tmpfs mounts",
			paths: []string{"/tmp", "/dev/shm"},
			want: map[string][]string{
				"/dev/shm": {currentUser.Username},
				"/tmp":     {currentUser.Username},
			},
			wantErr: false,
		},
		{
			name:    "path does not exist",
			paths:   []string{"/aaaaaaaaaaaaaaaaaaaaaaa"},
			want:    map[string][]string{},
			wantErr: true,
		},
		{
			name:    "no paths",
			paths:   []string{},
			want:    map[string][]string{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := getMountpointsUsers([]string{defaultFuserCmd}, tt.paths)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("getMountpointUsers() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("getMountpointUsers() succeeded unexpectedly")
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getMountpointUsers() = %v, want %v", got, tt.want)
			}
		})
	}
}
