'use strict';

angular.module('newshoundApp')
	.controller('NavbarCtrl', ['$scope', '$location',
		function($scope, $location) {
			$scope.isActive = function(viewLocation, root) {
                if(root) {
                    return $location.path() === viewLocation;
                }
				return $location.path().indexOf(viewLocation) != -1;
			};
		}
	]);
