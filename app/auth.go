package app

import "github.com/gin-gonic/gin"
import (
	u "bitescrow/utils"
	"strings"
	"bitescrow/models"
	"github.com/dgrijalva/jwt-go"
	"os"
)

var GinJwt = func(c *gin.Context) {

	noAuth := []string {"/api/authenticate", "/api/user/new", "/api/ws", "/api/ws/connect"}
	path := c.Request.RequestURI

	for _, val := range noAuth {
		if val == path {
			c.Next()
			return
		}
	}

	headerValue := c.GetHeader("Authorization")
	if headerValue == "" {
		c.JSON(403, u.Message(false, "UnAuthorized"))
		return
	}

	values := strings.Split(headerValue, " ")
	if len(values) < 2 || len(values) > 2 {
		c.JSON(403, u.Message(false, "Invalid/Malformed token"))
		return
	}

	tokenValue := values[1]
	token := &models.Token{}
	claim, err := jwt.ParseWithClaims(tokenValue, token, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("tk_password")), nil
	})

	if err != nil {
		c.JSON(403, u.Message(false, "Failed to recognize token"))
		return
	}
	if !claim.Valid {
		c.JSON(403, u.Message(false, "Failed to proceed. Invalid token"))
		return
	}

	c.Set("user", token.UserId)
	c.Next()
}
