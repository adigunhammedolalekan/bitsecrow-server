package models

type TransactionRequest struct {

	From uint `json:"from"`
	To uint `json:"to"`
	Amount float64 `json:"amount"`

}


