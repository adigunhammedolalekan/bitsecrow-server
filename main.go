package main

import (
	"github.com/gin-gonic/gin"
	"bitescrow/app"
	"os"
	"bitescrow/controllers"
	"melody"
	"bitescrow/models"
	"encoding/json"
	"fmt"
)

var (
	channels = make(map[string] *models.Channel)
)
func main() {


	r := gin.New()
	m := melody.New()
	r.Use(gin.Recovery())
	r.Use(gin.Logger())
	r.Use(app.GinJwt)

	gin.SetMode(gin.ReleaseMode)

	r.POST("/api/user/new", controllers.CreateAccount)
	r.POST("/api/me/update", controllers.UpdateAccount)
	r.POST("/api/authenticate", controllers.Login)
	r.POST("/api/escrow/new", controllers.NewEscrow)
	r.POST("/api/escrow/cancel/:id", controllers.CancelEscrow)
	r.POST("/api/escrow/release/:id", controllers.ReleaseCoin)
	r.GET("/api/me/escrows", controllers.UserEscrows)
	r.GET("/api/me/escrows/linked", controllers.Escrows)
	r.POST("/api/me/auth", controllers.GenOTP)
	r.POST("/api/me/code/verify", controllers.VerifyOTP)
	r.GET("/api/me/wallet", controllers.Wallet)
	r.GET("/api/me/wallet/stats", controllers.WalletStats)
	r.POST("/api/escrow/message/:id", controllers.NewMessage)
	r.POST("/api/feedback/new", controllers.NewFeedBack)
	r.GET("/api/feedbacks/:page", controllers.Feedbacks)
	r.POST("/api/invites/new", controllers.NewInvites)
	r.GET("/api/s", controllers.S)

	r.GET("/api/ws/connect", func(context *gin.Context) {
		m.HandleRequest(context.Writer, context.Request)
	})

	m.HandleConnect(func(session *melody.Session) {
		fmt.Println("New Connection")
	})
	m.HandleMessage(func(session *melody.Session, bytes []byte) {

		message := &models.Message{}
		err := json.Unmarshal(bytes, message)
		if err == nil {

			action := message.Action
			switch action {
			case "sub":
				_, ok := channels[message.ChannelName]
				if !ok {
					c := models.CreateChannel(message.ChannelName)
					channels[c.Name] = c
				}

				channel := channels[message.ChannelName]
				if channel != nil {
					channel.Sessions[session] = true
					fmt.Println(len(channel.Sessions))
				}
				break

			case "message":
				c := message.ChannelName
				channel, ok := channels[c]
				if ok {
					channel.Send(message)
				}
				break
			case "unsub":
				channel, ok := channels[message.ChannelName]
				if ok {
					channel.UnSubscribe(session)
				}
				break
			default:
				fmt.Println("No channel with name " + message.ChannelName)
			}
		}else {
			fmt.Println("Failed to decode message => ", err)
		}

	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8009"
	}

	r.Run(":" + port)
}
