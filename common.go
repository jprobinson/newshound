package newshound

import (
	"encoding/json"
	"io/ioutil"
	"log"

	"labix.org/v2/mgo"
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
