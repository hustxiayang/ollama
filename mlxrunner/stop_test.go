package mlxrunner

import "testing"

func TestFindStop(t *testing.T) {
	cases := []struct {
		name  string
		text  string
		stops []string
		want  int
		found bool
	}{
		{"no stops", "hello", nil, -1, false},
		{"empty stops ignored", "hello", []string{""}, -1, false},
		{"no match", "hello world", []string{"STOP"}, -1, false},
		{"single match", "hello STOP world", []string{"STOP"}, 6, true},
		{"earliest of many", "aXbY", []string{"Y", "X"}, 1, true},
		{"earliest across stops", "the end. done", []string{"done", "end"}, 4, true},
		{"match at start", "STOPnow", []string{"STOP"}, 0, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, found := findStop(tc.text, tc.stops)
			if got != tc.want || found != tc.found {
				t.Fatalf("findStop(%q, %v) = (%d, %t), want (%d, %t)", tc.text, tc.stops, got, found, tc.want, tc.found)
			}
		})
	}
}

func TestPartialStopSuffix(t *testing.T) {
	cases := []struct {
		name  string
		text  string
		stops []string
		want  int
	}{
		{"no partial", "hello", []string{"STOP"}, 0},
		{"one char prefix", "fooS", []string{"STOP"}, 1},
		{"multi char prefix", "fooST", []string{"STOP"}, 2},
		{"full stop not counted here", "STOP", []string{"STOP"}, 0},
		{"prefix longer than remaining stop", "abcSTO", []string{"STOP"}, 3},
		{"longest across stops", "xAB", []string{"ABC", "B"}, 2},
		{"text shorter than stop", "S", []string{"STOP"}, 1},
		{"empty stop ignored", "fooS", []string{""}, 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := partialStopSuffix(tc.text, tc.stops); got != tc.want {
				t.Fatalf("partialStopSuffix(%q, %v) = %d, want %d", tc.text, tc.stops, got, tc.want)
			}
		})
	}
}

func feedApplyStops(stops []string, chunks ...string) (emitted string, stopped bool, tail string) {
	d := detokenizer{stops: stops}
	for _, c := range chunks {
		emit, stop := d.applyStops(c)
		emitted += emit
		if stop {
			return emitted, true, ""
		}
	}
	if resp, ok := d.flush(); ok {
		tail = resp.Content
	}
	return emitted, false, tail
}

func TestApplyStops(t *testing.T) {
	t.Run("whole in one chunk", func(t *testing.T) {
		emitted, stopped, _ := feedApplyStops([]string{"STOP"}, "hello STOP world")
		if !stopped || emitted != "hello " {
			t.Fatalf("got (%q, stop=%t), want (%q, true)", emitted, stopped, "hello ")
		}
	})

	t.Run("split across chunks", func(t *testing.T) {
		emitted, stopped, _ := feedApplyStops([]string{"STOP"}, "hello ST", "OP world")
		if !stopped || emitted != "hello " {
			t.Fatalf("got (%q, stop=%t), want (%q, true)", emitted, stopped, "hello ")
		}
	})

	t.Run("split one byte at a time", func(t *testing.T) {
		emitted, stopped, _ := feedApplyStops([]string{"STOP"}, "a", "S", "T", "O", "P", "b")
		if !stopped || emitted != "a" {
			t.Fatalf("got (%q, stop=%t), want (%q, true)", emitted, stopped, "a")
		}
	})

	t.Run("stop at start", func(t *testing.T) {
		emitted, stopped, _ := feedApplyStops([]string{"STOP"}, "STOP after")
		if !stopped || emitted != "" {
			t.Fatalf("got (%q, stop=%t), want (%q, true)", emitted, stopped, "")
		}
	})

	t.Run("partial that never completes is flushed", func(t *testing.T) {
		emitted, stopped, tail := feedApplyStops([]string{"STOP"}, "done ST")
		if stopped || emitted != "done " || tail != "ST" {
			t.Fatalf("got (%q, stop=%t, tail=%q), want (%q, false, %q)", emitted, stopped, tail, "done ", "ST")
		}
	})

	t.Run("no stops emits everything unchanged", func(t *testing.T) {
		emitted, stopped, tail := feedApplyStops(nil, "hello STOP world")
		if stopped || emitted != "hello STOP world" || tail != "" {
			t.Fatalf("got (%q, stop=%t, tail=%q), want passthrough", emitted, stopped, tail)
		}
	})
}
