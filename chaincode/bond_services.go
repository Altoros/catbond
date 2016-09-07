package main

import (
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"errors"
	"strings"
	"fmt"
)

//issuerId: 'issuer0',
//id: 'issuer0.2017.6.13.600',
//principal: 500000,
//term: 12,
//maturityDate: '2017.6.13',
//rate: 600,
//trigger: 'hurricane 2 FL',
//state: 'offer'


type bond struct {
	IssuerId       string `json:"issuerId"`
	Id             string `json:"id"`
	Principal      uint64 `json:"principal"`
	Term           uint64 `json:"term"`
	MaturityDate   string `json:"maturityDate"`
	Rate           uint64 `json:"rate"`
	Trigger        string `json:"trigger"`
	State          string `json:"state"`
	CouponsPaid    uint64 `json:"couponsPaid"`
}

func (bond_ *bond) toRow() (shim.Row) {
	return shim.Row{
		Columns: []*shim.Column{
			&shim.Column{Value: &shim.Column_String_{String_: bond_.IssuerId}},
			&shim.Column{Value: &shim.Column_String_{String_: bond_.Id}},
			&shim.Column{Value: &shim.Column_Uint64{Uint64: bond_.Principal}},
			&shim.Column{Value: &shim.Column_Uint64{Uint64: bond_.Term}},
			&shim.Column{Value: &shim.Column_String_{String_: bond_.MaturityDate}},
			&shim.Column{Value: &shim.Column_Uint64{Uint64: bond_.Rate}},
			&shim.Column{Value: &shim.Column_String_{String_: bond_.Trigger}},
			&shim.Column{Value: &shim.Column_String_{String_: bond_.State}},
			&shim.Column{Value: &shim.Column_Uint64{Uint64: bond_.CouponsPaid}}},
	}
}



func (t *BondChaincode) initBonds(stub shim.ChaincodeStubInterface) (error) {
	// Create bonds table
	err := stub.CreateTable("Bonds", []*shim.ColumnDefinition{
		&shim.ColumnDefinition{Name: "IssuerId", Type: shim.ColumnDefinition_STRING, Key: true},
		&shim.ColumnDefinition{Name: "ID", Type: shim.ColumnDefinition_STRING, Key: true},
		&shim.ColumnDefinition{Name: "Principal", Type: shim.ColumnDefinition_UINT64, Key: false},
		&shim.ColumnDefinition{Name: "Term", Type: shim.ColumnDefinition_UINT64, Key: false},
		&shim.ColumnDefinition{Name: "MaturityDate", Type: shim.ColumnDefinition_STRING, Key: false},
		&shim.ColumnDefinition{Name: "Rate", Type: shim.ColumnDefinition_UINT64, Key: false},
		&shim.ColumnDefinition{Name: "Trigger", Type: shim.ColumnDefinition_STRING, Key: false},
		&shim.ColumnDefinition{Name: "State", Type: shim.ColumnDefinition_STRING, Key: false},
		&shim.ColumnDefinition{Name: "CouponsPaid", Type: shim.ColumnDefinition_UINT64, Key: false},
	})
	if err != nil {
		log.Criticalf("Cannot initialize Bonds")
		return errors.New("Failed creating Bonds table.")
	}

	return nil
}

func (t *BondChaincode) getBonds(stub shim.ChaincodeStubInterface, issuerID string) ([]bond, error) {
	var columns []shim.Column
	if issuerID != "" {
		columnIssuerIDs := shim.Column{Value: &shim.Column_String_{String_: issuerID}}
		columns = append(columns, columnIssuerIDs)
	}

	rows, err := stub.GetRows("Bonds", columns)
	if err != nil {
		message := "Failed retrieving bonds. Error: " + err.Error()
		log.Error(message)
		return nil, errors.New(message)
	}

	var bonds []bond

	for row := range rows {
		result := bond{
			IssuerId:       row.Columns[0].GetString_(),
			Id:             row.Columns[1].GetString_(),
			Principal:      row.Columns[2].GetUint64(),
			Term:           row.Columns[3].GetUint64(),
			MaturityDate:   row.Columns[4].GetString_(),
			Rate:           row.Columns[5].GetUint64(),
			Trigger:        row.Columns[6].GetString_(),
			State:          row.Columns[7].GetString_(),
			CouponsPaid:    row.Columns[8].GetUint64()}

		log.Debugf("getBonds result includes: %+v", result)
		bonds = append(bonds, result)
	}

	return bonds, nil
}

