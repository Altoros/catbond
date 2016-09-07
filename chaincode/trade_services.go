package main

import (
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"errors"
	"strconv"
	"fmt"
)

//trades: [{
//id: 1000,
//contractId: 'issuer0.2017.6.13.600.0',
//sellerId: 'issuer0',
//price: 100,
//state: 'offer'
//},

type trade struct {
	Id 		uint64 `json:"id"`
	ContractId 	string `json:"contractId"`
	SellerId 	string `json:"sellerId"`
	Price 		uint64 `json:"price"`
	State 		string `json:"state"`
}

func (t *BondChaincode) GetSwiftChaincodeToCall() string {
	chainCodeToCall := "823c5031c771067239dcedffed1f63c25800e13c61b242c573b2d77ac2efa73ab2cfc4a85d5d989c4290c0a0b19cdf958e30c8c645763ac057e917b81d15ec88"
	return chainCodeToCall
}

func (trade_ *trade) readFromRow(row shim.Row) {
	log.Debugf("readFromRow: %+v", row)
	trade_.Id 		= row.Columns[0].GetUint64()
	trade_.ContractId 	= row.Columns[1].GetString_()
	trade_.SellerId 	= row.Columns[2].GetString_()
	trade_.Price 		= row.Columns[3].GetUint64()
	trade_.State 		= row.Columns[4].GetString_()
}

func (trade_ *trade) toRow() (shim.Row) {
	return shim.Row{
		Columns: []*shim.Column{
			&shim.Column{Value: &shim.Column_Uint64{Uint64: trade_.Id}},
			&shim.Column{Value: &shim.Column_String_{String_: trade_.ContractId}},
			&shim.Column{Value: &shim.Column_String_{String_: trade_.SellerId}},
			&shim.Column{Value: &shim.Column_Uint64{Uint64: trade_.Price}},
			&shim.Column{Value: &shim.Column_String_{String_: trade_.State}}},
	}
}

func (t *BondChaincode) initTrades(stub shim.ChaincodeStubInterface) (error) {
	// Create trades table
	err := stub.CreateTable("Trades", []*shim.ColumnDefinition{
		&shim.ColumnDefinition{Name: "ID", Type: shim.ColumnDefinition_UINT64, Key: true},
		&shim.ColumnDefinition{Name: "ContractId", Type: shim.ColumnDefinition_STRING, Key: false},
		&shim.ColumnDefinition{Name: "SellerId", Type: shim.ColumnDefinition_STRING, Key: false},
		&shim.ColumnDefinition{Name: "Price", Type: shim.ColumnDefinition_UINT64, Key: false},
		&shim.ColumnDefinition{Name: "State", Type: shim.ColumnDefinition_STRING, Key: false},
	})
	if err != nil {
		log.Criticalf("Cannot initialize Trades")
		return errors.New("Failed creating Trades table.")
	}

	err = stub.PutState("TradesCounter", []byte(strconv.FormatUint(0, 10)))
	if err != nil {
		return err
	}

	return nil
}


func (t *BondChaincode) archiveTrade(stub shim.ChaincodeStubInterface, id uint64) (error) {

	var columns []shim.Column
	columnID := shim.Column{Value: &shim.Column_Uint64{Uint64: id}}
	columns = append(columns, columnID)

	err := stub.DeleteRow("Trades", columns)
	if err != nil {
		return fmt.Errorf("archiveTrade operation failed. %s", err)
	}

	return nil
}

func (t *BondChaincode) createTradeForContract(stub shim.ChaincodeStubInterface, contract_ contract, price uint64) ([]byte, error) {
	log.Debugf("function: %s, args: %s", "createTradeForContract", contract_.Id)
	var trade_ trade
	trade_.State = "offer"
	trade_.ContractId = contract_.Id

	counter, err := t.incrementAndGetCounter(stub, "TradesCounter")
	if err != nil {
		return nil, err
	}

	trade_.Id = counter
	trade_.SellerId = contract_.OwnerId
	trade_.Price = price

	if ok, err := stub.InsertRow("Trades", trade_.toRow()); !ok {
		log.Error("Failed inserting new trade: " + err.Error())
		return nil, err
	}

	_, err = t.changeContractState(stub, contract_.IssuerId, contract_.Id, "offer")
	return nil, err
}

