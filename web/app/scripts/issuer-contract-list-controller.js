/**
 * @class IssuerContractListController
 * @classdesc
 * @ngInject
 */
function IssuerContractListController($scope, $log, $interval, PeerService, $rootScope) {

  var ctl = this;


  var init = function() {
    ctl.reload();
    $rootScope.$on('chainblock', function(payload){
          ctl.reload();
    });
  };

  ctl.reload = function(){
    PeerService.getIssuerContracts().then(function(list) {
      ctl.list = list;
    });
  };

  $scope.$on('$viewContentLoaded', init);
}

angular.module('issuerContractListController', [])
.controller('IssuerContractListController', IssuerContractListController);