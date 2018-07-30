package models

import (
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/jinzhu/gorm"
	"os"
	"github.com/joho/godotenv"
	"fmt"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"math/rand"
	"time"
	"encoding/json"
	u "bitescrow/utils"
	"github.com/NaySoftware/go-fcm"
)

var (
	Db *gorm.DB
	WorkQueue = make(chan *MailRequest, 9009)
	MessageQueue = make(chan *Message, 9000)
	retryCount = 0

)

func init() {

	e := godotenv.Load()
	if e != nil {
		fmt.Print(e)
	}

	username := os.Getenv("db_user")
	password := os.Getenv("db_pass")
	dbName := os.Getenv("db_name")
	dbHost := os.Getenv("db_host")


	dbUri := fmt.Sprintf("host=%s user=%s dbname=%s sslmode=disable password=%s", dbHost, username, dbName, password)
	//fmt.Println(dbUri)

	conn, err := gorm.Open("postgres", dbUri)
	if err != nil {
		fmt.Print(err)
	}

	rand.Seed(time.Now().UnixNano())
	go StartWorker()
	go MessageWorker()

	Db = conn
	Db.Debug().AutoMigrate(&User{}, &Wallet{}, &Escrow{}, &Auth{}, &Message{}, &FeedBack{})
}


func StartWorker()  {

	fmt.Println("Worker Started!")
	for  {
		select {
		case req, ok := <- WorkQueue:
			if ok {
				fmt.Println("Recieved send mail request => " + req.To)
				req.Send()
			}

		}
	}
}

func MessageWorker() {

	fmt.Println("Message Worker Started")
	for {
		select {
		case m, ok := <- MessageQueue: //Pop the next message off queue
			if ok {
				fmt.Println("New Message => " + m.Text)
				err := SendMessage(m) //try sending message
				if err != nil {

					for retryCount < 5 {//Retry sending message 5 more time if there was error
						err = SendMessage(m)
						if err == nil { //Sent!
							break
						}
					}
				}
			}

		}
	}
}

func GetDB() *gorm.DB {
	return Db
}

type AuthCode struct {
	Code json.Number `json:"code"`
}

type MailRequest struct {

	Subject string `json:"subject"`
	Body string `json:"body"`
	To string `json:"to"`
	Name string `json:"name"`

}

func (request *MailRequest) Send() (error) {
	return SendEmail(request)
}

func SendEmail(request *MailRequest) error {

	from := mail.NewEmail("BitsEscrow", os.Getenv("email"))
	to := mail.NewEmail(request.Name, request.To)

	message := mail.NewSingleEmail(from, request.Subject, to, request.Body, request.Body)
	client := sendgrid.NewSendClient(os.Getenv("SENDGRID_API_KEY"))
	_, err := client.Send(message)
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

func SendMessage(m *Message) (error) {


	client := fcm.NewFcmClient(os.Getenv("fcm_client_key"))
	user := GetUser(m.FromId)
	to := GetUser(m.ToId)
	if user != nil {

		data := map[string]interface{}{
			"from_user": user.Username,
			"to_user":   to.Username,
			"text": m.Text,
			"message_type": m.MessageType,
			"time_when": u.GetReadableTime(m.CreatedAt),
		}

		topic := fmt.Sprintf("%s%d", "/topics/escrow", m.EsId)
		fmt.Println(topic)
		client.NewFcmMsgTo(topic, data)
		status, err := client.Send()
		if err == nil {
			status.PrintResults()
			retryCount = 0//Success, re-init retryCount to save the Logic

			//Save message into the DB
			Db.Create(m)
			return err //err == nil
		}
		retryCount++
		fmt.Println(err)
		return err
	}

	return nil //No user
}

type Invite struct {
	User uint `json:"user"`
	Emails []string `json:"emails"`
}

func (iv *Invite) Send() {

	user := GetUser(iv.User)

	subject := "BitsEscrow Invitation"
	for _, val := range iv.Emails {

		m := MailRequest{
			Subject: subject,
			Body : user.ValidName() + " has invited you to join BitsEscrow",
			To : val,
		}

		WorkQueue <- &m
	}
}

