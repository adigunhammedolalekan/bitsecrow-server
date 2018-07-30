package models

import (
	"melody"
	"github.com/gin-gonic/gin/json"
	"fmt"
)

type Channel struct {
	Name string
	Sessions map[*melody.Session] bool
}

func CreateChannel(name string) *Channel {

	return &Channel{
		Name: name, Sessions: make(map[*melody.Session] bool),
	}
}

func (c *Channel) Send(m *Message) error {

	var err error
	for s := range c.Sessions {
		data, err := json.Marshal(m)
		if err != nil {
			return err
		}
		err = s.Write(data)
	}

	return err
}

func (c *Channel)  UnSubscribe(session *melody.Session)  {

	ok, _ := c.Sessions[session]
	if ok {
		delete(c.Sessions, session)
	}
	fmt.Println(len(c.Sessions))
}
