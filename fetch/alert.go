package fetch

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/mail"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/jasonmoo/toget"
	"github.com/jprobinson/eazye"
	"golang.org/x/net/html"
	"gopkg.in/mgo.v2/bson"

	"github.com/jprobinson/newshound"
)

var Senders = map[string]bool{
	"cnn":                     true,
	"foxnews.com":             true,
	"foxbusiness.com":         true,
	"nbcnews.com":             true,
	"nytimes.com":             true,
	"the washington post":     true,
	"w:sj.com":                true,
	"the wall street journal": true,
	"politico":                true,
	"los angeles times":       true,
	"cbs":                     true,
	"abc":                     true,
	"usatoday.com":            true,
	"yahoo":                   true,
	"ft":                      true,
	"bbc":                     true,
	"npr":                     true,
	"time":                    true,
	"bloomberg.com":           true,
}

func NewNewsAlert(msg eazye.Email, address string) (newshound.NewsAlert, error) {
	sender := findSender(msg.From)

	// default to HTML, but grab text if we must.
	body := msg.HTML
	if len(body) == 0 {
		body = msg.Text
	}

	na := newshound.NewsAlert{
		NewsAlertLite: newshound.NewsAlertLite{
			ID:         bson.NewObjectId(),
			Sender:     sender,
			Subject:    msg.Subject,
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

	news := findNews(text, address, sender)
	if _, bad := badSubjects[sender]; !bad {
		news = periodCheck(news)
		news = append(news, blankSpace...)
		news = append(news, []byte(na.Subject)...)
	}
	na.Tags, na.Sentences, na.TopSentence, err = callNP(news)
	return na, err
}

func ReParseNewsAlert(na newshound.NewsAlert, address string) (newshound.NewsAlert, error) {
	body := []byte(na.RawBody)
	na.ArticleUrl = findArticleUrl(na.Sender, body)
	na.Body = scrubBody(body, address)

	text, err := eazye.VisibleText(bytes.NewReader(body))
	if err != nil {
		log.Print("unable to get visible text: ", err)
		return na, err
	}

	news := findNews(text, address, na.Sender)
	if _, bad := badSubjects[na.Sender]; !bad {
		news = periodCheck(news)
		news = append(news, blankSpace...)
		news = append(news, []byte(na.Subject)...)
	}
	na.Tags, na.Sentences, na.TopSentence, err = callNP(news)
	return na, err
}

var (
	blankSpace      = []byte(" ")
	period          = []byte(".")
	comma           = []byte(",")
	periodWithSpace = []byte(". ")
)

func periodCheck(line []byte) []byte {
	if len(line) > 0 &&
		!bytes.HasSuffix(bytes.TrimSpace(line), period) &&
		!bytes.HasSuffix(bytes.TrimSpace(line), comma) {
		line = append(line, periodWithSpace...)
	}
	return line
}

func findNews(text [][]byte, address, sender string) []byte {
	// prep the address for searching against text
	addr := []byte(address)
	addrStart := bytes.SplitN(addr, []byte("@"), 2)[0]
	// so we can reuse similar addresses?
	if len(addrStart) > 15 {
		addrStart = addrStart[:15]
	}

	var news [][]byte
	badLines := 0
	senderBytes := bytes.ToLower([]byte(sender))
	for _, line := range text {
		line := bytes.Trim(line, "-| ?")
		if isNews(line, addr, addrStart, senderBytes) {
			badLines = 0
			news = append(news, line)

		} else if (len(news) > 0) && (len(line) > 0) {
			badLines++
		}

		// get at most 3 or quit if we have bad rows and at least 2 or over 2 bads and 1 good
		if (len(news) >= 3) || ((badLines > 0) && (len(news) >= 2)) || ((badLines >= 2) && (len(news) >= 1)) {
			break
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
		[]byte("to your address book"),
		[]byte("to ensure delivery"),
		[]byte("having difficulty"),
		[]byte("having trouble"),
		[]byte("to view this"),
		[]byte("please add"),
		[]byte("get complete coverage"),
		[]byte("for more"),
		[]byte("more on this"),
		[]byte("for breaking news,"),
		[]byte("you are currently subscribed"),
		[]byte("for the latest"),
		[]byte("view this"),
		[]byte("and watch"),
		[]byte("watch cnn live or on demand"),
		[]byte("read more"),
		[]byte(", cnn tv"),
		[]byte("or on the"),
		[]byte("you received this"),
		[]byte("cnn is now live"),
		[]byte("you can watch"),
		[]byte("fox business never sends"),
		[]byte("you have opt"),
		[]byte("fox news never"),
		[]byte("more top stories"),
		[]byte("for further development"),
		[]byte("this is a developing"),
		[]byte("to unsubscribe"),
		[]byte("privacy policy"),
		[]byte("7950 jones branch drive"),
		[]byte("\u00a0731 lexington avenue"),
	}

	// maps for lookups! hooray!
	junkFillers = map[string]struct{}{
		"this email":                   struct{}{},
		"click here":                   struct{}{},
		"all rights reserved":          struct{}{},
		"go here":                      struct{}{},
		"ft exclusive":                 struct{}{},
		"advertisement":                struct{}{},
		"advertise":                    struct{}{},
		"for more":                     struct{}{},
		"share this":                   struct{}{},
		"view this":                    struct{}{},
		"—":                            struct{}{},
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
		"manage portfolio":             struct{}{},
		"forward this email":           struct{}{},
		"subscribe to":                 struct{}{},
		"view it in your browser":      struct{}{},
		"you are currently subscribed": struct{}{},
		"manage alerts":                struct{}{},
		"alerts":                       struct{}{},
		"alerts.":                      struct{}{},
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
		"nytimes.com":                  struct{}{},
		"usatoday.com":                 struct{}{},
		"foxnews.com":                  struct{}{},
		"home delivery":                struct{}{},
		"the wall street journal":      struct{}{},
		"go to":                        struct{}{},
		"et":                           struct{}{},
		"pt":                           struct{}{},
		"nbcnews.com":                  struct{}{},
		"sponsored\u00a0by":            struct{}{},
		"sponsored\xa0by":              struct{}{},
		"@nbcnews":                     struct{}{},
		"@nbcnews.":                    struct{}{},
		"cnn.com":                      struct{}{},
		"watch cnn":                    struct{}{},
		"businessweek.com":             struct{}{},
		"foxbusiness.com":              struct{}{},
		"fox business never sends":     struct{}{},
		"nytdirect@nytimes.com":        struct{}{},
		"share on facebook":            struct{}{},
		"video alerts":                 struct{}{},
		"on your cell phone":           struct{}{},
		"more coverage":                struct{}{},
		"you received this message":    struct{}{},
		"you received this email":      struct{}{},
		"visit":                        struct{}{},
	}

	breakingNewsFiller = map[string]struct{}{
		"breaking news":       struct{}{},
		"breaking news.":      struct{}{},
		"breaking\xa0news":    struct{}{},
		"breaking\u00a0news":  struct{}{},
		"breaking news alert": struct{}{},
		"news alert":          struct{}{},
		"national news alert": struct{}{},
		"sports alert":        struct{}{},
	}
	nationalJunk = []byte("national")
	dotJunk      = []byte("•")
	likeJunk     = []byte("like")
	followJunk   = []byte("follow")
	twitter      = []byte("twitter")
	facebook     = []byte("facebook")
)

func isNews(line []byte, address []byte, addrStart, sender []byte) bool {
	// less than 5 chars? very likely crap
	if len(line) < 5 {
		return false
	}

	lower := bytes.ToLower(line)
	if bytes.HasPrefix(lower, nationalJunk) && bytes.Contains(lower, dotJunk) {
		return false
	}

	// if it's just the sender and little else, likely crap
	if len(bytes.Replace(lower, sender, []byte{}, -1)) < 5 {
		return false
	}

	if (bytes.HasPrefix(lower, likeJunk) || bytes.HasPrefix(lower, followJunk)) &&
		(bytes.HasSuffix(lower, facebook) || bytes.HasSuffix(lower, twitter)) {
		return false
	}

	for _, start := range junkStarts {
		if bytes.HasPrefix(lower, start) {
			return false
		}
	}

	if len(lower) < 30 {
		// string conversion..OUCH! at least its only on small strings...
		lowerStr := string(lower)
		if _, isJunk := junkFillers[lowerStr]; isJunk {
			return false
		}
		if _, isBreaking := breakingNewsFiller[lowerStr]; isBreaking {
			return false
		}
	}

	if bytes.HasPrefix(lower, []byte("www.")) {
		return false
	}

	if bytes.HasPrefix(lower, []byte("http")) {
		return false
	}

	if bytes.Contains(lower, []byte("|")) ||
		bytes.Contains(lower, []byte("©")) ||
		bytes.Contains(lower, []byte("=")) {
		return false
	}

	// only run the date check on smaller strings
	if len(line) < 50 {
		if isDate(lower) {
			return false
		}
	}

	if bytes.Contains(line, addrStart) || bytes.Contains(line, address) {
		return false
	}

	return true
}

var (
	daysOfWeek = [][]byte{
		[]byte("sunday"),
		[]byte("monday"),
		[]byte("tuesday"),
		[]byte("wednesday"),
		[]byte("thursday"),
		[]byte("friday"),
		[]byte("saturday"),
	}
	daysOfWeekShort = [][]byte{
		[]byte("sun"),
		[]byte("mon"),
		[]byte("tue"),
		[]byte("wed"),
		[]byte("thu"),
		[]byte("thur"),
		[]byte("fri"),
		[]byte("sat"),
	}
	months = [][]byte{
		[]byte("jan"), []byte("january"),
		[]byte("feb"), []byte("february"),
		[]byte("mar"), []byte("march"),
		[]byte("apr"), []byte("april"),
		[]byte("may"),
		[]byte("jun"), []byte("june"),
		[]byte("jul"), []byte("july"),
		[]byte("aug"), []byte("august"),
		[]byte("sep"), []byte("sept"), []byte("september"),
		[]byte("oct"), []byte("october"),
		[]byte("nov"), []byte("november"),
		[]byte("dec"), []byte("december"),
	}
	timezones = [][]byte{
		[]byte("est"), []byte("edt"), []byte("et"),
		[]byte("pst"), []byte("pdt"), []byte("pt"),
	}
	amPM = [][]byte{[]byte("a.m"), []byte("p.m"), []byte("am"), []byte("pm")}
)

// isDate detects if we see one of the following formats:
// August 12, 2014
// Aug 10, 2014  1:02 PM EDT
// Sunday August 10 2014
// Sunday, August 10, 2014 2:36 PM EDT
// Monday, August 11, 2014 9:18:59 AM
// Sat., Feb. 7, 2015 04:35 PM
// Tue., Apr. 21, 2015 4:17 p.m.
func isDate(line []byte) bool {
	// Trim dots 'n periods
	line = bytes.Trim(line, "• .\u00a0")
	// check if it starts with a day or month
	dateStart := false
	for _, day := range daysOfWeek {
		if bytes.HasPrefix(line, day) {
			dateStart = true
			break
		}
	}
	if !dateStart {
		for _, day := range daysOfWeekShort {
			if bytes.HasPrefix(line, day) {
				dateStart = true
				break
			}
		}
	}
	if !dateStart {
		for _, month := range months {
			if bytes.HasPrefix(line, month) {
				dateStart = true
				break
			}
		}
	}

	if !dateStart {
		return false
	}

	// check if it ends with a timezone/daytime/year
	dateEnd := false
	for _, ap := range amPM {
		if bytes.HasSuffix(line, ap) {
			dateEnd = true
			break
		}
	}
	if !dateEnd {
		// newshound started in 2012. adjust if you want older data
		for i := 2012; i <= time.Now().Year(); i++ {
			if bytes.HasSuffix(line, []byte(strconv.Itoa(i))) {
				dateEnd = true
				break
			}
		}
	}
	if !dateEnd {
		for _, zone := range timezones {
			if bytes.HasSuffix(line, zone) {
				dateEnd = true
				break
			}
		}
	}

	return dateEnd
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
		[]byte("UNSUBSCRIBE"),
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
	body = replaceHREFs(body)
	return string(body)
}

var (
	senderStoryURLs = map[string]int{
		"BBC":                     4,
		"CBS":                     2,
		"FT":                      2,
		"WSJ.com":                 1,
		"The Wall Street Journal": 3,
		"USATODAY.com":            6,
		"NYTimes.com":             3,
		"The Washington Post":     2,
		"FoxNews.com":             0,
		"NPR":                     3,
	}
	badSubjects = map[string]struct{}{
		"ABC":      struct{}{},
		"CNN":      struct{}{},
		"POLITICO": struct{}{},
	}
)

func findArticleUrl(sender string, body []byte) string {
	var aUrl string
	if index, ok := senderStoryURLs[sender]; ok {
		hrefs := findHREFs(body)
		// if we didnt find enough, give up
		if len(hrefs) < (index + 1) {
			log.Print("not enough urls: ", hrefs)
			return aUrl
		}

		aUrl = hrefs[index]

		// ignore if it is a doubleclick link
		if strings.Contains(aUrl, "doubleclick") {
			return ""
		}

		// hit url and try to grab the result url
		resp, err := toget.Get(aUrl, 3*time.Second)
		if err != nil {
			if resp != nil {
				loc := resp.Header.Get("Location")
				if len(loc) == 0 {
					return aUrl
				}
				resp.Body.Close()
				var rUrl *url.URL
				if rUrl, err = url.Parse(loc); err != nil {
					return aUrl
				}
				if aUrl = rUrl.Query().Get("URI"); len(aUrl) == 0 {
					return aUrl
				}
			}
		} else {
			resp.Body.Close()
			// grab url after all redirects
			aUrl = resp.Request.URL.String()
		}

		// and cut off any parameters
		aUrl = strings.Split(aUrl, "?")[0]
	}
	return aUrl
}

var (
	anchorTag  = []byte("a")
	hrefAttr   = []byte("href")
	classAttr  = []byte("class")
	styleAttr  = []byte("style")
	httpPrefix = []byte("http")
	blank      = []byte("")

	urlRegex = regexp.MustCompile(`http[s]?://(?:[a-zA-Z]|[0-9]|[$-_@.&+]|[!*\(\),]|(?:%[0-9a-fA-F][0-9a-fA-F]))+`)
)

func replaceHREFs(body []byte) []byte {
	z := html.NewTokenizer(bytes.NewReader(body))
	var out bytes.Buffer
loop:
	for {
		tt := z.Next()
		switch tt {
		case html.ErrorToken:
			if err := z.Err(); err != nil && err != io.EOF {
				log.Print("unexpected error parsing html: ", err)
			}
			break loop
		case html.TextToken:
			// replace all URLs in the text
			out.Write(urlRegex.ReplaceAll(z.Raw(), blank))
		case html.StartTagToken:
			tn, hasAttr := z.TagName()
			if bytes.Equal(tn, anchorTag) && hasAttr {
				// keep the styling, drop everything else
				var class []byte
				var style []byte
				for {
					key, val, more := z.TagAttr()
					if bytes.Equal(classAttr, key) {
						class = val
					}
					if bytes.Equal(styleAttr, key) {
						style = val
					}
					if !more {
						break
					}
				}
				// write our fake anchor tag
				fmt.Fprintf(&out, `<a href="#" style="%s" class="%s">`, style, class)
				continue
			}
			// just write what we got if not an anchor
			out.Write(z.Raw())
		default:
			out.Write(z.Raw())
		}
	}

	return out.Bytes()
}
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
