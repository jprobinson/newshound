// api package contains the Newshound API.
package api

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/jprobinson/go-utils/web"
	"labix.org/v2/mgo"
)

var ErrDB = errors.New("problems accessing database")

// NewshoundAPI is a struct that keeps a handle on the mgo session
type NewshoundAPI struct {
	session *mgo.Session
}

// NewNewshoundAPI creates a new NewshoundAPI struct to run the newshound API.
func NewNewshoundAPI(conn string, user string, pw string) *NewshoundAPI {
	// make conn pass it to data
	session, err := mgo.Dial(conn)
	if err != nil {
		log.Fatalf("Unable to connect to newshound db! - %s", err)
	}

	db := session.DB("newshound")
	err = db.Login(user, pw)
	if err != nil {
		log.Fatalf("Unable to connect to newshound db! - %s", err)
	}
	session.SetMode(mgo.Eventual, true)
	return &NewshoundAPI{session}
}

// UrlPrefix is a function meant to implement the PrefixHandler interface for paperboy-api.
func (n NewshoundAPI) UrlPrefix() string {
	return "/svc/newshound-api/v1"
}

// Handle is a function meant to implement the PrefixHandler interface for paperboy-api.
// It accepts a mux.Router and adds all the required Handlers for the Newshound API.
func (n NewshoundAPI) Handle(subRouter *mux.Router) {
	// ALERTS
	subRouter.HandleFunc("/find_alerts/{start}/{end}", n.findAlertsByDate).Methods("GET")
	subRouter.HandleFunc("/ordered_alerts/{alert_ids}", n.findOrderedAlerts).Methods("GET")
	subRouter.HandleFunc("/alert/{alert_id}", n.findAlert).Methods("GET")
	subRouter.HandleFunc("/alert_html/{alert_id}", n.findAlertHtml).Methods("GET")

	// EVENTS
	subRouter.HandleFunc("/find_events/{start}/{end}", n.findEventsByDate).Methods("GET")
	subRouter.HandleFunc("/event_feed/{start}/{end}", n.eventFeed).Methods("GET")
	subRouter.HandleFunc("/event/{event_id}", n.findEvent).Methods("GET")

	// REPORTS
	subRouter.HandleFunc("/report/alerts_per_week", n.getAlertsPerWeek).Methods("GET")
	subRouter.HandleFunc("/report/events_per_week", n.getEventsPerWeek).Methods("GET")
	subRouter.HandleFunc("/report/event_attendance", n.getEventAttendance).Methods("GET")
	subRouter.HandleFunc("/sender_info/{sender}", n.findSenderInfo).Methods("GET")
}

// findAlertsByDate is an http.Handler that will expect a 'start' and 'end' date in the URL
// and will return a list of News Alerts that occured in that timeframe.
func (n NewshoundAPI) findAlertsByDate(w http.ResponseWriter, r *http.Request) {
	setCommonHeaders(w, r, "")
	vars := mux.Vars(r)
	startTime, endTime, err := web.ParseDateRangeFullDay(vars)
	if err != nil {
		web.ErrorResponse(w, err, http.StatusBadRequest)
		return
	}

	s, db := n.getDB()
	defer s.Close()

	alerts, err := FindAlertsByDate(db, startTime, endTime)
	if err != nil {
		log.Printf("Unable to access alerts by date! - %s", err.Error())
		web.ErrorResponse(w, ErrDB, http.StatusBadRequest)
		return
	}

	fmt.Fprint(w, web.JsonResponseWrapper{Response: alerts})
}

// findOrderedAlerts is an http.Handler that will expect a comma delmited list of News Alert IDs
// in the URL and will return a chronologically ordered of those News Alerts' information.
func (n NewshoundAPI) findOrderedAlerts(w http.ResponseWriter, r *http.Request) {
	setCommonHeaders(w, r, "")
	vars := mux.Vars(r)
	alertIDs := strings.Split(vars["alert_ids"], ",")

	s, db := n.getDB()
	defer s.Close()

	alerts, err := FindOrderedAlerts(db, alertIDs)
	if err != nil {
		log.Printf("Unable to access alerts by multiple alert_id's! - %s", err.Error())
		web.ErrorResponse(w, ErrDB, http.StatusBadRequest)
		return
	}

	fmt.Fprint(w, web.JsonResponseWrapper{Response: alerts})
}

// findAlert is an http.Handler that expects a News Alert ID in the URL and if the
// alert exists, it will return it's information.
func (n NewshoundAPI) findAlert(w http.ResponseWriter, r *http.Request) {
	setCommonHeaders(w, r, "")
	vars := mux.Vars(r)
	alertID := vars["alert_id"]

	s, db := n.getDB()
	defer s.Close()

	alert, err := FindAlertByID(db, alertID)
	if err != nil {
		log.Printf("Unable to access alerts by alert_id! - %s", err.Error())
		web.ErrorResponse(w, ErrDB, http.StatusBadRequest)
		return
	}

	fmt.Fprint(w, web.JsonResponseWrapper{Response: alert})
}

// findAlertHtml is an http.Handler that expects a News Alert ID in the URL and if the
// alert exists, it will return it's HTML with a 'text/html' content-type.
func (n NewshoundAPI) findAlertHtml(w http.ResponseWriter, r *http.Request) {
	setCommonHeaders(w, r, "text/html")
	vars := mux.Vars(r)
	alertID := vars["alert_id"]

	s, db := n.getDB()
	defer s.Close()

	alertHtml, err := FindAlertHtmlByID(db, alertID)
	if err != nil {
		log.Printf("Unable to access alerts by alert_id! - %s", err.Error())
		web.ErrorResponse(w, ErrDB, http.StatusBadRequest)
		return
	}

	fmt.Fprint(w, alertHtml)
}

