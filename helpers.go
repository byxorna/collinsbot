package main

/*
  Helpers to construct messages for the bot
*/

import (
	"fmt"
	c "github.com/byxorna/collinsbot/collins"
	"github.com/nlopes/slack"
	"log"
	"math/rand"
	"strings"
	"time"
)

// I dont know if there is a better way to do this, but scan the text of the message
// for anything mentioning our @userid or just user to say if someone mentioned us
func isBotMention(m *slack.MessageEvent) bool {
	//log.Printf("Looking for bot mention in %+v (@%s or %s)\n", m, botIdentity.UserId, botIdentity.User)
	mention := strings.Contains(m.Msg.Text, fmt.Sprintf("@%s", botIdentity.UserId)) || strings.Contains(m.Msg.Text, botIdentity.User)
	//log.Printf("Was mention: %v\n", mention)
	return mention
}

func isBotMessage(m *slack.MessageEvent) bool {
	return m.Msg.SubType == "bot_message"
}

func lookupAssetsFromTags(tags []string) []c.Asset {
	var assets []c.Asset
	for _, t := range tags {
		log.Printf("Attempting to resolve %s to a collins asset\n", t)
		a, err := collins.Get(t)
		if err != nil {
			log.Printf("Error resolving tag %s: %s", t, err.Error())
		} else {
			assets = append(assets, *a)
		}
	}
	return assets
}

func assetStringForSlack(asset c.Asset) string {
	// this is crazy and hacky. There has to be a better way to format this
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
	//TODO fix me! if an attribute is missing we should omit the link
	return fmt.Sprintf("<%s|%s> <http://%s|%s> [<%s|%s>/<%s|%s>/<%s|%s>/<%s|%s>] <%s|%s:%s>",
		collins.Link(asset), asset.Asset.Tag,
		*hostname, *hostname,
		collins.LinkFromAttributeWeb("NODECLASS", *nodeclass), *nodeclass,
		collins.LinkFromAttributeWeb("POOL", *pool), *pool,
		collins.LinkFromAttributeWeb("PRIMARY_ROLE", *primary_role), *primary_role,
		collins.LinkFromAttributeWeb("SECONDARY_ROLE", *secondary_role), *secondary_role,
		collins.LinkFromAttributesWeb(map[string]string{
			"status": status,
			"state":  state,
		}, map[string]string{}), status, state,
	)
}

func slackAssetsAttachment(assets []c.Asset) *slack.Attachment {
	if len(assets) == 0 {
		return &slack.Attachment{
			Title:    "No assets found",
			Text:     "I couldn't find any assets matching that query!",
			Fallback: "I couldn't find any assets matching that query!",
			Color:    "danger",
		}
	}
	var fallbackparts = make([]string, len(assets))
	var textparts = make([]string, len(assets))
	for i, a := range assets {
		fallbackparts[i] = collins.Link(a)
		textparts[i] = assetStringForSlack(a)
	}
	return &slack.Attachment{
		Title:    fmt.Sprintf("%d Assets in Collins", len(assets)),
		Color:    "good",
		Fallback: strings.Join(fallbackparts, "\n"),
		Text:     strings.Join(textparts, "\n"),
	}
}

// return a random string from an array of strings
func random(arr []string) string {
	rand.Seed(time.Now().Unix())
	return arr[rand.Intn(len(arr))]
}
