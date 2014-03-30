#!/usr/bin/env python
# encoding: utf-8

'''
newshound-fetch.py

Created by JP Robinson on 2012-08-22.
'''
import re
import sys
import math
import email
import string
import getopt
import imaplib
import chardet
import pymongo
import requests
from dateutil import parser
from np_extractor.service import np_extract 
from bson.objectid import ObjectId
from collections import defaultdict
from BeautifulSoup import BeautifulSoup
from ConfigParser import SafeConfigParser
from BeautifulSoup import BeautifulStoneSoup
from email.header import decode_header, make_header
from datetime import datetime, time, timedelta, date, MAXYEAR, MINYEAR

class NewsAlert:

    SENDER_STORY_URL = {
            'CBS'				:	1,
            'FT'				:	3,
            'WSJ.com'				:	0,
            'USATODAY.com'			:	1,
            'NYTimes.com'			:	6,
            'The Washington Post'		:	3,
            'FoxNews.com'			:	0,
            }		 

    SENDER_EXCLUDE_SUBJECT = ['ABC','CNN','POLITICO']

    BREAKING_NEWS = ["breaking news", "breaking news alert", "news alert", "national news alert", "sports alert"]

    EMAIL_FILLERS = ["===", "*", u"©", "this email", "click here", "go here", "for more", "share this", "view this", "to unsubscribe", 
                "e-mail alerts", "view it online", "complete coverage", "paste the link", "paste this link", "is a developing story",
                "for further developments", "for breaking news", "the moment it happens", "keep reading", "connect with diane",  
                "on the go?", "more newsletters", "manage email", 'text "breaking', "read more", "(c)", "contact your cable", 
                "share this", "for the latest", "|", "to your address book", "unsubscribe", "and watch ", "if this message", 
                "to view this email", "more on this", "more stories", "go to nbcnews", "to ensure", "privacy policy",
                "manage portfolio", "forward this email", "subscribe to",
                "share on facebook", "video alerts", "on your cell phone", "more coverage"]

    def __init__(self,raw_message, alert_info={}):

        if raw_message:
            message = email.message_from_string(raw_message[0][1])
            alert_info = {}
            alert_info["sender"] = self.__find_sender_name(message)
            alert_info["subject"] = self.__parse_subject(message)
            alert_info["raw_body"] = self.__get_body(message)
            alert_info["timestamp"] = self.__get_date(message)
            alert_info["article_url"] = self.__find_article_url(alert_info["sender"],alert_info["raw_body"])
            instance_id = self.__get_instance_id(message)
            if(instance_id):
                alert_info["instance_id"] = instance_id
            else:
                alert_info["instance_id"] = None

        self.alert_info	 = alert_info			
        self.alert_info["body"] = self.__strip_unsub_links(alert_info["raw_body"])
        self.alert_info["tags"] = self.__create_tags(alert_info["sender"],alert_info["raw_body"],alert_info["subject"])

    def get_id(self):
        return self.alert_info["_id"]

    def set_id(self,id):
        self.alert_info["_id"] = id

    def get_timestamp(self):
        return self.alert_info["timestamp"]

    def get_tags(self):
        return self.alert_info["tags"]

    def get_sender(self):
        return self.alert_info["sender"]

    def get_subject(self):
        return self.alert_info["subject"]

    def __parse_subject(self,message):
        subject, encoding = decode_header(message['subject'])[0]
        subject_str = str()
        if encoding==None:
            charset = chardet.detect(str(subject))['encoding']
            text = unicode(subject,charset,'ignore').encode('utf8','replace')
            subject_str = text.strip()
        else:
            subject_str = subject.decode(encoding)

        if subject_str.strip() == str():
            subject_str = message['Subject']

        return subject_str

    def __find_sender_name(self,message):
        name_info = message['From'].split('"')
        name_str = str()
        if len(name_info) > 1:
            name_str = name_info[1]
        else:
            name_str = message['From'].split()[0]

        name = name_str.split()[0]
        # DEAL WITH LA TIMES OR ANYONE WITH 'THE'
        if name.lower() == "the" or name.lower() == "los":
            name = name_str

        if name == 'LA':
            name = 'Los Angeles Times'

        return name

    def __get_instance_id(self,message):
        instance_id = ""
        if message.has_key("X-InstanceId"):
            instance_id = message["X-InstanceId"]

        return instance_id

    def __get_date(self,message):
        date_string = message['Received'].split(';')[1].strip()
        return parser.parse(date_string)

    def __get_body(self,message):
        if message.is_multipart():
            html = None
	    text = ""
            for part in message.get_payload():
                if part.get_content_charset() is None:
                    charset = chardet.detect(str(part))['encoding']
                else:
                    charset = part.get_content_charset()
                if part.get_content_type() == 'text/plain':
                    text = unicode(part.get_payload(decode=True),str(charset),"ignore").encode('utf8','replace')
                if part.get_content_type() == 'text/html':
                    html = unicode(part.get_payload(decode=True),str(charset),"ignore").encode('utf8','replace')
            if html is None:
                return text.strip()
            else:
                return html.strip()

        text = unicode(message.get_payload(decode=True),errors='ignore').encode('utf8','replace')
        if text.find("table") == -1 and text.find("<table>") == -1:
            text = text.replace("\n","<br/>").lstrip("<br/>\n")
            if text.startswith('p>'):
                text = '<' + text

        return text.strip()

    def __strip_unsub_links(self,body):
        body = body.replace('breaking.news.catcher','')
        body = body.replace('unsubscribe','')
        body = body.replace('Unsubscribe','')
        body = body.replace('dyn.politico.com','')
        body = body.replace('email.foxnews.com','')
        body = body.replace('reg.e.usatoday.com','')
        body = body.replace('click.e.usatoday.com','')
        body = body.replace('cheetahmail','')
        body = body.replace('EmailSubMgr','')
        body = body.replace('newsvine','')
        body = body.replace('ts.go.com','')
        return body

    def __visible_html_text(self,element):
        if element.parent.name.lower() in ['style','script','head','meta','title','doctype', '!doctype', 'v:shape','v:imagedata','!']:
            return False
        elif re.match('<!--.*-->', str(element)):
            return False
        elif element.string.startswith('[if'):
            return False
        elif 'doctype' in element.lower():
            return False
        else:
            return element != '\n'

    def __find_article_url(self,sender,body):
        url = str()

        # we can only access nyt @ www so dont try anyone else for now
        #if sender != "NYTimes.com":
        #    return url 

        if sender in self.SENDER_STORY_URL:
            soup = BeautifulSoup(body)
            anchors = soup.findAll(href=True)
            hrefs = [anchor['href'] for anchor in anchors]
            try:
                url = hrefs[self.SENDER_STORY_URL[sender]]
            except:
                # lets just try this in case of plain text
                urls = re.findall('http[s]?://(?:[a-zA-Z]|[0-9]|[$-_@.&+]|[!*\(\),]|(?:%[0-9a-fA-F][0-9a-fA-F]))+',body)
                if len(urls) > 0:
                    url = urls[0]

            if (len(url.strip()) == 0) or ("doubleclick" in url):		 
                return str()

            # attempt to find final url...most have click tracking 
            try:
                new_url = requests.get(url)
                url = new_url.url
            except:
                # log("unable to open url!")
                pass

        return url

    def __find_tagable_text(self,sender,body,subject = ''):
        tags = set()
        soup = BeautifulSoup(body,convertEntities=BeautifulStoneSoup.ALL_ENTITIES)
        texts = soup.findAll(text=True)

        return filter(self.__visible_html_text,texts)

    def __is_news_line(self, line, sender):

        if len(line) == 0:
            return False

        if line.lower() in self.BREAKING_NEWS:
            return False

        for filler_phrase in self.EMAIL_FILLERS:
            if filler_phrase in line.lower():
                return False

        if line.endswith("-"):
            line = line[-1:]

        try:
            parser.parse(line)
            return False
        except:
            # do nothing...we dont want this to be a valid date
            pass

        if line.startswith("www."):
            return False

        if sender.lower() in line.lower():
            if ("told" not in line.lower()) and ("tell" not in line.lower()):
                return False

        return not line.startswith("http")

    def __create_tags(self,sender,body,subject = ''):
        tagable_text = self.__find_tagable_text(sender,body,subject)
        valid_lines = [] 
        bad_line_count = 0
        for line in tagable_text:
            line = line.strip()
            if self.__is_news_line(line, sender):				
                bad_line_count = 0
                valid_lines.append(line)
            elif len(valid_lines) > 0 and line != "":
                bad_line_count += 1

            if (len(valid_lines) == 3) or ((bad_line_count > 0) and (len(valid_lines) >= 2)):
                break

        if sender not in self.SENDER_EXCLUDE_SUBJECT:
            # lowercase the subject so we dont get false pronouns with title (lookin at you, nyt...)
            # add a period on the end so it appears to be a sentence 
            valid_lines.append(subject.lower()+".")

        article_text = " ".join(valid_lines)
        
        tag_results = np_extract(article_text.encode("utf-8"))
        tags = tag_results["noun_phrases"].keys()
        tags = [tag.replace(" 's", "'s") for tag in tags]

        return tags

