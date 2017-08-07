package fetch

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"log"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/bitly/go-nsq"
	"github.com/jprobinson/newshound"
)

var (
	eventTimeframe = 1 * time.Hour

	minOccurPerc = 0.4

	minSenders = 2

	minAlerts = 3

	minLikePhrases = 2
)

func minOccurances(alertCount int) int {
	return int(math.Max(math.Ceil(float64(alertCount)*minOccurPerc), 2.0))
}

func EventRefresh(ctx context.Context, db DB, eventTime time.Time, producer *nsq.Producer) error {
	// find all alerts within a event timeframe of the given time and refresh the events
	start := eventTime.Add(-eventTimeframe)
	end := eventTime.Add(eventTimeframe)
	eligible, err := db.FindAlertsByTimeframe(ctx, start, end)
	if err != nil {
		return err
	}

	for _, alert := range eligible {
		if err = UpdateEvents(ctx, db, alert, producer); err != nil {
			return err
		}
	}

	return nil
}

func UpdateEvents(ctx context.Context, db DB, a *newshound.NewsAlert, producer *nsq.Producer) error {
	cluster, tags, err := findLikeAlertCluster(ctx, db, a)
	if err != nil {
		return fmt.Errorf("unable to create possible alert cluster for event: %s", err)
	}

	// at least 3 alerts for significance
	if len(cluster) <= 2 {
		return nil
	}

	// grab any events related to the alerts we've got (where alertID in $alerts)
	existingEvents, err := db.FindEventsByAlertIDs(ctx, cluster)
	if err != nil {
		return err
	}

	// merge any existing alerts into ours, reuse id if possible
	eventID, newID, eventUpdate, alertIDs, eventPhrases, staleEventIDs := mergeEvents(cluster, tags, existingEvents)

	// get the data of all the alerts
	nas, err := db.GetAlertsByID(ctx, alertIDs)
	if err != nil {
		return err
	}

	// we need at least 3 alerts total
	if len(nas) <= 2 {
		//log.Printf("does not have enough alerts! %#v", eventPhrases)
		return nil
	}

	// verify it has the min sender count
	if !hasMinSenders(nas) {
		//log.Printf("does not have enough senders! %#v", eventPhrases)
		return nil
	}

	// create the event (all the metrics and sorting and whatnot and save it
	event := NewNewsEvent(eventID, nas, eventPhrases)
	log.Printf("event found with %d alerts and tags: %#v", len(event.NewsAlerts), event.TopPhrases)

	err = db.UpsertEvent(ctx, event)
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
				if err = producer.Publish(newshound.NewsEventUpdateTopic, buff.Bytes()); err != nil {
					log.Print("unable to publish event update: ", err)
				}
			}
		}
	}

	// clean up stale events
	if len(staleEventIDs) > 0 {
		err = db.DeleteEvents(ctx, staleEventIDs)
	}
	return err
}

func hasMinSenders(alerts []*newshound.NewsAlert) bool {
	senders := map[string]struct{}{}
	for _, alert := range alerts {
		senders[alert.Sender] = struct{}{}
	}
	// quit if we dont have enough senders
	return len(senders) >= minSenders
}

func NewNewsEvent(id int64, alerts []*newshound.NewsAlert, eventPhrases []string) *newshound.NewsEvent {
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
			alertPhraseCount := 0
			for _, phrase := range s.Phrases {
				for _, tag := range eventPhrases {
					// increment the score for any tag/phrase intersection
					if strings.EqualFold(tag, phrase) {
						alertPhraseCount++
					}
				}
			}

			if alertPhraseCount > topCount {
				topSender = a.Sender
				topSentence = s.Value
				topCount = alertPhraseCount
			}
		}

		ea := newshound.NewsEventAlert{
			AlertID:     a.ID,
			ArticleUrl:  a.ArticleUrl,
			Sender:      a.Sender,
			TopPhrases:  a.TopPhrases,
			Subject:     a.Subject,
			TopSentence: a.TopSentence,
			TimeLapsed:  int64(a.Timestamp.Sub(start).Seconds()),
			Order:       int64(order),
		}
		eas = append(eas, ea)
	}
	sort.Strings(eventPhrases)
	return &newshound.NewsEvent{
		ID:          id,
		TopPhrases:  eventPhrases,
		EventStart:  start,
		EventEnd:    end,
		NewsAlerts:  eas,
		TopSentence: topSentence,
		TopSender:   topSender,
	}
}

type naByTimestamp []*newshound.NewsAlert

