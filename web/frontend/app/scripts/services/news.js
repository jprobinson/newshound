'use strict';

angular.module('newshoundApp')
	.factory('news', ['$http', '$filter', '$q', 'config', 'senders',
		function($http, $filter, $q, config, senders) {

			// Service logic
			var buildAlertEventTitle = function(alert) {
				var output = alert.sender + '\n\n';
				output += alert.subject;
				return output;
			};

			var getSenderClassName = function(sender) {
				return sender.replace(/(\s|\.|\!)/g, '').toLowerCase();
			};


			// Public API here
			return {
				// expose it elsewhere
				getSenderClassName: getSenderClassName,

				getAlerts: function(start, end) {
					var deferred = $q.defer();

					var startString = start.format("YYYY-MM-DD");
					var endString = end.format("YYYY-MM-DD");
					$http({
						url: config.apiHost() + "/find_alerts/" + startString + "/" + endString,
						method: "GET",
					}).success(function(data, status, headers, config) {
						var events = [];
						if (data != "null") {
							angular.forEach(data, function(alert, index) {
								var senderClass = getSenderClassName(alert.sender);
								var senderClasses = [senderClass];
								if (alert.hasOwnProperty("instance_id")) {
									senderClasses = [senderClass, "instance_" + alert.instance_id, "object_" + alert.id]
								}
								senders[senderClass] = alert.sender;
								events.push({
									title: buildAlertEventTitle(alert),
									start: alert.timestamp,
									obj_id: alert.id,
									allDay: false,
									className: senderClasses
								});
							});
						}
						deferred.resolve(events);
					}).error(function(data, status, headers, config) {
						console.log(status);
						console.log(headers);
						console.log(data);
						deferred.reject('we had a problem fetching alerts');
					});

					return deferred.promise;
				},

				getAlert: function(id) {
					var deferred = $q.defer();

					$http({
						url: config.apiHost() + "/alert/" + id,
						method: "GET",
					}).success(function(data, status, headers, config) {
						deferred.resolve(data);
					}).error(function(data, status, headers, config) {
						console.log(status);
						console.log(headers);
						console.log(data);
						deferred.reject('we had a problem fetching an alert!');
					});

					return deferred.promise;
				},

				getEvents: function(start, end, feed) {
					var deferred = $q.defer();
                    var uri = "/find_events/";
                    if (feed) {
                        uri = "/event_feed/";
                    }
					var startString = $filter('date')(start, "yyyy-MM-dd");
					var endString = $filter('date')(end, "yyyy-MM-dd");
					$http({
						url: config.apiHost() + uri + startString + "/" + endString,
						method: "GET",
					}).success(function(data, status, headers, config) {
						var events = [];
						if (data != "null") {
							angular.forEach(data, function(event, index) {
								var newsAlerts = event.news_alerts;
								var senderClasses = [];
                                var sizeClass = "";
								if (newsAlerts.length <= 3) {
                                    sizeClass = 'news_event_cal_small';
								} else if (newsAlerts.length >= 10) {
                                    sizeClass = 'news_event_cal_large';
								} else {
                                    sizeClass = 'news_event_cal_' + new String(newsAlerts.length);
								}
                                senderClasses.push(sizeClass);
								senderClasses.push('news_event_cal');

								angular.forEach(newsAlerts, function(alert, index) {
									var senderClass = getSenderClassName(alert.sender);
									senders[senderClass] = alert.sender;
									senderClasses.push(senderClass);
								});
                                var senderList = [];
                                for(var sender in senders) {
                                    senderList.push(sender);
                                }
                                
								events.push({
									title: event.top_sentence,
									start: event.event_start,
									obj_id: event.id,
									event_start: event.event_start,
									news_alerts: newsAlerts,
                                    top_sentence: event.top_sentence,
                                    top_sender: event.top_sender,
									tags: event.tags,
									end: event.event_end,
									allDay: false,
                                    size_class: sizeClass,
									className: senderClasses,
                                    senders:senderList 
								});
							});
						}
						deferred.resolve(events);
					}).error(function(data, status, headers, config) {
						console.log(status);
						console.log(headers);
						console.log(data);
						deferred.reject('we had a problem fetching events. check the console.')
					});

					return deferred.promise;
				},

				getEvent: function(id) {
					var deferred = $q.defer();

					$http({
						url: config.apiHost() + "/event/" + id,
						method: "GET",
					}).success(function(data, status, headers, config) {
						deferred.resolve(data);
					}).error(function(data, status, headers, config) {
						console.log(status);
						console.log(headers);
						console.log(data);
						deferred.reject('we had a problem fetching an event!');
					});

					return deferred.promise;
				},
			};
		}
	]);
