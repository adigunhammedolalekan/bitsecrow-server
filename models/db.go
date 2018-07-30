package models

import "fmt"

func GetUserByEmail(email string) *User {

	user := &User{}
	err := Db.Table("users").Where("email = ?", email).First(user).Error
	if err != nil {
		return nil
	}

	return user
}

func GetEscrow(id uint) *Escrow {

	es := &Escrow{}
	err := Db.Table("escrows").Where("id = ?", id).First(es).Error
	if err != nil {
		return nil
	}

	return es
}

func GetUserEscrows(user uint) []*Escrow {

	fmt.Println(user)
	data := make([]*Escrow, 0)
	err := Db.Debug().Table("escrows").Where("from_id = ?", user).Or("to_id = ?", user).Find(&data).Error
	if err != nil {
		return nil
	}

	r := make([]*Escrow, 0)
	for _, val := range data {
		if val != nil {
			val.From = GetUser(val.FromId)
			val.To = GetUser(val.ToId)
			val.FormattedTime = val.CreatedAt.Format("2018-01-01, 09:09")

			r = append(r, val)
		}
	}
	return r
}

func GetLinkedEscrows(user uint) ([]*Escrow) {

	data := make([]*Escrow, 0)
	err := Db.Table("escrows").Where("to_id = ?", user).Find(&data).Error
	if err != nil {
		return nil
	}

	r := make([]*Escrow, 0)
	for _, val := range data {
		if val != nil {
			val.From = GetUser(val.FromId)
			val.To = GetUser(val.ToId)
			val.FormattedTime = val.CreatedAt.Format("2018-01-01, 09:09")

			r = append(r, val)
		}
	}
	return r
}
