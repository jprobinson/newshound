package newshound

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"time"

	"github.com/jprobinson/eazye"
	"github.com/pkg/errors"
)

const (
	configFile = "/opt/newshound/etc/config.json"

	ServerLog = "/var/log/newshound/server.log"
	FetchLog  = "/var/log/newshond/fetchd.log"
	AccessLog = "/var/log/newshound/access.log"

	WebDir = "/opt/newshound/www"

	NewsAlertTopic       = "news-alerts"
	NewsEventTopic       = "news-events"
	NewsEventUpdateTopic = "news-event-updates"
)

type Config struct {
	DBHost     string `json:"db-host"`
	DBUser     string `json:"db-user"`
	DBName     string `json:"db-name"`
	DBPassword string `json:"db-pw"`

	MarkRead          bool `json:"mark_as_read"`
	eazye.MailboxInfo `,inline`

	NSQDAddr     string `json:"nsqd-addr"`
	NSQLAddr     string `json:"nsqlookup-addr"`
	BarkdChannel string `json:"barkd-channel"`

	SlackAlerts []struct {
		Key string `json:"key"`
		Bot string `json:"bot"`
	} `json:"slack-alerts"`

	SlackEvents []struct {
		Key string `json:"key"`
		Bot string `json:"bot"`
	} `json:"slack-events"`

	Twitter []struct {
		ConsumerKey       string `json:"consumer-key"`
		ConsumerSecret    string `json:"consumer-secret"`
		AccessToken       string `json:"access-token"`
		AccessTokenSecret string `json:"access-token-secret"`
	} `json:"twitter"`

	WSPort int `json:"ws-port"`
}

func (c *Config) DB() (*sql.DB, error) {
	db, err := sql.Open("postgres", fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable",
		c.DBUser, c.DBPassword, c.DBHost, c.DBName))
	return db, errors.Wrap(err, "unable to connect to DB")
}

func NewConfig() *Config {
	config := Config{}

	readBytes, err := ioutil.ReadFile(configFile)
	if err != nil {
		log.Fatalf("Cannot read config file: %s %s", config, err)
	}

	err = json.Unmarshal(readBytes, &config)
	if err != nil {
		log.Fatalf("Cannot parse JSON in config file: %s %s", config, err)
	}

	return &config
}

// NewsAlertLite is a struct that contains partial News Alert
// data. This struct lacks a Body and Raw Body to reduce the size
// when pulling large lists of Alerts. Mainly used for 'findByDate' scenarios.
type NewsAlertLite struct {
	ID          int64     `json:"id"`
	ArticleUrl  string    `json:"article_url"`
	Sender      string    `json:"sender"`
	Timestamp   time.Time `json:"timestamp"`
	TopPhrases  []string  `json:"phrases"`
	Subject     string    `json:"subject"`
	TopSentence string    `json:"top_sentence"`
}

// NewsAlertFull is a struct that contains all News Alert
// data. This struct is used for access to a single Alert's information.
type NewsAlert struct {
	NewsAlertLite `,inline`
	RawBody       string     `json:"-"`
	Body          string     `json:"body"`
	Sentences     []Sentence `json:"sentences"`
}

type Sentence struct {
	Value   string   `json:"sentence"`
	Phrases []string `json:"noun_phrases"`
}

// NewsEvent is a struct that contains all the information for
// a particular News Event.
type NewsEvent struct {
	ID          int64            `json:"id"`
	TopPhrases  []string         `json:"phrases"`
	EventStart  time.Time        `json:"event_start"`
	EventEnd    time.Time        `json:"event_end"`
	NewsAlerts  []NewsEventAlert `json:"news_alerts"`
	TopSentence string           `json:"top_sentence"`
	TopSender   string           `json:"top_sender"`
}

// NewsEventAlert is a struct for holding a smaller version of
// News Alert data. This struct has extra fields for determining the order
// and time differences of the News Alerts within the News Event.
type NewsEventAlert struct {
	AlertID     int64    `json:"alert_id"`
	InstanceID  string   `json:"instance_id"`
	ArticleUrl  string   `json:"article_url"`
	Sender      string   `json:"sender"`
	TopPhrases  []string `json:"phrases"`
	Subject     string   `json:"subject"`
	TopSentence string   `json:"top_sentence"`
	Order       int64    `json:"order"`
	TimeLapsed  int64    `json:"time_lapsed"`
}