// eventFeed is an http.Handler that will expect a 'start' and 'end' date in the URL
// and will return a list of News Events that occured in that timeframe order by time desc.
func (n NewshoundAPI) eventFeed(w http.ResponseWriter, r *http.Request) {
	setCommonHeaders(w, r, "")
	vars := mux.Vars(r)
	startTime, endTime, err := web.ParseDateRangeFullDay(vars)
	if err != nil {
		web.ErrorResponse(w, err, http.StatusBadRequest)
		return
	}

	s, db := n.getDB()
	defer s.Close()

	events, err := FindEventsByDateReverse(db, startTime, endTime)
	if err != nil {
		log.Printf("Unable to access events by date! - %s", err.Error())
		web.ErrorResponse(w, ErrDB, http.StatusBadRequest)
		return
	}

	fmt.Fprint(w, web.JsonResponseWrapper{Response: events})
}

// findEventsByDate is an http.Handler that will expect a 'start' and 'end' date in the URL
// and will return a list of News Events that occured in that timeframe.
func (n NewshoundAPI) findEventsByDate(w http.ResponseWriter, r *http.Request) {
	setCommonHeaders(w, r, "")
	vars := mux.Vars(r)
	startTime, endTime, err := web.ParseDateRangeFullDay(vars)
	if err != nil {
		web.ErrorResponse(w, err, http.StatusBadRequest)
		return
	}

	s, db := n.getDB()
	defer s.Close()

	events, err := FindEventsByDate(db, startTime, endTime)
	if err != nil {
		log.Printf("Unable to access events by date! - %s", err.Error())
		web.ErrorResponse(w, ErrDB, http.StatusBadRequest)
		return
	}

	fmt.Fprint(w, web.JsonResponseWrapper{Response: events})
}

// findEvent is an http.Handler that expects a News Event ID in the URL and if the
// event exists, it will return it's information.
func (n NewshoundAPI) findEvent(w http.ResponseWriter, r *http.Request) {
	setCommonHeaders(w, r, "")
	vars := mux.Vars(r)
	eventID := vars["event_id"]

	s, db := n.getDB()
	defer s.Close()

	event, err := FindEventByID(db, eventID)
	if err != nil {
		log.Printf("Unable to access event by event_id! - %s", err.Error())
		web.ErrorResponse(w, ErrDB, http.StatusBadRequest)
		return
	}

	fmt.Fprint(w, web.JsonResponseWrapper{Response: event})
}
func (n NewshoundAPI) getAlertsPerWeek(w http.ResponseWriter, r *http.Request) {
	setCommonHeaders(w, r, "")
	s, db := n.getDB()
	defer s.Close()

	sendersReport, err := GetAlertsPerWeek(db)
	if err != nil {
		log.Printf("Unable to retrieve sender alerts per week! - %s", err.Error())
		web.ErrorResponse(w, ErrDB, http.StatusBadRequest)
		return
	}

	fmt.Fprint(w, web.JsonResponseWrapper{Response: sendersReport})
}

func (n NewshoundAPI) getEventAttendance(w http.ResponseWriter, r *http.Request) {
	setCommonHeaders(w, r, "")
	s, db := n.getDB()
	defer s.Close()

	sendersReport, err := GetEventAttendance(db)
	if err != nil {
		log.Printf("Unable to retrieve sender event attendance! - %s", err.Error())
		web.ErrorResponse(w, ErrDB, http.StatusBadRequest)
		return
	}

	fmt.Fprint(w, web.JsonResponseWrapper{Response: sendersReport})
}
func (n NewshoundAPI) getEventsPerWeek(w http.ResponseWriter, r *http.Request) {
	setCommonHeaders(w, r, "")
	s, db := n.getDB()
	defer s.Close()

	sendersReport, err := GetEventsPerWeek(db)
	if err != nil {
		log.Printf("Unable to retrieve sender events per week! - %s", err.Error())
		web.ErrorResponse(w, ErrDB, http.StatusBadRequest)
		return
	}

	fmt.Fprint(w, web.JsonResponseWrapper{Response: sendersReport})
}

// findSenderInfo is an http.Handler that will expect a Sender name in the URL and if
// the sender exists, it will return the Sender Info report for the past 3 months.
func (n NewshoundAPI) findSenderInfo(w http.ResponseWriter, r *http.Request) {
	setCommonHeaders(w, r, "")
	vars := mux.Vars(r)
	sender := vars["sender"]

	s, db := n.getDB()
	defer s.Close()

	senderInfo, err := FindSenderInfo(db, sender)
	if err != nil {
		log.Printf("Unable to retrieve sender info report! - %s", err.Error())
		web.ErrorResponse(w, ErrDB, http.StatusBadRequest)
		return
	}

	fmt.Fprint(w, web.JsonResponseWrapper{Response: senderInfo})
}

// setCommondHeaders is a utility function to set the 'Access-Control-Allow-Origin' to * and
// set the Content-Type to the given input. If not Content-Type is given, it defaults to
// 'application/json'.
func setCommonHeaders(w http.ResponseWriter, r *http.Request, contentType string) {
	origin := r.Header.Get("Origin")
	if len(origin) == 0 {
		origin = "*"
	}
	w.Header().Set("Access-Control-Allow-Origin", origin)
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, *")
	w.Header().Set("Cache-Control", "no-cache")
	if len(contentType) == 0 {
		w.Header().Set("Content-Type", web.JsonContentType)
	} else {
		w.Header().Set("Content-Type", contentType)
	}
}

func (n NewshoundAPI) getDB() (*mgo.Session, *mgo.Database) {
	s := n.session.Copy()
	return s, s.DB("newshound")
}
