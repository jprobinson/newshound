'use strict';

angular.module('newshoundApp')
    .controller('CalendarCtrl', ['$filter', '$route', '$scope', '$sce', '$window', '$location', '$document', '$modal', 'config', 'news', 'senders', 'senderColors',
        function($filter, $route, $scope, $sce, $window, $location, $document, $modal, config, news, senders, senderColors) {
            var urlDisplay = $location.search().display;
            var defaultDisplay = "alerts";
            var displayInit = true;

            if (urlDisplay) {
                defaultDisplay = urlDisplay;
            }
            $scope.calDisplay = defaultDisplay;

            $scope.$on('$routeUpdate', function() {
                var searchVals = $location.search();
                var newStart = searchVals.start;
                var newEnd = searchVals.end;
                var newDisplay = searchVals.display;
                var newAlert = searchVals.alert;
                var newEvent = searchVals.event;

                if ((newStart != $scope.startDate) ||
                    (newEnd != $scope.endDate) && newStart) {

                    if (newDisplay != $scope.calDisplay) {
                        $scope.calDisplay = newDisplay;
                    }

                    goToDate(newStart);
                } else if (newDisplay != $scope.calDisplay) {
                    $scope.theCalendar.fullCalendar('refetchEvents');
                }

                if (newEvent != $scope.eventId) {
                    if (newEvent) {
                        displayEventDialog(newEvent);
                    } else if ($scope.eventDialog) {
                        $scope.eventDialog.close();
                    }
                }

                if (newAlert != $scope.alertId) {
                    if (newAlert) {
                        displayAlertDialog(newAlert);
                    } else if ($scope.alertDialog) {
                        $scope.alertDialog.close();
                    }
                }

            });

            var goToDate = function(dateStr) {
                if (dateStr) {
                    var dateVals = dateStr.split('-');
                    if (dateVals.length != 3) {
                        alert('invalid date!');
                    } else {
                        var startDate = new Date(dateVals[0], dateVals[1] - 1, dateVals[2]);
                        $scope.theCalendar.fullCalendar('gotoDate', startDate);
                    }
                } else {
                    $route.reload();
                }
            };

            

            $scope.senders = senders;
            $scope.$watch('calDisplay', function() {
                if (displayInit) {
                    displayInit = false;
                    return;
                }

                var searchVars = $location.search();
                var oldDisplay = searchVars.display;
                if (oldDisplay != $scope.calDisplay) {
                    searchVars.display = $scope.calDisplay;
                    $location.search(searchVars);

                    $scope.theCalendar.fullCalendar('refetchEvents');
                }
            });

            $scope.senderFilter = "";
            var filterCalEvents = function() {
                var selectedSender = $scope.senderFilter;
                var hideEm = true;
                if (!selectedSender.length) {
                    hideEm = false;
                }
                // not a very 'angular' approach, I know, but cal events are dynamically generated so I can't 
                // drop an "ng-show" on them
                angular.forEach($window.document.getElementsByClassName('fc-event'), function(event, indx) {
                    event = angular.element(event);
                    if (hideEm) {
                        if (event.hasClass(selectedSender)) {
                            event.show();
                        } else {
                            event.hide();
                        }
                    } else {
                        event.show();
                    }
                });
            };
            $scope.$watch('senderFilter', filterCalEvents);

            var dateChange = function(newDate, oldDate) {
                var newDateStr = $filter('date')(newDate, "yyyy-MM-dd");
                if(newDate && 
                    ($scope.startDate >= newDateStr) || 
                    ($scope.endDate <= newDateStr)) {
                    $scope.theCalendar.fullCalendar('gotoDate', newDate);
                }
            }
            $scope.$watch('startInput', dateChange);
            $scope.$watch('endInput', dateChange);

            $scope.calEvents = [];
            var getCalEvents = function(start, end, callback) {
                var promise;
                var searchVals = $location.search();

                if (searchVals.display == "alerts") {
                    promise = news.getAlerts(start, end);
                } else {
                    promise = news.getEvents(start, end);
                }

                promise.then(function(events) {
                    callback(events);
                }, function(reason) {
                    alert('Failed getting event data!: ' + reason);
                    callback([]);
                });

            };

            var viewRender = function(view, element) {
                var start = $filter('date')(view.start, "yyyy-MM-dd");
                var end = $filter('date')(view.end, "yyyy-MM-dd");
                var searchVals = $location.search();
                if (!$scope.startDate && searchVals.start) {
                    start = searchVals.start;
                    $scope.startDate = start;
                    $scope.endDate = end;

                    var newEvent = searchVals.event;
                    if (newEvent != $scope.eventId) {
                        if (newEvent) {
                            displayEventDialog(newEvent);
                        } else if ($scope.eventDialog) {
                            $scope.eventDialog.close();
                        }
                    }

                    var newAlert = searchVals.alert;
                    if (newAlert != $scope.alertId) {
                        if (newAlert) {
                            displayAlertDialog(newAlert);
                        } else if ($scope.alertDialog) {
                            $scope.alertDialog.close();
                        }
                    }
                    goToDate(start);

                } else if ((start != $scope.startDate) || (end != $scope.endDate)) {
                    $scope.startDate = start;
                    $scope.startInput = start;
                    $scope.endDate = end;
                    $scope.endInput = end;
                    searchVals.start = $filter('date')(view.start, "yyyy-MM-dd");
                    searchVals.end = $filter('date')(view.end, "yyyy-MM-dd");
                    searchVals.display = $scope.calDisplay;
                    $location.search(searchVals);
                }
            };

            var viewNewsModal = function(event, jsEvent, view) {
                var id = event.obj_id;
                if ($scope.calDisplay == "alerts") {
                    displayAlertDialog(id);
                } else {
                    displayEventDialog(id);
                }
            };

            var displayAlertDialog = function(id) {
                var modalPromise = news.getAlert(id);
                modalPromise.then(function(alert) {
                    addAlertLocation(id);
                    $scope.alertDialog = $modal.open({
                        templateUrl: 'alertModal.html',
                        windowClass: 'large-modal',
                        controller: ['$scope', '$modalInstance', 'alert', function($scope, $modalInstance, alert) {
                            var htmlUrl = config.apiHost() + "/alert_html/" + alert.id;
                            $scope.alertHtmlUrl = $sce.trustAsResourceUrl(htmlUrl);
                            $scope.alert = alert;
                            $scope.senderClass = news.getSenderClassName(alert.sender);
                        }],
                        resolve: {
                            alert: function() {
                                return alert;
                            }
                        }
                    });

                    $scope.alertDialog.result.then(function(modal) {
                        clearAlertLocation();
                    }, function() {
                        clearAlertLocation();
                    });

                }, function(reason) {
                    alert('Failed getting alert: ' + reason);
                });
            };

            var addAlertLocation = function(id) {
                var searchVals = $location.search();
                searchVals.alert = id;
                $scope.alertId = id;
                $location.search(searchVals);
            };

            var clearAlertLocation = function() {
                $scope.alertId = undefined;
                $scope.alertDialog = undefined;
                var searchVals = $location.search();
                delete searchVals['alert'];
                $location.search(searchVals);
            };

            var addEventLocation = function(id) {
                var searchVals = $location.search();
                searchVals.event = id;
                $scope.eventId = id;
                $location.search(searchVals);
            };

            var clearEventLocation = function() {
                $scope.eventId = undefined;
                $scope.eventDialog = undefined;
                var searchVals = $location.search();
                delete searchVals['event'];
                $location.search(searchVals);
            };

            var displayEventDialog = function(id) {
                var modalPromise = news.getEvent(id);
                modalPromise.then(function(event) {
                    addEventLocation(id)
                    $scope.eventDialog = $modal.open({
                        templateUrl: 'eventModal.html',
                        windowClass: 'large-modal',
                        controller: ['$scope', '$modalInstance', 'event', function($scope, $modalInstance, event) {
                            $scope.event = event;

                            var maxLapsed = 0.0;
                            var seriesData = [];
                            var eventSenders = [];
                            $scope.eventChartSeries = [];
                            $.each(event.news_alerts, function(index, alert) {
                                var minDiff = alert.time_lapsed / 60;
                                var secDiff = alert.time_lapsed % 60;
                                var timeDiffStr = Math.floor(minDiff) + " minute(s), " + secDiff + " seconds";
                                seriesData.push({
                                    subject: alert.subject,
                                    time_diff_str: timeDiffStr,
                                    name: alert.sender,
                                    obj_id: alert.alert_id,
                                    y: minDiff,
                                    color: senderColors[news.getSenderClassName(alert.sender)],
                                    events: {
                                        click: function(event) {
                                            displayAlertDialog(event.point.obj_id);
                                        }
                                    }
                                });
                                eventSenders.push(alert.sender);
                                if (minDiff > maxLapsed) {
                                    maxLapsed = minDiff;
                                }
                            });
                            // bump it up a lil bit so theres something to click on.
                            seriesData[0].y = (maxLapsed / 100);

                            $scope.eventChartSeries.push({
                                name: 'News Alerts',
                                data: seriesData
                            });

                            $scope.eventChartConfig = {
                                options: {
                                    chart: {
                                        type: 'bar'
                                    },
                                    credits: {
                                        enabled: false
                                    },
                                    plotOptions: {
                                        bar: {
                                            cursor: 'pointer'
                                        }
                                    },
                                    title: {
                                        text: 'News Alert Time Differences'
                                    },
                                    tooltip: {
                                        formatter: function(something) {
                                            var time_str;
                                            if (this.y == seriesData[0].y) time_str = '<span style="font-weight:bold;">This was the first alert to arrive!</span>';
                                            else time_str = 'Arrived ' + this.point.time_diff_str + ' after first alert';
                                            return this.key + '<br/>' + time_str + '<br/>' + this.point.subject.substr(0, 80);
                                        }
                                    },
                                    xAxis: {
                                        categories: eventSenders,
                                    },
                                    yAxis: {
                                        minPadding: 1.5,
                                        title: {
                                            text: 'Time Elapsed (minutes)'
                                        }
                                    },
                                    legend: {
                                        enabled: false
                                    }
                                },
                                series: $scope.eventChartSeries
                            };

                        }],
                        resolve: {
                            event: function() {
                                return event;
                            }
                        }
                    });

                    $scope.eventDialog.result.then(function(modal) {
                        clearEventLocation();
                    }, function() {
                        clearEventLocation();
                    });

                }, function(reason) {
                    alert('Failed getting alert: ' + reason);
                });
            };

            var getCalViews = function(current){
                var views = "agendaWeek,agendaDay";
                var view = current;
                var width = $(document).width();
                if(width < 800){
                    views = "basicWeek,basicDay";
                    view = "basicDay";
                }
                return {view:view, views:views};
            };

            $scope.uiConfig = {
                calendar: {
                    theme: false,
                    defaultView: getCalViews('agendaWeek').view,
                    editable: false,
                    contentHeight: Math.max($(window).height() - 165, 300),
                    allDaySlot: false,
                    header: {
                        left: 'today prev,next',
                        center: '',
                        right: getCalViews('agendaWeek').views
                    },
                    windowResize: function(view) {
                        view.setHeight(Math.max($(document).height() - 250, 300));
                        var views = getCalViews(view); 
                        if(view != views.view){
                            view.changeView(views.view);
                        }
                    },
                    ignoreTimezone: false,
                    viewRender: viewRender,
                    eventAfterAllRender: filterCalEvents,
                    events: getCalEvents,
                    eventClick: viewNewsModal
                }
            };

        }
    ]);
