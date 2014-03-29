'use strict';

angular.module('newshoundApp')
	.factory('config', ['$location',
		function($location) {
			return {
				apiHost: function() {
					return 'svc/newshound-api/v1';
				}
			};
		}
	]);
