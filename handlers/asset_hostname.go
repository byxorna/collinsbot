package handlers

import (
	"github.com/byxorna/collinsbot/collins"
	"github.com/nlopes/slack"
	"log"
)

func (c *Context) AssetHostnameHandler(m *slack.MessageEvent) (bool, error) {
	// handle messages with any hostnames present - if assets, link them
	if isBotMessage(m) {
		return false, nil
	}
	hosts := extractHostnames(m.Msg.Text)
	if len(hosts) > 0 {
		assets := []collins.Asset{}
		for _, host := range hosts {
			//TODO!
			log.Printf("Found host %s in message: %s\n", host, m.Msg.Text)
			a, err := c.Collins.Find(map[string]string{}, map[string]string{"hostname": host})
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
		if len(assets) > 0 {
			log.Printf("Found assets: %+v\n", assets)
			p := slack.NewPostMessageParameters()
			//old style list assets
			//p.Attachments = []slack.Attachment{*slackAssetsAttachment(assets)}
			p.Attachments = make([]slack.Attachment, len(assets))
			for i, asset := range assets {
				p.Attachments[i] = c.slackAssetAttachment(asset)
			}
			_, _, err := c.Slack.PostMessage(m.ChannelId, "", p)
			return true, err
		}
	}
	return false, nil
}