class NewsEvent:

    def __init__(self,news_alerts,start_time,end_time,tags):
        self.event_info = {}
        self.event_info["news_alerts"] = news_alerts
        self.event_info["event_start"] = start_time
        self.event_info["event_end"] = end_time
        self.event_info["tags"] = list(tags)


class NewsAlertService:

    def __init__(self,db_replica_set,db_user,db_pw):
        # connection = pymongo.ReplicaSetConnection(db_replica_set,replicaSet='newshound')
	connection = pymongo.MongoClient(db_replica_set)
        self.db = connection.newshound
        self.db.authenticate(db_user,db_pw)
        self.news_alerts = self.db["news_alerts"]
        self.news_events = self.db["news_events"]
        self.__ensure_indices()

    def __ensure_indices(self):
        self.news_alerts.ensure_index([("_id",pymongo.ASCENDING),("timestamp",pymongo.ASCENDING), ("tags",pymongo.ASCENDING)])
        self.news_alerts.ensure_index([("_id",pymongo.ASCENDING),("timestamp",pymongo.ASCENDING)])
        self.news_events.ensure_index("news_alerts.alert_id")

    def commit_alert(self,alert):
        alert_id = self.news_alerts.insert(alert.alert_info)
        alert.set_id(alert_id)
        return alert

    def __find_all_alerts(self):
        results = self.news_alerts.find()
        for alert in results:
            yield NewsAlert(None,alert)

    def create_events(self,alerts):
        for alert in alerts:
            self.__create_events(alert)

    def rebuild_events(self):
        self.news_events.remove()
        all_alerts = []
        for alert in self.__find_all_alerts():
            obj_id = ObjectId(alert.alert_info["_id"])
            self.news_alerts.update({"_id":obj_id},
                        {"body":alert.alert_info["body"],
                        "raw_body":alert.alert_info["raw_body"],
                        "timestamp":alert.alert_info["timestamp"],
                        "subject":alert.alert_info["subject"],
                        "sender":alert.alert_info["sender"],
                        "tags":alert.alert_info["tags"],
                        "instance_id":alert.alert_info.get("instance_id",None),
                        "article_url":alert.alert_info["article_url"]})
            all_alerts.append(alert)

        self.create_events(all_alerts)

    def __find_simple_like_alerts(self,alert):
        '''
        Finds alerts that match any tag and fall within the time range.
        '''
        alert_date = alert.get_timestamp()
        hour_range = timedelta(seconds=3600*2)
        before_date = alert_date - hour_range
        after_date = alert_date + hour_range
        possible_like_alerts = self.news_alerts.find({ "_id" : { "$ne" : alert.get_id() },
                                                    "timestamp": { "$gte" : before_date, "$lte" : after_date },
                                                    "tags" : { "$in": alert.get_tags() }})


        like_alerts = []
        for like_alert in possible_like_alerts:
            like_alerts.append(like_alert)

        return like_alerts

    def __tag_contains_tag(self, tag, existing_tag):
        #simple check
        if tag == existing_tag:
            return True

        # word check
        tag_word_count = len(tag.split())
        existing_words = existing_tag.split()
        for x in range(0,len(existing_words)):
            partial_tag = " ".join(existing_words[x:(x+tag_word_count)])
            if partial_tag.endswith("'s"):
                partial_tag = partial_tag[:-3]
            if tag == partial_tag:
                return True

        return False

    def __find_event_tags(self,like_alerts,alert_tags):
        #BUILD TAG MAP FROM ALL ALERTS
        tag_map = defaultdict(int)
        for tag in alert_tags: tag_map[tag] += 1
        for match_alert in like_alerts:
            for tag in match_alert["tags"]:
                # N^2 errrrmaagerrrrd... find a better way
                tags_to_increment = set()
                for existing_tag in tag_map:
                    # look for partial word or whole word matches
                    if self.__tag_contains_tag(tag, existing_tag) or self.__tag_contains_tag(existing_tag, tag):
                        tags_to_increment.add(existing_tag)
                        tags_to_increment.add(tag)
                for tag_to_increment in tags_to_increment:
                    tag_map[tag_to_increment] += 1						

        min_tag_count = math.ceil(len(like_alerts) * 0.5)

        #FILTER OUT TAGS THAT APPEAR IN < 50% OF ALERTS
        # log("total alerts: %d" % len(like_alerts))
        # log("min tags: %d" % min_tag_count)
        # log(tag_map)
        return set([tag for tag,count in filter(lambda (tag,count): count >=min_tag_count, tag_map.items())])

    def __filter_alerts_by_event_tags(self,possible_like_alerts,alert):
        event_alerts = []

        event_tags = self.__find_event_tags(possible_like_alerts,alert.get_tags())

        if len(event_tags) >= 2:
            #FILTER OUT ALERTS THAT DONT HAVE ANY EVENT TAGS
            like_alerts = [like_alert["_id"] for like_alert in possible_like_alerts]
            like_alerts.append(alert.get_id())
            matched_alerts = self.news_alerts.find({ "_id" : { "$in" : like_alerts },
                "tags" : { "$in": list(event_tags) }})

            #FILTER OUT ALERTS THAT HAVE LESS THAN 3 EVENT TAGS
            for matched_alert in matched_alerts:
                if self.__alert_belongs_in_event(matched_alert["tags"],event_tags):
                    event_alerts.append(matched_alert)

        return (event_alerts,event_tags)

    def __minimum_sender_count(self,event_alerts):
        senders = set()
        for alert in event_alerts:
            senders.add(alert['sender'])

        return len(senders) >= 2

    def __build_event_alert_data(self,alert_ids):
        ordered_alerts = self.news_alerts.find({ "_id" : { "$in" : list(alert_ids) }}).sort( "timestamp", 1 )

        first_time = 0
        alert_data = []
        order = 1
        for alert in ordered_alerts:
            if first_time == 0:
                first_time = alert["timestamp"]

            time_lapsed = alert["timestamp"] - first_time
            alert_data.append({ "subject": alert["subject"],
                "order": order, "sender":alert["sender"],
                "alert_id":alert["_id"],
                "article_url":alert["article_url"],
                "time_lapsed": time_lapsed.seconds,
                "tags":alert["tags"]})
            order += 1

        return alert_data

    def __find_date_range(self,first_time,last_time,alert):
        if first_time > alert["timestamp"]:
            first_time = alert["timestamp"]
        if last_time < alert["timestamp"]:
            last_time = alert["timestamp"]

        return (first_time,last_time)

    def __alert_belongs_in_event(self,alert_tags,event_tags):
        '''
        MAKE SURE ORIGINAL ALERT HAS AT LEAST 3 OF THESE TAGS
        '''
        original_tag_score = 0
        for orig_tag in alert_tags:
            if orig_tag in event_tags:
                original_tag_score += 1 #len(orig_tag.split())

        #include if it has a score of 3+ OR if ALL tags in alert match (2 if only 2 exist)
        return (original_tag_score >= 3) or  (original_tag_score == len(alert_tags))

    def __find_like_events(self,alert_ids):
        return self.news_events.find({ "news_alerts.alert_id" : { "$in" : alert_ids } })

    def __create_or_merge_event(self,event_alerts,event_tags):
        matched_alert_ids = set([matched_alert["_id"] for matched_alert in event_alerts])

        #Abosrb and delete any like events
        existing_events = self.__find_like_events(list(matched_alert_ids))
        for existing_event in existing_events:
            event_tags = set(event_tags.union(existing_event["tags"]))
            matched_alert_ids = matched_alert_ids.union([alert["alert_id"] for alert in existing_event["news_alerts"]])
            self.news_events.remove({"_id":existing_event["_id"]})

        matched_alert_ids = list(matched_alert_ids) 
        final_alert_set = self.news_alerts.find({ "_id" : { "$in" : matched_alert_ids }})

        first_time = datetime.combine(date(MAXYEAR, 1, 1),time())
        last_time = datetime.combine(date(MINYEAR, 1, 1),time())
        for alert in final_alert_set:
            (first_time,last_time) = self.__find_date_range(first_time,last_time,alert)

        log("updating/creating event!:" + str(len(matched_alert_ids)))
        log(str(event_tags) + "\n\n")

        new_event = NewsEvent(self.__build_event_alert_data(matched_alert_ids),first_time,last_time,list(event_tags))
        self.news_events.insert(new_event.event_info)		 

    def __create_events(self,alert):
        #FIND ALERTS THAT MATCH 'ANY' TAG IN TIME RANGE
        like_alerts = self.__find_simple_like_alerts(alert)
        if len(like_alerts) > 2:
            (event_alerts,event_tags) = self.__filter_alerts_by_event_tags(like_alerts,alert)

            #EVENTS MUST HAVE MORE THAN 1 NEWS SOURCE & AT LEAST 3 ALERTS
            if self.__minimum_sender_count(event_alerts) and (len(event_alerts) > 2):
                self.__create_or_merge_event(event_alerts,event_tags)


