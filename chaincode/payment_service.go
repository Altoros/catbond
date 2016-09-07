package main

import (
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"errors"
	"strconv"
	"fmt"
)

func (t *BondChaincode) submitPaymentInstruction(stub shim.ChaincodeStubInterface, payer string, payee string, price uint64, paymentType string, instruction string, callback string, payload string) (error) {
	log.Debugf("payment instructions")
	var args [][]byte

	args = append(args, []byte("submitPayment"))
	args = append(args, []byte(payer))
	args = append(args, []byte(payee))
	args = append(args, []byte(strconv.FormatUint(price , 10)))
	args = append(args, []byte(paymentType))
	args = append(args, []byte(instruction))
	chainId, _ := t.getChaincodeId(stub)
	args = append(args, []byte(chainId))
	args = append(args, []byte(callback))
	args = append(args, []byte(payload))

	response, err := stub.InvokeChaincode(t.GetSwiftChaincodeToCall(), args)
	if err != nil {
		errStr := fmt.Sprintf("Failed to invoke chaincode. Got error: %s", err.Error())
		fmt.Printf(errStr)
		return errors.New(errStr)
	}

	log.Debugf("Invoke chaincode successful. Got response %s", string(response))

	return nil
}
