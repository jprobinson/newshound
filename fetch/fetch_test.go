package fetch

import (
	"reflect"
	"testing"
)

func TestFindHREFs(t *testing.T) {
	tests := []struct {
		given string
		want  []string
	}{
		{
			`<html>
				<body>
					<div>
					random text
					</div>
					<div>
						<a href="http://newshound.test.link.com">a link!</a>
					</div>
					<ol>
						<li>
						<a href="http://newshound.test.link.com/123">another link!</a>
						</li>
					</ol>
				</body>
			</html>`,
			[]string{
				"http://newshound.test.link.com",
				"http://newshound.test.link.com/123",
			},
		},
		{
			`A plain text email with a simple link <a href="http://newshound.test.link.com">a link!</a>`,
			[]string{
				"http://newshound.test.link.com",
			},
		},
		{
			`A plain text email with a simple link http://newshound.test.link.com/123?abc `,
			[]string{
				"http://newshound.test.link.com/123?abc",
			},
		},
	}

	for _, test := range tests {
		got := findHREFs([]byte(test.given))
		if !reflect.DeepEqual(got, test.want) {
			t.Errorf("findHREFs() got:%s want:%s", got, test.want)
		}
	}
}

func TestScrubBody(t *testing.T) {
	tests := []struct {
		given string
		email string
		want  string
	}{
		{
			`unsubscribe Unsubscribe blah@gmail.com blah`,
			"blah@gmail.com",
			"   ",
		},
	}

	for _, test := range tests {
		got := scrubBody([]byte(test.given), test.email)
		if got != test.want {
			t.Errorf("scrubBody(%s, %s) got:%s want:%s", test.given, test.email, got, test.want)
		}
	}
}

func TestFindSender(t *testing.T) {
	tests := []struct {
		given string
		want  string
	}{
		{
			"USATODAY.com <newsletters@e.usatoday.com>",
			"USATODAY.com",
		},
		{
			"FT Exclusive <FT@emailbriefings.ft.com>",
			"FT",
		},
		{
			`"Los Angeles Times" <news@e.latimes.com>`,
			"Los Angeles Times",
		},
		{
			"LA Times",
			"Los Angeles Times",
		},
		{
			"\"NYTimes.com\"",
			"NYTimes.com",
		},
		{
			"The Washington Post",
			"The Washington Post",
		},
		{
			"\"POLITICO Breaking News\" \u003cbreakingnews@politico.com\u003e",
			"POLITICO",
		},
	}

	for _, test := range tests {
		got := findSender(test.given)
		if got != test.want {
			t.Errorf("findSender(%s) got:%s want:%s", test.given, got, test.want)
		}
	}
}
