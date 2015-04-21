package fetch

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"log"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/bitly/go-nsq"
	"github.com/jprobinson/newshound"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var (
	eventTimeframe = 1 * time.Hour

	minOccurPerc = 0.4

	minSenders = 2

	minAlerts = 3

	minLikeTags = 2
)

func minOccurances(alertCount int) int {
	return int(math.Max(math.Ceil(float64(alertCount)*minOccurPerc), 2.0))
}

func EventRefresh(na *mgo.Collection, ne *mgo.Collection, eventTime time.Time, producer *nsq.Producer) error {
	// find all alerts within a event timeframe of the given time and refresh the events
	start := eventTime.Add(-eventTimeframe)
	end := eventTime.Add(eventTimeframe)
	query := bson.M{"timestamp": bson.M{"$gte": start, "$lte": end}}
	var eligible []newshound.NewsAlert
	err := na.Find(query).All(&eligible)
	if err != nil {
		return err
	}

	for _, alert := range eligible {
		if err = UpdateEvents(na, ne, alert, producer); err != nil {
			return err
		}
	}

	return nil
}

func UpdateEvents(na *mgo.Collection, ne *mgo.Collection, a newshound.NewsAlert, producer *nsq.Producer) error {

	cluster, tags, err := findLikeAlertCluster(na, a)
	if err != nil {
		return fmt.Errorf("unable to create possible alert cluster for event: %s", err)
	}

	// at least 3 alerts for significance
	if len(cluster) <= 2 {
		return nil
	}

	// grab any events related to the alerts we've got (where alertID in $alerts)
	var existingEvents []newshound.NewsEvent
	query := bson.M{"news_alerts.alert_id": bson.M{"$in": cluster}}
	err = ne.Find(query).All(&existingEvents)

	// merge any existing alerts into ours, reuse id if possible
	eventID, newID, eventUpdate, alertIDs, eventTags, staleEventIDs := mergeEvents(cluster, tags, existingEvents)

	// get the data of all the alerts
	var nas []newshound.NewsAlert
	err = na.Find(bson.M{"_id": bson.M{"$in": alertIDs}}).All(&nas)
	if err != nil {
		return err
	}
	// we need at least 3 alerts total
	if len(nas) <= 2 {
		//log.Printf("does not have enough alerts! %#v", eventTags)
		return nil
	}

	// verify it has the min sender count
	if !hasMinSenders(nas) {
		//log.Printf("does not have enough senders! %#v", eventTags)
		return nil
	}

	// create the event (all the metrics and sorting and whatnot and save it
	event := NewNewsEvent(eventID, nas, eventTags)
	log.Printf("event found with %d alerts and tags: %#v", len(event.NewsAlerts), event.Tags)
	_, err = ne.UpsertId(event.ID, event)
	if err != nil {
		return err
	}

	// emit event notification if new event
	if producer != nil {
		if newID {
			var buff bytes.Buffer
			err = gob.NewEncoder(&buff).Encode(&event)
			if err != nil {
				log.Print("unable to gob event: ", err)
			} else {
				if err = producer.Publish(newshound.NewsEventTopic, buff.Bytes()); err != nil {
					log.Print("unable to publish event: ", err)
				}
			}
		}
		if !newID && eventUpdate {
			var buff bytes.Buffer
			err = gob.NewEncoder(&buff).Encode(&event)
			if err != nil {
				log.Print("unable to gob event: ", err)
			} else {
				if err = producer.Publish(newshound.NewsEventUpdatesTopic, buff.Bytes()); err != nil {
					log.Print("unable to publish event update: ", err)
				}
			}
		}
	}

	// clean up stale events
	if len(staleEventIDs) > 0 {
		_, err = ne.RemoveAll(bson.M{"_id": bson.M{"$in": staleEventIDs}})
	}
	return err
}

func hasMinSenders(alerts []newshound.NewsAlert) bool {
	senders := map[string]struct{}{}
	for _, alert := range alerts {
		senders[alert.Sender] = struct{}{}
	}
	// quit if we dont have enough senders
	return len(senders) >= minSenders
}

