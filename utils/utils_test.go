package utils

import "testing"

func TestRemoveWhitespace(t *testing.T) {
	raw := "\t\n  this 	\n contains    too \n\n much  space.   \n\t"
	got := RemoveWhitespace(raw)
	expect := "this contains too much space."
	if got != expect {
		t.Fatalf("remove whitespace failed: expected [%s] got [%s]", expect, got)
	}
}
