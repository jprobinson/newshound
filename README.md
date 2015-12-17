[![Newshound: The Breaking News Email Aggregator](https://raw.githubusercontent.com/jprobinson/newshound/master/web/frontend/app/images/newshound_logo.png)](http://newshound.jprbnsn.com)
Newshound: The Breaking News Email Aggregator
=========
[![GoDoc](https://godoc.org/github.com/jprobinson/newshound?status.svg)](https://godoc.org/github.com/jprobinson/newshound)
[![Build Status](https://travis-ci.org/jprobinson/newshound.svg?branch=master)](https://travis-ci.org/jprobinson/newshound)

Newshound is a tool to analyze, visualize and share breaking news email alerts.

This repository contains a [service to pull and parse breaking news alerts from an email inbox](https://github.com/jprobinson/newshound/tree/master/fetch) and a [fast noun-phrase extracting 'microservice'](https://github.com/jprobinson/newshound/tree/master/lib/np_extractor) to extract important phrases and help [detect any News Events](https://github.com/jprobinson/newshound/tree/master/common.go#L124) that may have occurred. That News Event data is then used to [generate historic reports for each news source.](https://github.com/jprobinson/newshound/blob/master/fetch/mapreduce.go) 

To emit alert notifications to Slack, Twitter or WebSocket connections, [fetchd](https://github.com/jprobinson/newshound/tree/master/fetch/fetchd) can pass information to [barkd](https://github.com/jprobinson/newshound/tree/master/bark/barkd) via [NSQ.](http://nsq.io/) 

There is also a [web server](https://github.com/jprobinson/newshound/tree/master/web/webserver) that can host a [UI](https://github.com/jprobinson/newshound/tree/master/web/frontend) and an [API](https://github.com/jprobinson/newshound/tree/master/web/webserver/api) for displaying and sharing Newshound information. 
