package api

import (
	"time"

	"github.com/jprobinson/newshound"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// FindByDate accepts a start and end date and returns all the News Events that occured in that timeframe.
func FindEventsByDate(db *mgo.Database, start time.Time, end time.Time) (events []newshound.NewsEvent, err error) {
	c := getNE(db)
	err = c.Find(bson.M{"event_start": bson.M{"$gte": start, "$lte": end}}).All(&events)
	if err != nil {
		return
	}

	return
}

// FindByDateReverse accepts a start and end date and returns all the News Events that occured in that timeframe order by time desc.
func FindEventsByDateReverse(db *mgo.Database, start time.Time, end time.Time) (events []newshound.NewsEvent, err error) {
	c := getNE(db)
	err = c.Find(bson.M{"event_start": bson.M{"$gte": start, "$lte": end}}).Sort("-event_start").All(&events)
	if err != nil {
		return
	}

	return
}

// FindEventByID accepts a News Event ID and returns the full information for that Event.
func FindEventByID(db *mgo.Database, eventID string) (event newshound.NewsEvent, err error) {
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
