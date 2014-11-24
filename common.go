package newshound

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"time"

	"github.com/jprobinson/eazye"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const (
	configFile = "/opt/newshound/etc/config.json"

	ServerLog = "/var/log/newshound/server.log"
	FetchLog  = "/var/log/newshond/fetchd.log"
	AccessLog = "/var/log/newshound/access.log"

	WebDir = "/opt/newshound/www"
)

type Config struct {
	DBURL      string `json:"db-url"`
	DBUser     string `json:"db-user"`
	DBPassword string `json:"db-pw"`

	MarkRead          bool `json:"mark_as_read"`
	eazye.MailboxInfo `,inline`
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
	ID          bson.ObjectId `json:"id" bson:"_id"`
	InstanceID  string        `json:"instance_id"bson:"instance_id"`
	ArticleUrl  string        `json:"article_url"bson:"article_url"`
	Sender      string        `json:"sender"`
	Timestamp   time.Time     `json:"timestamp"`
	Tags        []string      `json:"tags"`
	Subject     string        `json:"subject"`
	TopSentence string        `json:"top_sentence"bson:"top_sentence"`
}

// NewsAlertFull is a struct that contains all News Alert
// data. This struct is used for access to a single Alert's information.
type NewsAlert struct {
	NewsAlertLite `,inline`
	RawBody       string     `json:"-"bson:"raw_body"`
	Body          string     `json:"body"`
	Sentences     []Sentence `json:"sentences"`
}

type Sentence struct {
	Value   string   `json:"sentence"bson:"sentence"`
	Phrases []string `json:"noun_phrases"bson:"noun_phrases"`
}
