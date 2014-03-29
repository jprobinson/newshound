'use strict';

angular.module('newshoundApp')
	.controller('NavbarCtrl', ['$scope', '$location',
		function($scope, $location) {
			$scope.isActive = function(viewLocation) {
				return $location.path().indexOf(viewLocation) != -1;
			};
		}
	]);
