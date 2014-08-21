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
                if (index == 0) {
                    return true;
                }

                var prevDate = new Date($scope.events[index - 1].event_start);
                var currDate = new Date($scope.events[index].event_start);
                return (prevDate.getDate() != currDate.getDate()) ||
                    (prevDate.getFullYear() != currDate.getFullYear()) ||
                    (prevDate.getMonth() != currDate.getMonth());
            }

            $scope.toggleEmail = function(alert_id) {
                if ($scope.collapseEmails[alert_id]) {
                    $('#alert-email-'+alert_id).attr('src',config.apiHost() + "/alert_html/" + alert_id);
                    $scope.collapseEmails[alert_id] = false;
                } else {
                    $scope.collapseEmails[alert_id] = true;
                    $('#alert-email-'+alert_id).attr('src','');
                }
            };

            $scope.senderClass = news.getSenderClassName;
            $scope.collapseAlerts = {};
            $scope.collapseEmails = {};
            var getEvents = function(start, end) {
                var promise = news.getEvents(start, end, true);
                promise.then(function(events) {
                    angular.forEach(events, function(ev, idx) {
                        $scope.collapseAlerts[ev.obj_id] = true;
                        $scope.events.push(ev);
                        angular.forEach(ev.news_alerts, function(al, idex) {
                            $scope.collapseEmails[al.alert_id] = true;
                        });
                    });
                }, function(reason) {
                    console.log('Failed getting event data!: ' + reason);
                    $scope.events = [];
                });
            };

            $scope.prevDate = new Date();
            $scope.loadMore = function() {
                var today = $scope.prevDate;
                var endDate = new Date(today.getFullYear(), today.getMonth(), today.getDate() - 7);
                getEvents(endDate, today);
                $scope.prevDate = new Date(endDate.getFullYear(), endDate.getMonth(), endDate.getDate() - 1);
            }
            $scope.loadMore();
        }
    ]);
