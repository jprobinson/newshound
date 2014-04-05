#!/usr/bin/env python
# encoding: utf-8

import nltk
import simplejson as json
import requests
from BaseHTTPServer import BaseHTTPRequestHandler,HTTPServer

from np_extractor import NPExtractor

sent_detector = None 
np_extractor = None 

PORT=1029

class NPService(BaseHTTPRequestHandler):

    def do_POST(self):
        content_len = int(self.headers.getheader('content-length'))
        raw_text = self.rfile.read(content_len)

        np_results = self.extract(raw_text)

        self.send_response(200)
        self.send_header("Content-Type", "text/javascript; charset=UTF-8") 
        self.end_headers()

        self.wfile.write(json.dumps(np_results))
        return

    def extract(self, raw_text):
        sentences = sent_detector.tokenize(raw_text)
        results = dict() 
        max_phrases = -1
        top_sent = ""
        sents = list() 
        for sent in sentences:
            result = np_extractor.extract(sent)

            if len(result.keys()) > max_phrases:
                max_phrases = len(result.keys())
                top_sent = sent 

            sents.append({'sentence':sent, 'noun_phrases': result.keys()})

            for key in result.keys():
                results[key] = result[key]

            
        return {'noun_phrases':results, 'top_sentence': top_sent, 'sentences': sents, 'count':len(results)}


def np_extract(text):
    r = requests.post("http://localhost:%d/" % (PORT), data=text, headers={'content-type': 'text/plain; chartset=utf-8'})
    if r.status_code != requests.codes.ok:
        r.raise_for_status()

    return json.loads(r.text)

def main():
    # only setup extractors if we're running the server
    global np_extractor
    np_extractor = NPExtractor()
    global sent_detector
    sent_detector = nltk.data.load('tokenizers/punkt/english.pickle')

    server = HTTPServer(('', PORT), NPService)
    server.serve_forever()

if __name__ == "__main__":
    main() 
