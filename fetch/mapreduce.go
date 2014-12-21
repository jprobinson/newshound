package fetch

import (
	"errors"
	"fmt"
	"log"
	"time"

	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
)

func MapReduce(sess *mgo.Session) error {
	startTime := time.Now()
	updateTimeframes()

	if err := ensureIndices(sess); err != nil {
		return err
	}

	if err := generateAlertsPerWeekBySender(sess); err != nil {
		return err
	}

	if err := generateAlertsPerWeek(sess); err != nil {
		return err
	}

	if err := generateAvgAlertsPerWeekBySender(sess); err != nil {
		return err
	}

	if err := generateEventsPerWeekBySender(sess); err != nil {
		return err
	}

	if err := generateEventsPerWeek(sess); err != nil {
		return err
	}

	if err := generateAvgEventsPerWeekBySender(sess); err != nil {
		return err
	}

	if err := generateSenderEventAttendance(sess); err != nil {
		return err
	}

	if err := generateAlertsPerHour(sess); err != nil {
		return err
	}

	log.Printf("MapReduce complete in %s", time.Since(startTime))
	return nil
}

var indices = map[string][][]string{
	"news_alerts": [][]string{
		[]string{"timestamp"}},

	"news_events": [][]string{
		[]string{"news_alerts.sender", "event_start"},
		[]string{"event_start"}},

	"alerts_per_week": [][]string{
		[]string{"_id.week_start"}},

	"events_per_week": [][]string{
		[]string{"_id.week_start"}},

	"avg_alerts_per_week_by_sender": [][]string{
		[]string{"_id.sender", "_id.timeframe"}},

	"avg_events_per_week_by_sender": [][]string{
		[]string{"_id.sender", "_id.timeframe"}},

	"alerts_per_week_by_sender": [][]string{
		[]string{"_id.sender", "_id.week_start"}},

	"events_per_week_by_sender": [][]string{
		[]string{"_id.sender", "_id.week_start"}},

	"sender_event_attendance": [][]string{
		[]string{"_id.sender", "_id.timeframe"}},

	"sender_alerts_per_hour": [][]string{
		[]string{"_id.sender"}},
}

func ensureIndices(sess *mgo.Session) error {
	db := sess.DB("newshound")
	log.Print("ensuring indices")

	for colName, inds := range indices {
		coll := db.C(colName)
		for _, indx := range inds {
			if err := coll.EnsureIndexKey(indx...); err != nil {
				return err
			}
		}
	}
	return nil
}

const (
	LastSevenDays = "7days"
	ThreeMonths   = "3months"
	SixMonths     = "6months"
	TwelveMonths  = "12months"
)

var Timeframes map[string][]time.Time

func updateTimeframes() {
	var startDate = time.Now()
	Timeframes = map[string][]time.Time{
		LastSevenDays: []time.Time{startDate.AddDate(0, 0, -7), startDate},
		ThreeMonths:   []time.Time{startDate.AddDate(0, -3, 0), startDate},
		SixMonths:     []time.Time{startDate.AddDate(0, -6, 0), startDate},
		TwelveMonths:  []time.Time{startDate.AddDate(0, -12, 0), startDate},
	}
}

func generateData(sess *mgo.Session, mapper, reducer, finalize, sourceCollection, resultCollection string, query bson.M) error {
	db := sess.DB("newshound")
	source := db.C(sourceCollection)
	tempResultCollection := resultCollection + "_tmp"
	tempResult := db.C(tempResultCollection)
	_, err := tempResult.RemoveAll(nil)
	if err != nil {
		return errors.New(fmt.Sprintf("remove error for %s:  %s", tempResultCollection, err.Error()))
	}

	log.Printf("calculating %s", resultCollection)

	job := &mgo.MapReduce{
		Map:      mapper,
		Reduce:   reducer,
		Out:      bson.M{"replace": tempResultCollection},
		Finalize: finalize,
	}

	_, err = source.Find(query).MapReduce(job, nil)
	if err != nil {
		log.Printf("unable mapreduce for %s:  %s", resultCollection, err.Error())
		return err
	}

	if err = renameCollection(sess, tempResultCollection, resultCollection); err != nil {
		return err
	}

	return nil
}

