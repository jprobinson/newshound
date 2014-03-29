'use strict';

angular.module('newshoundApp', [
	'ngResource',
	'ngSanitize',
	'ngRoute',
	'ui.calendar',
	'ui.bootstrap',
	'ui.select2',
	'highcharts-ng'
])
	.config(['$locationProvider', '$routeProvider', '$httpProvider',
		function($locationProvider, $routeProvider, $httpProvider) {
			$routeProvider
				.when('/', {
					// templateUrl: 'views/main.html',
					// controller: 'MainCtrl'
					redirectTo: '/calendar'
				})
				.when('/calendar', {
					templateUrl: 'views/calendar.html',
					controller: 'CalendarCtrl',
					reloadOnSearch: false
				})
				.when('/report', {
					templateUrl: 'views/report.html',
					controller: 'ReportCtrl',
					reloadOnSearch: false					
				})
				.otherwise({
					redirectTo: '/calendar'
				});

			//Enable cross domain calls
			$httpProvider.defaults.useXDomain = true;
			//Remove the header used to identify ajax call  that would prevent CORS from working
			delete $httpProvider.defaults.headers.common['X-Requested-With'];

		}
	]);
