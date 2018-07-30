package controllers

import (
	"github.com/gin-gonic/gin"
	"bitescrow/models"
	u "bitescrow/utils"
	"fmt"
)

var NewInvites = func(c *gin.Context) {

	iv := &models.Invite{}
	if err := c.ShouldBindJSON(iv); err != nil {
		c.JSON(200, u.Message(false, "Invalid payload"))
		return
	}

	id, ok := c.Get("user")
	if !ok {
		c.JSON(200, u.Message(false, "UnAuthorized"))
		return
	}

	fmt.Println(len(iv.Emails))
	iv.User = id . (uint)
	iv.Send()
	c.JSON(200, u.Message(true, "success"))
}