func NewNewsEvent(id bson.ObjectId, alerts []newshound.NewsAlert, eventTags []string) newshound.NewsEvent {
	// sort by timestamp
	sort.Sort(naByTimestamp(alerts))

	// grab our start n end since we're sorted
	start := alerts[0].Timestamp
	end := alerts[len(alerts)-1].Timestamp

	// find the sentence that reaches the highest score first
	// and also add our alerts along the way!
	topCount := 0
	topSender := ""
	topSentence := ""
	var eas []newshound.NewsEventAlert
	for order, a := range alerts {
		for _, s := range a.Sentences {
			alertTagCount := 0
			for _, phrase := range s.Phrases {
				for _, tag := range eventTags {
					// increment the score for any tag/phrase intersection
					if strings.EqualFold(tag, phrase) {
						alertTagCount++
					}
				}
			}

			if alertTagCount > topCount {
				topSender = a.Sender
				topSentence = s.Value
				topCount = alertTagCount
			}
		}

		ea := newshound.NewsEventAlert{
			AlertID:     a.ID,
			InstanceID:  a.InstanceID,
			ArticleUrl:  a.ArticleUrl,
			Sender:      a.Sender,
			Tags:        a.Tags,
			Subject:     a.Subject,
			TopSentence: a.TopSentence,
			TimeLapsed:  int64(a.Timestamp.Sub(start).Seconds()),
			Order:       int64(order),
		}
		eas = append(eas, ea)
	}
	sort.Strings(eventTags)
	return newshound.NewsEvent{
		ID:          id,
		Tags:        eventTags,
		EventStart:  start,
		EventEnd:    end,
		NewsAlerts:  eas,
		TopSentence: topSentence,
		TopSender:   topSender,
	}
}

type naByTimestamp []newshound.NewsAlert

func (n naByTimestamp) Len() int           { return len(n) }
func (n naByTimestamp) Swap(i, j int)      { n[i], n[j] = n[j], n[i] }
func (n naByTimestamp) Less(i, j int) bool { return n[j].Timestamp.After(n[i].Timestamp) }

func mergeEvents(alerts []bson.ObjectId, tags []string, events []newshound.NewsEvent) (eventID bson.ObjectId, newID bool, eventUpdated bool, eventAlerts []bson.ObjectId, eventTags []string, staleEventIDs []bson.ObjectId) {
	originalEventSize := 0
	// add IDs of new event cluster
	idSet := map[string]struct{}{}
	for _, alert := range alerts {
		idSet[alert.Hex()] = struct{}{}
	}
	// add tags to event set
	tagSet := map[string]struct{}{}
	for _, tag := range tags {
		tagSet[tag] = struct{}{}
	}

	// default to way in future to grab oldest date
	oldest := time.Now().Add(10 * 24 * 365 * time.Hour)
	// grab all event IDs so we can remove stale ones afterwards
	var eventIDs []bson.ObjectId

	// add all IDs and tags to sets
	for _, event := range events {
		for _, alert := range event.NewsAlerts {
			idSet[alert.AlertID.Hex()] = struct{}{}
		}
		for _, tag := range event.Tags {
			tagSet[tag] = struct{}{}
		}
		// grab the id of the oldest event
		if oldest.After(event.EventStart) {
			oldest = event.EventStart
			eventID = event.ID
			originalEventSize = len(event.NewsAlerts)
		}
		eventIDs = append(eventIDs, event.ID)
	}

	// remove the event we're going to keep
	for _, eID := range eventIDs {
		if eID.Hex() != eventID.Hex() {
			staleEventIDs = append(staleEventIDs, eID)
		}
	}

	// make the event ID if we dont have one
	if len(eventID.Hex()) == 0 {
		newID = true
		eventID = bson.NewObjectId()
	}

	// convert all ids to objectids
	for id, _ := range idSet {
		eventAlerts = append(eventAlerts, bson.ObjectIdHex(id))
	}

	// generate tag output list
	for tag, _ := range tagSet {
		eventTags = append(eventTags, tag)
	}

	eventUpdated = len(eventAlerts) > originalEventSize
	return eventID, newID, eventUpdated, eventAlerts, eventTags, staleEventIDs
}

