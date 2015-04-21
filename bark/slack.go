package bark

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/jprobinson/newshound"
)

type slackKey struct {
	key     string
	botName string
}

func AddSlackAlertBot(d *Distributor, key, botName string) {
	skey := slackKey{key, botName}
	d.AddAlertBarker(&SlackAlertBarker{skey})
}

func AddSlackEventBot(d *Distributor, key, botName string) {
	skey := slackKey{key, botName}
	d.AddEventBarker(&SlackEventBarker{skey})
}

type SlackAlertBarker struct {
	key slackKey
}

func (s *SlackAlertBarker) Bark(alert newshound.NewsAlertLite) error {
	title := fmt.Sprintf("%s - %s", strings.TrimSuffix(alert.Sender, ".com"), alert.Subject)

	link := alertLink(alert)
	message := fmt.Sprintf("\n%s\n<%s|more...>", alert.TopSentence, link)
	color := SenderColors[strings.ToLower(alert.Sender)]
	return sendSlack(s.key.botName, s.key.key, title, link, message, color)
}

type SlackEventBarker struct {
	key slackKey
}

func (s *SlackEventBarker) Bark(event newshound.NewsEvent) error {
	title := fmt.Sprintf("New Event With %d Alerts!", len(event.NewsAlerts))
	link := eventLink(event)
	message := fmt.Sprintf("_key quote_\n%s\n_from_\n%s\n<%s|more info...>",
		event.TopSentence,
		strings.TrimSuffix(event.TopSender, ".com"),
		link)
	return sendSlack(s.key.botName, s.key.key, title, link, message, "#439FE0")
}

type slackAttachment struct {
	Title     string   `json:"title"`
	TitleLink string   `json:"title_link,omitempty"`
	Text      string   `json:"text"`
	Fallback  string   `json:"fallback"`
	Color     string   `json:"color"`
	MrkDownIn []string `json:"mrkdwn_in"`
}

func sendSlack(bot, key, title, link, message, color string) error {
	data := struct {
		Username    string            `json:"username"`
		Unfurl      bool              `json:"unfurl_links"`
		MrkDwn      bool              `json:"mrkdwn"`
		Attachments []slackAttachment `json:"attachments"`
	}{
		bot,
		false,
		true,
		[]slackAttachment{
			slackAttachment{
				title,
				link,
				message,
				message,
				color,
				[]string{"text", "fallback"},
			},
		},
	}

	var payload bytes.Buffer
	err := json.NewEncoder(&payload).Encode(data)
	if err != nil {
		log.Print("unable to encode slack json:", err)
		return err
	}

	slackURL := fmt.Sprintf("https://hooks.slack.com/services/%s", key)
	r, err := http.Post(slackURL, "application/json", &payload)
	defer r.Body.Close()
	if err != nil {
		resp, _ := ioutil.ReadAll(r.Body)
		log.Printf("unable to send slack notification: %s\nresponse:\n%s", err, string(resp))
	}

	return err
}

func alertLink(alert newshound.NewsAlertLite) string {
	return fmt.Sprintf("http://newshound.jprbnsn.com/#/calendar?start=%s&display=alerts&alert=%s",
		alert.Timestamp.Format("2006-01-02"),
		alert.ID.Hex())
}

func eventLink(event newshound.NewsEvent) string {
	return fmt.Sprintf("http://newshound.jprbnsn.com/#/calendar?start=%s&display=events&event=%s",
		event.EventStart.Format("2006-01-02"),
		event.ID.Hex())
}
