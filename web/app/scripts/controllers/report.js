'use strict';

angular.module('newshoundApp')
    .controller('ReportCtrl', ['$scope', '$window', '$location', '$modal', '$filter', 'news', 'report',
        function($scope, $window, $location, $modal, $filter, news, report) {
            $scope.$on('$routeUpdate', function() {
                var searchVals = $location.search();
                var newSender = searchVals.sender;
                if (newSender != $scope.senderName) {
                    if (newSender) {
                        $scope.senderName = newSender;
                        displaySenderDialog();
                    } else if ($scope.senderDialog) {
                        $scope.senderDialog.close();
                    }
                }
            });

            $scope.curr_report = 'avg_alerts';
            $scope.curr_report_name = 'Average Alerts Per Week';

            $scope.update = function(curr_report) {
                // build initial report rows
                var senders = report.getSenderReport(curr_report);
                senders.then(function(data) {
                    $scope.senderReport = data;
                }, function(reason) {
                    alert(reason);
                });
            }

            $scope.update($scope.curr_report);

            // handle table sorting
            $scope.sort = {
                column: 'sender',
                descending: false
            };
            $scope.changeSorting = function(column) {
                var sort = $scope.sort;
                if (sort.column == column) {
                    $scope.sort.descending = !sort.descending;
                    $scope.sort.column = column;
                } else {
                    $scope.sort.column = column;
                    $scope.sort.descending = !sort.descending;
                }
            };

            var createPhrasesChart = function(senderName, info) {
                var frequencies = [];
                var words = [];
                angular.forEach(info.tag_array, function(tag_info, indx) {
                    frequencies.push({
                        y: tag_info.frequency
                    });
                    words.push(tag_info.tag);
                });
                var chart_data = [{
                    name: senderName,
                    data: frequencies
                }];

                return {
                    options: {
                        chart: {
                            type: 'bar'
                        },
                        credits: {
                            enabled: false
                        },
                        title: {
                            text: 'Top 15 Key Phrases Over Last 3 Months'
                        },
                        xAxis: {
                            categories: words,
                            title: {
                                text: 'Key Phrases'
                            }
                        },
                        yAxis: {
                            minPadding: 0.2,
                            allowDecimals: false,
                            title: {
                                text: 'Occurences'
                            },
                            min: 0
                        },
                        legend: {
                            enabled: false
                        },
                    },
                    series: chart_data
                };
            };

            var createAlertsChart = function(senderName, info) {
                var alerts = [];
                var weeks = [];
                angular.forEach(info.alerts_per_week, function(alert_week, indx) {
                    alerts.push({
                        y: alert_week.value.alerts,
                        date: $filter('date')(alert_week._id.week_start, "yyyy-MM-dd")
                    });

                    var week_str = $filter('date')(alert_week._id.week_start, "MM/dd");
                    weeks.push(week_str);
                });
                var chart_data = [{
                    name: senderName,
                    data: alerts
                }];
                return {
                    options: {
                        chart: {
                            type: 'line'
                        },
                        credits: {
                            enabled: false
                        },
                        plotOptions: {
                            line: {
                                allowPointSelect: true,
                                events: {
                                    click: function(event) {
                                        var url = $location.absUrl();
                                        var urlBase = url.split('#')[0];
                                        $window.open(urlBase + '#/calendar?display=alerts&start=' + event.point.date + '');
                                    }
                                }
                            }
                        },
                        title: {
                            text: 'Alerts Per Week Over Last 3 Months'
                        },
                        xAxis: {
                            categories: weeks,
                            title: {
                                text: 'Week'
                            }
                        },
                        yAxis: {
                            minPadding: 0.2,
                            allowDecimals: false,
                            title: {
                                text: '# of Alerts'
                            },
                            min: 0
                        },
                        legend: {
                            enabled: false
                        }
                    },
                    series: chart_data
                };
            };

            var createEventsChart = function(senderName, info) {
                var events = [];
                var weeks = [];
                angular.forEach(info.events_per_week, function(event_week, indx) {
                    events.push({
                        y: event_week.value.total_events,
                        date: $filter('date')(event_week._id.week_start, "yyyy-MM-dd")
                    });

                    var week_str = $filter('date')(event_week._id.week_start, "MM/dd");
                    weeks.push(week_str);
                });
                var chart_data = [{
                    name: senderName,
                    data: events
                }];
                return {
                    options: {
                        chart: {
                            type: 'line'
                        },
                        credits: {
                            enabled: false
                        },
                        plotOptions: {
                            line: {
                                allowPointSelect: true,
                                events: {
                                    click: function(event) {
                                        var url = $location.absUrl();
                                        var urlBase = url.split('#')[0];
                                        $window.open(urlBase + '#/calendar?display=events&start=' + event.point.date + '');
                                    }
                                }
                            }
                        },
                        title: {
                            text: 'Events Per Week Over Last 3 Months'
                        },
                        xAxis: {
                            categories: weeks,
                            title: {
                                text: 'Week'
                            }
                        },
                        yAxis: {
                            minPadding: 0.2,
                            allowDecimals: false,
                            title: {
                                text: '# of Events'
                            },
                            min: 0
                        },
                        legend: {
                            enabled: false
                        }
                    },
                    series: chart_data
                };
            };

            var createPlacementChart = function(senderName, info) {
                var events = [];
                var weeks = [];
                angular.forEach(info.events_per_week, function(event_week, indx) {
                    events.push({
                        y: event_week.value.avg_rank,
                        date: $filter('date')(event_week._id.week_start, "yyyy-MM-dd")
                    });

                    var week_str = $filter('date')(event_week._id.week_start, "MM/dd");
                    weeks.push(week_str);
                });
                var chart_data = [{
                    name: senderName,
                    data: events
                }];
                return {
                    options: {
                        chart: {
                            type: 'line'
                        },
                        credits: {
                            enabled: false
                        },
                        plotOptions: {
                            line: {
                                allowPointSelect: true,
                                events: {
                                    click: function(event) {
                                        var url = $location.absUrl();
                                        var urlBase = url.split('#')[0];
                                        $window.open(urlBase + '#/calendar?display=events&start=' + event.point.date + '');
                                    }
                                }
                            }
                        },
                        title: {
                            text: 'Avg Placement Per Week Over Last 3 Months'
                        },
                        tooltip: {
                            formatter: function() {
                                var time_str = this.y.toPrecision(3);
                                return this.x + '<br/>' + this.series.name + ':  ' + time_str;
                            }
                        },
                        xAxis: {
                            categories: weeks,
                            title: {
                                text: 'Week'
                            }
                        },
                        yAxis: {
                            minPadding: 0.2,
                            allowDecimals: false,
                            title: {
                                text: 'Avg Placement'
                            },
                            min: 1
                        },
                        legend: {
                            enabled: false
                        }
                    },
                    series: chart_data
                };
            };

            var createArrivalChart = function(senderName, info) {
                var events = [];
                var weeks = [];
                angular.forEach(info.events_per_week, function(event_week, index) {
                    var time_lapsed = 0.0;
                    if (event_week.value.avg_time_lapsed != 0) {
                        time_lapsed = event_week.value.avg_time_lapsed / 60;
                    }

                    events.push({
                        y: time_lapsed,
                        date: $filter('date')(event_week._id.week_start, "yyyy-MM-dd")
                    });
                    var week_str = $filter('date')(event_week._id.week_start, "MM/dd");
                    weeks.push(week_str);
                });
                var chart_data = [{
                    name: senderName,
                    data: events
                }];
                return {
                    options: {
                        chart: {
                            type: 'line'
                        },
                        credits: {
                            enabled: false
                        },
                        plotOptions: {
                            line: {
                                allowPointSelect: true,
                                events: {
                                    click: function(event) {
                                        var url = $location.absUrl();
                                        var urlBase = url.split('#')[0];
                                        $window.open(urlBase + '#/calendar?display=events&start=' + event.point.date + '');
                                    }
                                }
                            }
                        },
                        title: {
                            text: 'Avg Time Lapsed Per Week Over Last 3 Months'
                        },
                        tooltip: {
                            formatter: function() {
                                var time_str = this.y.toPrecision(3);
                                return this.x + '<br/>' + this.series.name + ':  ' + time_str;
                            }
                        },
                        xAxis: {
                            categories: weeks,
                            title: {
                                text: 'Week'
                            }
                        },
                        yAxis: {
                            minPadding: 0.2,
                            allowDecimals: false,
                            title: {
                                text: 'Time Lapsed (min)'
                            },
                            min: 0
                        },
                        legend: {
                            enabled: false
                        }
                    },
                    series: chart_data
                };
            };

            var createHoursChart = function(senderName, info) {
                var hour_counts = [];
                var hours = [];
                angular.forEach(info.alerts_per_hour, function(count, hour) {
                    hour_counts.push({
                        y: count
                    });
                    hours.push(hour);
                });
                var chart_data = [{
                    name: senderName,
                    data: hour_counts
                }];
                return {
                    options: {
                        chart: {
                            type: 'column'
                        },
                        credits: {
                            enabled: false
                        },
                        plotOptions: {
                            bar: {
                                borderRadius: 5
                            }
                        },
                        title: {
                            text: 'Alerts Per Hour Over Last 3 Months'
                        },
                        xAxis: {
                            categories: hours,
                            title: {
                                text: 'Hour'
                            }
                        },
                        yAxis: {
                            minPadding: 0.2,
                            allowDecimals: false,
                            title: {
                                text: '# Alerts'
                            },
                            min: 0
                        },
                        legend: {
                            enabled: false
                        },
                    },
                    series: chart_data
                };
            };

            var senderDialogCtrl = ['$scope', '$modalInstance', 'info', 'senderReport', 'senderName',
                function($scope, $modalInstance, info, senderReport, senderName) {
                    $scope.senderReport = senderReport;
                    $scope.info = info;
                    $scope.senderName = senderName;
                    $scope.senderClass = news.getSenderClassName(senderName);
                    $scope.filterBySender = function(sender) {
                        return sender.sender == $scope.senderName;
                    }
                    $scope.senderPhrasesChartConfig = createPhrasesChart(senderName, info);
                    $scope.senderAlertsChartConfig = createAlertsChart(senderName, info);
                    $scope.senderEventsChartConfig = createEventsChart(senderName, info);
                    $scope.senderPlacementChartConfig = createPlacementChart(senderName, info);
                    $scope.senderArrivalChartConfig = createArrivalChart(senderName, info);
                    $scope.senderHoursChartConfig = createHoursChart(senderName, info);
                }
            ];

            // handle sender dialog
            var displaySenderDialog = function() {
                var modalPromise = report.getSenderInfo($scope.senderName);
                modalPromise.then(function(info) {
                    addSenderLocation($scope.senderName);
                    $scope.senderDialog = $modal.open({
                        templateUrl: 'senderInfoModal.html',
                        windowClass: 'xlarge-modal',
                        controller: senderDialogCtrl,
                        resolve: {
                            info: function() {
                                return info;
                            },
                            senderReport: function() {
                                return $scope.senderReport;
                            },
                            senderName: function() {
                                return $scope.senderName;
                            }
                        }
                    });

                    $scope.senderDialog.result.then(function(modal) {
                        clearSenderLocation();
                    }, function() {
                        clearSenderLocation();
                    });

                }, function(reason) {
                    alert('Failed getting sender info: ' + reason);
                });
            };

            // on load
            if (($location.search().sender) && !$scope.senderName) {
                $scope.senderName = $location.search().sender;
                displaySenderDialog();
            }
        }
    ]);
