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

					var startString = $filter('date')(start, "yyyy-MM-dd");
					var endString = $filter('date')(end, "yyyy-MM-dd");
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

				getEvents: function(start, end) {
					var deferred = $q.defer();

					var startString = $filter('date')(start, "yyyy-MM-dd");
					var endString = $filter('date')(end, "yyyy-MM-dd");
					$http({
						url: config.apiHost() + "/find_events/" + startString + "/" + endString,
						method: "GET",
					}).success(function(data, status, headers, config) {
						var events = [];
						if (data != "null") {
							angular.forEach(data, function(event, index) {
								var newsAlerts = event.news_alerts;
								var senderClasses = [];
								if (newsAlerts.length <= 3) {
									senderClasses.push('news_event_cal_small');
								} else if (newsAlerts.length >= 10) {
									senderClasses.push('news_event_cal_large');
								} else {
									senderClasses.push('news_event_cal_' + new String(newsAlerts.length));
								}
								senderClasses.push('news_event_cal');

								angular.forEach(newsAlerts, function(alert, index) {
									var senderClass = getSenderClassName(alert.sender);
									senders[senderClass] = alert.sender;
									senderClasses.push(senderClass);
								});

								events.push({
									title: event.tags.join(),
									start: event.event_start,
									obj_id: event.id,
									event_start: event.event_start,
									news_alerts: newsAlerts,
									tags: event.tags,
									end: event.event_end,
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
