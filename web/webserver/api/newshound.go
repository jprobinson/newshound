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

var DBErr = errors.New("problems accessing database")

// NewshoundAPI is a struct that keeps a handle on collections
// to serve API data.
type NewshoundAPI struct {
	alerts  NewsAlerts
	events  NewsEvents
	reports NewsReports
}

// NewNewshoundAPI creates a new NewshoundAPI struct to run the newshound API.
func NewNewshoundAPI(conn string, user string, pw string) *NewshoundAPI {
	// make conn pass it to data
	session, err := mgo.Dial(conn)
	if err != nil {
		panic(fmt.Errorf("Unable to connect to newshound db! - %s", err.Error()))
	}
	db := session.DB("newshound")
	err = db.Login(user, pw)
	if err != nil {
		panic(fmt.Errorf("Unable to connect to newshound db! - %s", err.Error()))
	}
	session.SetMode(mgo.Eventual, true)
	return &NewshoundAPI{NewNewsAlerts(db), NewNewsEvents(db), NewNewsReports(db)}
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
	subRouter.HandleFunc("/event/{event_id}", n.findEvent).Methods("GET")

	// REPORTS
	subRouter.HandleFunc("/total_report", n.getTotalReport).Methods("GET")
	subRouter.HandleFunc("/sender_report", n.getSenderReport).Methods("GET")
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

	alerts, err := n.alerts.FindByDate(startTime, endTime)
	if err != nil {
		log.Printf("Unable to access alerts by date! - %s", err.Error())
		web.ErrorResponse(w, DBErr, http.StatusBadRequest)
		return
	}

	fmt.Fprint(w, web.JsonResponseWrapper{Response: alerts})
}

// findOrderedAlerts is an http.Handler that will expect a comma delmited list of News Alert IDs
// in the URL and will return a chronologically ordered of those News Alerts' information.
func (n NewshoundAPI) findOrderedAlerts(w http.ResponseWriter, r *http.Request) {
	setCommonHeaders(w, r, "")
	vars := mux.Vars(r)
	alert_ids := strings.Split(vars["alert_ids"], ",")

	alerts, err := n.alerts.FindOrderedAlerts(alert_ids)
	if err != nil {
		log.Printf("Unable to access alerts by multiple alert_id's! - %s", err.Error())
		web.ErrorResponse(w, DBErr, http.StatusBadRequest)
		return
	}

	fmt.Fprint(w, web.JsonResponseWrapper{Response: alerts})
}

// findAlert is an http.Handler that expects a News Alert ID in the URL and if the
// alert exists, it will return it's information.
func (n NewshoundAPI) findAlert(w http.ResponseWriter, r *http.Request) {
	setCommonHeaders(w, r, "")
	vars := mux.Vars(r)
	alert_id := vars["alert_id"]

	alert, err := n.alerts.FindAlertByID(alert_id)
	if err != nil {
		log.Printf("Unable to access alerts by alert_id! - %s", err.Error())
		web.ErrorResponse(w, DBErr, http.StatusBadRequest)
		return
	}

	fmt.Fprint(w, web.JsonResponseWrapper{Response: alert})
}

// findAlertHtml is an http.Handler that expects a News Alert ID in the URL and if the
// alert exists, it will return it's HTML with a 'text/html' content-type.
func (n NewshoundAPI) findAlertHtml(w http.ResponseWriter, r *http.Request) {
	setCommonHeaders(w, r, "text/html")
	vars := mux.Vars(r)
	alert_id := vars["alert_id"]

	alert_html, err := n.alerts.FindAlertHtmlByID(alert_id)
	if err != nil {
		log.Printf("Unable to access alerts by alert_id! - %s", err.Error())
		web.ErrorResponse(w, DBErr, http.StatusBadRequest)
		return
	}

	fmt.Fprint(w, alert_html)
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

	events, err := n.events.FindByDate(startTime, endTime)
	if err != nil {
		log.Printf("Unable to access events by date! - %s", err.Error())
		web.ErrorResponse(w, DBErr, http.StatusBadRequest)
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

	event, err := n.events.FindEventByID(eventID)
	if err != nil {
		log.Printf("Unable to access event by event_id! - %s", err.Error())
		web.ErrorResponse(w, DBErr, http.StatusBadRequest)
		return
	}

	fmt.Fprint(w, web.JsonResponseWrapper{Response: event})
}

// getTotalReport is an http.Handler that will return overall News Alert & News Event statistics
// for the past week and for the past 3 months.
func (n NewshoundAPI) getTotalReport(w http.ResponseWriter, r *http.Request) {
	setCommonHeaders(w, r, "")

	totals, err := n.reports.GetTotalSummaryReport()
	if err != nil {
		log.Printf("Unable to retrieve total summary report! - %s", err.Error())
		web.ErrorResponse(w, DBErr, http.StatusBadRequest)
		return
	}

	fmt.Fprint(w, web.JsonResponseWrapper{Response: totals})
}

// getTotalReport is an http.Handler that will return a list News Alert & News Event statistics
// for each Sender Newshound tracks for the past week and for the past 3 months.
func (n NewshoundAPI) getSenderReport(w http.ResponseWriter, r *http.Request) {
	setCommonHeaders(w, r, "")

	sendersReport, err := n.reports.GetSenderSummaryReport()
	if err != nil {
		log.Printf("Unable to retrieve sender summary report! - %s", err.Error())
		web.ErrorResponse(w, DBErr, http.StatusBadRequest)
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

	senderInfo, err := n.reports.FindSenderInfo(sender)
	if err != nil {
		log.Printf("Unable to retrieve sender info report! - %s", err.Error())
		web.ErrorResponse(w, DBErr, http.StatusBadRequest)
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
