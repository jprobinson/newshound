package api

import (
	"time"

	"github.com/jprobinson/newshound"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// FindByDate will accept a date range and return any News Alerts that occured within it. News Alert information
// returned will be of the 'lite' form without the raw and scrubbed bodies.
func FindAlertsByDate(db *mgo.Database, start time.Time, end time.Time) ([]newshound.NewsAlertLite, error) {
	c := getNA(db)
	var alerts []newshound.NewsAlertLite
	if err := c.Find(bson.M{"timestamp": bson.M{"$gte": start, "$lte": end}}).All(&alerts); err != nil {
		return alerts, err
	}

	return alerts, nil
}

// FindByDate accepts a slice of News Alert IDs and returns a chronologically ordered list of
// the 'lite' version of News Alerts.
func FindOrderedAlerts(db *mgo.Database, alertIDs []string) ([]newshound.NewsAlertLite, error) {
	var alertObjectIDs []bson.ObjectId
	var alerts []newshound.NewsAlertLite
	for _, alertID := range alertIDs {
		alertObjectIDs = append(alertObjectIDs, bson.ObjectIdHex(alertID))
	}
	c := getNA(db)
	if err := c.Find(bson.M{"_id": bson.M{"$in": alertObjectIDs}}).Sort("timestamp").All(&alerts); err != nil {
		return alerts, err
	}

	return alerts, nil
}

// FindAlertByID accepts a News Alert ID and returns the full version of that New Alert's information.
func FindAlertByID(db *mgo.Database, alertID string) (newshound.NewsAlert, error) {
	c := getNA(db)
	var alert newshound.NewsAlert
	if err := c.Find(bson.M{"_id": bson.ObjectIdHex(alertID)}).One(&alert); err != nil {
		return alert, err
	}

	return alert, nil
}

// FindAlertHtmlByID accepts a News Alert ID and just the body of the given News Alert.
func FindAlertHtmlByID(db *mgo.Database, alertID string) (string, error) {
	alert, err := FindAlertByID(db, alertID)
	if err != nil {
		return "", err
	}

	return alert.Body, nil
}

func getNA(db *mgo.Database) *mgo.Collection {
	return db.C("news_alerts")
}
