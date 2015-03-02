# -*- coding: utf-8 -*-
import nltk
from nltk.corpus import brown
from nltk.corpus import stopwords
import string
import re

PUNCS = dict((ord(char), None) for char in filter(lambda x: (x != '-' and x != '.' and x != '\''),string.punctuation+"*"))
TIME_REGEX = re.compile(r'^\d+ (am|pm)$')
STOP_WORDS = stopwords.words('english') + [ u'—', u'»', u'000', u'8211',
        u'8217', u'a12013', u'a.m', u'at least', u'am', u'amp', u'abcnews',
        u'abcnews.com', u'abcs', u'according', u'accused', u'affair',
        u'afternoon', u'alert', u'alerts', u'also', u'am', u'announce',
        u'announced', u'ap', u'april', u'ask', u'associat', u'associated',
        u'aug', u'august', u'bbc', u'begin', u'believe', u'break', u'breaking',
        u'breaking news', u'bloomberg.com', u'case', u'cbs', u'cbsnews', u'cbsnewscom',
        u'[cbsnews] u', u'cbss', u'cdt', u'charged', u'charges', u'cite', u'coverage',
        u'cites', u'citing', u'cliff', u'cnn', u'cnncom', u'cnn mobile',
        u'come', u'congress', u'control', u'counts', u'cst', u'ct', u'cut',
        u'day', u'days', u'deal', u'dealbook','dealbook alert', u'dec',
        u'december', u'democrat', u'earn', u'east', u'edt', u'emailfoxnews',
        u'est', u'et', u'evening', u'expected', u'extramarital', u'fall',
        u'feb', u'february', u'file', u'filed', u'fill', u'fire', u'fired',
        u'fires', u'first', u'found', u'fox', u'full story', u'foxbusiness',
        u'foxnews', u'foxnewscom', u'foxnews.com', u'foxs', u'fri', u'friday',
        u'full', u'get', u'gmt', u'hln', u'home','hour', u'hours', u'http',
        u'invite', u'invites', u'jan', u'history', u'january', u'july', u'jul',
        u'jun', u'june', u'l.a', u'least', u'level', u'live', u'long', 
        u'los angeles', u'lowest', u'mar', u'march', u'may', u'gmail.com',
        u'million', u'minute', u'minutes', u'mon', u'monday', u'month',
        u'morning', u'must', u'name', u'nbc', u'nbcnews',  u'nbc news',
        u'nbcs', u'nbsp', u'new', u'news', u'news alert', u'next', 
        u'news update', u'north', u'nov', u'november', u'oct', u'october', u'office',
        u'old', u'p.m', u'pdt', u'people', u'percent', u'plan', u'plans',
        u'planned', u'pm', u'point', u'police', u'politic', u'politico',
        u'politics', u'post', u'presid', u'president', u'press', u'prior',
        u'pst', u'pt', u'rate', u're', u'n\'t', u'reached', u'reaches',
        u'read', u'rep', u'report', u'reported', u'reportedly', u'reports',
        u'representing', u'republican', u'right', u'rise', u'rqbdmg', u'said',
        u'saturday', u'say', u'says', u'scoop', u'sen', u'sept', u'september',u'sponsored\u00a0by',
        u'set', u'since', u'source', u'sources', u'south', u'speak', u'speaks',
        u'state', u'step', u'sunday', u'take', u'talk', u'talks', u'team',
        u'thanks', u'thu', u'thur', u'thurs', 'wall street journal',
        u'thursday', u'time', u'times', u'today', u'told', u'top', u'trade',
        u'trading', u'tue', u'tues', u'tuesday', u'two', u'wsj newalert', 
        u'wsj news', u'\'s',u"'s", u'undisclosed', u'unsubscribe', u'unveil',
        u'usatoday', u'usatoday.com', u'washington', u'washington post',
        u'washingtonpost', u'watch', u'watch live', u'way', u'wed',
        u'wednesday', u'week', u'west', u'without', u'wnbc', u'wsj', 
        u'wsj news alert', u'year', u'years', u'pm edt']

