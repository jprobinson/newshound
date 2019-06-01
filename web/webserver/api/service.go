package api

import (
	"net/http"

	"github.com/NYTimes/gizmo/server"
	"gopkg.in/mgo.v2"
)

var _ server.MixedService = &service{}

type service struct {
	sess *mgo.Session
}

func (s *service) Prefix() string {
	return "/svc/newshound/v1/"
}

func (s *service) Middleware(h http.Handler) http.Handler {
	return h
}

func (s *service) Endpoints() map[string]map[string]http.HandlerFunc {
	return map[string]map[string]server.HandlerFunc{
		"/svc/newshound-api/v1/alert_html/{alert_id}": {
			"GET": s.findAlertHTML,
		},
	}
}

func (s *service) JSONEndpoints() map[string]map[string]server.JSONEndpoint {
	return map[string]map[string]server.JSONEndpoint{
		"/svc/newshound-api/v1/find_alerts/{start}/{end}": {
			"GET": s.findOrderedAlerts,
		},
		"/svc/newshound-api/v1/ordered_alerts/{alert_ids}": {
			"GET": s.findOrderedAlerts,
		},
		"/svc/newshound-api/v1/alert/{alert_id}": {
			"GET": s.findAlert,
		},
		"/svc/newshound-api/v1/find_events/{start}/{end}": {
			"GET": s.findEventsByDate,
		},
		"/svc/newshound-api/v1/event_feed/{start}/{end}": {
			"GET": s.eventFeed,
		},
		"/svc/newshound-api/v1/event/{event_id}": {
			"GET": s.findEvent,
		},
		"/svc/newshound-api/v1/report/alerts_per_week": {
			"GET": s.getAlertsPerWeek,
		},
		"/svc/newshound-api/v1/report/events_per_week": {
			"GET": s.getEventsPerWeek,
		},
		"/svc/newshound-api/v1/report/event_attendance": {
			"GET": s.getEventsAttendance,
		},
		"/svc/newshound-api/v1/report/sender_info/{sender}": {
			"GET": s.findSenderInfo,
		},
	}
}

func (s *service) JSONMiddleware(e server.JSONEndpoint) server.JSONEndpoint {
	return e
}

type config struct {
	DBURL      string `envconfig:"DB_URL"`
	DBUser     string `envconfig:"DB_USER"`
	DBPassword string `envconfig:"DB_PASSWORD"`
}

func (c *config) MgoSession() (*mgo.Session, error) {
	// make conn pass it to data
	sess, err := mgo.Dial(c.DBURL)
	if err != nil {
		return sess, err
	}

	db := sess.DB("newshound")
	err = db.Login(c.DBUser, c.DBPassword)
	if err != nil {
		return sess, err
	}
	sess.SetMode(mgo.Eventual, true)
	return sess, nil
}
