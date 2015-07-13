package handlers

import (
	"fmt"
	"github.com/nlopes/slack"
	"log"
	"strings"
)

func YouAlive(c *Context, m *slack.MessageEvent) (bool, error) {
	matched := c.isBotMention(m) && strings.Contains(m.Msg.Text, "yt?")
	if matched {
		log.Printf("Got yt? message %+v", m.Msg)
		u, err := c.Slack.GetUserInfo(m.Msg.UserId)
		if err != nil {
			return false, err
		}
		//_, _, err = api.PostMessage(m.ChannelId, fmt.Sprintf("Not dead yet, @%s", u.Name), postParams)
		c.chOutgoingMessages <- *c.ws.NewOutgoingMessage(fmt.Sprintf("Not dead yet, <@%s|%s>", m.Msg.UserId, u.Name), m.ChannelId)
		return true, err
	}
	return false, nil
}
