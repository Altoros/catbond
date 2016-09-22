window.disableThemeSettings = true;
angular.module('config', [])
.constant('cfg', 
    {
  endpoint: 'http://localhost:7050/chaincode',
  secureContext: 'user_type1_deadbeef',
  chaincodeID: "32e8308c13e647f8e430096e148a767cf22376fa76a519f5fdfe8c5758154ab03650b87e65ff7148b5289f7d9bdb9c7b7c829be7c7564a9abf512124a0b337a3",
//  users: [{id: 'issuer0', role: 'issuer', endpoint:'http://localhost:7050/chaincode'},
//          {id: 'issuer1', role: 'issuer', endpoint:'http://localhost:7050/chaincode'},
//          {id: 'investor0', role: 'investor', endpoint:'http://localhost:7050/chaincode'},
//          {id: 'investor1', role: 'investor', endpoint:'http://localhost:7050/chaincode'},
//          {id: 'auditor0', role: 'auditor', endpoint:'http://localhost:7050/chaincode'}],
  users: [{id: 'issuer0', role: 'issuer', endpoint:'http://catbond.altoros.com:7050/chaincode'},
          {id: 'issuer1', role: 'issuer', endpoint:'http://catbond.altoros.com:7050/chaincode'},
          {id: 'investor0', role: 'investor', endpoint:'http://swift.altoros.com:7050/chaincode'},
          {id: 'investor1', role: 'investor', endpoint:'http://swift.altoros.com:7050/chaincode'},
          {id: 'auditor0', role: 'auditor', endpoint:'http://52.91.72.177:7050/chaincode'}],
  triggers: ['hurricane 2 FL', 'earthquake 5 CA'],
  bonds: [{
            id: 'issuer0.2017.6.13.600',
            issuerId: 'issuer0',
            principal: 500000,
            term: 12,
            maturityDate: '2017.6.13',
            rate: 600,
            trigger: 'hurricane 2 FL',
            state: 'offer'
          }],
  contracts: [{
            id: 'issuer0.2017.6.13.600.0',
            issuerId: 'issuer0',
            bondId: 'issuer0.2017.6.13.600',
            ownerId: 'issuer0',
            couponsPaid: 0,
            state: 'offer'
          },
          {
            id: 'issuer0.2017.6.13.600.1',
            issuerId: 'issuer0',
            bondId: 'issuer0.2017.6.13.600',
            ownerId: 'issuer0',
            couponsPaid: 0,
            state: 'offer'
          },
          {
            id: 'issuer0.2017.6.13.600.2',
            issuerId: 'issuer0',
            bondId: 'issuer0.2017.6.13.600',
            ownerId: 'issuer0',
            couponsPaid: 0,
            state: 'offer'
          },
          {
            id: 'issuer0.2017.6.13.600.3',
            issuerId: 'issuer0',
            bondId: 'issuer0.2017.6.13.600',
            ownerId: 'issuer0',
            couponsPaid: 0,
            state: 'offer'
          },
          {
            id: 'issuer0.2017.6.13.600.4',
            issuerId: 'issuer0',
            bondId: 'issuer0.2017.6.13.600',
            ownerId: 'issuer0',
            couponsPaid: 0,
            state: 'offer'
          }],
    trades: [{
            id: 1000,
            contractId: 'issuer0.2017.6.13.600.0',
            sellerId: 'issuer0',
            price: 100,
            state: 'offer'
          },
          {
            id: 1001,
            contractId: 'issuer0.2017.6.13.600.1',
            sellerId: 'issuer0',
            price: 100,
            state: 'offer'
          },
          {
            id: 1002,
            contractId: 'issuer0.2017.6.13.600.2',
            sellerId: 'issuer0',
            price: 100,
            state: 'offer'
          },
          {
            id: 1003,
            contractId: 'issuer0.2017.6.13.600.3',
            sellerId: 'issuer0',
            price: 100,
            state: 'offer'
          },
          {
            id: 1004,
            contractId: 'issuer0.2017.6.13.600.4',
            sellerId: 'issuer0',
            price: 100,
            state: 'offer'
          }]
    }
);
