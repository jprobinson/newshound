package fetch

import (
	"bytes"
	"encoding/base64"
	"io"
	"log"
	"net/mail"
	"regexp"
	"strings"
	"time"

	"github.com/jasonmoo/toget"
	"github.com/jprobinson/eazye"
	"golang.org/x/net/html"

	"github.com/jprobinson/newshound"
)

func NewNewsAlert(msg eazye.Email, address string) (newshound.NewsAlert, error) {
	sender := findSender(msg.From)

	// default to HTML, but grab text if we must.
	body := msg.HTML
	if len(body) == 0 {
		body = msg.Text
	}

	na := newshound.NewsAlert{
		NewsAlertLite: newshound.NewsAlertLite{
			Sender:     sender,
			Subject:    parseSubject(msg.Subject),
			Timestamp:  msg.InternalDate,
			ArticleUrl: findArticleUrl(sender, body),
			InstanceID: msg.Message.Header.Get("X-InstanceId"),
		},
		RawBody: string(body),
		Body:    scrubBody(body, address),
	}

	text, err := msg.VisibleText()
	if err != nil {
		log.Print("unable to get visible text: ", err)
		return na, err
	}

	news := findNews(text)
	na.Tags, na.Sentences, na.TopSentence, err = callNP(news)
	return na, err
}

func findNews(text [][]byte) []byte {
	var news [][]byte
	for _, line := range text {
		line := bytes.Trim(line, "-| ")
		if isNews(line) {
			news = append(news, line)
		}
	}

	return bytes.Join(news, []byte(" "))
}

var (
	junkStarts = [][]byte{
		[]byte("follow"),
		[]byte("like"),
		[]byte("national/global news alert  •"),
		[]byte("technology news alert  •"),
		[]byte("sports news alert  •"),
		[]byte("economy/business news alert  •"),
		[]byte("national politics news alert  •"),
	}

	// maps for lookups! hooray!
	junkFillers = map[string]struct{}{
		"this email":                   struct{}{},
		"click here":                   struct{}{},
		"go here":                      struct{}{},
		"advertisement":                struct{}{},
		"for more":                     struct{}{},
		"share this":                   struct{}{},
		"view this":                    struct{}{},
		"to unsubscribe":               struct{}{},
		"e-mail alerts":                struct{}{},
		"view it online":               struct{}{},
		"complete coverage":            struct{}{},
		"paste the link":               struct{}{},
		"paste this link":              struct{}{},
		"is a developing story":        struct{}{},
		"for further developments":     struct{}{},
		"for breaking news":            struct{}{},
		"the moment it happens":        struct{}{},
		"keep reading":                 struct{}{},
		"connect with diane":           struct{}{},
		"sponsored by":                 struct{}{},
		"on the go?":                   struct{}{},
		"more newsletters":             struct{}{},
		"manage email":                 struct{}{},
		"manage subscriptions":         struct{}{},
		"text \"breaking":              struct{}{},
		"read more":                    struct{}{},
		"contact your cable":           struct{}{},
		"for the latest":               struct{}{},
		"to your address book":         struct{}{},
		"unsubscribe":                  struct{}{},
		"and watch ":                   struct{}{},
		"if this message":              struct{}{},
		"bloomberg news\xa0on twitter": struct{}{},
		"to view this email":           struct{}{},
		"more on this":                 struct{}{},
		"more stories":                 struct{}{},
		"go to nbcnews":                struct{}{},
		"to ensure":                    struct{}{},
		"privacy policy":               struct{}{},
		"read this story":              struct{}{},
		"sponsored\xa0by":              struct{}{},
		"manage portfolio":             struct{}{},
		"forward this email":           struct{}{},
		"subscribe to":                 struct{}{},
		"view it in your browser":      struct{}{},
		"you are currently subscribed": struct{}{},
		"manage alerts":                struct{}{},
		"manage preferences":           struct{}{},
		"update your profile":          struct{}{},
		"send to a friend":             struct{}{},
		"contact us":                   struct{}{},
		"731 lexington ave":            struct{}{},
		"view in your web browser":     struct{}{},
		"update preferences":           struct{}{},
		"feedback":                     struct{}{},
		"bloomberg tv+":                struct{}{},
		"bloomberg.com":                struct{}{},
		"businessweek.com":             struct{}{},
		"share on facebook":            struct{}{},
		"video alerts":                 struct{}{},
		"on your cell phone":           struct{}{},
		"more coverage":                struct{}{},
		"you received this message":    struct{}{},
	}
	nationalJunk = []byte("national")
	dotJunk      = []byte("•")
)