func (t *BondChaincode) archiveBond(stub shim.ChaincodeStubInterface, bond_ bond) (error) {

	// Archive related contracts
	contracts, err := t.getIssuerContracts(stub, bond_.IssuerId)
	if err != nil {
		return fmt.Errorf("archiveBond operation failed. cannot get contracts %s", err)
	}
	for _, contract_ := range contracts {
		if bond_.Id == contract_.Id[:strings.LastIndex(contract_.Id, ".")]{
			err = t.archiveContract(stub, contract_)
			if err != nil {
				return err
			}
		}
	}

	// Archive bond
	var columns []shim.Column
	columnIssuerIDs := shim.Column{Value: &shim.Column_String_{String_: bond_.IssuerId}}
	columns = append(columns, columnIssuerIDs)
	columnID := shim.Column{Value: &shim.Column_String_{String_: bond_.Id}}
	columns = append(columns, columnID)

	err = stub.DeleteRow("Bonds", columns)
	if err != nil {
		return fmt.Errorf("archiveBond operation failed. %s", err)
	}

	return nil
}


func (t *BondChaincode) getBond(stub shim.ChaincodeStubInterface, issuerID string, bondId string) (bond, error) {
	var columns []shim.Column
	columnIssuerIDs := shim.Column{Value: &shim.Column_String_{String_: issuerID}}
	columns = append(columns, columnIssuerIDs)
	columnBondId := shim.Column{Value: &shim.Column_String_{String_: bondId}}
	columns = append(columns, columnBondId)

	row, err := stub.GetRow("Bonds", columns)
	if err != nil {
		message := "Failed retrieving bonds. Error: " + err.Error()
		log.Error(message)
		return bond{}, errors.New(message)
	}

	result := bond{
		IssuerId:       row.Columns[0].GetString_(),
		Id:             row.Columns[1].GetString_(),
		Principal:      row.Columns[2].GetUint64(),
		Term:           row.Columns[3].GetUint64(),
		MaturityDate:   row.Columns[4].GetString_(),
		Rate:           row.Columns[5].GetUint64(),
		Trigger:        row.Columns[6].GetString_(),
		State:          row.Columns[7].GetString_(),
		CouponsPaid:    row.Columns[8].GetUint64()}

	log.Debugf("getBonds result includes: %+v", result)

	return result, nil
}

func (t *BondChaincode) createBond(stub shim.ChaincodeStubInterface, bond_ bond) ([]byte, error) {
	//TODO Verify if bond with such id is created already

	if ok, err := stub.InsertRow("Bonds", shim.Row{
		Columns: []*shim.Column{
			&shim.Column{Value: &shim.Column_String_{String_: bond_.IssuerId}},
			&shim.Column{Value: &shim.Column_String_{String_: bond_.Id}},
			&shim.Column{Value: &shim.Column_Uint64{Uint64: bond_.Principal}},
			&shim.Column{Value: &shim.Column_Uint64{Uint64: bond_.Term}},
			&shim.Column{Value: &shim.Column_String_{String_: bond_.MaturityDate}},
			&shim.Column{Value: &shim.Column_Uint64{Uint64: bond_.Rate}},
			&shim.Column{Value: &shim.Column_String_{String_: bond_.Trigger}},
			&shim.Column{Value: &shim.Column_String_{String_: bond_.State}},
			&shim.Column{Value: &shim.Column_Uint64{Uint64: bond_.CouponsPaid}}},
	}); !ok {
		log.Error("Failed inserting new bond: " + err.Error())
		return nil, err
	}

	return nil, nil
}

