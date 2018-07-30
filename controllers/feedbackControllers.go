package controllers

import (
	"github.com/gin-gonic/gin"
	"bitescrow/models"
	u "bitescrow/utils"
	"strconv"
)

var NewFeedBack = func(c *gin.Context) {

	feedback := &models.FeedBack{}
	if err := c.ShouldBindJSON(feedback); err != nil {
		c.JSON(200, u.Message(false, "Invalid payload"))
		return
	}

	p, ok := c.Get("user")
	if !ok {
		c.JSON(200, u.Message(false, "Invalid payload"))
		return
	}

	feedback.UserId = p . (uint)
	r := feedback.Create()
	c.JSON(200, r)
}

var Feedbacks = func(c *gin.Context) {

	p := c.Param("page")
	page, err := strconv.Atoi(p)
	if err != nil {
		c.JSON(200, u.Message(false, "Invalid page number"))
	}

	data := models.GetFeedBacks(page)
	r := u.Message(true, "success")
	r["data"] = data
	c.JSON(200, r)
}
