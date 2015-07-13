package handlers

import (
	"github.com/nlopes/slack"
)

func AssetTag(c *Context, m *slack.MessageEvent) (bool, error) {
	// dont process if this came from a bot (like ourselves). avoids looping
	if isBotMessage(m) {
		return false, nil
	}
	// handle messages with any asset tags present - we will turn them into collins links
	tags := extractAssetTags(m.Msg.Text)
	if len(tags) > 0 {
		assets := c.lookupAssetsFromTags(tags)
		// send a message back to that channel with the links to the assets
		p := slack.NewPostMessageParameters()
		p.Attachments = make([]slack.Attachment, len(assets))
		for i, asset := range assets {
			p.Attachments[i] = c.slackAssetAttachment(asset)
		}
		_, _, err := c.Slack.PostMessage(m.ChannelId, "", p)
		return true, err
	}
	return false, nil
}