func (t *BondChaincode) sell(stub shim.ChaincodeStubInterface, contractId string, price uint64) ([]byte, error) {
	log.Debugf("function: %s, args: %s", "sell", contractId)

	// Get Contract
	contract_, err := t.getContractById(stub, contractId)
	if err != nil {
		message := "Failed retrieving contract. Error: " + err.Error()
		log.Error(message)
		return nil, errors.New(message)
	}

	if _, err := t.createTradeForContract(stub, contract_, price); err != nil {
		message := "createTradeForContract failed. Error: " + err.Error()
		log.Error(message)
		return nil, errors.New(message)
	}

	return nil, nil
}

func (t *BondChaincode) buy(stub shim.ChaincodeStubInterface, tradeId uint64, newOwnerId string) ([]byte, error) {
	log.Debugf("function: %s, args: %s", "buy", tradeId)

	trade_, err := t.getTradeByType(stub, "offer", tradeId)
	if err != nil {
		message := "Failed buying trade. Error: " + err.Error()
		log.Error(message)
		return nil, errors.New(message)
	}

	// Get Contract
	contract_, err := t.getContractById(stub, trade_.ContractId)
	if err != nil {
		message := "Failed retrieving contract. Error: " + err.Error()
		log.Error(message)
		return nil, errors.New(message)
	}

	// Transfer Contract ownership
	if _, err := t.reserveContract(stub, contract_, newOwnerId); err != nil {
		message := "Failed transfering contract ownership. Error: " + err.Error()
		log.Error(message)
		return nil, errors.New(message)
	}

	err = t.sendPaymentInstruction(stub, trade_, newOwnerId)
	if err != nil {
		errStr := fmt.Sprintf("Failed to invoke swift chaincode. Got error: %s", err.Error())
		fmt.Printf(errStr)
		return nil, err
	}

	// Create new trade entry with "settled" state
	trade_.State = "reserved"
	if ok, err := stub.ReplaceRow("Trades", trade_.toRow()); !ok {
		log.Error("Failed inserting new trade: " + err.Error())
		return nil, err
	}

	return nil, nil
}

func (t *BondChaincode) sendPaymentInstruction(stub shim.ChaincodeStubInterface, trade_ trade, newOwnerId string) (error) {
	log.Debugf("payment instructions for payment:%+v", trade_)


	var args [][]byte

	args = append(args, []byte("submitPayment"))
	args = append(args, []byte(newOwnerId))
	args = append(args, []byte(trade_.SellerId))
	//  1000 * trade_.Price =  ( 100000 / 100 ) * trade_.Price
	args = append(args, []byte(strconv.FormatUint(1000 * trade_.Price , 10)))
	args = append(args, []byte("payment"))
	args = append(args, []byte(trade_.ContractId))
	chainId, _ := t.getChaincodeId(stub)
	args = append(args, []byte(chainId))
	args = append(args, []byte("confirm"))
	args = append(args, []byte(trade_.ContractId))

	response, err := stub.InvokeChaincode(t.GetSwiftChaincodeToCall(), args)
	if err != nil {
		errStr := fmt.Sprintf("Failed to invoke chaincode. Got error: %s", err.Error())
		fmt.Printf(errStr)
		return errors.New(errStr)
	}

	log.Debugf("Invoke chaincode successful. Got response %s", string(response))

	return nil
}


