Newshound Python lib
===========

This library contains the code to:

1. Connect to an IMAP server pull data for unread alert emails
2. Use the np_extractor service to find Noun Phrases for each alert
3. Determine if other alerts in the same timeframe are about similar topics
4. Create news events and track them over time as more alerts come in.
5. Use MongoDB mapreduce to generate aggregated statistics about each news source


This library expects the following Python modules to be installed:

- numpy
- chardet
- imaplib
- pymongo
- requests
- BeautifulSoup

