'use strict';

angular.module('newshoundApp')
    .factory('config', ['$location',
        function($location) {
            return {
                apiHost: function() {
                    var apiHost = "svc/newshound-api/v1";
                    if ($location.host().indexOf('appspot.com') == -1) {
                        apiHost = "https://newshound.jprbnsn.com/" + apiHost;
                    }

                    return apiHost;
                }
            };
        }
    ]);