class EmailFetcher:

    def __init__(self):
        configParser = SafeConfigParser()
        configParser.read('/opt/newshound/etc/config.ini')
        self.imap_server_info = {
            "user":configParser.get("imap_server_info", "user"),
            "password":configParser.get("imap_server_info", "password"),
            "server_name":configParser.get("imap_server_info", "server_name")
        }
        self.db_replica_set = configParser.get("newshound_db","replica_set").replace("/newshound","")
        self.db_user = configParser.get("newshound_db","user")
        self.db_pw = configParser.get("newshound_db","password")
        self.mark_as_read = (configParser.get("newshound_info","mark_as_read") == "true")

    def get_the_mail(self):
        if not self.imap_server_info.get("server_name"):
            log('No server configured, not running.')
            return

        try:
            imap_conn = imaplib.IMAP4_SSL(self.imap_server_info['server_name'])
        except:
            log('Unable to connect to imap server :' + self.imap_server_info['server_name'])
            sys.exit(1)

        try:
            imap_conn.login(self.imap_server_info['user'],self.imap_server_info['password'])
            log('logged into IMAP server')
        except:
            log('Unable to authenticate with pop server:'+self.imap_server_info['server_name']+' invalid creds:'+self.imap_server_info['user'])
            sys.exit(1)

        # Moving to inbox
        imap_conn.select('inbox')
        # Finding UNREAD messages
        typ, msg_data = imap_conn.search(None,'UnSeen')

        alerts = []
        message_nums = msg_data[0].split()
        service = NewsAlertService(self.db_replica_set,self.db_user,self.db_pw)
        log('pulling %d emails from inbox' % len(message_nums))
        for message_num in message_nums:

            typ, raw_message = imap_conn.fetch(message_num,'(RFC822)')

            # mark the mail as 'unread' if we're in dev
            if not self.mark_as_read:
                imap_conn.store(message_num,'-FLAGS','\\Seen')

            # pull out info, save to mongo and put into list
            alerts.append(service.commit_alert(NewsAlert(raw_message)))

        imap_conn.close()

        #load the creates events from the list of alerts
        service.create_events(alerts)

        log('Done fetching emails, boss')

    def rebuild_events(self):
        log('Re-parsing alert data and rebuilding events.')
        service = NewsAlertService(self.db_replica_set,self.db_user,self.db_pw)
        service.rebuild_events()
        log('Rebuild complete.')

def log(text):
    print '%s - %s' % (datetime.now().strftime('%Y-%m-%d %H:%M:%S'),text)		 

