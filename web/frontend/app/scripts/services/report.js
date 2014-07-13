'use strict';

angular.module('newshoundApp')
	.factory('report', ['$http', '$q', 'config',
		function($http, $q, config) {
			// Service logic

			// Public API here
			return {
				getSenderReport: function(report_type) {
                    console.log("WHAAA");
                    var report_name = "";
                    switch (report_type) {
                        case "avg_events":
                        case "avg_rank":
                        case "avg_time_lapsed":
                            report_name = "events_per_week";
                            break;
                        case "avg_alerts":
                            report_name = "alerts_per_week";
                            break;
                        case "attendance":
                            report_name = "event_attendance";
                            break;
                    }
					var deferred = $q.defer();
					$http({
						url: config.apiHost() + "/report/" + report_name,
						method: "GET",
					}).success(function(data, status, headers, config) {
						deferred.resolve(data);
					}).error(function(data, status, headers, config) {
						console.log(status);
						console.log(headers);
						console.log(data);
						deferred.reject('we had a problem fetching the full sender report');
					});

					return deferred.promise;
				},

				getTotalReport: function() {
					var deferred = $q.defer();
					$http({
						url: config.apiHost() + "/total_report",
						method: "GET",
					}).success(function(data, status, headers, config) {
						deferred.resolve(data);
					}).error(function(data, status, headers, config) {
						console.log(status);
						console.log(headers);
						console.log(data);
						deferred.reject('we had a problem fetching the totals report');
					});

					return deferred.promise;
				},
				
				getSenderInfo: function(sender) {
					var deferred = $q.defer();
					$http({
						url: config.apiHost() + "/sender_info/"+sender,
						method: "GET",
					}).success(function(data, status, headers, config) {
						deferred.resolve(data);
					}).error(function(data, status, headers, config) {
						console.log(status);
						console.log(headers);
						console.log(data);
						deferred.reject('we had a problem fetching the sender info report');
					});

					return deferred.promise;
				}
			};
		}
	]);
