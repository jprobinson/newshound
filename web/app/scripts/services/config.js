'use strict';

angular.module('newshoundApp')
    .factory('config', ['$location',
        function($location) {
            return {
                apiHost: function() {
                    return "https://api.newshound.email/svc/newshound-api/v1";
                }
            };
        }
    ]);
