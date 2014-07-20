'use strict';

angular.module('newshoundApp')
    .controller('FeedCtrl', ['$filter', '$route', '$scope', '$sce', '$window', '$location', '$document', '$modal', 'config', 'news', 'senders', 'senderColors',
        function($filter, $route, $scope, $sce, $window, $location, $document, $modal, config, news, senders, senderColors) {
            $scope.events = [];
            $scope.timeDiff = function(time) {
                var minDiff = time / 60;
                var secDiff = time % 60;
                return Math.floor(minDiff) + " minute(s), " + secDiff + " seconds";
            };

            $scope.showDate = function(index) {
                if(index == 0){
                    return true;
                }

                var prevDate = new Date($scope.events[index-1].event_start);
                var currDate = new Date($scope.events[index].event_start);
                return (prevDate.getDate() != currDate.getDate()) ||
                        (prevDate.getFullYear() != currDate.getFullYear()) ||
                        (prevDate.getMonth() != currDate.getMonth());
            }

            $scope.senderClass = news.getSenderClassName;
            $scope.collapseAlerts = {};
            var getEvents = function(start, end) {
                var promise = news.getEvents(start, end, true);
                promise.then(function(events) {
                    console.log(events);
                    angular.forEach(events, function(ev, idx) {
                        $scope.collapseAlerts[ev.obj_id] = true;
                        $scope.events.push(ev);
                        console.log(events);
                    });
                }, function(reason) {
                    console.log('Failed getting event data!: ' + reason);
                    $scope.events = [];
                });
            };

            $scope.prevDate = new Date();
            $scope.loadMore = function() {
                console.log('hi');
                var today = $scope.prevDate;
                var endDate = new Date(today.getFullYear(), today.getMonth(), today.getDate() - 7);
                console.log(today);
                console.log(endDate);
                getEvents(endDate, today);
                $scope.prevDate = endDate;
            }
            $scope.loadMore();
        }
    ]);
