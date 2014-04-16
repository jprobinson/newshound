'use strict';

angular.module('newshoundApp')
    .factory('config', ['$location',
        function($location) {
            return {
                apiHost: function() {
                    var apiHost = "svc/newshound-api/v1";
                    if ($location.host().indexOf('jprbnsn.com') == -1) {
                        //apiHost = "http://localhost:8080/"+ apiHost;
                        apiHost = "http://newshound.jprbnsn.com/" + apiHost;
                    }
                    return apiHost;
                }
            };
        }
    ]);
