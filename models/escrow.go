package models

import (
	"github.com/jinzhu/gorm"
	u "bitescrow/utils"
	"fmt"
)

type Escrow struct {

	gorm.Model
	FromId uint `json:"from_id"`
	ToId uint `json:"to_id"`
	Amount float64 `json:"amount"`
	Status string `json:"status"`
	Narration string `json:"narration"`
	From *User `gorm:"-";sql:"-";json:"from"`
	To *User `gorm:"-";sql:"-";json:"to"`
	FormattedTime string `gorm:"-";sql:"-";json:"formatted_time"`
}

func New(request *EscroRequest) (*Escrow, map[string] interface{}) {

	from := GetUserByEmail(request.FromEmail)
	if from == nil {
		return nil, u.Message(false, fmt.Sprintf("User with email %s not found", request.FromEmail))
	}
	to := GetUserByEmail(request.ToEmail)
	if to == nil {
		return nil, u.Message(false, fmt.Sprintf("User with email %s not found", request.ToEmail))
	}

	wallet := GetOnlineWallet(request.User)
	if wallet == nil {
		return nil, u.Message(false, "Failed to access user wallet balance")
	}

	amount, err := request.Amount.Float64()
	if err != nil {
		return nil, u.Message(false, "Invalid request object")
	}

	if amount <= 0 {
		return nil, u.Message(false, "Invalid amount")
	}


	if amount > wallet.Balance {
		return nil, u.Message(false, "Insufficient balance. Please, fund your wallet first")
	}
	escrow := &Escrow{}
	escrow.Amount = amount
	escrow.FromId = from.ID
	escrow.ToId = to.ID

	return escrow, nil
}

func (es *Escrow) Create() (map[string] interface{}) {

	to := GetUser(es.ToId)
	if to == nil {
		//Should never happen
		return u.Message(false, "Email address not found")
	}
	from := GetUser(es.FromId)
	if from == nil {
		return u.Message(false, "Email address not found for master")
	}

	tx := Db.Begin()
	err := tx.Create(es).Error
	if err != nil {
		tx.Rollback()
		return u.Message(false, "Failed to create escrow. Please, retry")
	}

	wallet := GetWallet(from.ID)
	wallet.Balance -= es.Amount

	err = tx.Table("wallets").Where("id = ?", wallet.ID).UpdateColumn("balance", wallet.Balance).Error
	if err != nil {
		tx.Rollback()
		return u.Message(false, "Failed to create escrow. Please, retry")
	}

	txRequest := &TransactionRequest{
		From: es.FromId,
		Amount: es.Amount,
	}

	err = SendToEscrow(txRequest)
	if err != nil {
		tx.Rollback()
		fmt.Println("Error ", err)
		return u.Message(false, "Failed to send transaction at this time. Please, retry")
	}

	emailBody := fmt.Sprintf("User %s has funded an escrow with %f BTC, Open your BitsEscrow app to view this escrow",
		from.ValidName(), es.Amount)
	mailRequest := &MailRequest{
		To: to.Email,
		Subject: "BitsEscrow - New Escrow Funded",
		Body: emailBody,
		Name: to.ValidName(),
	}

	WorkQueue <- mailRequest

	emailBody2 := fmt.Sprintf("You funded an escrow with %f BTC, Open your BitsEscrow app to view this escrow",
		es.Amount)
	mailRequest2 := &MailRequest{
		To: from.Email,
		Subject: "BitsEscrow - New Escrow Funded",
		Body: emailBody2,
		Name: to.ValidName(),
	}

	WorkQueue <- mailRequest2

	tx.Commit()
	resp := u.Message(true, "success")
	es.From = GetUser(es.FromId)
	es.To = GetUser(es.ToId)
	resp["escrow"] = es
	return resp
}

