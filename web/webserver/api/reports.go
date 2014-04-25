package api

import (
	"fmt"
	"time"

	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
)

// TotalSummaryReport is a struct to hold overall News Alert & News Event statistics
// for the past week and for the past 3 months.
type TotalSummaryReport struct {
	Sender           string  `json:"sender"`
	AvgAlertsPerWeek float64 `json:"avg_alerts_per_week"bson:"avg_alerts_per_week"`
	AlertsLastWeek   int64   `json:"alerts_last_week"bson:"alerts_last_week"`
	AvgEventsPerWeek float64 `json:"avg_events_per_week"bson:"avg_events_per_week"`
	EventsLastWeek   int64   `json:"events_last_week"bson:"events_last_week"`
}

// SenderSummaryReport is a struct to hold News Alert & News Event statistics
// for a specific Sender Newshound tracks for the past week and for the past 3 months.
type SenderSummaryReport struct {
	Sender                        string  `json:"sender"`
	AvgAlertsPerWeek              float64 `json:"avg_alerts_per_week"bson:"avg_alerts_per_week"`
	AlertsLastWeek                int64   `json:"alerts_last_week"bson:"alerts_last_week"`
	AvgEventsPerWeek              float64 `json:"avg_events_per_week"bson:"avg_events_per_week"`
	EventsLastWeek                int64   `json:"events_last_week"bson:"events_last_week"`
	EventAttendance               float64 `json:"event_attendance"bson:"event_attendance"`
	EventAttendanceRatingLastWeek float64 `json:"event_attendance_rating_last_week"bson:"event_attendance_rating_last_week"`
	AvgEventRankLastWeek          float64 `json:"avg_event_rank_last_week"bson:"avg_event_rank_last_week"`
	EventAttendanceRating         float64 `json:"event_attendance_rating"bson:"event_attendance_rating"`
	EventAttendanceLast_Week      int64   `json:"event_attendance_last_week"bson:"event_attendance_last_week"`
	AvgEventRank                  float64 `json:"avg_event_rank"bson:"avg_event_rank"`
	AvgEventArrival               float64 `json:"avg_event_arrival"bson:"avg_event_arrival"`
	AvgEventArrivalLastWeek       float64 `json:"avg_event_arrival_last_week"bson:"avg_event_arrival_last_week"`
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
		TotalEvents     int64            `json:"total_events"bson:"total_events"`
		TotalRank       int64            `json:"total_rank"bson:"total_rank"`
		AvgRank         float64          `json:"avg_rank"bson:"avg_rank"`
		TotalTimeLapsed int64            `json:"total_time_lapsed"bson:"total_time_lapsed"`
		AvgTimeLapsed   float64          `json:"avg_time_lapsed"bson:"avg_time_lapsed"`
		TagMap          map[string]int64 `json:"tag_map"bson:"tag_map"`
	} `json:"value"`
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

// GetTotalSummaryReport returns the current Totals Report that summarizes all Alerts/Events in the past
// week and the last 3 months.
func GetTotalSummaryReport(db *mgo.Database) (totalReport TotalSummaryReport, err error) {
	c := db.C("news_report_by_sender")
	err = c.Find(bson.M{"sender": "total"}).One(&totalReport)
	if err != nil {
		return
	}

	return
}

// GetSenderSummaryReport returns the current Sender Report that summarizes all Alerts/Events in the past
// week and the last 3 months for each Sender.
func GetSenderSummaryReport(db *mgo.Database) (senderReports []SenderSummaryReport, err error) {
	c := db.C("news_report_by_sender")
	err = c.Find(bson.M{"sender": bson.M{"$ne": "total"}}).Sort("avg_alerts_per_week").All(&senderReports)
	if err != nil {
		return
	}

	return
}

// FindSenderInfo returns the full Sender Info report for the given sender over the past 3 months.
func FindSenderInfo(db *mgo.Database, sender string) (senderInfo SenderInfo, err error) {
	query := bson.M{"_id.sender": sender}
	sort := "_id.week_start"

	var tagResult TagArrayResult
	avgAlertsPerWeek := db.C("avg_alerts_per_week_by_sender")
	err = avgAlertsPerWeek.Find(query).One(&tagResult)
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
