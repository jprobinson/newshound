package api

import (
	"time"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// NewsEvent is a struct that contains all the information for
// a particular News Event.
type NewsEvent struct {
	Id          bson.ObjectId    `json:"id" bson:"_id"`
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
	AlertId     bson.ObjectId `json:"alert_id"bson:"alert_id"`
	InstanceID  string        `json:"instance_id"bson:"instance_id"`
	ArticleUrl  string        `json:"article_url"bson:"article_url"`
	Sender      string        `json:"sender"`
	Tags        []string      `json:"tags"`
	Subject     string        `json:"subject"`
	TopSentence string        `json:"top_sentence"bson:"top_sentence"`
	Order       int64         `json:"order"`
	TimeLapsed  int64         `json:"time_lapsed"bson:"time_lapsed"`
}

// FindByDate accepts a start and end date and returns all the News Events that occured in that timeframe.
func FindEventsByDate(db *mgo.Database, start time.Time, end time.Time) (events []NewsEvent, err error) {
	c := getNE(db)
	err = c.Find(bson.M{"event_start": bson.M{"$gte": start, "$lte": end}}).All(&events)
	if err != nil {
		return
	}

	return
}

// FindByDateReverse accepts a start and end date and returns all the News Events that occured in that timeframe order by time desc.
func FindEventsByDateReverse(db *mgo.Database, start time.Time, end time.Time) (events []NewsEvent, err error) {
	c := getNE(db)
	err = c.Find(bson.M{"event_start": bson.M{"$gte": start, "$lte": end}}).Sort("-event_start").All(&events)
	if err != nil {
		return
	}

	return
}

// FindEventByID accepts a News Event ID and returns the full information for that Event.
func FindEventByID(db *mgo.Database, eventID string) (event NewsEvent, err error) {
	c := getNE(db)
	err = c.Find(bson.M{"_id": bson.ObjectIdHex(eventID)}).One(&event)
	if err != nil {
		return
	}

	return
}

func getNE(db *mgo.Database) *mgo.Collection {
	return db.C("news_events")
}