func generateAvgData(sess *mgo.Session, mapper, reducer, finalize, sourceCollection, resultCollection string, useWeeks bool) error {
	db := sess.DB("newshound")
	source := db.C(sourceCollection)
	tempResultCollection := resultCollection + "_tmp"
	tempResult := db.C(tempResultCollection)
	_, err := tempResult.RemoveAll(nil)
	if err != nil {
		return errors.New(fmt.Sprintf("remove error for %s:  %s", tempResultCollection, err.Error()))
	}

	// find all senders
	var senders []string
	err = db.C("news_alerts").Find(nil).Distinct("sender", &senders)
	if err != nil {
		log.Printf("Unable to find all distinct senders: %s", err.Error())
		return err
	}

	for timekey, timeframe := range Timeframes {
		log.Printf("calculating %s over '%s' thru '%s' as %s", resultCollection, timeframe[0], timeframe[1], timekey)
		var extraParam interface{}
		if useWeeks {
			// # weeks between timeframe
			extraParam = weeksBetween(timeframe[0], timeframe[1])
		} else {
			// # events in timeframe
			var tEvents int
			tEvents, err = totalEvents(sess, timeframe[0], timeframe[1])
			if err != nil {
				log.Print("unable to find total event count: ", err.Error())
				return err
			}
			extraParam = tEvents
		}

		job := &mgo.MapReduce{
			Map:      fmt.Sprintf(mapper, timekey, extraParam),
			Reduce:   fmt.Sprintf(reducer, extraParam),
			Out:      bson.M{"merge": tempResultCollection},
			Finalize: finalize,
		}

		_, err := source.Find(bson.M{"_id.week_start": bson.M{"$gte": timeframe[0], "$lt": timeframe[1]}}).MapReduce(job, nil)
		if err != nil {
			log.Printf("unable to mapreduce %s for %s:  %s", resultCollection, timekey, err.Error())
			return err
		}

		// add blank records for each sender missing
		var rSenders []string
		if err = tempResult.Find(bson.M{"_id.timeframe": timekey}).Distinct("_id.sender", &rSenders); err != nil {
			return err
		}
		// put results in a map for easy searching
		resultSenders := make(map[string]struct{})
		for _, sender := range rSenders {
			resultSenders[sender] = struct{}{}
		}
		// any senders that dont exist get a blank record for this timeframe.
		for _, sender := range senders {
			if _, found := resultSenders[sender]; !found {
				// insert blank record
				err = tempResult.Insert(bson.M{"_id": bson.M{"sender": sender, "timeframe": timekey}, "value": bson.M{}})
				if err != nil {
					log.Printf("Unable to insert blank %s record for %s: %s", resultCollection, sender, err.Error())
					return err
				}
			}
		}
	}

	if err = renameCollection(sess, tempResultCollection, resultCollection); err != nil {
		return err
	}

	return nil

}

var hoursPerWeek = float64(24 * 7)

func weeksBetween(from, to time.Time) float64 {
	dur := to.Sub(from)
	weeks := 0.0
	if hours := dur.Hours(); hours > 0 {
		weeks = hours / hoursPerWeek
	}
	return weeks
}

func totalEvents(sess *mgo.Session, from, to time.Time) (int, error) {
	newsEvents := sess.DB("newshound").C("news_events")
	return newsEvents.Find(bson.M{"event_start": bson.M{"$gte": from, "$lt": to}}).Count()
}

func renameCollection(sess *mgo.Session, from, to string) error {
	inds := indices[to]
	ndb := sess.DB("newshound")
	tempResult := ndb.C(to)
	for _, indx := range inds {
		tempResult.EnsureIndexKey(indx...)
	}

	admin := sess.DB("admin")
	err := admin.Run(bson.D{{"renameCollection", "newshound." + from}, {"to", "newshound." + to}, {"dropTarget", true}}, nil)
	if err != nil {
		return err
	}

	return nil

}

func generateAlertsPerWeekBySender(sess *mgo.Session) error {
	return generateData(sess, `function() {
							if(!this.timestamp){
								return;
							}
                            var lastSunday = new Date();
                            lastSunday.setHours(0,0,0,0);
						    lastSunday.setYear(this.timestamp.getFullYear());
                            lastSunday.setMonth(this.timestamp.getMonth());
                            lastSunday.setDate(this.timestamp.getDate() - this.timestamp.getDay());
                            var tag_map = {};
                            this.tags.forEach(function(tag){
                                tag_map[tag.replace(/\./g,'&#46;')] = 1;
                            });

                            emit({sender:this.sender,week_start:lastSunday},
                                {alerts:1,tag_map:tag_map});
                    }`,
		`function(key,values){
							var result = {alerts:0,tag_map:{}};
							values.forEach(function(value){
								result.alerts += value.alerts;

								for(var tag_key in value.tag_map){
									if(result.tag_map.hasOwnProperty(tag_key)){
										result.tag_map[tag_key] += value.tag_map[tag_key];
									}
									else{
										result.tag_map[tag_key] = value.tag_map[tag_key];
									}
								}
							});

							return result;
					  }`,
		"",
		"news_alerts",
		"alerts_per_week_by_sender",
		bson.M{"timestamp": bson.M{"$gte": Timeframes[TwelveMonths][0]}})
}

