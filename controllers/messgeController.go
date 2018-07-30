package controllers

import (
	"github.com/gin-gonic/gin"
	"bitescrow/models"
	u "bitescrow/utils"
	"strconv"
)

var NewMessage = func(c *gin.Context) {

	message := &models.Message{}
	if err := c.ShouldBindJSON(message); err != nil {
		c.JSON(200, u.Message(false, "Invalid message body"))
		return
	}

	id := c.Param("id")
	esId, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(200, u.Message(false, "Not a valid message body"))
		return
	}

	message.EsId = uint(esId)
	es := models.GetEscrow(message.EsId)

	if es == nil {
		c.JSON(200, u.Message(false, "No escrow found"))
		return
	}

	message.FromId = es.FromId
	message.ToId = es.ToId

	r := models.Create(message)
	c.JSON(200, r)
}
