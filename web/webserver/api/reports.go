package api

import (
	"fmt"
	"time"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type TimeframeID struct {
	Sender    string `json:"sender"bson:"sender"`
	Timeframe string `json:"timeframe"bson:"timeframe"`
}

type AvgAlertsPerWeek struct {
	ID    TimeframeID    `json:"id"bson:"_id"`
	Value AvgAlertsValue `json:"value"bson:"value"`
}

type AvgAlertsValue struct {
	AvgAlerts   float64 `json:"avg_alerts"bson:"avg_alerts"`
	TotalAlerts int     `json:"total_alerts"bson:"total_alerts"`
}

type AvgAlertsReport struct {
	Sender string                    `json:"sender"`
	Values map[string]AvgAlertsValue `json:"values"`
}

type AvgEventsPerWeek struct {
	ID    TimeframeID    `json:"id"bson:"_id"`
	Value AvgEventsValue `json:"value"bson:"value"`
}

type AvgEventsValue struct {
	AvgEvents       float64 `json:"avg_events"bson:"avg_events"`
	TotalEvents     int64   `json:"total_events"bson:"total_events"`
	TotalRank       int64   `json:"total_rank"bson:"total_rank"`
	AvgRank         float64 `json:"avg_rank"bson:"avg_rank"`
	TotalTimeLapsed int64   `json:"total_time_lapsed"bson:"total_time_lapsed"`
	AvgTimeLapsed   float64 `json:"avg_time_lapsed"bson:"avg_time_lapsed"`
}

type AvgEventsReport struct {
	Sender string                    `json:"sender"`
	Values map[string]AvgEventsValue `json:"values"`
}

type EventAttendance struct {
	ID    TimeframeID      `json:"id"bson:"_id"`
	Value EventAttendValue `json:"value"bson:"value"`
}

type EventAttendValue struct {
	Attendance float64 `json:"attendance"bson:"attendance"`
	Events     int     `json:"total_events"bson:"total_events"`
}

type EventAttendReport struct {
	Sender string                      `json:"sender"`
	Values map[string]EventAttendValue `json:"values"`
}

// SenderInfo is a struct for containing the Sender Info report for the past 3 months.
type SenderInfo struct {
	AlertsPerWeek []AlertWeekInfo `json:"alerts_per_week"bson:"alerts_per_week"`
	EventsPerWeek []EventWeekInfo `json:"events_per_week"bson:"events_per_week"`
	TagArray      []TagInfo       `json:"tag_array"bson:"tag_array"`
	AlertsPerHour []int64         `json:"alerts_per_hour"bson:"alerts_per_hour"`
}

// AlertWeekInfo holds the 'alerts per week' counts for a particular sender.
type AlertWeekInfo struct {
	Id    WeekInfoID `json:"_id" bson:"_id"`
	Value struct {
		Alerts int              `json:"alerts"`
		TagMap map[string]int64 `json:"tag_map"bson:"tag_map"`
	} `json:"value"`
}

// EventWeekInfo holds the 'events per week' counts for a particular sender.
type EventWeekInfo struct {
	Id    WeekInfoID `json:"_id" bson:"_id"`
	Value struct {
		TotalEvents     int64   `json:"total_events"bson:"total_events"`
		TotalRank       int64   `json:"total_rank"bson:"total_rank"`
		AvgRank         float64 `json:"avg_rank"bson:"avg_rank"`
		TotalTimeLapsed int64   `json:"total_time_lapsed"bson:"total_time_lapsed"`
		AvgTimeLapsed   float64 `json:"avg_time_lapsed"bson:"avg_time_lapsed"`
	} `json:"value"bson:"value"`
}

// WeekInfoID is used as a helper to pull 'alerts per week' and
// 'events per week' info from the database.
type WeekInfoID struct {
	Sender    string    `json:"sender"`
	WeekStart time.Time `json:"week_start"bson:"week_start"`
}

// TagInfo is used to hold the 'top tags' for a particular Sender.
type TagInfo struct {
	Tag       string `json:"tag"`
	Frequency int64  `json:"frequency"`
}

// TagArrayResult is used as a helper to pull 'top tags' information
// from the database.
type TagArrayResult struct {
	Value struct {
		TagArray []TagInfo `bson:"tag_array"`
	}
}

// AlertsPerHourResult is used as a helper struct to pull 'alerts per hour'
// information from the database.
type AlertsPerHourResult struct {
	ID    WeekInfoID `json:"_id" bson:"_id"`
	Value struct {
		Hours map[string]int64 `json:"hours"`
	}
}

func GetAlertsPerWeek(db *mgo.Database) ([]AvgAlertsReport, error) {
	coll := db.C("avg_alerts_per_week_by_sender")
	iter := coll.Find(nil).Iter()

	var results []AvgAlertsReport
	resultMap := make(map[string]map[string]AvgAlertsValue)
	var result AvgAlertsPerWeek
	for iter.Next(&result) {
		if _, exists := resultMap[result.ID.Sender]; exists {
			resultMap[result.ID.Sender][result.ID.Timeframe] = result.Value
		} else {
			resultMap[result.ID.Sender] = map[string]AvgAlertsValue{result.ID.Timeframe: result.Value}
		}
	}

	err := iter.Close()
	if err != nil {
		return results, err
	}

	for sender, values := range resultMap {
		result := AvgAlertsReport{sender, values}
		results = append(results, result)
	}
	return results, err
}

func GetEventsPerWeek(db *mgo.Database) ([]AvgEventsReport, error) {
	coll := db.C("avg_events_per_week_by_sender")
	iter := coll.Find(nil).Iter()

	var results []AvgEventsReport
	resultMap := make(map[string]map[string]AvgEventsValue)
	var result AvgEventsPerWeek
	for iter.Next(&result) {
		if _, exists := resultMap[result.ID.Sender]; exists {
			resultMap[result.ID.Sender][result.ID.Timeframe] = result.Value
		} else {
			resultMap[result.ID.Sender] = map[string]AvgEventsValue{result.ID.Timeframe: result.Value}
		}
	}

	err := iter.Close()
	if err != nil {
		return results, err
	}

	for sender, values := range resultMap {
		result := AvgEventsReport{sender, values}
		results = append(results, result)
	}

	return results, err
}

func GetEventAttendance(db *mgo.Database) ([]EventAttendReport, error) {
	coll := db.C("sender_event_attendance")
	iter := coll.Find(nil).Iter()

	var results []EventAttendReport
	resultMap := make(map[string]map[string]EventAttendValue)
	var result EventAttendance
	for iter.Next(&result) {
		if _, exists := resultMap[result.ID.Sender]; exists {
			resultMap[result.ID.Sender][result.ID.Timeframe] = result.Value
		} else {
			resultMap[result.ID.Sender] = map[string]EventAttendValue{result.ID.Timeframe: result.Value}
		}
	}

	err := iter.Close()
	if err != nil {
		return results, err
	}

	for sender, values := range resultMap {
		result := EventAttendReport{sender, values}
		results = append(results, result)
	}

	return results, err
}

// FindSenderInfo returns the full Sender Info report for the given sender over the past 3 months.
func FindSenderInfo(db *mgo.Database, sender string) (senderInfo SenderInfo, err error) {
	tfquery := bson.M{"_id.sender": sender, "_id.timeframe": "12months"}
	query := bson.M{"_id.sender": sender}
	sort := "_id.week_start"

	var tagResult TagArrayResult
	avgAlertsPerWeek := db.C("avg_alerts_per_week_by_sender")
	err = avgAlertsPerWeek.Find(tfquery).One(&tagResult)
	if err != nil {
		return
	}
	senderInfo.TagArray = tagResult.Value.TagArray

	var perHourResult AlertsPerHourResult
	alertsPerHour := db.C("sender_alerts_per_hour")
	err = alertsPerHour.Find(query).One(&perHourResult)
	if err != nil {
		return
	}
	// translate from map[string]int to []int
	for hour := 0; hour < 24; hour++ {
		senderInfo.AlertsPerHour = append(senderInfo.AlertsPerHour, perHourResult.Value.Hours[fmt.Sprintf("%d", hour)])
	}

	alertsPerWeek := db.C("alerts_per_week_by_sender")
	err = alertsPerWeek.Find(query).Sort(sort).All(&senderInfo.AlertsPerWeek)
	if err != nil {
		return
	}

	eventsPerWeek := db.C("events_per_week_by_sender")
	err = eventsPerWeek.Find(query).Sort(sort).All(&senderInfo.EventsPerWeek)
	if err != nil {
		return
	}

	return
}
