'use strict';

angular.module('newshoundApp')
    .factory('config', ['$location',
        function($location) {
            return {
                apiHost: function() {
                    return "https://api-dot-newshound.appspot.com/svc/newshound-api/v1";
                }
            };
        }
    ]);