func (t *BondChaincode) confirm(stub shim.ChaincodeStubInterface, contractId string) ([]byte, error) {
	log.Debugf("function: %s, args: %s", "buy", contractId)

	trade_, err := t.getTradeForContract(stub, contractId, "reserved")
	if err != nil {
		message := "Failed confirming trade. Error: " + err.Error()
		log.Error(message)
		return nil, errors.New(message)
	}

	// Get Contract
	contract_, err := t.getContractById(stub, trade_.ContractId)
	if err != nil {
		message := "Failed retrieving contract. Error: " + err.Error()
		log.Error(message)
		return nil, errors.New(message)
	}

	// Transfer Contract ownership
	if _, err := t.changeContractState(stub, contract_.IssuerId, contract_.Id, "active"); err != nil {
		message := "Failed transfering contract ownership. Error: " + err.Error()
		log.Error(message)
		return nil, errors.New(message)
	}

	// Create new trade entry with "settled" state
	trade_.State = "settled"
	if ok, err := stub.ReplaceRow("Trades", trade_.toRow()); !ok {
		log.Error("Failed inserting new trade: " + err.Error())
		return nil, err
	}

	return nil, nil
}

func (t *BondChaincode) getAllTrades(stub shim.ChaincodeStubInterface) (trades []trade, err error) {
	rows, err := stub.GetRows("Trades", []shim.Column{})
	if err != nil {
		message := "Failed retrieving trades. Error: " + err.Error()
		log.Error(message)
		return nil, errors.New(message)
	}

	for row := range rows {
		var result trade
		result.readFromRow(row)
		log.Debugf("getOfferTrades result includes: %+v", result)
		trades = append(trades, result)
	}

	return trades, nil
}

func (t *BondChaincode) getTradesByType(stub shim.ChaincodeStubInterface, state string) (trades []trade, err error) {
	rows, err := stub.GetRows("Trades", []shim.Column{})
	if err != nil {
		message := "Failed retrieving trades. Error: " + err.Error()
		log.Error(message)
		return nil, errors.New(message)
	}

	for row := range rows {
		var result trade
		result.readFromRow(row)
		if result.State != state {
			continue
		}
		log.Debugf("getOfferTrades result includes: %+v", result)
		trades = append(trades, result)
	}

	return trades, nil
}

func (t *BondChaincode) getTradeByType(stub shim.ChaincodeStubInterface, state string, tradeId uint64) (trade, error) {
	rows, err := stub.GetRows("Trades", []shim.Column{})
	if err != nil {
		message := "Failed retrieving trades. Error: " + err.Error()
		log.Error(message)
		return trade{}, errors.New(message)
	}

	for row := range rows {
		var result trade
		result.readFromRow(row)
		if result.State == state && result.Id == tradeId {
			log.Debugf("getOfferTradeForContract returns: %+v", result)
			return result, nil
		}
	}
	return trade{}, errors.New("No trades found for id " + strconv.FormatUint(tradeId, 10))
}

func (t *BondChaincode) getTradeForContract(stub shim.ChaincodeStubInterface, contractId string, state string) (trade, error) {
	rows, err := stub.GetRows("Trades", []shim.Column{})
	if err != nil {
		message := "Failed retrieving trades. Error: " + err.Error()
		log.Error(message)
		return trade{}, errors.New(message)
	}

	for row := range rows {
		var result trade
		result.readFromRow(row)
		if result.ContractId != contractId {
			continue
		}
		if state != "" && result.State != state  {
			continue
		}
		log.Debugf("getTradeForContract returns: %+v", result)
		return result, nil
	}
	return trade{}, errors.New("No trades found for contract " + contractId)
}

//func (t *BondChaincode) verifyTradeForContract(stub shim.ChaincodeStubInterface, contractId string, price uint64) (response) {
//	trade, err := t.getOfferTradeForContract(stub, contractId, "reserved")
//	log.Debugf("getOfferTradeForContract returns: %+v", trade)
//	var msg response
//	if err != nil {
//		msg.State = "ERROR"
//		msg.Msg = "Payment is not confirmed."
//		log.Debugf("contract id %s with state 'reserved' not found: %+v", contractId, err)
//		return msg
//	}
//	if trade.Price == price{
//		msg.State = "OK"
//		msg.Msg = "Approved"
//		log.Debugf("contract approved")
//	}else{
//		msg.State = "ERROR"
//		msg.Msg = "Incorrect price"
//		log.Debugf("contract incorrect")
//	}
//	return msg
//}