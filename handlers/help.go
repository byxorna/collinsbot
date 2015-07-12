package handlers

import (
	"fmt"
	"github.com/nlopes/slack"
	"log"
	"strings"
)

var helpinfo = map[string]string{
	"help": "show this help output",
	"yt?":  "see if I am still alive",
	"mention any asset tag or hostname": "get a link to the asset",
}

func (c *Context) Help(m *slack.MessageEvent) (bool, error) {
	if c.isBotMention(m) && strings.Contains(m.Msg.Text, "help") {
		log.Printf("Got help message\n")
		p := slack.NewPostMessageParameters()

		//TODO sort this nonsense, cause the help output ordering keeps changing...
		helpfallbackslice := make([]string, len(helpinfo))
		actionFields := make([]string, len(helpinfo))
		descriptionFields := make([]string, len(helpinfo))
		i := 0
		for k, v := range helpinfo {
			helpfallbackslice[i] = fmt.Sprintf("%s - %s", k, v)
			actionFields[i] = k
			descriptionFields[i] = v
			i = i + 1
		}
		/* this looks lame
		helpattachmentfields := []slack.AttachmentField{
			slack.AttachmentField{
				Title: "Action",
				Value: strings.Join(actionFields, "\n"),
				Short: true,
			},
			slack.AttachmentField{
				Title: "Description",
				Value: strings.Join(descriptionFields, "\n"),
				Short: true,
			},
		}
		*/

		p.Attachments = []slack.Attachment{
			slack.Attachment{
				Title:    fmt.Sprintf("%s help", c.Self.User),
				Fallback: strings.Join(helpfallbackslice, "\n"),
				Text:     strings.Join(helpfallbackslice, "\n"),
				Color:    "warning",
				//Fields:   helpattachmentfields,
			},
		}
		_, _, err := c.Slack.PostMessage(m.ChannelId, "", p)
		return true, err
	} else {
		return false, nil
	}
}