class NPExtractor(object):
    """
    Most of this code is from Slomi Babluki's very helpful blog post:
    http://thetokenizer.com/2013/05/09/efficient-way-to-extract-the-main-topics-of-a-sentence/
    """
    def __init__(self):
        # This is our fast Part of Speech tagger
        brown_train = brown.tagged_sents(categories=['news'])
        regexp_tagger = nltk.RegexpTagger(
                [(r'^-?[0-9]+(.[0-9]+)?$', 'CD'),
                    (r'(-|:|;)$', ':'),
                    (r'\'*$', 'MD'),
                    (r'(The|the|A|a|An|an)$', 'AT'),
                    (r'.*able$', 'JJ'),
                    (r'^[A-Z].*$', 'NNP'),
                    (r'.*ness$', 'NN'),
                    (r'.*ly$', 'RB'),
                    (r'.*s$', 'NNS'),
                    (r'.*ing$', 'VBG'),
                    (r'.*ed$', 'VBD'),
                    (r'.*', 'NN')
                    ])
        self.unigram_tagger = nltk.UnigramTagger(brown_train, backoff=regexp_tagger)
        self.bigram_tagger = nltk.BigramTagger(brown_train, backoff=self.unigram_tagger)

        # This is our semi-CFG; Extend it according to your own needs
        cfg = {}
        cfg["NNP+NNP"] = "NNP"
        cfg["CD+CD"] = "CD"
        cfg["NN+NN"] = "NNI"
        cfg["NNI+NN"] = "NNI"
        cfg["JJ+JJ"] = "JJ"
        cfg["JJ+NN"] = "NNI"
        cfg["VBN+NNS"] = "NNP"
        self.cfg = cfg

        for i, word in enumerate(STOP_WORDS):
            STOP_WORDS[i] = word

    # Split the sentence into singlw words/tokens
    def tokenize_sentence(self, sentence):
        #tokens = nltk.word_tokenize(sentence)
        #return tokens
        return re.split(r'[ \t\n]+', sentence)

    def __filter_tag(self,tag):
        return (tag.encode('utf-8').lower() not in STOP_WORDS) and ("=" not in tag) and (len(tag.strip()) > 0) and (not TIME_REGEX.match(tag)) 

    def __clean_tag(self,tag):
        clean_tag = tag.strip()
        try:
              clean_tag = tag.translate(PUNCS).lower()
        except:
              clean_tag = unicode(tag, "UTF-8").translate(PUNCS).lower()

        if clean_tag.startswith("'s "):
            clean_tag = clean_tag[3:]

        if 'http' in clean_tag:
            clean_tag = clean_tag.replace("http","")

        if clean_tag.startswith('"') or clean_tag.startswith(u"•")  or clean_tag.startswith("'") or clean_tag.startswith(u'“') or clean_tag.startswith(u'‘'):
            clean_tag = clean_tag[1:]

        if clean_tag.endswith("'") or clean_tag.endswith('"') or clean_tag.endswith(u'”') or clean_tag.endswith(u'’'):
            clean_tag = clean_tag[:-1]

        return clean_tag

    # Normalize brown corpus' tags ("NN", "NN-PL", "NNS" > "NN")
    def normalize_tags(self, tagged):
        n_tagged = []
        for t in tagged:
            if t[1] == "NP-TL" or t[1] == "NP":
                n_tagged.append((t[0], "NNP"))
                continue
            if t[1].endswith("-TL"):
                n_tagged.append((t[0], t[1][:-3]))
                continue
            if t[1].endswith("S"):
                n_tagged.append((t[0], t[1][:-1]))
                continue
            n_tagged.append((t[0], t[1]))
        return n_tagged

    # Extract the main topics from the sentence
    def extract(self, sentence):
        tokens = self.tokenize_sentence(sentence)
        tags = self.normalize_tags(self.bigram_tagger.tag(tokens))
        merge = True
        while merge:
            merge = False
            for x in range(0, len(tags) - 1):
                t1 = tags[x]
                t2 = tags[x + 1]
                key = "%s+%s" % (t1[1], t2[1])
                value = self.cfg.get(key, '')
                if value:
                    merge = True
                    tags.pop(x)
                    tags.pop(x)
                    match = "%s %s" % (t1[0], t2[0])
                    pos = value
                    tags.insert(x, (match, pos))
                    break
        matches = dict() 
        for t in tags:
            if t[1] == "NNP" or t[1] == "NNI" or t[1] == "NNS" or t[1] == "NP-HL" or t[1] == "NN" or t[1] == "CD":
                matches[self.__clean_tag(t[0])] = t[1]
        
        final_tags = filter(self.__filter_tag, matches.keys()) 

        final_matches = dict()
        for t in final_tags:
            final_matches[t] = matches[t]

        return final_matches