func findLikeAlertCluster(na *mgo.Collection, a newshound.NewsAlert) (alerts []bson.ObjectId, tags []string, err error) {
	var possible []newshound.NewsAlert
	// find any alerts in the eventTimeframe
	possible, err = findPossibleLikeAlerts(na, a)
	if err != nil {
		return alerts, tags, err
	}

	// build tag map around main alert's tags
	tagCounts := buildTagCounts(a.Tags, possible)

	// calc min tag limit
	minOccurs := minOccurances(len(possible))

	// filter out any tags that do not meet the limit
	for tag, count := range tagCounts {
		if count <= minOccurs {
			delete(tagCounts, tag)
		}
	}

	// we need at least 2 tags for an event
	if len(tagCounts) < 2 {
		return alerts, tags, nil
	}

	// make sure main alert goes the the same filtering
	possible = append(possible, a)

	// filter out any alerts that do not have minLikeTags
	for _, alert := range possible {
		likeTags := 0
		tagScore := 0
		for _, tag := range alert.Tags {
			if count, ok := tagCounts[tag]; ok {
				likeTags++
				tagScore += count
			}
		}

		if likeTags >= minLikeTags {
			alerts = append(alerts, alert.ID)
			continue
		}

		if likeTags >= len(possible) {
			//log.Printf("made it on tag count? %d - %d - %d - %d - %s - %s", len(possible), minLikeTags, likeTags, tagScore, alert.Sender, alert.Tags)
			alerts = append(alerts, alert.ID)
			continue
		}

		if float32(tagScore) >= float32(len(possible))*(1.2) {
			//log.Printf("made it on tag score? %d - %d - %d - %d - %s - %s", len(possible), minLikeTags, likeTags, tagScore, alert.Sender, alert.Tags)
			alerts = append(alerts, alert.ID)
			continue
		}
	}

	for tag, _ := range tagCounts {
		tags = append(tags, tag)
	}

	return alerts, tags, nil
}

func findPossibleLikeAlerts(na *mgo.Collection, a newshound.NewsAlert) (possible []newshound.NewsAlert, err error) {
	// find any alerts within a  timeframe
	start := a.Timestamp.Add(-eventTimeframe)
	end := a.Timestamp.Add(eventTimeframe)
	query := bson.M{"timestamp": bson.M{"$gte": start, "$lte": end}, "_id": bson.M{"$ne": a.ID}, "tags": bson.M{"$in": a.Tags}}
	err = na.Find(query).All(&possible)
	if err != nil {
		return possible, err
	}

	return possible, err
}

func buildTagCounts(mainTags []string, alerts []newshound.NewsAlert) (tagCounts map[string]int) {
	tagCounts = map[string]int{}
	// seed the map with the main alert's tags
	for _, tag := range mainTags {
		tagCounts[tag] = 1
	}
	// increment the map for any matches and add any new partial matches
	// N^2 ...better way?
	for _, alert := range alerts {
		for _, tag := range alert.Tags {
			// check if each tag matches any of our seeds
			toIncrement := map[string]struct{}{}
			for existing, _ := range tagCounts {
				// exact equality check
				if strings.EqualFold(existing, tag) {
					toIncrement[existing] = struct{}{}
					break
				}

				// partial match check
				if partialMatch(existing, tag) {
					toIncrement[existing] = struct{}{}
					toIncrement[tag] = struct{}{}
				}
			}

			// increment all elgibile tags
			for toInc, _ := range toIncrement {
				tagCounts[toInc]++
			}
		}
	}

	return tagCounts
}

func partialMatch(a, b string) bool {
	aWrds := strings.Fields(a)
	aLen := len(aWrds)
	bWrds := strings.Fields(b)
	bLen := len(bWrds)
	// if they have same # of words but did
	// not pass equality check, one will not be a partial match of the other
	if aLen == bLen {
		return false
	}
	if aLen > bLen {
		return phrasesMatch(b, a)
	}
	return phrasesMatch(a, b)
}

// checks if partial is any of the words or phrases within whole
func phrasesMatch(partial, whole string) bool {
	// create words of whole
	words := strings.Fields(whole)
	// tag partial apart and put it back together with standard space
	partials := strings.Fields(partial)
	partialWords := len(partials)
	partial = strings.Join(partials, " ")
	// check if partial matches any of whole's phrases
	for i := 0; (i + partialWords) <= len(words); i++ {

		slidingBit := strings.Join(words[i:(i+partialWords)], " ")

		if strings.EqualFold(slidingBit, partial) {
			return true
		}

		if strings.HasSuffix(slidingBit, "'s") {
			if strings.EqualFold(strings.TrimSuffix(slidingBit, "'s"), partial) {
				return true
			}
		}
	}

	return false
}
