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
		nodeclass, nodeclass_ok           = asset.Attr("NODECLASS")
		pool, pool_ok                     = asset.Attr("POOL")
		primary_role, primary_role_ok     = asset.Attr("PRIMARY_ROLE")
		secondary_role, secondary_role_ok = asset.Attr("SECONDARY_ROLE")
	)

	return fmt.Sprintf("%s %s/%s/%s/%s %s",
		slackLinkIfSet(collins.Link(asset), asset.Asset.Tag, true),
		slackLinkIfSet(collins.LinkFromAttributeWeb("NODECLASS", nodeclass), nodeclass, nodeclass_ok),
		slackLinkIfSet(collins.LinkFromAttributeWeb("POOL", pool), pool, pool_ok),
		slackLinkIfSet(collins.LinkFromAttributeWeb("PRIMARY_ROLE", primary_role), primary_role, primary_role_ok),
		slackLinkIfSet(collins.LinkFromAttributeWeb("SECONDARY_ROLE", secondary_role), secondary_role, secondary_role_ok),
		slackLinkIfSet(collins.LinkFromAttributesWeb(map[string]string{
			"status": asset.Asset.Status,
			"state":  asset.Asset.State.Name,
		}, map[string]string{}), fmt.Sprintf("%s:%s", asset.Asset.Status, asset.Asset.State.Name), true),
	)
}

func slackLinkIfSet(link string, thing string, ok bool) string {
	if !ok || thing == "" {
		return ""
	}
	return fmt.Sprintf("<%s|%s>", link, thing)
}

func assetAttachmentFields(asset c.Asset) []slack.AttachmentField {
	var (
		nodeclass, nodeclass_ok           = asset.Attr("NODECLASS")
		pool, pool_ok                     = asset.Attr("POOL")
		primary_role, primary_role_ok     = asset.Attr("PRIMARY_ROLE")
		secondary_role, secondary_role_ok = asset.Attr("SECONDARY_ROLE")
	)
	return []slack.AttachmentField{
		slack.AttachmentField{
			Title: "Tag",
			Value: slackLinkIfSet(collins.Link(asset), asset.Asset.Tag, true),
			Short: true,
		},
		slack.AttachmentField{
			Title: "Status:State",
			//Value: slackLinkIfSet(collins.LinkFromAttributeWeb("STATUS", asset.Asset.Status), asset.Asset.Status, true),
			Value: slackLinkIfSet(collins.LinkFromAttributesWeb(map[string]string{
				"status": asset.Asset.Status,
				"state":  asset.Asset.State.Name,
			}, map[string]string{}), fmt.Sprintf("%s:%s", asset.Asset.Status, asset.Asset.State.Name), true),
			Short: true,
		},
		slack.AttachmentField{
			Title: "Nodeclass",
			Value: slackLinkIfSet(collins.LinkFromAttributeWeb("NODECLASS", nodeclass), nodeclass, nodeclass_ok),
			Short: true,
		},
		slack.AttachmentField{
			Title: "Pool",
			Value: slackLinkIfSet(collins.LinkFromAttributeWeb("POOL", pool), pool, pool_ok),
			Short: true,
		},
		slack.AttachmentField{
			Title: "Primary Role",
			Value: slackLinkIfSet(collins.LinkFromAttributeWeb("PRIMARY_ROLE", primary_role), primary_role, primary_role_ok),
			Short: true,
		},
		slack.AttachmentField{
			Title: "Secondary Role",
			Value: slackLinkIfSet(collins.LinkFromAttributeWeb("SECONDARY_ROLE", secondary_role), secondary_role, secondary_role_ok),
			Short: true,
		},
	}
}

func slackAssetAttachment(asset c.Asset) slack.Attachment {
	title := asset.Asset.Tag
	if v, ok := asset.Attribs["0"]["HOSTNAME"]; ok {
		// if the title is a link, slack will ignore TitleLink so strip the domain
		title = strings.Split(v, ".")[0]
	}
	color := "good"
	switch asset.Asset.Status {
	case "Maintenance":
		color = "danger"
		break
	case "Provisioning", "Provisioned":
		color = "#439FE0"
		break
	}
	//Text:      assetStringForSlack(asset), // text isnt necessary here
	return slack.Attachment{
		Title:     title,
		TitleLink: collins.Link(asset),
		Color:     color,
		Fallback:  collins.Link(asset),
		Fields:    assetAttachmentFields(asset),
	}
}

/*
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
*/

// return a random string from an array of strings
func random(arr []string) string {
	rand.Seed(time.Now().Unix())
	return arr[rand.Intn(len(arr))]
}
