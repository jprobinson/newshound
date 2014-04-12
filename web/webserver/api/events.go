package api

import (
	"time"

	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
)

// NewsEvents is for accessing News Events information from
// the newshound mongo database.
type NewsEvents struct {
	c *mgo.Collection
}

// NewNewsEvents returns a new NewsEvents for accessing News Events information.
func NewNewsEvents(db *mgo.Database) NewsEvents {
	return NewsEvents{db.C("news_events")}
}

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
	//Sentences   []Sentence    `json:"sentences"`
	Order       int64         `json:"order"`
	TimeLapsed  int64         `json:"time_lapsed"bson:"time_lapsed"`
}

// FindByDate accepts a start and end date and returns all the News Events that occured in that timeframe.
func (ne NewsEvents) FindByDate(start time.Time, end time.Time) (events []NewsEvent, err error) {
	err = ne.c.Find(bson.M{"event_start": bson.M{"$gte": start, "$lte": end}}).All(&events)
	if err != nil {
		return
	}

	return
}

// FindEventByID accepts a News Event ID and returns the full information for that Event.
func (ne NewsEvents) FindEventByID(eventID string) (event NewsEvent, err error) {
	err = ne.c.Find(bson.M{"_id": bson.ObjectIdHex(eventID)}).One(&event)
	if err != nil {
		return
	}

	return
}
