/**
 * @class IssuerContractListController
 * @classdesc
 * @ngInject
 */
function IssuerContractListController($scope, $log, $interval, PeerService, $rootScope) {

  var ctl = this;

  $scope.$on('$viewContentLoaded', init);
  
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

}

angular.module('issuerContractListController', [])
.controller('IssuerContractListController', IssuerContractListController);