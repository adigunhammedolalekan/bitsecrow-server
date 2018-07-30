package models

import "github.com/jinzhu/gorm"
import u "bitescrow/utils"

type FeedBack struct {
	gorm.Model
	UserId uint `json:"user_id"`
	Title string `json:"title"`
	Body string `json:"body"`
	Attachment string `json:"attachment"`
	User *User `gorm:"-";sql:"-";json:"user"`
}

func (f *FeedBack) Create() (map[string] interface{}) {

	Db.Create(f)

	mailRequest := &MailRequest{
		To: "devpasky@gmail.com",
		Subject: f.Title,
		Body: f.Body,
	}

	WorkQueue <- mailRequest
	return u.Message(true, "success")
}

func GetFeedBacks(page int) []*FeedBack {

	data := make([]*FeedBack, 0)
	err := Db.Debug().Table("feed_backs").Limit(10).Offset(page * 10).Find(&data).Error
	if err != nil {
		return nil
	}

	feedbacks := make([]*FeedBack, 0)
	for _, v := range data {
		v.User = GetUser(v.UserId)
		feedbacks = append(feedbacks, v)
	}

	return feedbacks
}
