package fetch

import "testing"

func TestPartialMatch(t *testing.T) {
	tests := []struct {
		givenA string
		givenB string
		want   bool
	}{
		{
			"The Dude Dies",
			"The Dude",
			true,
		},
		{
			"Known As The Dude Dies",
			"The Dude",
			true,
		},
		{
			"A man named sue",
			"sue",
			true,
		},
		{
			"This is The dude's story",
			"the dude",
			true,
		},
		{
			"third baseman",
			"man",
			false,
		},
		{
			"something with a number 12345",
			"something with a similar number 12345.123",
			false,
		},
		{
			"no way does",
			"it does",
			false,
		},
	}

	for _, test := range tests {
		if got := partialMatch(test.givenA, test.givenB); got != test.want {
			t.Errorf("partialMatch(%q,%q) returned %v; expected %v", test.givenA, test.givenB, got, test.want)
		}
		if got := partialMatch(test.givenB, test.givenA); got != test.want {
			t.Errorf("partialMatch(%q,%q) returned %v; expected %v", test.givenB, test.givenA, got, test.want)
		}

	}
}
