#!/usr/bin/env python
# encoding: utf-8

import nltk
import simplejson as json
from BaseHTTPServer import BaseHTTPRequestHandler,HTTPServer

from np_extractor import NPExtractor

sent_detector = None 
np_extractor = None 

PORT=8080

class NPService(BaseHTTPRequestHandler):

    def do_POST(self):
        content_len = int(self.headers.getheader('content-length'))
        raw_text = self.rfile.read(content_len)
        raw_text = raw_text.decode("utf8")
        np_results = self.extract(raw_text)

        self.send_response(200)
        self.send_header("Content-Type", "text/javascript; charset=UTF-8") 
        self.end_headers()

        self.wfile.write(json.dumps(np_results))
        self.finish()
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
