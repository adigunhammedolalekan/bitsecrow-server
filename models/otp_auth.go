package models

import (
	"github.com/jinzhu/gorm"
	"time"
	"math"
	"math/rand"
	u "bitescrow/utils"
	"fmt"
)

type Auth struct {
	gorm.Model
	UserId uint `json:"user_id"`
	Code int `json:"code"`
	TryCount int `json:"try_count"`
	Status int `json:"status"` // 1 = done, 0 = Not Done
}

/*
	return age of AUTH in minutes
*/
func (auth *Auth) Age() (float64) {

	t := time.Since(auth.CreatedAt)
	return math.Round(t.Seconds() / 50)
}
/*
Create an {Auth} if it does not already exists. Update it otherwise,
*/
func (a *Auth) Create() (*Auth, bool) {

	auth := &Auth{}
	err := Db.Table("auths").Where("user_id = ?", a.UserId).First(auth).Error
	if err == gorm.ErrRecordNotFound {

		auth.UserId = a.UserId
		auth.Code = randomInt()
		auth.TryCount = 0
		auth.Status = 0

		Db.Create(auth)

		return auth, true
	}else if err == nil {

		if auth.ID > 0 {//Seen an auth? Refresh it!
			auth.Code = randomInt()
			auth.TryCount = 0
			auth.Status = 0
			auth.CreatedAt = time.Now()
			Db.Table("auths").Where("id = ?", auth.ID).Update(auth)
			Db.Table("auths").Where("id = ?", auth.ID).UpdateColumn("try_count", 0)
			Db.Table("auths").Where("id = ?", auth.ID).UpdateColumn("status", 0)
		}
		return auth, true
	}

	return nil, false
}

func Verify(user uint, code int64) (map[string]interface{}, bool) {

	auth := &Auth{}
	err := Db.Table("auths").Where("user_id = ?", user).First(auth).Error

	if err != nil && err == gorm.ErrRecordNotFound {
		return u.Message(false, "No auth created for this account"), false
	}

	if auth.Code == 0 {
		return u.Message(false, "No auth created for this account"), false
	}

	age := auth.Age()

	if age > 2 { // more than 2 mins ago
		return u.Message(false, "Auth has expired. Generate a new authentication code"), false
	}

	if auth.Code != int(code) {
		auth.TryCount += 1
		if auth.TryCount >= 3 {
			return u.Message(false, "Auth code does not match. You have exceeded the number of trials. Please, generate a new auth code"), false
		}

		tx := Db.Begin()
		tx.Table("auths").Where("user_id = ?", auth.UserId).UpdateColumn("try_count", auth.TryCount)
		tx.Commit()

		numTrials := 3 - auth.TryCount //number of trials remaining
		return u.Message(false, fmt.Sprintf("Authentication code does not match. %d Trials remaining", numTrials)), false
	}

	//Code Match!
	tx := Db.Begin()
	tx.Table("auths").Where("id = ?", auth.ID).UpdateColumn("status", 1)
	tx.Commit()

	return u.Message(true, "success"), true
}

func randomInt() int {

	i := rand.Intn(99999999)
	return i
}

func GetAuth(user uint) *Auth {

	auth := &Auth{}
	err := Db.Table("auths").Where("user_id = ?", user).First(auth).Error
	if err != nil {
		return nil
	}

	return auth
}
