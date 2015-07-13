package handlers

import (
	"fmt"
	"github.com/nlopes/slack"
)

var (
	wtfs = []string{
		"Que?", "Bist du bescheuert oder?", "Donde esta la biblioteca?", "Huh?", "WTF?", "Scientists better check their hypotenuses!",
	}
)

func (c *Context) WTFHandler(m *slack.MessageEvent) (bool, error) {
	if c.isBotMention(m) {
		c.chOutgoingMessages <- *c.ws.NewOutgoingMessage(fmt.Sprintf("%s Try 'help'", random(wtfs)), m.ChannelId)
		return true, nil
	}
	return false, nil
}
