package main

import (
	"errors"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/op/go-logging"

	"encoding/json"
	"strconv"
)

var log = logging.MustGetLogger("bond-traiding")
const PRICE_PER_CONTRACT uint64 = 100000

// SimpleChaincode example simple Chaincode implementation
type BondChaincode struct {
}


func (t *BondChaincode) Init(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	log.Debugf("function: %s, args: %s", function, args)

	// Create bonds table
	err := t.initBonds(stub)
	if err != nil {
		log.Criticalf("function: %s, args: %s", function, args)
		return nil, errors.New("Failed creating Bond table.")
	}
	// Create contracts table
	err = t.initContracts(stub)
	if err != nil {
		log.Criticalf("function: %s, args: %s", function, args)
		return nil, errors.New("Failed creating Contracts table.")
	}
	// Create trades table
	err = t.initTrades(stub)
	if err != nil {
		log.Criticalf("function: %s, args: %s", function, args)
		return nil, errors.New("Failed creating Trades table.")
	}

	return nil, nil
}

func (t *BondChaincode) Invoke(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	log.Debugf("function: %s, args: %s", function, args)

	callerName := t.getCallerName(stub)
	callerRole := t.getCallerRole(stub)

	log.Debugf("role: %s, name: %s", callerRole, callerName)

	// Handle different functions
	if function == "createBond" {
		if len(args) != 4 {
			return nil, errors.New("Incorrect arguments. Expecting maturityDate, principal, rate and term.")
		}
		if callerRole != "issuer" {
			return nil, errors.New("Incorrect caller role. Expecting issuer.")
		}

		var newBond bond

		newBond.IssuerId = callerName
		newBond.MaturityDate = args[0]

		principal, err := strconv.ParseUint(args[1], 10, 64)
		if err != nil {
			return nil, errors.New("Incorrect principa. Uint64 expected.")
		}
		newBond.Principal = principal

		rate, err := strconv.ParseUint(args[2], 10, 64)
		if err != nil {
			return nil, errors.New("Incorrect rate. Uint64 expected.")
		}
		newBond.Rate = rate

		term, err := strconv.ParseUint(args[3], 10, 64)
		if err != nil {
			return nil, errors.New("Incorrect term. Uint64 expected.")
		}
		newBond.Term = term
		newBond.State = "active"
		newBond.Id = newBond.IssuerId + "." + newBond.MaturityDate + "." + strconv.FormatUint(newBond.Rate, 10)
		newBond.CouponsPaid = 0

		if msg, err := t.createBond(stub, newBond); err != nil {
			return msg, err
		}
		return t.createContractsForBond(stub, newBond, principal/PRICE_PER_CONTRACT)

	} else if function == "buy" {
		if callerRole != "investor" {
			return nil, errors.New("Incorrect caller role. Expecting investor.")
		}

		if len(args) != 1 {
			return nil, errors.New("Incorrect arguments. Expecting tradeId.")
		}

		tradeId, err := strconv.ParseUint(args[0], 10, 64)
		if err != nil {
			return nil, errors.New("Incorrect tradeId. Uint64 expected.")
		}
		
		return t.buy(stub, tradeId, callerName)

	} else if function == "confirm" {
		//TODO: uncomment code below when SecurityContext will be propagated in cross chaincode requests
		//if callerRole != "swiftagent" {
		//	return nil, errors.New("Incorrect caller role. Expecting swiftagent.")
		//}
		if len(args) != 1 {
			return nil, errors.New("Incorrect arguments. Expecting contractId")
		}
		return t.confirm(stub, args[0])

	} else if function == "payContractCoupon" {
		//TODO: uncomment code below when SecurityContext will be propagated in cross chaincode requests
		//if callerRole != "swiftagent" {
		//	return nil, errors.New("Incorrect caller role. Expecting swiftagent.")
		//}
		if len(args) != 1 {
			return nil, errors.New("Incorrect arguments. Expecting contractId")
		}
		_, err := t.payContractCoupon(stub, args[0])
		return nil, err
	} else if function == "sell" {
		if callerRole != "investor" {
			return nil, errors.New("Incorrect caller role. Expecting investor.")
		}
		if len(args) != 2 {
			return nil, errors.New("Incorrect arguments. Expecting contractId, price.")
		}

		price, err := strconv.ParseUint(args[1], 10, 64)
		if err != nil {
			return nil, errors.New("Incorrect price. Uint64 expected.")
		}

		return t.sell(stub, args[0], price, callerName)

	} else if function == "payCoupons" {
		if callerRole != "system" {
			return nil, errors.New("Incorrect caller role. Expecting system.")
		}
		if len(args) != 0 {
			return nil, errors.New("Incorrect arguments. No arguments expected.")
		}

		t.removeExpiredBonds(stub)
		return t.payCoupons(stub)

	} else if function == "setChainCodeId" {
		if callerRole != "system" {
			return nil, errors.New("Incorrect caller role. Expecting system.")
		}
		if len(args) != 1 {
			return nil, errors.New("Incorrect arguments. Expecting chaincodeID.")
		}
		error := stub.PutState("chaincodeid", []byte(args[0]))

		return nil, error

	} else {
		log.Errorf("function: %s, args: %s", function, args)
		return nil, errors.New("Received unknown function invocation")
	}
}



