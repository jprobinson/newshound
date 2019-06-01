// api package contains the Newshound API.
package api

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/NYTimes/gizmo/server"
	"github.com/gorilla/mux"
	"github.com/jprobinson/go-utils/web"
	"gopkg.in/mgo.v2"
)

// findAlertsByDate is an http.Handler that will expect a 'start' and 'end' date in the URL
// and will return a list of News Alerts that occured in that timeframe.
func (s *service) findAlertsByDate(r *http.Request) (int, interface{}, error) {
	vars := server.Vars(r)
	startTime, endTime, err := web.ParseDateRangeFullDay(vars)
	if err != nil {
		return http.StatusBadRequest, "bad request", nil
	}

	sess, db := s.getDB()
	defer sess.Close()

	alerts, err := FindAlertsByDate(db, startTime, endTime)
	if err != nil {
		log.Printf("unable to access alerts by date - %s", err)
		return http.StatusInternalServerError, "server error", nil
	}

	return http.StatusOK, alerts, nil
}

// findOrderedAlerts is an http.Handler that will expect a comma delmited list of News Alert IDs
// in the URL and will return a chronologically ordered of those News Alerts' information.
func (s *service) findOrderedAlerts(r *http.Request) (int, interface{}, error) {
	vars := server.Vars(r)
	alertIDs := strings.Split(vars["alert_ids"], ",")

	sess, db := s.getDB()
	defer sess.Close()

	alerts, err := FindOrderedAlerts(db, alertIDs)
	if err != nil {
		log.Printf("unable to access alerts - %s", err)
		return http.StatusInternalServerError, "server error", nil
	}

	return http.StatusOK, alerts, nil
}

// findAlert is an http.Handler that expects a News Alert ID in the URL and if the
// alert exists, it will return it's information.
func (s *service) findAlert(r *http.Request) (int, interface{}, error) {
	vars := server.Vars(r)
	alertID := vars["alert_id"]

	sess, db := s.getDB()
	defer sess.Close()

	alert, err := FindAlertByID(db, alertID)
	if err != nil {
		log.Printf("unable to access alert - %s", err)
		return http.StatusInternalServerError, "server error", nil
	}

	return http.StatusOK, alert, nil
}

// findAlertHtml is an http.Handler that expects a News Alert ID in the URL and if the
// alert exists, it will return it's HTML with a 'text/html' content-type.
func (s *service) findAlertHtml(w http.ResponseWriter, r *http.Request) {
	alertID := server.Vars(r)["alert_id"]

	s, db := n.getDB()
	defer s.Close()

	alertHtml, err := FindAlertHtmlByID(db, alertID)
	if err != nil {
		log.Printf("unable to access alert HTML - %s", err)
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}

	fmt.Fprint(w, alertHtml)
}

// eventFeed is an http.Handler that will expect a 'start' and 'end' date in the URL
// and will return a list of News Events that occured in that timeframe order by time desc.
func (s *service) eventFeed(r *http.Request) (int, interface{}, error) {
	setCommonHeaders(w, r, "")
	vars := mux.Vars(r)
	startTime, endTime, err := web.ParseDateRangeFullDay(vars)
	if err != nil {
		web.ErrorResponse(w, err, http.StatusBadRequest)
		return
	}

	sess, db := s.getDB()
	defer sess.Close()

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
func (s *service) findEventsByDate(r *http.Request) (int, interface{}, error) {
	setCommonHeaders(w, r, "")
	vars := mux.Vars(r)
	startTime, endTime, err := web.ParseDateRangeFullDay(vars)
	if err != nil {
		return http.StatusBadRequest, "bad request", nil
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
func (s *service) findEvent(r *http.Request) {
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
func (s *service) getAlertsPerWeek(r *http.Request) (int, interface{}, error) {
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

func (s *service) getEventAttendance(r *http.Request) (int, interface{}, error) {
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
func (s *service) getEventsPerWeek(r *http.Request) (int, interface{}, error) {
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
func (s *service) findSenderInfo(r *http.Request) (int, interface{}, error) {
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

func (s *service) getDB() (*mgo.Session, *mgo.Database) {
	s := s.sess.Copy()
	return s, s.DB("newshound")
}
