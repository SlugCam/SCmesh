package wiflyparsers

import (
	"testing"
)

// TestMax tests the max function with several cases
func TestMax(t *testing.T) {
	cases := []struct {
		in1, in2, want int
	}{
		{1, 1, 1},
		{5, 1, 5},
		{0, 7, 7},
		{-5, 5, 5},
		{2, -7, 2},
	}
	for _, c := range cases {
		got := max(c.in1, c.in2)
		if got != c.want {
			t.Errorf("max(%d, %d) == %d, want %d", c.in1, c.in2, got, c.want)
		}
	}
}

// TestCouldBe tests the couldBe function with several cases
func TestCouldBe(t *testing.T) {
	cases := []struct {
		in1, in2 string
		want     bool
	}{
		{1, 1, 1},
		{5, 1, 5},
		{0, 7, 7},
		{-5, 5, 5},
		{2, -7, 2},
	}
	for _, c := range cases {
		got := max(c.in1, c.in2)
		if got != c.want {
			t.Errorf("max(%d, %d) == %d, want %d", c.in1, c.in2, got, c.want)
		}
	}
}