func generateAlertsPerWeek(sess *mgo.Session) error {
	return generateData(sess, `function() {
							if(!this.value){
								this.value = {alerts:0};
							}
                            emit({week_start:this._id.week_start},
                                {alerts:this.value.alerts});
                    }`,
		`function(key,values){
							var result = {alerts:0};
							values.forEach(function(value){
								result.alerts += value.alerts;
							});

							return result;
					  }`,
		"",
		"alerts_per_week_by_sender",
		"alerts_per_week",
		bson.M{})
}

func generateAvgAlertsPerWeekBySender(sess *mgo.Session) error {
	return generateAvgData(sess, `function() {
                            var tag_array = [];
                            for(var tag_key in this.value.tag_map){
                                tag_array.push({tag:tag_key,frequency:this.value.tag_map[tag_key]});
                            }
                            tag_array = tag_array.sort(function(a,b){
	                            return (a.frequency > b.frequency) ? -1 : ((a.frequency < b.frequency) ? 1 : 0);
                            });
                            tag_array = tag_array.slice(0,16);

                            emit({sender:this._id.sender, timeframe:'%s'},
                                {avg_alerts:(this.value.alerts/%f),tag_array:tag_array,tag_map:this.value.tag_map,total_alerts:this.value.alerts});
                    }`,
		`function(key,values){
                                var result = {avg_alerts:0.0,tag_map:{},tag_array:[],total_alerts:0.0};
                                var total_weeks = 0;

                                values.forEach(function(value){
                                    total_weeks += 1;
                                    result.total_alerts += value.total_alerts;
                                    for(var tag_key in value.tag_map){
                                        if(result.tag_map.hasOwnProperty(tag_key)){
                                            result.tag_map[tag_key] += value.tag_map[tag_key];
                                        }
                                        else{
                                            result.tag_map[tag_key] = value.tag_map[tag_key];
                                        }
                                    }
                                });

                                for(var tag_key in result.tag_map){
                                    result.tag_array.push({tag:tag_key,frequency:result.tag_map[tag_key]});
                                }
                                result.tag_array = result.tag_array.sort(function(a,b){
                                                    return (a.frequency > b.frequency) ? -1 : ((a.frequency < b.frequency) ? 1 : 0);
                                });
                                result.tag_array = result.tag_array.slice(0,16);
                                result.avg_alerts = result.total_alerts/%f;
                                return result;
                          }`,
		"",
		"alerts_per_week_by_sender",
		"avg_alerts_per_week_by_sender",
		true)
}

func generateEventsPerWeekBySender(sess *mgo.Session) error {
	return generateData(sess, `function() {
                                var lastSunday = new Date();
                                lastSunday.setHours(0,0,0,0);
                                lastSunday.setYear(this.event_start.getFullYear());
                                lastSunday.setMonth(this.event_start.getMonth());
                                lastSunday.setDate(this.event_start.getDate() - this.event_start.getDay());

                                var total_rank = 0;
                                var total_time_lapsed = 0;
                                var tag_map = {};
								var emitted = {};
                                this.news_alerts.forEach(function(alert){
									total_rank++;
									if(!emitted[alert.sender]){
										total_rank = alert.order;
										total_time_lapsed = alert.time_lapsed;
										alert.tags.forEach(function(tag){
											tag_map[tag.replace(/\./g,'&#46;')] = 1;
										});
										emitted[alert.sender] = 1;
										emit({sender:alert.sender,week_start:lastSunday},
                                    				{total_events:1,total_rank:total_rank,
                                                     avg_rank:total_rank,total_time_lapsed:total_time_lapsed,
						                             tag_map:tag_map,avg_time_lapsed:total_time_lapsed/60});
 
									}
                                });

                       }`, `function(key,values){
                                    var result = {total_events:0,total_rank:0,avg_rank:0.0,total_time_lapsed:0,avg_time_lapsed:0.0,tag_map:{}};

                                    values.forEach(function(value){
                                        result.total_events += value.total_events;
                                        result.total_rank += value.total_rank;
                                        result.total_time_lapsed += value.total_time_lapsed;
                                        for(var tag_key in value.tag_map){
                                            if(result.tag_map.hasOwnProperty(tag_key)){
                                                result.tag_map[tag_key] += value.tag_map[tag_key];
                                            }
                                            else{
                                                result.tag_map[tag_key] = value.tag_map[tag_key];
                                            }
                                        }
                                    });

                                    if(result.total_events != 0){
                                        result.avg_rank = (result.total_rank/result.total_events);
                                        result.avg_time_lapsed = (result.total_time_lapsed/result.total_events)/60;
                                    }

                                    return result;
                              }`,
		"",
		"news_events",
		"events_per_week_by_sender",
		bson.M{"event_start": bson.M{"$gte": Timeframes[TwelveMonths][0]}})
}

