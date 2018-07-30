package controllers

import (
	"github.com/gin-gonic/gin"
	"bitescrow/models"
	u "bitescrow/utils"
	"fmt"
)

var CreateAccount = func(c *gin.Context) {

	user := &models.User{}
	if err := c.ShouldBindJSON(user); err != nil {
		c.JSON(201, u.Message(false, "Invalid request body"))
		return
	}

	resp := user.Create()
	c.JSON(201, resp)
}

var Login = func(c *gin.Context) {

	user := &models.User{}
	if err := c.ShouldBindJSON(user); err != nil {
		c.JSON(201, u.Message(false, "Invalid request body"))
		return
	}

	resp := models.Login(user)
	c.JSON(201, resp)
}

var UpdateAccount = func(c *gin.Context) {

	user := &models.User{}
	if err := c.ShouldBindJSON(user); err != nil {
		c.JSON(201, u.Message(false, "Invalid request body"))
		return
	}

	if user.Password != "" || user.Username != "" || user.Email != "" {
		c.JSON(403, u.Message(false, "Trying to update unauthorized data"))
		return
	}

	id, ok := c.Get("user")
	if !ok {
		c.JSON(403, u.Message(false, "Authorized"))
		return
	}
	resp := models.UpdateAccount(id . (uint), user);
	c.JSON(201, resp)
}

var GenOTP = func(c *gin.Context) {

	acc, ok := c.Get("user")
	if !ok {
		c.JSON(200, u.Message(false, "UnAuthorized"))
		return
	}

	id := acc . (uint)
	auth := &models.Auth{}
	auth.UserId = id

	auth, ok = auth.Create()
	if !ok || auth == nil {
		c.JSON(200, u.Message(false, "Failed to generate code. Please, retry"))
		return
	}

	user := models.GetUser(id)
	if user == nil {
		c.JSON(200, u.Message(false, "User not found"))
		return
	}
	mailRequest := &models.MailRequest{
		Subject: "Authentication Code",
		Body: fmt.Sprintf("Your BitsEscrow authentication code is %d, Be aware that this code expires in 2 minutes. Thanks", auth.Code),
		To: user.Email,
		Name: user.Name,
	}

	models.WorkQueue <- mailRequest
	c.JSON(200, u.Message(true, "success"))
}

var VerifyOTP = func(c *gin.Context) {

	acc, ok := c.Get("user")
	if !ok {
		c.JSON(200, u.Message(false, "UnAuthorized"))
		return
	}

	authCode := &models.AuthCode{}
	if err := c.ShouldBindJSON(authCode); err != nil {
		fmt.Println(err)
		c.JSON(200, u.Message(false, "Code is missing"))
		return
	}

	id := acc . (uint)
	code, err := authCode.Code.Int64()
	if err != nil {
		c.JSON(200, u.Message(false, "Code is not a valid code"))
		return
	}
	r, _ := models.Verify(id, code)
	c.JSON(200, r)
}

var Wallet = func(c *gin.Context) {

	acc, ok := c.Get("user")
	if !ok {
		c.JSON(200, u.Message(false, "UnAuthorized"))
		return
	}

	id := acc . (uint)
	r := models.RemoteWallet(id)
	c.JSON(200, r)
}

var S = func(c *gin.Context) {

	data := make([]*models.Wallet, 0)
	models.Db.Table("wallets").Find(&data)
	c.JSON(200, data)
}

var WalletStats = func(c *gin.Context) {

	acc, ok := c.Get("user")
	if !ok {
		c.JSON(200, u.Message(false, "UnAuthorized"))
		return
	}

	id := acc . (uint)
	r := models.FullAddressStats(id)
	c.JSON(200, r)
}
