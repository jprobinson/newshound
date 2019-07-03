package api

import (
	"log"
	"net/http"

	sdpropagation "contrib.go.opencensus.io/exporter/stackdriver/propagation"
	"github.com/NYTimes/gizmo/observe"
	"github.com/NYTimes/gizmo/server"
	"github.com/pkg/errors"
	"github.com/rs/cors"
	"go.opencensus.io/plugin/ochttp"
	"gopkg.in/mgo.v2"
)

var _ server.MixedService = &service{}

func NewService() (server.MixedService, error) {
	observe.RegisterAndObserveGCP(func(err error) {
		log.Printf("exporter client encountered an error: %s", err)
	})
	cfg := NewConfig()
	sess, err := cfg.MgoSession()
	if err != nil {
		return nil, errors.Wrap(err, "unable to init mgo")
	}
	return &service{sess: sess}, nil
}

type service struct {
	sess *mgo.Session
}

func (s *service) Prefix() string {
	return ""
}

func (s *service) Middleware(h http.Handler) http.Handler {
	return &ochttp.Handler{
		Handler:     cors.Default().Handler(h),
		Propagation: &sdpropagation.HTTPFormat{},
	}
}

func (s *service) Endpoints() map[string]map[string]http.HandlerFunc {
	return map[string]map[string]http.HandlerFunc{
		"/svc/newshound-api/v1/alert_html/{alert_id}": {
			"GET": s.findAlertHTML,
		},
	}
}

func (s *service) JSONEndpoints() map[string]map[string]server.JSONEndpoint {
	return map[string]map[string]server.JSONEndpoint{
		"/svc/newshound-api/v1/find_alerts/{start}/{end}": {
			"GET": s.findAlertsByDate,
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
			"GET": s.getEventAttendance,
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