//func (t *BondChaincode) couponsPaid(stub shim.ChaincodeStubInterface, issuerId string, bondId string) ([]byte, error) {
//	log.Debugf("couponsPaid called with issuerId:%s, bondId:%s", issuerId, bondId)
//
//	// Get all contracts issued by issuerId
//	contracts, err := t.getIssuerContracts(stub, issuerId)
//	if err != nil {
//		log.Error("couponsPaid failed on retrieving contracts: " + err.Error())
//		return nil, err
//	}
//
//	// Iterate over the contracts and increment those that match bondId
//	matchCounter := 0
//	for _, contract_ := range contracts {
//		// "issuer0.2017.6.13.600" expected after trimming a suffix from "issuer0.2017.6.13.600.42"
//		if bondId == contract_.Id[:strings.LastIndex(contract_.Id, ".")] && contract_.State=="active"  {
//			matchCounter++
//			if _, err := t.payContractCoupon(stub, contract_); err != nil {
//				log.Errorf("couponsPaid failed on paying coupon for %s: %s", contract_.Id, err.Error())
//				return nil, err
//			}
//		}
//	}
//	log.Debugf("couponsPaid: %d out of %d issued by %s matched %s and were paid",
//		   matchCounter, len(contracts), issuerId, bondId)
//
//	return nil, nil
//}


func (t *BondChaincode) payCoupons(stub shim.ChaincodeStubInterface) ([]byte, error) {
	log.Debugf("couponsPaid called ")

	// Get all contracts issued by issuerId
	contracts, err := t.getAllContracts(stub)
	if err != nil {
		log.Error("couponsPaid failed on retrieving contracts: " + err.Error())
		return nil, err
	}

	// Iterate over the contracts and increment those that match bondId
	matchCounter := 0
	for _, contract_ := range contracts {
		// "issuer0.2017.6.13.600" expected after trimming a suffix from "issuer0.2017.6.13.600.42"
		if contract_.State=="active"  {
			bondId := contract_.Id[:strings.LastIndex(contract_.Id, ".")]
			log.Debugf("try to load bond with id %s ", bondId)
			bond, err := t.getBond(stub, contract_.IssuerId, bondId)
			log.Debugf("getBonds: %+v ", bond)
			if err != nil {
				log.Error("cannot load bond for contract: " + err.Error())
				continue
			}
			price := (PRICE_PER_CONTRACT / bond.Term) + (PRICE_PER_CONTRACT / bond.Term ) * uint64((bond.Rate / 100.0) / 12.0)
			t.submitPaymentInstruction(stub, contract_.IssuerId, contract_.OwnerId, price, "coupon", contract_.Id, "payContractCoupon", contract_.Id)
		}
	}
	log.Debugf("couponsPaid: %d out of %d i were paid",
		matchCounter, len(contracts))

	t.recordCouponPayment(stub)

	return nil, nil
}

func (t *BondChaincode) removeExpiredBonds(stub shim.ChaincodeStubInterface) (error) {
	log.Debugf("removeExpiredBonds called ")

	// Get all bonds
	bonds, err := t.getBonds(stub, "")
	if err != nil {
		log.Error("removeExpiredBonds failed on retrieving bonds: " + err.Error())
		return err
	}
	count := 0

	for _, bond_ := range bonds {
		if bond_.Term<=bond_.CouponsPaid  {
			err = t.archiveBond(stub, bond_)
			if err != nil {
				return err
			}
			count++
		}
	}
	log.Debugf("Expired Bonds Removed: %d out of %d",
		count, len(bonds))

	return nil
}

func (t *BondChaincode) recordCouponPayment(stub shim.ChaincodeStubInterface) (error) {
	log.Debugf("recordCouponPayment called ")

	// Get all bonds
	bonds, err := t.getBonds(stub, "")
	if err != nil {
		log.Error("recordCouponPayment failed on retrieving bonds: " + err.Error())
		return err
	}

	for _, bond_ := range bonds {
		bond_.CouponsPaid = bond_.CouponsPaid + 1
		if ok, err := stub.ReplaceRow("Bonds", bond_.toRow()); !ok {
			log.Error("Failed inserting CouponsPaid number: " + err.Error())
			return err
		}

	}

	return nil
}