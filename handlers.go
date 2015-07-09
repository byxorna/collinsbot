package main

import (
	"fmt"
	"github.com/nlopes/slack"
	"log"
	"regexp"
	"strings"
)

func YouAliveHandler(m *slack.MessageEvent) (bool, error) {
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
		_, _, err = api.PostMessage(m.ChannelId, fmt.Sprintf("Not dead yet, @%s", u.Name), postParams)
		return true, err
	}
	return false, nil
}

func AssetTagHandler(m *slack.MessageEvent) (bool, error) {
	// handle messages with any asset tags present - we will turn them into collins links
	tags := extractAssetTags(m.Msg.Text)
	if len(tags) > 0 {
		assets := lookupAssetsFromTags(tags)
		items := []string{}
		for _, asset := range assets {
			var (
				emptystr       = ""
				hostname       = asset.AttrFetch("HOSTNAME", "0", &emptystr)
				pool           = asset.AttrFetch("POOL", "0", &emptystr)
				primary_role   = asset.AttrFetch("PRIMARY_ROLE", "0", &emptystr)
				secondary_role = asset.AttrFetch("SECONDARY_ROLE", "0", &emptystr)
				nodeclass      = asset.AttrFetch("NODECLASS", "0", &emptystr)
				status         = asset.Asset.Status
				state          = asset.Asset.State.Name
			)
			items = append(items, fmt.Sprintf("<%s|%s> %s [%s/%s/%s/%s] <fixme|%s:%s>", collins.Link(*asset), asset.Asset.Tag, *hostname, *nodeclass, *pool, *primary_role, *secondary_role, status, state))
		}
		// send a message back to that channel with the links to the assets
		if len(items) > 0 {
			_, _, err := api.PostMessage(m.ChannelId, strings.Join(items, "\n"), postParams)
			return true, err
		}
	}
	return false, nil
}

func AssetHostnameHandler(m *slack.MessageEvent) (bool, error) {
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
