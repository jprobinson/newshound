package api

import (
	"time"

	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
)

// NewsAlerts is for accessing News Alerts information from
// the newshound mongo database.
type NewsAlerts struct {
	c *mgo.Collection
}

// NewNewsAlertsDAO returns a new news_alerts_dao for accessing News Alerts information.
func NewNewsAlerts(db *mgo.Database) NewsAlerts {
	return NewsAlerts{db.C("news_alerts")}
}

// NewsAlertLite is a struct that contains partial News Alert
// data. This struct lacks a Body and Raw Body to reduce the size
// when pulling large lists of Alerts. Mainly used for 'findByDate' scenarios.
type NewsAlertLite struct {
	Id          bson.ObjectId `json:"id" bson:"_id"`
	InstanceID  string        `json:"instance_id"bson:"instance_id"`
	ArticleUrl  string        `json:"article_url"bson:"article_url"`
	Sender      string        `json:"sender"`
	Timestamp   time.Time     `json:"timestamp"`
	Tags        []string      `json:"tags"`
	Subject     string        `json:"subject"`
	TopSentence string        `json:"top_sentence"bson:"top_sentence"`
	Sentences   []Sentence    `json:"sentences"`
}

// NewsAlertFull is a struct that contains all News Alert
// data. This struct is used for access to a single Alert's information.
type NewsAlertFull struct {
	Id          bson.ObjectId `json:"id" bson:"_id"`
	InstanceID  string        `json:"instance_id"bson:"instance_id"`
	Article_Url string        `json:"article_url"bson:"article_url"`
	Sender      string        `json:"sender"`
	Timestamp   time.Time     `json:"timestamp"`
	Tags        []string      `json:"tags"`
	Subject     string        `json:"subject"`
	Body        string        `json:"body"`
}

type Sentence struct {
	Value   string   `json:"sentence"bson:"sentence"`
	Phrases []string `json:"noun_phrases"bson:"noun_phrases"`
}

// FindByDate will accept a date range and return any News Alerts that occured within it. News Alert information
// returned will be of the 'lite' form without the raw and scrubbed bodies.
func (na NewsAlerts) FindByDate(start time.Time, end time.Time) (alerts []NewsAlertLite, err error) {
	err = na.c.Find(bson.M{"timestamp": bson.M{"$gte": start, "$lte": end}}).All(&alerts)
	if err != nil {
		return []NewsAlertLite{}, err
	}

	return alerts, nil
}

// FindByDate accepts a slice of News Alert IDs and returns a chronologically ordered list of
// the 'lite' version of News Alerts.
func (na NewsAlerts) FindOrderedAlerts(alertIDs []string) (alerts []NewsAlertLite, err error) {
	var alert_object_ids []bson.ObjectId
	for _, alertID := range alertIDs {
		alert_object_ids = append(alert_object_ids, bson.ObjectIdHex(alertID))
	}

	err = na.c.Find(bson.M{"_id": bson.M{"$in": alert_object_ids}}).Sort("timestamp").All(&alerts)
	if err != nil {
		return
	}

	return
}

// FindAlertByID accepts a News Alert ID and returns the full version of that New Alert's information.
func (na NewsAlerts) FindAlertByID(alertID string) (alert NewsAlertFull, err error) {
	err = na.c.Find(bson.M{"_id": bson.ObjectIdHex(alertID)}).One(&alert)
	if err != nil {
		return
	}

	return
}

// FindAlertHtmlByID accepts a News Alert ID and just the body of the given News Alert.
func (na NewsAlerts) FindAlertHtmlByID(alertID string) (html string, err error) {
	var alert NewsAlertFull
	alert, err = na.FindAlertByID(alertID)
	if err != nil {
		return "", err
	}
	html = alert.Body

	return
}
