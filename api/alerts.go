package api

import (
	"context"
	"time"

	"github.com/jprobinson/newshound"
	"go.opencensus.io/trace"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// FindByDate will accept a date range and return any News Alerts that occured within it. News Alert information
// returned will be of the 'lite' form without the raw and scrubbed bodies.
func FindAlertsByDate(ctx context.Context, db *mgo.Database, start time.Time, end time.Time) ([]newshound.NewsAlertLite, error) {
	ctx, span := trace.StartSpan(ctx, "newshound/mongodb/find-alerts-by-date")
	defer span.End()

	c := getNA(db)
	var alerts []newshound.NewsAlertLite
	if err := c.Find(bson.M{"timestamp": bson.M{"$gte": start, "$lte": end}}).All(&alerts); err != nil {
		return alerts, err
	}

	return alerts, nil
}

// FindByDate accepts a slice of News Alert IDs and returns a chronologically ordered list of
// the 'lite' version of News Alerts.
func FindOrderedAlerts(ctx context.Context, db *mgo.Database, alertIDs []string) ([]newshound.NewsAlertLite, error) {
	ctx, span := trace.StartSpan(ctx, "newshound/mongodb/find-ordered-alerts")
	defer span.End()

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
func FindAlertByID(ctx context.Context, db *mgo.Database, alertID string) (newshound.NewsAlert, error) {
	ctx, span := trace.StartSpan(ctx, "newshound/mongodb/find-alerts-by-id")
	defer span.End()

	c := getNA(db)
	var alert newshound.NewsAlert
	if err := c.Find(bson.M{"_id": bson.ObjectIdHex(alertID)}).One(&alert); err != nil {
		return alert, err
	}

	return alert, nil
}

// FindAlertHtmlByID accepts a News Alert ID and just the body of the given News Alert.
func FindAlertHtmlByID(ctx context.Context, db *mgo.Database, alertID string) (string, error) {
	alert, err := FindAlertByID(ctx, db, alertID)
	if err != nil {
		return "", err
	}

	return alert.Body, nil
}

func getNA(db *mgo.Database) *mgo.Collection {
	return db.C("news_alerts")
}
