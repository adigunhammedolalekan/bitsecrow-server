package models

import "errors"
import (
	u "bitescrow/utils"
	"fmt"
	"os"
)

type Transaction struct {
	Req *TransactionRequest
}


func (txn *Transaction) Send() (error) {

	request := txn.Req
	fromWallet := GetWallet(request.From)
	toWallet := GetWallet(request.To)

	if fromWallet == nil || toWallet == nil {
		return errors.New("User has no wallet")
	}

	amountInShatoshi := u.BtcToShatoshi(request.Amount)

	return SendTx(fromWallet.Address, toWallet.Address, fromWallet.Private, int(amountInShatoshi))
}

func SendToEscrow(request *TransactionRequest) error {

	fromWallet := GetWallet(request.From)
	if fromWallet == nil {
		return errors.New("Wallet not found")
	}


	amountInShatoshi := u.BtcToShatoshi(request.Amount)

	fmt.Println("Amount ", int(amountInShatoshi))

	return SendTx(fromWallet.Address, os.Getenv("escrow_wallet_address"), fromWallet.Private, int(amountInShatoshi))
}

func SendTx(from, to, key string, amount int) error {

	bc := u.BC()
	bcTx := u.TempTx(from, to, amount)
	txSkeleton, err := bc.NewTX(bcTx, true)
	if err != nil {
		fmt.Println("NewTx ", err)
		return err
	}

	privKeys := make([]string, 0)
	for {

		if len(privKeys) == len(txSkeleton.ToSign) {
			break
		}
		privKeys = append(privKeys, key)
	}

	fmt.Println(len(privKeys), len(txSkeleton.ToSign))

	err = txSkeleton.Sign(privKeys)
	if err != nil {
		fmt.Println("SignTx ", err)
		return err
	}

	txSkeleton, err = bc.SendTX(txSkeleton)
	if err != nil {
		fmt.Println("SendTx ", err)
		return err
	}

	fmt.Println("Tx ", txSkeleton)
	return nil
}

func FundUser(user uint, amount float64) error {

	toFund := GetWallet(user)
	if toFund == nil {
		return errors.New("Wallet not found")
	}

	return SendTx(os.Getenv("escrow_wallet_address"), toFund.Address,
		os.Getenv("escrow_wallet_private_key"), int(amount))
}