func (n naByTimestamp) Len() int           { return len(n) }
func (n naByTimestamp) Swap(i, j int)      { n[i], n[j] = n[j], n[i] }
func (n naByTimestamp) Less(i, j int) bool { return n[j].Timestamp.After(n[i].Timestamp) }

func mergeEvents(alerts []int64, phrases []string, events []*newshound.NewsEvent) (eventID int64, newID bool, eventUpdated bool, eventAlerts []int64, eventPhrases []string, staleEventIDs []int64) {
	originalEventSize := 0
	// add IDs of new event cluster
	idSet := map[int64]struct{}{}
	for _, alert := range alerts {
		idSet[alert] = struct{}{}
	}
	// add phrases to event set
	phraseSet := map[string]struct{}{}
	for _, phrase := range phrases {
		phraseSet[phrase] = struct{}{}
	}

	// default to way in future to grab oldest date
	oldest := time.Now().Add(10 * 24 * 365 * time.Hour)
	// grab all event IDs so we can remove stale ones afterwards
	var eventIDs []int64

	// add all IDs and phrases to sets
	for _, event := range events {
		for _, alert := range event.NewsAlerts {
			idSet[alert.AlertID] = struct{}{}
		}
		for _, phrase := range event.TopPhrases {
			phraseSet[phrase] = struct{}{}
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
		if eID != eventID {
			staleEventIDs = append(staleEventIDs, eID)
		}
	}

	// make the event ID if we dont have one
	if eventID == 0 {
		newID = true
	}

	// convert all ids to objectids
	for id, _ := range idSet {
		eventAlerts = append(eventAlerts, id)
	}

	// generate tag output list
	for phrase, _ := range phraseSet {
		eventPhrases = append(eventPhrases, phrase)
	}

	eventUpdated = len(eventAlerts) > originalEventSize
	return eventID, newID, eventUpdated, eventAlerts, eventPhrases, staleEventIDs
}

func findLikeAlertCluster(ctx context.Context, db DB, a *newshound.NewsAlert) (alerts []int64, phrases []string, err error) {
	// find any alerts in the eventTimeframe
	possible, err := db.FindPossibleLikeAlerts(ctx, a)
	if err != nil {
		return alerts, phrases, err
	}

	// build tag map around main alert's tags
	phraseCounts := buildPhraseCounts(a.TopPhrases, possible)

	// calc min tag limit
	minOccurs := minOccurances(len(possible))

	// filter out any tags that do not meet the limit
	for phrase, count := range phraseCounts {
		if count <= minOccurs {
			delete(phraseCounts, phrase)
		}
	}

	// we need at least 2 tags for an event
	if len(phraseCounts) < 2 {
		return alerts, phrases, nil
	}

	// make sure main alert goes the the same filtering
	possible = append(possible, a)

	// filter out any alerts that do not have minLikePhrases
	for _, alert := range possible {
		likePhrases := 0
		phraseScore := 0
		for _, phrase := range alert.TopPhrases {
			if count, ok := phraseCounts[phrase]; ok {
				likePhrases++
				phraseScore += count
			}
		}

		if likePhrases >= minLikePhrases {
			alerts = append(alerts, alert.ID)
			continue
		}

		if likePhrases >= len(possible) {
			alerts = append(alerts, alert.ID)
			continue
		}

		if float32(phraseScore) >= float32(len(possible))*(1.2) {
			alerts = append(alerts, alert.ID)
			continue
		}
	}

	for phrase, _ := range phraseCounts {
		phrases = append(phrases, phrase)
	}

	return alerts, phrases, nil
}

func buildPhraseCounts(mainPhrases []string, alerts []*newshound.NewsAlert) (phraseCounts map[string]int) {
	phraseCounts = map[string]int{}
	// seed the map with the main alert's phrases
	for _, phrase := range mainPhrases {
		phraseCounts[phrase] = 1
	}
	// increment the map for any matches and add any new partial matches
	// N^2 ...better way?
	for _, alert := range alerts {
		for _, phrase := range alert.TopPhrases {
			// check if each phrase matches any of our seeds
			toIncrement := map[string]struct{}{}
			for existing, _ := range phraseCounts {
				// exact equality check
				if strings.EqualFold(existing, phrase) {
					toIncrement[existing] = struct{}{}
					break
				}

				// partial match check
				if partialMatch(existing, phrase) {
					toIncrement[existing] = struct{}{}
					toIncrement[phrase] = struct{}{}
				}
			}

			// increment all elgibile phrases
			for toInc, _ := range toIncrement {
				phraseCounts[toInc]++
			}
		}
	}

	return phraseCounts
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
