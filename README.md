[![Newshound: The Breaking News Email Aggregator](https://raw.githubusercontent.com/jprobinson/newshound/master/web/app/images/newshound_logo.png)](http://newshound.jprbnsn.com)
Newshound: The Breaking News Email Aggregator
=========
[![GoDoc](https://godoc.org/github.com/jprobinson/newshound?status.svg)](https://godoc.org/github.com/jprobinson/newshound)
[![Build Status](https://cloud.drone.io/api/badges/jprobinson/newshound/status.svg)](https://cloud.drone.io/jprobinson/newshound)

Newshound is a tool to analyze, visualize and share breaking news email alerts.

This repository contains a [service to pull and parse breaking news alerts from an email inbox](https://github.com/jprobinson/newshound/tree/master/fetch) and a [fast noun-phrase extracting 'microservice'](https://github.com/jprobinson/newshound/tree/master/np_extractor) to extract important phrases and help detect any News Events that may have occurred. That News Event data is then used to generate historic reports for each news source.

To emit alert notifications to Slack or Twitter, [fetchd](https://github.com/jprobinson/newshound/tree/master/fetch/fetchd) can pass information to [barkd](https://github.com/jprobinson/newshound/tree/master/bark/barkd) via [Google Cloud Pub/Sub.](https://cloud.google.com/pubsub/docs/overview) 

There is also a [web server](https://github.com/jprobinson/newshound/tree/master/web) and an [API](https://github.com/jprobinson/newshound/tree/master/api) for displaying and sharing Newshound information. 
