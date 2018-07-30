package models

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/jinzhu/gorm"
	u "bitescrow/utils"
	"golang.org/x/crypto/bcrypt"
	"os"
	"strings"
	"fmt"
)

type Token struct {
	UserId uint
	jwt.StandardClaims
}

type User struct {

	gorm.Model
	Email string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
	Name string `json:"name"`
	Picture string `json:"picture"`
	Phone string `json:"phone"`
	Token string `gorm:"-";sql:"-";json:"token"`
	FcmToken string `json:"fcm_token"`
}

func (user *User) Validate () (map[string] interface{}, bool) {

	if user.Email == "" {
		return u.Message(false, "Email is required"), false
	}

	if user.Username == "" {
		return u.Message(false, "Username is required"), false
	}

	usernames := strings.Split(user.Username, " ")
	if len(usernames) > 1 {
		return u.Message(false, "Username must not contain any spaces"), false
	}

	var count int
	err := Db.Table("users").Where("email = ?", user.Email).Count(&count).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return u.Message(false, "Cannot process account at this time. Please, retry."), false
	}

	if count > 0 {
		return u.Message(false, "Email address already exists"), false
	}
	err = Db.Table("users").Where("username = ?", user.Username).Count(&count).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return u.Message(false, "Cannot process account at this time. Please, retry."), false
	}

	if count > 0 {
		return u.Message(false, "Username already exists"), false
	}

	if len(user.Password) < 6 {
		return u.Message(false, "Password must be at least 6 characters long"), false
	}

	return u.Message(true, "success"), true
}

func (user *User) Create() (map[string] interface{}) {

	if response, ok := user.Validate(); !ok {
		return response
	}

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	user.Password = string(hashedPassword)
	err := Db.Create(user).Error
	if err != nil {
		return u.Message(false, "Failed to create user account. Please, retry")
	}

	//create wallet
	wallet := &Wallet{}
	wallet.UserId = user.ID
	bc := u.BC()
	addr, err := bc.GenAddrKeychain()
	if err == nil {
		wallet.Address = addr.Address
		wallet.Private = addr.Private
		Db.Create(wallet)

		hash, err := bc.Faucet(addr, 45000)
		if err != nil {
			fmt.Println("Faucets ", err)
		}
		fmt.Println("Hash ", hash)
	}

	mailRequest := &MailRequest{
		To: user.Email,
		Name: user.Username,
		Subject: fmt.Sprintf("Hi %s, Welcome to BitsEscrow", user.ValidName()),
		Body: "We are very excited to get started with you. Our escrow system is the on you can trust",
	}
	WorkQueue <- mailRequest

	//Account created
	tk := &Token{UserId: user.ID}
	j := jwt.NewWithClaims(jwt.GetSigningMethod("HS256"), tk)
	user.Token, _ = j.SignedString([]byte(os.Getenv("tk_password")))

	user.Password = ""
	resp := u.Message(true, "success")
	resp["user"] = user
	return resp
}

func (user *User) ValidName() string {

	if user.Name == ""{
		return user.Username
	}

	return user.Name
}

func Login(user *User) (map[string]interface{}) {

	account := &User{}
	err := GetDB().Table("users").Where("email = ?", user.Email).First(account).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return u.Message(false, "Email address not found")
		}
		return u.Message(false, "Connection error. Please retry")
	}

	err = bcrypt.CompareHashAndPassword([]byte(account.Password), []byte(user.Password))
	if err != nil && err == bcrypt.ErrMismatchedHashAndPassword {
		return u.Message(false, "Invalid login credentials. Please try again")
	}
	//Logged In
	account.Password = ""

	//Create JWT token
	tk := &Token{UserId: account.ID}
	token := jwt.NewWithClaims(jwt.GetSigningMethod("HS256"), tk)
	tokenString, _ := token.SignedString([]byte(os.Getenv("tk_password")))
	account.Token = tokenString

	resp := u.Message(true, "Logged In")
	resp["user"] = account
	return resp
}

func UpdateAccount(id uint, user *User) (map[string]interface{}) {

	acc := GetUser(id)
	if acc == nil {
		return u.Message(false, "Account does not exists")
	}

	err := Db.Table("users").Where("id = ?", id).UpdateColumn(user).Error
	if err != nil {
		return u.Message(false, "Failed to update data. Please, retry")
	}

	user = GetUser(id)
	resp := u.Message(true, "success")
	resp["user"] = user
	return resp
}

func GetUser(u uint) *User {

	acc := &User{}
	Db.Table("users").Where("id = ?", u).First(acc)
	if acc.Username == "" {
		return nil
	}

	acc.Password = ""
	return acc
}