func ReleaseCoin(user uint, esId uint) (map[string] interface{}) {

	auth := GetAuth(user)
	if auth == nil {
		return u.Message(false, "OTP authentication required for this operation")
	}

	if auth.Status != 1 {
		return u.Message(false, "Please verify operation through otp")
	}

	es := GetEscrow(esId)
	if es.FromId != user {
		return u.Message(false, "Attempt to release coin from escrow you do not have access to")
	}

	if es.Status == "completed" {
		return u.Message(false, "Escrow has already been released")
	}

	if es.Status == "cancelled" {
		return u.Message(false, "Escrow has been cancelled")
	}

	toPay := (1 * es.Amount) / 100
	toPay = es.Amount - toPay //99%
	toPay = u.BtcToShatoshi(toPay)

	tx := Db.Begin()
	err := FundUser(es.ToId, toPay)
	if err != nil {
		tx.Rollback()
		return u.Message(false, "Failed to release coin. Server error. Please, retry")
	}

	wallet := GetWallet(es.ToId)//Get wallet of {TO}
	wallet.Balance += es.Amount

	es.Status = "completed"

	err = tx.Table("wallets").Where("id = ?", wallet.ID).UpdateColumn("balance", wallet.Balance).Error
	if err != nil {
		tx.Rollback()
		return u.Message(false, "Failed to release coin. Server error. Please, retry")
	}

	err = tx.Table("escrows").Where("id = ?", es.ID).UpdateColumn(es).Error
	if err != nil {
		tx.Rollback()
		return u.Message(false, "Failed to release coin. Server error. Please, retry")
	}

	to := GetUser(es.ToId)
	mailRequest := &MailRequest{
		To: to.Email,
		Subject: "BitsEscrow - Coin Released",
		Body: fmt.Sprintf("Coin %f BTC has been release from Escrow. The coin has been added to your wallet. " +
			"Open your BitsEscrow app to see the changes", es.Amount),
		Name: to.ValidName(),
		}

		WorkQueue <- mailRequest

	from := GetUser(es.FromId)
	mailRequest = &MailRequest{
		To: from.Email,
		Subject: "BitsEscrow - Coin Released",
		Body: fmt.Sprintf("You released Coin worth %f BTC to %s. The coin has been added to your wallet. " +
			"Open your BitsEscrow app to see the changes", es.Amount, to.ValidName()),
		Name: to.ValidName(),
	}
	WorkQueue <- mailRequest

	tx.Table("auths").Where("user_id = ?", es.FromId).UpdateColumn("status", 0)
	tx.Commit()
	return u.Message(true, "success")
}

func CancelEscrow(user uint, esId uint) (map[string]interface{}) {

	es := GetEscrow(esId)
	if es.ToId != user {
		return u.Message(false, "The only person that can cancel this escrow is not here yet!")
	}

	tx := Db.Begin()
	fmt.Println("Amount ", es.Amount)
	toPay := (1 * es.Amount) / 100
	toPay = es.Amount - toPay //99%

	fmt.Println("ToPay ", toPay)
	fmt.Println("Platform Gains in sato ", u.BtcToShatoshi(es.Amount - toPay))
	toPay = u.BtcToShatoshi(toPay)

	fmt.Println("To Pay in sato ", toPay)

	err := FundUser(es.FromId, toPay)
	if err != nil {
		tx.Rollback()
		return u.Message(false, "Failed to cancel escrow. Please, retry")
	}

	wallet := GetWallet(es.FromId)
	wallet.Balance += es.Amount
	err = tx.Table("wallets").Where("id = ?", wallet.ID).UpdateColumn("balance", wallet.Balance).Error
	if err != nil {
		tx.Rollback()
		return u.Message(false, "Failed to cancel escrow. Please, retry.")
	}

	es.Status = "cancelled"
	err = tx.Table("escrows").Where("id = ?", es.ID).UpdateColumn(es).Error
	if err != nil {
		tx.Rollback()
		return u.Message(false, "Failed to cancel escrow. Please, retry")
	}

	from := GetUser(es.FromId)
	to := GetUser(es.ToId)
	mailRequest := &MailRequest{
		To: from.Email,
		Subject: fmt.Sprintf("BitsEscrow - Escrow Cancelled by %s", to.ValidName()),
		Body: fmt.Sprintf("Escrow has been cancelled by %s, %f BTC has been added back to your wallet", to.ValidName(), es.Amount),
	}

	WorkQueue <- mailRequest

	mailRequest = &MailRequest{
		To: to.Email,
		Subject: fmt.Sprintf("BitsEscrow - Escrow Cancelled by you"),
		Body: fmt.Sprintf("Escrow has been cancelled by you, %f BTC has been added back to %s wallet", es.Amount, from.ValidName()),
	}

	WorkQueue <- mailRequest

	tx.Commit()
	return u.Message(true, "success")
}