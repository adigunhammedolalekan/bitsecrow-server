package models

import "github.com/jinzhu/gorm"
import u "bitescrow/utils"

type Message struct {
	gorm.Model
	EsId uint `json:"es_id"`
	Text string `json:"text"`
	FromId uint `json:"from_id"`
	ToId uint `json:"to_id"`
	MessageType string `json:"message_type"`
	ChannelName string `json:"channel_name"`
	Action string `json:"action"`
}

func (m *Message) IsValid() bool {

	if m.Text == "" {
		return false
	}
	if m.EsId <= 0 {
		return false
	}

	return true
}

func Create(m *Message) (map[string]interface{}) {

	if !m.IsValid() {
		return u.Message(false, "Invalid message body")
	}

	MessageQueue <- m
	return u.Message(true, "message sent")
}