// Query callback representing the query of a chaincode
func (t *BondChaincode) Query(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	log.Debugf("function: %s, args: %s", function, args)

	role := t.getCallerRole(stub)
	user := t.getCallerName(stub)

	// Handle different functions
	if function == "getBonds" {
		if len(args) != 0 {
			return nil, errors.New("Incorrect arguments. Expecting no arguments.")
		}
		if role != "issuer" {
			return nil, errors.New("Incorrect caller role. Expecting issuer.")
		}

		bonds, err := t.getBonds(stub, user)
		if err != nil {
			return nil, err
		}

		return json.Marshal(bonds)

	} else if function == "getContracts" {
		if len(args) != 0 {
			return nil, errors.New("Incorrect arguments. Expecting no arguments.")
		}
		if role == "issuer" {
			contracts, err := t.getIssuerContracts(stub, user)
			if err != nil {
				return nil, err
			}
			return json.Marshal(contracts)

		} else if role == "investor" {
			contracts, err := t.getOwnerContracts(stub, user)
			if err != nil {
				return nil, err
			}
			return json.Marshal(contracts)

		} else if role == "auditor" {
			contracts, err := t.getAllContracts(stub)
			if err != nil {
				return nil, err
			}
			return json.Marshal(contracts)
		} else {
			return nil, errors.New("Incorrect caller role. Expecting investor, issuer or auditor.")
		}
	} else if function == "getTrades" {
		if len(args) != 0 {
			return nil, errors.New("Incorrect arguments. Expecting no arguments.")
		}
		if role == "auditor" {
			trades, err := t.getAllTrades(stub)
			if err != nil {
				return nil, err
			}

			return json.Marshal(trades)
		} else if role == "investor"{
			trades, err := t.getTradesByType(stub, "offer")
			if err != nil {
				return nil, err
			}

			return json.Marshal(trades)
		} else {
			return nil, errors.New("Incorrect caller role. Expecting investor or auditor.")
		}
	} else {
		log.Errorf("function: %s, args: %s", function, args)
		return nil, errors.New("Received unknown function invocation")
	}
}

func (t *BondChaincode) incrementAndGetCounter(stub shim.ChaincodeStubInterface, counterName string) (result uint64, err error) {
	if contractIDBytes, err := stub.GetState(counterName); err != nil {
		log.Errorf("Failed retrieving %s.", counterName)
		return result, err
	} else {
		result, _ = strconv.ParseUint(string(contractIDBytes), 10, 64)
	}
	result++
	if err = stub.PutState(counterName, []byte(strconv.FormatUint(result, 10))); err != nil {
		log.Errorf("Failed saving %s!", counterName)
		return result, err
	}
	return result, err
}

func (t *BondChaincode) getCallerAttribute(stub shim.ChaincodeStubInterface, attr string) (string) {
	value, err := stub.ReadCertAttribute(attr)
	if err != nil {
		log.Error("Failed fetching caller's attribute. Error: " + err.Error())
		return ""
	}
	log.Debugf("Caller %s is: %s", attr, value)
	return string(value)
}

func (t *BondChaincode) getCallerCompany(stub shim.ChaincodeStubInterface) (string) {
	return t.getCallerAttribute(stub, "company")
}

func (t *BondChaincode) getCallerName(stub shim.ChaincodeStubInterface) (string) {
	return t.getCallerAttribute(stub, "name")
}

func (t *BondChaincode) getCallerRole(stub shim.ChaincodeStubInterface) (string) {
	return t.getCallerAttribute(stub, "role")
}

func (t *BondChaincode) getCallBackChaincodeId(stub shim.ChaincodeStubInterface) (string, error) {
	chaincodeid, err := stub.GetState("chaincodeid")
	return string(chaincodeid), err
}

func main() {
	err := shim.Start(new(BondChaincode))
	if err != nil {
		log.Critical("Error starting CatbondChaincode: %s", err)
	}
}