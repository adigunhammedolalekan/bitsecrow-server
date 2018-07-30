package controllers

import (
	"github.com/gin-gonic/gin"
	"bitescrow/models"
	u "bitescrow/utils"
	"strconv"
)

var NewEscrow = func(c *gin.Context) {

	esReq := &models.EscroRequest{}
	if err := c.ShouldBindJSON(esReq); err != nil {
		c.JSON(200, u.Message(false, "Malformed parameters"))
		return
	}

	user, ok := c.Get("user")
	if !ok {
		c.JSON(200, u.Message(false, "unAuthorized"))
		return
	}

	from := models.GetUser(user . (uint))
	if from == nil {
		c.JSON(200, u.Message(false, "User not found"))
		return
	}

	esReq.FromEmail = from.Email
	esReq.User = from.ID
	escrow, message := models.New(esReq)
	if escrow == nil {
		c.JSON(200, message)
		return
	}

	resp := escrow.Create()
	c.JSON(200, resp)
}

var ReleaseCoin = func(c *gin.Context) {

	user, ok := c.Get("user")
	if !ok {
		c.JSON(403, u.Message(false, "Authorized"))
		return
	}

	esId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(200, u.Message(false, "Not a valid data"))
		return
	}
	id := user . (uint)

	resp := models.ReleaseCoin(id, uint(esId))
	c.JSON(200, resp)
}

var CancelEscrow = func(c *gin.Context) {

	user, ok := c.Get("user")
	if !ok {
		c.JSON(403, u.Message(false, "Authorized"))
		return
	}

	esId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(200, u.Message(false, "Not a valid data"))
		return
	}
	id := user . (uint)

	resp := models.CancelEscrow(id, uint(esId))
	c.JSON(200, resp)
}

var UserEscrows = func(c *gin.Context) {

	user, ok := c.Get("user")
	if !ok {
		c.JSON(403, u.Message(false, "Authorized"))
		return
	}

	data := models.GetUserEscrows(user . (uint))
	resp := u.Message(true, "succes")
	resp["data"] = data
	c.JSON(200, resp)
}

var Escrows = func(c *gin.Context) {

	user, ok := c.Get("user")
	if !ok {
		c.JSON(403, u.Message(false, "Authorized"))
		return
	}

	data := models.GetLinkedEscrows(user . (uint))
	resp := u.Message(true, "succes")
	resp["data"] = data
	c.JSON(200, resp)
}
