package models

import "encoding/json"

type EscroRequest struct {
	FromEmail string `json:"from_email"`
	ToEmail string `json:"to_email"`
	Amount json.Number `json:"amount"`
	User uint `json:"user"`
}


