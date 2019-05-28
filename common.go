package newshound

import (
	"log"
	"time"

	"github.com/jprobinson/eazye"
	"github.com/kelseyhightower/envconfig"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const (
	WebDir = "/opt/newshound/www"

	NewsAlertTopic       = "news-alerts"
	NewsEventTopic       = "news-events"
	NewsEventUpdateTopic = "news-event-updates"
)

type Config struct {
	DBURL      string `json:"db-url"`
	DBUser     string `json:"db-user"`
	DBPassword string `json:"db-pw"`

	MarkRead          bool `json:"mark_as_read"`
	eazye.MailboxInfo `,inline`

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

func (c *Config) MgoSession() (*mgo.Session, error) {
	// make conn pass it to data
	sess, err := mgo.Dial(c.DBURL)
	if err != nil {
		log.Printf("Unable to connect to newshound db! - %s", err.Error())
		return sess, err
	}

	db := sess.DB("newshound")
	err = db.Login(c.DBUser, c.DBPassword)
	if err != nil {
		log.Printf("Unable to connect to newshound db! - %s", err.Error())
		return sess, err
	}
	sess.SetMode(mgo.Eventual, true)
	return sess, nil
}

func NewConfig() *Config {
	var config Config
	envconfig.MustProcess("", &config)
	return &config
}

// NewsAlertLite is a struct that contains partial News Alert
// data. This struct lacks a Body and Raw Body to reduce the size
// when pulling large lists of Alerts. Mainly used for 'findByDate' scenarios.
type NewsAlertLite struct {
	ID          bson.ObjectId `json:"id" bson:"_id"`
	InstanceID  string        `json:"instance_id"bson:"instance_id"`
	ArticleUrl  string        `json:"article_url"bson:"article_url"`
	Sender      string        `json:"sender"bson:"sender"`
	Timestamp   time.Time     `json:"timestamp"bson:"timestamp"`
	Tags        []string      `json:"tags"bson:"tags"`
	Subject     string        `json:"subject"bson:"subject"`
	TopSentence string        `json:"top_sentence"bson:"top_sentence"`
}

// NewsAlertFull is a struct that contains all News Alert
// data. This struct is used for access to a single Alert's information.
type NewsAlert struct {
	NewsAlertLite `,inline`
	RawBody       string     `json:"-"bson:"raw_body"`
	Body          string     `json:"body"bson:"body"`
	Sentences     []Sentence `json:"sentences"bson:"sentences"`
}

type Sentence struct {
	Value   string   `json:"sentence"bson:"sentence"`
	Phrases []string `json:"noun_phrases"bson:"noun_phrases"`
}

// NewsEvent is a struct that contains all the information for
// a particular News Event.
type NewsEvent struct {
	ID          bson.ObjectId    `json:"id" bson:"_id"`
	Tags        []string         `json:"tags"`
	EventStart  time.Time        `json:"event_start"bson:"event_start"`
	EventEnd    time.Time        `json:"event_end"bson:"event_end"`
	NewsAlerts  []NewsEventAlert `json:"news_alerts"bson:"news_alerts"`
	TopSentence string           `json:"top_sentence"bson:"top_sentence"`
	TopSender   string           `json:"top_sender"bson:"top_sender"`
}

// NewsEventAlert is a struct for holding a smaller version of
// News Alert data. This struct has extra fields for determining the order
// and time differences of the News Alerts within the News Event.
type NewsEventAlert struct {
	AlertID     bson.ObjectId `json:"alert_id"bson:"alert_id"`
	InstanceID  string        `json:"instance_id"bson:"instance_id"`
	ArticleUrl  string        `json:"article_url"bson:"article_url"`
	Sender      string        `json:"sender"bson:"sender"`
	Tags        []string      `json:"tags"bson:"tags"`
	Subject     string        `json:"subject"bson:"subject"`
	TopSentence string        `json:"top_sentence"bson:"top_sentence"`
	Order       int64         `json:"order"bson:"order"`
	TimeLapsed  int64         `json:"time_lapsed"bson:"time_lapsed"`
}
