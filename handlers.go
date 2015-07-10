package main

import (
	"fmt"
	c "github.com/byxorna/collinsbot/collins"
	"github.com/nlopes/slack"
	"log"
	"strings"
)

var (
	wtfs = []string{
		"Que?", "Bist du bescheuert oder?", "Donde esta la biblioteca?", "Huh?", "WTF?", "Scientists better check their hypotenuses!",
	}
)

type Handler struct {
	Name     string
	Function func(*slack.MessageEvent, chan<- slack.OutgoingMessage) (bool, error)
}

func WTFHandler(m *slack.MessageEvent, q chan<- slack.OutgoingMessage) (bool, error) {
	if isBotMention(m) {
		q <- *ws.NewOutgoingMessage(fmt.Sprintf("%s Try 'help'", random(wtfs)), m.ChannelId)
		return true, nil
	}
	return false, nil
}

func HelpHandler(m *slack.MessageEvent, q chan<- slack.OutgoingMessage) (bool, error) {
	if isBotMention(m) && strings.Contains(m.Msg.Text, "help") {
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
				Title:    fmt.Sprintf("%s help", botIdentity.User),
				Fallback: strings.Join(helpfallbackslice, "\n"),
				Text:     strings.Join(helpfallbackslice, "\n"),
				Color:    "warning",
				//Fields:   helpattachmentfields,
			},
		}
		_, _, err := api.PostMessage(m.ChannelId, "", p)
		return true, err
	} else {
		return false, nil
	}
}

func YouAliveHandler(m *slack.MessageEvent, q chan<- slack.OutgoingMessage) (bool, error) {
	matched := isBotMention(m) && strings.Contains(m.Msg.Text, "yt?")
	if matched {
		log.Printf("Got yt? message %+v", m.Msg)
		u, err := api.GetUserInfo(m.Msg.UserId)
		if err != nil {
			return false, err
		}
		//_, _, err = api.PostMessage(m.ChannelId, fmt.Sprintf("Not dead yet, @%s", u.Name), postParams)
		q <- *ws.NewOutgoingMessage(fmt.Sprintf("Not dead yet, <@%s|%s>", m.Msg.UserId, u.Name), m.ChannelId)
		return true, err
	}
	return false, nil
}

func AssetTagHandler(m *slack.MessageEvent, q chan<- slack.OutgoingMessage) (bool, error) {
	// dont process if this came from a bot (like ourselves). avoids looping
	if isBotMessage(m) {
		return false, nil
	}
	// handle messages with any asset tags present - we will turn them into collins links
	tags := extractAssetTags(m.Msg.Text)
	if len(tags) > 0 {
		assets := lookupAssetsFromTags(tags)
		items := []string{}
		for _, asset := range assets {
			items = append(items, assetStringForSlack(asset))
		}
		// send a message back to that channel with the links to the assets
		p := slack.NewPostMessageParameters()
		p.Attachments = []slack.Attachment{*slackAssetsAttachment(assets)}
		_, _, err := api.PostMessage(m.ChannelId, "", p)
		return true, err
	}
	return false, nil
}

func AssetHostnameHandler(m *slack.MessageEvent, q chan<- slack.OutgoingMessage) (bool, error) {
	// handle messages with any hostnames present - if assets, link them
	if isBotMessage(m) {
		return false, nil
	}
	hosts := extractHostnames(m.Msg.Text)
	if len(hosts) > 0 {
		assets := []c.Asset{}
		for _, host := range hosts {
			//TODO!
			log.Printf("Found host %s in message: %s\n", host, m.Msg.Text)
			a, err := collins.Find(map[string]string{}, map[string]string{"hostname": host})
			if err != nil || a == nil {
				log.Printf("Error trying to find host %s: %s\n", host, err.Error())
				continue
			}
			if len(a) > 1 {
				log.Printf("Multiple assets found matching hostname %s: %+v\n", a)
			} else if len(a) == 0 {
				log.Printf("Nothing found for hostname %s\n", host)
			} else {
				assets = append(assets, a[0])
			}
		}
		log.Printf("Found assets: %+v\n", assets)
		p := slack.NewPostMessageParameters()
		p.Attachments = []slack.Attachment{*slackAssetsAttachment(assets)}
		_, _, err := api.PostMessage(m.ChannelId, "", p)
		return true, err
	}
	return false, nil
}
