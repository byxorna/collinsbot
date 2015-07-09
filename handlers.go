package main

import (
	"fmt"
	"github.com/nlopes/slack"
	"log"
	"regexp"
	"strings"
)

func YouAliveHandler(m *slack.MessageEvent, q chan<- slack.OutgoingMessage) (bool, error) {
	//TODO we should only respond to prompts directed at us! how do we know? Look for our userid in the text?
	matched, err := regexp.MatchString(`yt\?`, m.Msg.Text)
	if err != nil {
		return false, err
	}
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
	// handle messages with any asset tags present - we will turn them into collins links
	tags := extractAssetTags(m.Msg.Text)
	if len(tags) > 0 {
		assets := lookupAssetsFromTags(tags)
		items := []string{}
		for _, asset := range assets {
			items = append(items, assetStringForSlack(asset))
		}
		// send a message back to that channel with the links to the assets
		if len(items) > 0 {
			_, _, err := api.PostMessage(m.ChannelId, strings.Join(items, "\n"), postParams)
			//q <- *ws.NewOutgoingMessage(strings.Join(items, "\n"), m.ChannelId)
			return true, err
		}
	}
	return false, nil
}

func AssetHostnameHandler(m *slack.MessageEvent, q chan<- slack.OutgoingMessage) (bool, error) {
	// handle messages with any hostnames present - if assets, link them
	hosts := extractHostnames(m.Msg.Text)
	if len(hosts) > 0 {
		for _, host := range hosts {
			//TODO!
			log.Printf("Found host %s in %s\n", host, m.Msg.Text)
		}
		return true, nil
	}
	return false, nil
}
