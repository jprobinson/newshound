package fetch

import (
	"log"
	"os"

	"github.com/jprobinson/eazye"
	"github.com/kelseyhightower/envconfig"
	"gopkg.in/mgo.v2"
)

type Config struct {
	DBURL      string `envconfig:"DB_URL"`
	DBUser     string `envconfig:"DB_USER"`
	DBPassword string `envconfig:"DB_PASSWORD"`

	MarkRead bool `envconfig:"MARK_READ"`

	Mailbox eazye.MailboxInfo `envconfig:"MAILBOX"`

	NPHost string `envconfig:"NP_HOST"`
}

func NewConfig() *Config {
	var cfg Config
	envconfig.MustProcess("", &cfg)
	envconfig.MustProcess("", &cfg.Mailbox)
	cfg.Mailbox.Host = os.Getenv("MAIL_HOST")
	cfg.Mailbox.User = os.Getenv("MAIL_USER")
	cfg.Mailbox.Pwd = os.Getenv("MAIL_PWD")
	log.Printf("config: %#v", cfg)
	return &cfg
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
