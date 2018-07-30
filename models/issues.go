package models

import "github.com/jinzhu/gorm"

type Issue struct {
	gorm.Model
	EsId uint `json:"es_id"`
	Admin uint `json:"admin"`
	Status string `json:"status"`
}


