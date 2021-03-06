package main

import (
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"errors"
	"strconv"
	"strings"
	"fmt"
)

type contract struct {
	IssuerId       string `json:"issuerId"`
	Id             string `json:"id"`
	OwnerId        string `json:"ownerId"`
	CouponsPaid    uint64 `json:"couponsPaid"`
	State          string `json:"state"`
	BondId	       string `json:"bondid"`
}

func (contract_ *contract) readFromRow(row shim.Row) {
	contract_.IssuerId 	= row.Columns[0].GetString_()
	contract_.Id 		= row.Columns[1].GetString_()
	contract_.OwnerId 	= row.Columns[2].GetString_()
	contract_.CouponsPaid 	= row.Columns[3].GetUint64()
	contract_.State 	= row.Columns[4].GetString_()
	contract_.BondId	= row.Columns[5].GetString_()

}

func (t *BondChaincode) initContracts(stub shim.ChaincodeStubInterface) (error) {
	// Create contracts table
	err := stub.CreateTable("Contracts", []*shim.ColumnDefinition{
		&shim.ColumnDefinition{Name: "IssuerId", Type: shim.ColumnDefinition_STRING, Key: true},
		&shim.ColumnDefinition{Name: "ID", Type: shim.ColumnDefinition_STRING, Key: true},
		&shim.ColumnDefinition{Name: "OwnerId", Type: shim.ColumnDefinition_STRING, Key: false},
		&shim.ColumnDefinition{Name: "CouponsPaid", Type: shim.ColumnDefinition_UINT64, Key: false},
		&shim.ColumnDefinition{Name: "State", Type: shim.ColumnDefinition_STRING, Key: false},
		&shim.ColumnDefinition{Name: "BondId", Type: shim.ColumnDefinition_STRING, Key: false},
	})
	if err != nil {
		log.Criticalf("Cannot initialize Contracts")
		return errors.New("Failed creating Contracts table.")
	}

	return nil
}

func (t *BondChaincode) createContractsForBond(stub shim.ChaincodeStubInterface, bond_ bond, numberOfContracts uint64) ([]byte, error) {

	log.Debugf("function: %s, args: %s", "createContractsForBond", bond_.Id)
	if numberOfContracts > 128 {
		return nil, errors.New("Wrong number of contracts to create for bond.")
	}

	contract_ := contract{IssuerId: bond_.IssuerId, OwnerId: bond_.IssuerId, State: "offer", BondId:bond_.Id}
	for numberOfContracts > 0 {
		numberOfContracts--
		contract_.Id = bond_.Id + "." + strconv.FormatUint(numberOfContracts, 10)
		if _, err := t.createContract(stub, contract_); err != nil {
			return nil, err
		}
		if _, err := t.createTradeForContract(stub, contract_, 100); err != nil {
			return nil, err
		}
	}
	return nil, nil
}

func (t *BondChaincode) createContract(stub shim.ChaincodeStubInterface, contract_ contract) ([]byte, error) {
	//TODO Verify if contract with such id is created already

	log.Debugf("function: %s, args: %s", "createContract", contract_.Id)

	if ok, err := stub.InsertRow("Contracts", shim.Row{
		Columns: []*shim.Column{
			&shim.Column{Value: &shim.Column_String_{String_: contract_.IssuerId}},
			&shim.Column{Value: &shim.Column_String_{String_: contract_.Id}},
			&shim.Column{Value: &shim.Column_String_{String_: contract_.OwnerId}},
			&shim.Column{Value: &shim.Column_Uint64{Uint64: contract_.CouponsPaid}},
			&shim.Column{Value: &shim.Column_String_{String_: contract_.State}},
			&shim.Column{Value: &shim.Column_String_{String_: contract_.BondId}}},
	}); !ok {
		log.Error("Failed inserting new contract: " + err.Error())
		return nil, err
	}

	return nil, nil
}

func (t *BondChaincode) getContract(stub shim.ChaincodeStubInterface, issuerId string, contractId string) (contract, error) {
	var columns []shim.Column
	col1 := shim.Column{Value: &shim.Column_String_{String_: issuerId}}
	columns = append(columns, col1)
	col2 := shim.Column{Value: &shim.Column_String_{String_: contractId}}
	columns = append(columns, col2)

	row, err := stub.GetRow("Contracts", columns)
	if err != nil {
		message := "Failed retrieving contract ID " + contractId + ". Error: " + err.Error()
		log.Error(message)
		return contract{}, errors.New(message)
	}

	var result contract
	result.readFromRow(row)
	log.Debugf("getContract result: %+v", result)
	return result, nil
}

func (t *BondChaincode) archiveContract(stub shim.ChaincodeStubInterface, contract_ contract) (error) {

	trade_, err := t.getTradeForContract(stub, contract_.Id, "")
	if err != nil {
		return fmt.Errorf("archiveContract operation failed. cannot get trade %s", err)
	}
	err = t.archiveTrade(stub, trade_.Id)
	if err != nil {
		return err
	}

	var columns []shim.Column
	columnIssuerIDs := shim.Column{Value: &shim.Column_String_{String_: contract_.IssuerId}}
	columns = append(columns, columnIssuerIDs)
	columnID := shim.Column{Value: &shim.Column_String_{String_: contract_.Id}}
	columns = append(columns, columnID)

	err = stub.DeleteRow("Contracts", columns)
	if err != nil {
		return fmt.Errorf("archiveContract operation failed. %s", err)
	}

	return nil
}

