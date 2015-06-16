package bark

import "testing"

func TestTwitterize(t *testing.T) {
	tests := []struct {
		given string
		want  string
	}{
		{
			"howdy",
			"howdy ",
		},
		{
			"FoxNews - Chicago defeats the Tampa Bay Lightning, 2-0, in Game 6 of the Stanley Cup Final to win the franchise\u2019s third NHL title in six seasons.",
			"FoxNews - Chicago defeats the Tampa Bay Lightning, 2-0, in Game 6 of the Stanley Cup Final to win the franchise’s... ",
		},
	}

	for i, test := range tests {
		got := twitterize(test.given)
		if got != test.want {
			t.Errorf("%d:twitterize(%s) =\n%q\n expected\n%q", i, test.given, got, test.want)
		}
	}
}
