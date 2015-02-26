package util

import (
	"fmt"
)

// max takes two integers and returns the larger of the two.
func max(a int, b int) int {
	if a >= b {
		return a
	} else {
		return b
	}
}

// MakeSequenceValidator returns a function that takes a string and ensures that
// each one matches the current string in the inputs array in order. If the
// string matches, the function returns true, otherwise it returns false and an
// error string. If a string is provided past the strings in inputs the function
// returns false and an error string to indicate failure as well.
func MakeSequenceValidator(inputs []string) func(string) (valid bool, errorText string) {
	i := 0
	return func(b string) (valid bool, errorText string) {
		if i >= len(inputs) {
			valid = false
			errorText = "Received more input than expected"
		} else if inputs[i] != b {
			valid = false
			errorText = fmt.Sprintf("Wanted %#v, got %#v", inputs[i], b)
		} else {
			valid = true
		}
		i = i + 1
		return
	}

}
