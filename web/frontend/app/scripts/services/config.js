'use strict';

angular.module('newshoundApp')
	.factory('config', ['$location',
		function($location) {
			return {
				apiHost: function() {
                    var apiHost = "svc/newshound-api/v1";
                    switch($location.host()){
                        // if you're doing local dev work, just point to prd API
                        case '0.0.0.0':
                            apiHost = "http://newshound.jprbnsn.com/" + apiHost;
                            break;
                    }
					return apiHost;
				}
			};
		}
	]);
