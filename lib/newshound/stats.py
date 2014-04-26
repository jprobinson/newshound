#!/usr/bin/env python
# encoding: utf-8 

'''
stats.py

Created by JP Robinson on 2012-11-08.
'''
from ConfigParser import SafeConfigParser
import pymongo
from bson.code import Code
from bson.son import SON
from datetime import timedelta, datetime, date, time

class NewsStats:
    
    def __init__(self):
        configParser = SafeConfigParser()
        configParser.read('/opt/newshound/etc/config.ini')
        db_replica_set = configParser.get("newshound_db","replica_set").replace("/newshound","")
        db_pw = configParser.get("newshound_db","password")
        db_user = configParser.get("newshound_db","user") 
        #connection = pymongo.ReplicaSetConnection(db_replica_set,replicaSet='paperboy-newshound')
	connection = pymongo.MongoClient(db_replica_set)
        self.db = connection.newshound
        self.db.authenticate(db_user,db_pw)
        self.news_alerts = self.db["news_alerts"]
        self.news_events = self.db["news_events"]
        self.alerts_per_week = self.db["alerts_per_week"]
        self.events_per_week = self.db["events_per_week"]
        self.news_report_by_sender = self.db["news_report_by_sender"]
        self.sender_event_attendance = self.db["sender_event_attendance"]
        self.alerts_per_week_by_sender = self.db["alerts_per_week_by_sender"]
        self.events_per_week_by_sender = self.db["events_per_week_by_sender"]       
        self.avg_alerts_per_week_by_sender = self.db["avg_alerts_per_week_by_sender"]
        self.avg_events_per_week_by_sender = self.db["avg_events_per_week_by_sender"]

        today = datetime.today()
        self.three_months_ago = today - timedelta((365/12)*3)
        
        self.__ensure_indices()
            
    def run_stats(self):
        self.__create_alerts_per_week_by_sender()
        self.__create_alerts_per_week()
        self.__create_avg_alerts_per_week_by_sender()
        self.__create_events_per_week_by_sender()
        self.__create_events_per_week()
        self.__create_avg_events_per_week_by_sender()
        self.__create_sender_event_attendance()
        self.__create_sender_alerts_per_hour()
        self.__create_final_report_collection()
        self.__build_total_report_table()
        
    def __ensure_indices(self):
        self.news_alerts.ensure_index("timestamp")
        self.news_events.ensure_index([("news_alerts.sender",pymongo.ASCENDING),('event_start',pymongo.ASCENDING)])
        self.news_events.ensure_index("event_start")
        self.alerts_per_week.ensure_index("_id.week_start")
        self.events_per_week.ensure_index("_id.week_start")
        self.avg_alerts_per_week_by_sender.ensure_index("_id.sender")
        self.avg_events_per_week_by_sender.ensure_index("_id.sender")
        self.alerts_per_week_by_sender.ensure_index([("_id.sender",pymongo.ASCENDING),("_id.week_start",pymongo.ASCENDING)])
        self.events_per_week_by_sender.ensure_index([("_id.sender",pymongo.ASCENDING),("_id.week_start",pymongo.ASCENDING)])
        self.sender_event_attendance.ensure_index("_id.sender")
        self.news_report_by_sender.ensure_index("sender")
        self.db["sender_alerts_per_hour"].ensure_index("_id.sender")

        
    def __create_alerts_per_week_by_sender(self):

        #________________________MESSAGES PER WEEK_________________________________
        # FIRST PASS - GENERATE # MESSAGES FOR EACH SOURCE FOR EACH WEEK -- LOOP FOR EACH WEEK IN LAST 6 MONTHS
        map = Code("""function() {
                            var lastSunday = new Date();
                            lastSunday.setHours(0,0,0,0);
                            lastSunday.setYear(this.timestamp.getFullYear());
                            lastSunday.setMonth(this.timestamp.getMonth());
                            lastSunday.setDate(this.timestamp.getDate() - this.timestamp.getDay());
                            var tag_map = {};
                            this.tags.forEach(function(tag){
                                tag_map[tag] = 1;
                            });
                    
                            emit({sender:this.sender,week_start:lastSunday},
                                {alerts:1,tag_map:tag_map});
                    }""")

        reduce = Code("""function(key,values){
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
                          }""")
        self.news_alerts.map_reduce(map,reduce,out=SON([("replace", "alerts_per_week_by_sender"), ("db", "newshound")]),query={"timestamp":{'$gte':self.three_months_ago}})

        
    def __create_alerts_per_week(self):
        # SECOND PASS - GENERATE # MESSAGES PER WEEK TOTAL
        map = Code("""function() {
                            emit({week_start:this._id.week_start},
                                {alerts:this.value.alerts});
                    }""")

        reduce = Code("""function(key,values){
                                var result = {alerts:0};
                                values.forEach(function(value){
                                    result.alerts += value.alerts;
                                });
                        
                                return result;
                          }""")
        self.alerts_per_week_by_sender.map_reduce(map,reduce,out=SON([("replace", "alerts_per_week"), ("db", "newshound")]),query={})


    def __create_avg_alerts_per_week_by_sender(self):
        # THIRD PASS - GENERATE AVG MESSAGES PER WEEK BY SENDER
        map = Code("""function() {
                            var tag_array = [];
                            for(var tag_key in this.value.tag_map){
                                tag_array.push({tag:tag_key,frequency:this.value.tag_map[tag_key]});
                            }
                            tag_array = tag_array.sort(function(a,b){
                                                return (a.frequency > b.frequency) ? -1 : ((a.frequency < b.frequency) ? 1 : 0);
                            });
                            tag_array = tag_array.slice(0,16);

                            emit({sender:this._id.sender},
                                {avg_alerts:(this.value.alerts/%s),tag_array:tag_array,tag_map:this.value.tag_map,total_alerts:this.value.alerts});
                    }"""% self.__get_week_count())

        reduce = Code("""function(key,values){
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
                                print(key.sender);
                                print(result.total_alerts);
                                result.avg_alerts = result.total_alerts/%s;
                                print(result.avg_alerts);
                                return result;
                          }"""% self.__get_week_count())
        self.alerts_per_week_by_sender.map_reduce(map,reduce,out=SON([("replace", "avg_alerts_per_week_by_sender"), ("db", "newshound")]),query={})

    def __create_events_per_week_by_sender(self):
        #________________________EVENTS PER WEEK_________________________________
        # FIRST PASS - GENERATE # EVENT FOR EACH SOURCE FOR EACH WEEK
        #   -get senders -> loop each sender and do work
        self.senders = self.news_alerts.distinct("sender")
        self.events_per_week_by_sender.remove()
        for sender in self.senders:
            map = Code("""function() {
                                var lastSunday = new Date();
                                lastSunday.setHours(0,0,0,0);
                                lastSunday.setYear(this.event_start.getFullYear());
                                lastSunday.setMonth(this.event_start.getMonth());
                                lastSunday.setDate(this.event_start.getDate() - this.event_start.getDay());
                        
                                var total_rank = 0;
                                var total_time_lapsed = 0;
                                var tag_map = {};
                                this.news_alerts.forEach(function(alert){
                                    if(alert.sender == "%s"){
                                        if(total_rank == 0){
                                            total_rank = alert.order;
                                            total_time_lapsed = alert.time_lapsed;
                                            alert.tags.forEach(function(tag){
                                                tag_map[tag.replace(/\./g,'&#46;')] = 1;
                                            });
                                        }
                                    }
                                });
                        
                                emit({sender:'%s',week_start:lastSunday},
                                    {total_events:1,total_rank:total_rank,avg_rank:total_rank,total_time_lapsed:total_time_lapsed,tag_map:tag_map,avg_time_lapsed:total_time_lapsed});
                        }"""% (sender,sender))
    
            reduce = Code("""function(key,values){
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
                                        result.avg_time_lapsed = (result.total_time_lapsed/result.total_events);
                                    }
                            
                                    return result;
                              }""")
            self.news_events.map_reduce(map,reduce,out=SON([("merge", "events_per_week_by_sender"), ("db", "newshound")]),query={"news_alerts.sender":sender,"event_start" :{ "$gte":self.three_months_ago }})

    def __create_events_per_week(self):
        # GENERATE # EVENTS PER WEEK TOTAL
        map = Code("""function() {
                         var lastSunday = new Date();
                         lastSunday.setHours(0,0,0,0);
                         lastSunday.setYear(this.event_start.getFullYear());
                         lastSunday.setMonth(this.event_start.getMonth());
                         lastSunday.setDate(this.event_start.getDate() - this.event_start.getDay());
                    
                        emit({week_start:lastSunday},
                            {events:1});
                    }""")

        reduce = Code("""function(key,values){
                                var result = {events:0};
                                values.forEach(function(value){
                                    result.events += value.events;
                                });
                
                                return result;
                          }""")
        self.news_events.map_reduce(map,reduce,out=SON([("replace", "events_per_week"), ("db", "newshound")]),query={})

    def __get_week_count(self):
        distinct_weeks = self.alerts_per_week.distinct("_id.week_start")
        week_count = 0
        for week in distinct_weeks:
            week_count += 1

        return week_count

    def __create_avg_events_per_week_by_sender(self):
        
        # AVG EVENTS PER WEEK + PER SOURCE
        map = Code("""function() {
                            emit({sender:this._id.sender},
                                {avg_events:(this.value.total_events/%s),
                                    total_events:this.value.total_events,
                                    avg_rank:this.value.total_rank,
                                    total_rank:this.value.total_rank,
                                    total_weeks:1,
                                    tag_map:this.value.tag_map,
                                    avg_time_lapsed:this.value.total_time_lapsed,
                                    total_time_lapsed:this.value.total_time_lapsed});
                    }""" % self.__get_week_count())

        reduce = Code("""function(key,values){
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
                        
                                result.avg_events = (result.total_events/%s);
                                
                                if(result.total_events != 0){
                                    result.avg_rank = (result.total_rank/result.total_events);
                                    result.avg_time_lapsed = (result.total_time_lapsed/result.total_events);
                                }
                        
                                return result;
                          }""" % self.__get_week_count())
        self.events_per_week_by_sender.map_reduce(map,reduce,out=SON([("replace", "avg_events_per_week_by_sender"), ("db", "newshound")]),query={})


    def __create_sender_event_attendance(self):
        #_______________________EVENT ATTENDANCE_________________________________
        # FIRST PASS - FOR EACH CREATE TOTAL EVENT ATTENDANCE
        total_events = self.news_events.find({"event_start" :{ "$gte":self.three_months_ago }}).count();
        map = Code("""function() {
                            emit({sender:this._id.sender},
                                {total_events:this.value.total_events,attendance:((this.value.total_events/%s)*100)});
                    }"""% (total_events))

        reduce = Code("""function(key,values){
                                var result = {total_events:0,attendance:0.0};
                                var total_events = 0;
                                values.forEach(function(value){
                                    result.total_events += value.total_events;
                                });
                        
                                if(%s != 0){
                                    result.attendance = (result.total_events/%s)*100;
                                }
                        
                                return result;
                          }"""% (total_events,total_events))
        self.events_per_week_by_sender.map_reduce(map,reduce,out=SON([("replace", "sender_event_attendance"), ("db", "newshound")]),query={})

    def __create_sender_alerts_per_hour(self):
        map = Code("""function() {
                            var hours = {0:0,1:0,2:0,3:0,4:0,5:0,6:0,7:0,8:0,9:0,10:0,11:0,12:0,
                                            13:0,14:0,15:0,16:0,17:0,18:0,19:0,20:0,21:0,22:0,23:0};
                            //adding 5 to deal with time zones
                            var temp_hour = (this.timestamp.getHours() + 5) % 24;
                            hours[temp_hour] = 1
                            emit({sender:this.sender},
                                {hours:hours});
                    }""")

        reduce = Code("""function(key,values){
                                var result = {hours:{0:0,1:0,2:0,3:0,4:0,5:0,6:0,7:0,8:0,9:0,10:0,11:0,12:0,
                                                13:0,14:0,15:0,16:0,17:0,18:0,19:0,20:0,21:0,22:0,23:0}};
                                
                                values.forEach(function(value){
                                    for(var hour in value.hours){
                                        result.hours[hour] += value.hours[hour];
                                    }
                                });
                                return result;
                          }""")
        self.news_alerts.map_reduce(map,reduce,out=SON([("replace", "sender_alerts_per_hour"), ("db", "newshound")]),query={"timestamp":{'$gte':self.three_months_ago}})       

    def __create_final_report_collection(self):

        #______________________CREATE FINAL REPORT TABLE_________________________


        #CLEAR REPORT COLLECTION
        self.news_report_by_sender.remove()

        # BUILD REPORT FOR EACH SENDER
        today = date.today()
        self.last_week_start = datetime.combine(today + timedelta(days=-today.weekday()-8),time(0))
        self.last_week_end = datetime.combine(today + timedelta(days=-today.weekday()-1),time(0))
        total_events_last_week = self.news_events.find({ "event_start" : { "$gte" : self.last_week_start, "$lte" : self.last_week_end }}).count()
        for sender in self.senders:
            sender_report = { "sender" : sender }
            #MESSAGE PER WEEK
            avg_alerts_this_sender = self.avg_alerts_per_week_by_sender.find_one({ "_id" : { "sender" : sender }})
            if avg_alerts_this_sender:
                sender_report["avg_alerts_per_week"] = avg_alerts_this_sender["value"]["avg_alerts"]
                alerts_per_week_this_sender = self.alerts_per_week_by_sender.find_one({ "_id.sender" :  sender , "_id.week_start" : { "$gte" : self.last_week_start, "$lte" : self.last_week_end }})
                if alerts_per_week_this_sender is not None:
                    sender_report["alerts_last_week"] = alerts_per_week_this_sender["value"]["alerts"]
                else:
                    sender_report["alerts_last_week"] = 0
            
                #EVENTS PER WEEK 
                avg_events_this_sender = self.avg_events_per_week_by_sender.find_one({ "_id.sender" : sender })
                if avg_events_this_sender:
                    sender_report["avg_events_per_week"] = avg_events_this_sender["value"]["avg_events"]
                    events_per_week_this_sender = self.events_per_week_by_sender.find_one({ "_id.sender" :  sender , "_id.week_start" : { "$gte" : self.last_week_start, "$lte" : self.last_week_end }})
                    if events_per_week_this_sender is not None:
                        sender_report["events_last_week"] = events_per_week_this_sender["value"]["total_events"]
                    else:
                        sender_report["events_last_week"] = 0
                else:
                    sender_report["events_last_week"] = 0
                    sender_report["avg_events_per_week"] = 0.0

            
                #EVENT ATTENDANCE
                event_attendance_this_sender = self.sender_event_attendance.find_one({ "_id.sender" : sender })
                if event_attendance_this_sender:
                    sender_report["event_attendance"] = event_attendance_this_sender["value"]["attendance"]
                    if total_events_last_week != 0:
                        sender_report["event_attendance_last_week"] = (sender_report["events_last_week"]/total_events_last_week)*100
                    else:
                        sender_report["event_attendance_last_week"] = 0
                else:
                    sender_report["event_attendance_last_week"] = 0
                    sender_report["event_attendance"] = 0.0

                #AVG EVENT RANKING
                if avg_events_this_sender:
                    sender_report["avg_event_rank"] = avg_events_this_sender["value"]["avg_rank"]
                    if events_per_week_this_sender is not None:
                        sender_report["avg_event_rank_last_week"] = events_per_week_this_sender["value"]["avg_rank"]
                    else:
                        sender_report["avg_event_rank_last_week"] = 0.0
                else:
                    sender_report["avg_event_rank_last_week"] = 0.0
                    sender_report["avg_event_rank"] = 0.0            
    
                #EVENT PARTICIPATION RATING
                if (sender_report["avg_alerts_per_week"] != 0) and (sender_report["avg_event_rank"] != 0):
                    sender_report["event_attendance_rating"] = (sender_report["event_attendance"]/sender_report["avg_alerts_per_week"])/sender_report["avg_event_rank"]
                else:
                    sender_report["event_attendance_rating"] = 0.0
    
                if (sender_report["alerts_last_week"] != 0) and (sender_report["avg_event_rank_last_week"] != 0):
                    sender_report["event_attendance_rating_last_week"] = (sender_report["event_attendance_last_week"]/sender_report["alerts_last_week"])/sender_report["avg_event_rank_last_week"]
                else:
                    sender_report["event_attendance_rating_last_week"] = 0.0
    
                #AVG EVENT ARRIVAL
                if avg_events_this_sender:
                    sender_report["avg_event_arrival"] = avg_events_this_sender["value"]["avg_time_lapsed"]
                    if events_per_week_this_sender is not None:
                        sender_report["avg_event_arrival_last_week"] = events_per_week_this_sender["value"]["avg_time_lapsed"]
                    else:
                        sender_report["avg_event_arrival_last_week"] = 0.0
                else:
                    sender_report["avg_event_arrival_last_week"] = 0.0
                    sender_report["avg_event_arrival"] = 0.0

                self.news_report_by_sender.insert(sender_report)

    def __build_total_report_table(self):

        # BUILD REPORT TOTALS
        #MESSAGE PER WEEK
        alerts_last_week = self.alerts_per_week.find_one({ "_id.week_start" : { "$gte" : self.last_week_start, "$lte" : self.last_week_end }})
        if alerts_last_week is not None:
            total_alerts_last_week = alerts_last_week["value"]["alerts"]
        else:
            total_alerts_last_week = 0
        total_weeks = 0
        total_alerts = 0
        alerts_last_6_mos = self.alerts_per_week.find()
        for week in alerts_last_6_mos:
            total_weeks += 1
            total_alerts += week["value"]["alerts"]

        #EVENTS PER WEEK
        events_last_week = self.events_per_week.find_one({ "_id.week_start" : { "$gte" : self.last_week_start, "$lte" : self.last_week_end }})
        if events_last_week is not None:
            total_events_last_week = events_last_week["value"]["events"]
        else:
            total_events_last_week = 0
        total_events = 0
        events_last_6_mos = self.events_per_week.find()
        for week in events_last_6_mos:
            total_events += week["value"]["events"]

        self.news_report_by_sender.insert({"sender":"total",
                                    "avg_alerts_per_week":(total_alerts/total_weeks),
                                    "alerts_last_week":total_alerts_last_week,
                                    "avg_events_per_week":(total_events/total_weeks),
                                    "events_last_week":total_events_last_week})

def run():
    print "starting mapreduce processes"
    news_stats = NewsStats()
    news_stats.run_stats()
    print "mapreduce complete"

if __name__ == '__main__':
	run()                        
