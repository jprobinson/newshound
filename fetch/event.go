package fetch

import (
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"github.com/jprobinson/newshound"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var (
	eventTimeframe = 2 * time.Hour

	minLikeTags = 2

	minSenders = 3
)

func UpdateEvents(db *mgo.Database, a newshound.NewsAlert) error {

	alerts, err := findPossibleEvent(db, a)
	if err != nil {
		return fmt.Errorf("unable to find possible like alerts for event: %s", err)
	}

	// if we have no alerts beyond our original give up
	if len(alerts) < 2 {
		return nil
	}

	// grab any events related to the alerts we've got (where alertID in $alerts)
	var existingEvents []newshound.NewsEvent
	ne := newsEvents(db)
	query := bson.M{"news_alerts.alert_id": bson.M{"$in": alerts}}
	err = ne.Find(query).All(&existingEvents)

	// merge any existing alerts into ours, reuse id if possible
	eventID, alertIDs := mergeAlerts(alerts, existingEvents)

	// get the data of all the alerts
	var nas []newshound.NewsAlert
	err = newsAlerts(db).Find(bson.M{"_id": bson.M{"$in": alertIDs}}).All(&nas)
	if err != nil {
		return err
	}

	// verify it has the min sender count
	senders := map[string]struct{}{}
	for _, alert := range nas {
		senders[alert.Sender] = struct{}{}
	}
	// quit if we dont have enough senders
	if len(senders) < minSenders {
		return nil
	}

	// create the event (all the metrics and sorting and whatnot and save it
	event := NewNewsEvent(eventID, nas)

	log.Printf("event found with %d alerts and tags: %s", len(event.NewsAlerts), event.Tags)
	_, err = ne.UpsertId(event.ID, event)
	return err
}

func NewNewsEvent(id bson.ObjectId, alerts []newshound.NewsAlert) newshound.NewsEvent {
	var (
		tags        []string
		eas         []newshound.NewsEventAlert
		topSentence string
		topSender   string
		start, end  time.Time
	)
	tagCounts := map[string]int{}

	// sort by timestamp
	sort.Sort(naByTimestamp(alerts))

	// grab our start n end since we're sorted
	start = alerts[0].Timestamp
	end = alerts[len(alerts)-1].Timestamp

	// create tag counts and eventalerts
	for _, a := range alerts {
		for _, tag := range a.Tags {
			added := false
			for ttag, _ := range tagCounts {
				if ttag == tag || strings.Contains(ttag, tag) {
					added = true
					tagCounts[tag]++
					continue
				}
				if strings.Contains(tag, ttag) {
					added = true
					tagCounts[ttag]++
				}
			}
			if !added {
				tagCounts[tag]++
			}

		}
	}

	// cut out bad tags tags
	min := minLikeTags
	if len(alerts) > 5 {
		min = len(alerts) / 2
	}

	for tag, count := range tagCounts {
		if count < min {
			delete(tagCounts, tag)
		}
	}

	log.Printf("final tag counts!\n%+v\n", tagCounts)

	// find the top sentence/sender!
	topScore := 0
	// go through each phrase of each sentence of each alert
	// find the sentence that reaches the highest score first
	order := 0
	for _, a := range alerts {
		alertTagCount := 0
		for _, s := range a.Sentences {
			score := 0
			for _, phrase := range s.Phrases {
				for tag, count := range tagCounts {
					// increment the score for any tag/phrase intersection
					if strings.Contains(tag, phrase) ||
						strings.Contains(phrase, tag) {
						score += count
						alertTagCount++
					}
				}
			}

			if score > topScore {
				topSender = a.Sender
				topSentence = s.Value
			}
		}

		if alertTagCount < minLikeTags {
			log.Print("bad score: ", alertTagCount, " alert: ", a.Sender, " tags: ", a.Tags)
			continue
		}

		order++
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

	for tag, _ := range tagCounts {
		tags = append(tags, tag)
	}

	return newshound.NewsEvent{
		ID:          id,
		Tags:        tags,
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

func mergeAlerts(alerts []bson.ObjectId, events []newshound.NewsEvent) (bson.ObjectId, []bson.ObjectId) {
	idSet := map[string]struct{}{}
	for _, alert := range alerts {
		idSet[alert.Hex()] = struct{}{}
	}

	oldest := time.Now().Add(6 * time.Hour)
	var eventID string
	for _, event := range events {
		for _, alert := range event.NewsAlerts {
			idSet[alert.AlertID.Hex()] = struct{}{}
		}
		// grab the id of the oldest event
		if oldest.After(event.EventStart) {
			oldest = event.EventStart
			eventID = event.ID.Hex()
		}
	}

	var outputID bson.ObjectId
	// make the event ID if we dont have one
	if len(eventID) == 0 {
		outputID = bson.NewObjectId()
	} else {
		log.Print("REUSING EVENT ID! ", eventID)
		outputID = bson.ObjectIdHex(eventID)
	}

	// convert all ids to objectids
	var output []bson.ObjectId
	for id, _ := range idSet {
		output = append(output, bson.ObjectIdHex(id))
	}
	return outputID, output
}

func findPossibleEvent(db *mgo.Database, a newshound.NewsAlert) ([]bson.ObjectId, error) {
	var alerts []bson.ObjectId
	var possible []newshound.NewsAlert

	col := newsAlerts(db)
	// find any alerts within a 4hr timeframe
	start := a.Timestamp.Add(-eventTimeframe)
	end := a.Timestamp.Add(eventTimeframe)
	query := bson.M{"timestamp": bson.M{"$gte": start, "$lte": end}, "_id": bson.M{"$ne": a.ID}}
	err := col.Find(query).All(&possible)
	if err != nil {
		return alerts, err
	}

	// tag maps along the way: map[alertID][]tag
	alertTags := map[string][]string{}
	for _, alert := range possible {
		alertTags[alert.ID.Hex()] = findLikeTags(a.Tags, alert.Tags)
	}

	finalLikeTags := map[string][]string{}
	// filter out alerts with not enough matches len(alertTags[id]) < minLikeTags
	for id, tags := range alertTags {
		if len(tags) >= minLikeTags {
			alerts = append(alerts, bson.ObjectIdHex(id))
			finalLikeTags[id] = tags
		}
	}

	return alerts, nil
}

func findLikeTags(a, b []string) []string {
	tags := map[string]struct{}{}

	for _, tagA := range a {
		for _, tagB := range b {
			if tagA == tagB {
				tags[tagA] = struct{}{}
				continue
			}
			if len(tagB) < 3 || len(tagA) < 3 {
				continue
			}
			if strings.Contains(tagA, tagB) {
				tags[tagB] = struct{}{}
				continue
			}
			if strings.Contains(tagB, tagA) {
				tags[tagA] = struct{}{}
				continue
			}
		}
	}

	var output []string
	for tag, _ := range tags {
		output = append(output, tag)
	}

	return output
}
