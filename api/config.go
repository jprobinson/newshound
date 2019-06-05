package api

import (
	"log"

	"github.com/kelseyhightower/envconfig"
	"gopkg.in/mgo.v2"
)

type Config struct {
	DBURL      string `envconfig:"DB_URL"`
	DBUser     string `envconfig:"DB_USER"`
	DBPassword string `envconfig:"DB_PASSWORD"`
}

func NewConfig() *Config {
	var cfg Config
	envconfig.MustProcess("", &cfg)
	return &cfg
}

func (c *Config) MgoSession() (*mgo.Session, error) {
	// make conn pass it to data
	log.Printf("connecting to %s", c.DBURL)
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