func (t *BondChaincode) getContractById(stub shim.ChaincodeStubInterface, contractId string) (contract, error) {
	// TODO: check this method
	bondId := strings.Split(contractId, ".")[0]
	issuerId := strings.Split(bondId, ".")[0]
	log.Debugf("getContractById with contractId:%s, bondId:%s, issuerId:%s", contractId, bondId, issuerId)
	return t.getContract(stub, issuerId, contractId)
}

func (t *BondChaincode) updateContract(stub shim.ChaincodeStubInterface, contract_ contract) (bool, error) {
	log.Debugf("updateContract: %+v", contract_)

	return stub.ReplaceRow("Contracts", shim.Row{
		Columns: []*shim.Column{
			&shim.Column{Value: &shim.Column_String_{String_: contract_.IssuerId}},
			&shim.Column{Value: &shim.Column_String_{String_: contract_.Id}},
			&shim.Column{Value: &shim.Column_String_{String_: contract_.OwnerId}},
			&shim.Column{Value: &shim.Column_Uint64{Uint64: contract_.CouponsPaid}},
			&shim.Column{Value: &shim.Column_String_{String_: contract_.State}},
			&shim.Column{Value: &shim.Column_String_{String_: contract_.BondId}}},
	})
}


//func (t *BondChaincode) changeContractState(stub shim.ChaincodeStubInterface, issuerId string, contract_ contract, newState string) (bool, error) {
//	log.Debugf("changeContractState with issuerId:%s and contractId:%s to %s", issuerId, contract_.Id, newState)
//
//	contract_.State = newState
//	return stub.ReplaceRow("Contracts", shim.Row{
//		Columns: []*shim.Column{
//			&shim.Column{Value: &shim.Column_String_{String_: contract_.IssuerId}},
//			&shim.Column{Value: &shim.Column_String_{String_: contract_.Id}},
//			&shim.Column{Value: &shim.Column_String_{String_: contract_.OwnerId}},
//			&shim.Column{Value: &shim.Column_Uint64{Uint64: contract_.CouponsPaid}},
//			&shim.Column{Value: &shim.Column_String_{String_: contract_.State}}},
//	})
//}
//
//func (t *BondChaincode) reserveContract(stub shim.ChaincodeStubInterface, contract_ contract, newOwner string) (bool, error) {
//	log.Debugf("reserveContract to %s", newOwner)
//
//	contract_.OwnerId = newOwner
//	contract_.State = "reserved"
//	return stub.ReplaceRow("Contracts", shim.Row{
//		Columns: []*shim.Column{
//			&shim.Column{Value: &shim.Column_String_{String_: contract_.IssuerId}},
//			&shim.Column{Value: &shim.Column_String_{String_: contract_.Id}},
//			&shim.Column{Value: &shim.Column_String_{String_: contract_.OwnerId}},
//			&shim.Column{Value: &shim.Column_Uint64{Uint64: contract_.CouponsPaid}},
//			&shim.Column{Value: &shim.Column_String_{String_: contract_.State}}},
//	})
//}


func (t *BondChaincode) payContractCoupon(stub shim.ChaincodeStubInterface, contractId string) (bool, error) {
	log.Debugf("payContractCoupon for: %s", contractId)

	contract_, err := t.getContractById(stub, contractId)

	if err != nil {
		log.Error("payContractCoupon failed on retrieving contract: " + err.Error())
		return false, err
	}

	contract_.CouponsPaid++
	return t.updateContract(stub, contract_)
}

func (t *BondChaincode) getIssuerContracts(stub shim.ChaincodeStubInterface, issuerId string) (contracts []contract, err error) {
	var columns []shim.Column
	if issuerId != "" {
		columnIssuerIDs := shim.Column{Value: &shim.Column_String_{String_: issuerId}}
		columns = append(columns, columnIssuerIDs)
	}

	rows, err := stub.GetRows("Contracts", columns)
	if err != nil {
		message := "Failed retrieving contracts. Error: " + err.Error()
		log.Error(message)
		return nil, errors.New(message)
	}

	for row := range rows {
		var result contract
		result.readFromRow(row)

		log.Debugf("getIssuerContracts result includes: %+v", result)
		contracts = append(contracts, result)
	}

	return contracts, nil
}

func (t *BondChaincode) getOwnerContracts(stub shim.ChaincodeStubInterface, ownerId string) (contracts []contract, err error) {
	rows, err := stub.GetRows("Contracts", []shim.Column{})
	if err != nil {
		message := "Failed retrieving contracts. Error: " + err.Error()
		log.Error(message)
		return nil, errors.New(message)
	}

	for row := range rows {
		if row.Columns[2].GetString_() != ownerId {
			continue
		}
		var result contract
		result.readFromRow(row)

		contracts = append(contracts, result)
		log.Debugf("getOwnerContracts result includes: %+v", result)
	}

	return contracts, nil
}

func (t *BondChaincode) getAllContracts(stub shim.ChaincodeStubInterface) (contracts []contract, err error) {
	rows, err := stub.GetRows("Contracts", []shim.Column{})
	if err != nil {
		message := "Failed retrieving contracts. Error: " + err.Error()
		log.Error(message)
		return nil, errors.New(message)
	}

	for row := range rows {
		var result contract
		result.readFromRow(row)

		log.Debugf("getIssuerContracts result includes: %+v", result)
		contracts = append(contracts, result)
	}

	return contracts, nil
}