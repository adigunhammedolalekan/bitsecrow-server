package models

import "github.com/jinzhu/gorm"
import (
	u "bitescrow/utils"
	"fmt"
)

type Wallet struct {

	gorm.Model
	UserId uint `json:"user_id"`
	Address string `json:"address"`
	Private string `json:"private"`
	Balance float64 `json:"balance"`

}

func GetWallet(id uint) *Wallet {

	wallet := &Wallet{}
	err := Db.Table("wallets").Where("user_id = ?", id).First(wallet).Error
	if err != nil {
		return nil
	}

	return wallet
}

func RemoteWallet(user uint) (map[string] interface{}) {

	r := u.Message(true, "success")
	wallet := GetWallet(user)

	if wallet == nil {
		r["wallet"] = wallet;
		return r
	}

	wallet.Private = ""

	bc := u.BC()
	addr, err := bc.GetAddrBal(wallet.Address, nil)
	if err == nil {
		r["data"] = addr
	}

	fmt.Println(err)

	return r
}

func GetOnlineWallet(user uint) *Wallet {

	local := GetWallet(user)
	if local == nil {
		return nil
	}

	bc := u.BC()
	addr, err := bc.GetAddrBal(local.Address, nil)
	if err != nil {
		return nil
	}

	local.Balance = float64(u.ShatoshiToBtc(float64(addr.FinalBalance)))
	return local
}

func FullAddressStats(user uint) (map[string]interface{}) {

	r := u.Message(true, "success")
	wallet := GetWallet(user)
	if wallet == nil {
		r["wallet"] = wallet
		return r
	}

	wallet.Private = ""
	bc := u.BC()
	addr, err := bc.GetAddrFull(wallet.Address, nil)
	if err == nil {
		r["data"] = addr
	}

	return r
}