func isNews(line []byte) bool {
	// less than 5 chars? very likely crap
	if len(line) < 5 {
		return false
	}

	lower := bytes.ToLower(line)
	if bytes.HasPrefix(lower, nationalJunk) && bytes.Contains(lower, dotJunk) {
		return false
	}

	for _, start := range junkStarts {
		if bytes.HasPrefix(lower, start) {
			return false
		}
	}

	if len(lower) < 30 {
		// string conversion..OUCH! at least its only on small strings...
		if _, isJunk := junkFillers[string(lower)]; isJunk {
			return false
		}
	}

	return true
}

func parseSubject(subject string) string {
	if strings.HasPrefix(subject, "=?UTF-8?") {
		subs := strings.SplitN(subject, " ", -1)
		var result string
		for _, sub := range subs {
			sub = strings.Replace(sub, "=?UTF-8?", "", -1)
			sub = strings.Replace(sub, "?=", "", -1)
			data, err := base64.StdEncoding.DecodeString(sub)
			if err != nil {
				return ""
			}
			result += string(data)
		}
	}
	return subject
}

func findSender(from *mail.Address) string {
	// by default, just grab the first word
	sender := strings.SplitN(from.Name, " ", -1)[0]

	lower := strings.ToLower(sender)
	if lower == "la" {
		sender = "Los Angeles Times"
	}

	// deal with any 'the's and 'los'
	if lower == "the" || lower == "los" {
		sender = from.Name
	}

	return sender
}

var (
	unsubText = [][]byte{
		[]byte("unsubscribe"),
		[]byte("Unsubscribe"),
		[]byte("dyn.politico.com"),
		[]byte("email.foxnews.com"),
		[]byte("reg.e.usatoday.com"),
		[]byte("click.e.usatoday.com"),
		[]byte("cheetahmail"),
		[]byte("EmailSubMgr"),
		[]byte("newsvine"),
		[]byte("ts.go.com"),
	}
	emptyString = []byte("")
)

func scrubBody(body []byte, address string) string {
	for _, unsub := range unsubText {
		body = bytes.Replace(body, unsub, emptyString, -1)
	}
	body = bytes.Replace(body, []byte(address), emptyString, -1)
	name := []byte(strings.Split(address, "@")[0])
	body = bytes.Replace(body, name, emptyString, -1)
	return string(body)
}

var (
	senderStoryURLs = map[string]int{
		"BBC":                 4,
		"CBS":                 2,
		"FT":                  2,
		"WSJ.com":             1,
		"USATODAY.com":        6,
		"NYTimes.com":         6,
		"The Washington Post": 3,
		"FoxNews.com":         0,
		"NPR":                 3,
	}
)

func findArticleUrl(sender string, body []byte) string {
	var url string
	if index, ok := senderStoryURLs[sender]; ok {
		hrefs := findHREFs(body)
		// if we didnt find enough, give up
		if len(hrefs) < (index + 1) {
			log.Print("not enough urls: ", hrefs)
			return url
		}

		url = hrefs[index]

		// ignore if it is a doubleclick link
		if strings.Contains(url, "doubleclick") {
			return ""
		}

		// hit url and try to grab the result url
		resp, err := toget.Get(url, time.Second*20)
		if err != nil {
			log.Print("unable to get URL: ", err)
			return url
		}
		resp.Body.Close()
		// grab url after all redirects
		url = resp.Request.URL.String()

		// and cut off any parameters
		url = strings.Split(url, "?")[0]
	}
	return url
}

var (
	anchorTag  = []byte("a")
	hrefAttr   = []byte("href")
	httpPrefix = []byte("http")

	urlRegex = regexp.MustCompile(`http[s]?://(?:[a-zA-Z]|[0-9]|[$-_@.&+]|[!*\(\),]|(?:%[0-9a-fA-F][0-9a-fA-F]))+`)
)

func findHREFs(body []byte) []string {
	var hrefs []string

	z := html.NewTokenizer(bytes.NewReader(body))
loop:
	for {
		tt := z.Next()
		switch tt {
		case html.ErrorToken:
			if err := z.Err(); err != nil && err != io.EOF {
				log.Print("unexpected error parsing html: ", err)
			}
			break loop
		case html.StartTagToken:
			tn, hasAttr := z.TagName()
			if bytes.Equal(tn, anchorTag) && hasAttr {
				// loop til we find an href attr or the end
				for {
					key, val, more := z.TagAttr()
					if bytes.Equal(hrefAttr, key) && bytes.HasPrefix(val, httpPrefix) {
						hrefs = append(hrefs, string(val))
						break
					}
					if !more {
						break
					}
				}
			}
		}
	}

	// found nothing? maybe regex for it?
	if len(hrefs) == 0 {
		matches := urlRegex.FindAll(body, -1)
		for _, match := range matches {
			hrefs = append(hrefs, string(match))
		}
	}
	return hrefs
}
