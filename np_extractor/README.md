Noun-Phrase Extractor Service
==========

This is a Python service meant to encapsulate Python's NLTK so I have the option to rewrite the rest of Newshound's backend in Go :)

When the service starts, it will setup a quick noun-phrase extractor that has been trained with NLTK's Brown corpora and expose it on port 1029. 

You can post to the server directly or import the np_extractor.service module and run `service.np_extract("text")` as long as the service is running locally.

This library expects the following Python modules to be installed:

- nltk
- nltk's brown corpora, stopwords and punkt tokenizers
- simplejson