func generateEventsPerWeek(sess *mgo.Session) error {
	return generateData(sess, `function() {
                         var lastSunday = new Date();
                         lastSunday.setHours(0,0,0,0);
                         lastSunday.setYear(this.event_start.getFullYear());
                         lastSunday.setMonth(this.event_start.getMonth());
                         lastSunday.setDate(this.event_start.getDate() - this.event_start.getDay());

                        emit({week_start:lastSunday},
                            {events:1});
                    }`,
		`function(key,values){
						var result = {events:0};
						values.forEach(function(value){
							result.events += value.events;
						});

						return result;
				  }`,
		"",
		"news_events",
		"events_per_week",
		bson.M{})
}

func generateAvgEventsPerWeekBySender(sess *mgo.Session) error {
	return generateAvgData(sess, `function() {
                            emit({sender:this._id.sender,timeframe:'%s'},
                                {avg_events:(this.value.total_events/%f),
                                    total_events:this.value.total_events,
                                    avg_rank:this.value.total_rank,
                                    total_rank:this.value.total_rank,
                                    total_weeks:1,
                                    tag_map:this.value.tag_map,
                                    avg_time_lapsed:this.value.total_time_lapsed/60,
                                    total_time_lapsed:this.value.total_time_lapsed});
                    }`,
		`function(key,values){
							var result = {avg_events:0.0,
											total_events:0,
											avg_rank:0.0,
											total_rank:0,
											total_weeks:0,
											tag_map:{},
											avg_time_lapsed:0.0,
											total_time_lapsed:0};

							values.forEach(function(value){
								result.total_weeks += value.total_weeks;
								result.total_events += value.total_events;
								result.total_rank += value.total_rank;
								result.total_time_lapsed += value.total_time_lapsed;

								for(var tag_key in value.tag_map){
									if(result.tag_map.hasOwnProperty(tag_key)){
										result.tag_map[tag_key] += value.tag_map[tag_key];
									}
									else{
										result.tag_map[tag_key] = value.tag_map[tag_key];
									}
								}
							});

							result.avg_events = (result.total_events/%f);

							if(result.total_events != 0){
								result.avg_rank = (result.total_rank/result.total_events);
								result.avg_time_lapsed = (result.total_time_lapsed/result.total_events)/60;
							}

							return result;
					  }`,
		"",
		"events_per_week_by_sender",
		"avg_events_per_week_by_sender",
		true)
}

func generateSenderEventAttendance(sess *mgo.Session) error {
	return generateAvgData(sess, `function() {
                            emit({sender:this._id.sender, timeframe:'%s'},
                                {total_events:this.value.total_events,attendance:((this.value.total_events/%d)*100.0)});
                    }`,
		`function(key,values){
                                var result = {total_events:0,attendance:0.0};
                                var total_events = 0;
                                values.forEach(function(value){
                                    result.total_events += value.total_events;
                                });

                                if(result.total_events != 0){
                                    result.attendance = (result.total_events/%d)*100.0;
                                }

                                return result;
                          }`,
		"",
		"events_per_week_by_sender",
		"sender_event_attendance",
		false)
}

func generateAlertsPerHour(sess *mgo.Session) error {
	return generateData(sess, `function() {
                            var hours = {0:0,1:0,2:0,3:0,4:0,5:0,6:0,7:0,8:0,9:0,10:0,11:0,12:0,
                                            13:0,14:0,15:0,16:0,17:0,18:0,19:0,20:0,21:0,22:0,23:0};
                            //adding 5 to deal with time zones
                            var temp_hour = this.timestamp.getHours();
                            hours[temp_hour] = 1
                            emit({sender:this.sender},
                                {hours:hours});
                    }`,
		`function(key,values){
							var result = {hours:{0:0,1:0,2:0,3:0,4:0,5:0,6:0,7:0,8:0,9:0,10:0,11:0,12:0,
											13:0,14:0,15:0,16:0,17:0,18:0,19:0,20:0,21:0,22:0,23:0}};

							values.forEach(function(value){
								for(var hour in value.hours){
									result.hours[hour] += value.hours[hour];
								}
							});
							return result;
                          }`,
		"",
		"news_alerts",
		"sender_alerts_per_hour",
		bson.M{})
}
