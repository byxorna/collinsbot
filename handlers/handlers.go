package handlers

import (
	"fmt"
	"github.com/byxorna/collinsbot/collins"
	"github.com/nlopes/slack"
	"log"
	"math/rand"
	"reflect"
	"strconv"
	"strings"
	"time"
)

type HandlerFn func(*slack.MessageEvent) (bool, error)
type Handler struct {
	Name     string
	Function reflect.Value
}

type Context struct {
	Collins            *collins.Client
	Self               *slack.AuthTestResponse
	Slack              *slack.Slack
	Handlers           []Handler
	chIncomingEvents   chan slack.SlackEvent
	chOutgoingMessages chan slack.OutgoingMessage
	ws                 *slack.SlackWS
}

func New(client *collins.Client, self *slack.AuthTestResponse, api *slack.Slack, handlers []string) (*Context, error) {
	c := Context{
		Collins:            client,
		Self:               self,
		Slack:              api,
		Handlers:           []Handler{},
		chIncomingEvents:   make(chan slack.SlackEvent),
		chOutgoingMessages: make(chan slack.OutgoingMessage),
	}

	for _, h := range handlers {
		err := c.Register(h)
		if err != nil {
			return nil, err
		}
	}

	return &c, nil
}

func (c *Context) Register(handler string) error {
	// make sure we can invoke this handler method and it matches our signature
	m, ok := reflect.TypeOf(c).MethodByName(handler)
	if !ok {
		return fmt.Errorf("Method %s doesnt exist on this context", handler)
	}
	log.Printf("%+v\n", m)
	//_, ok := m.Interface().(func(*slack.MessageEvent) (bool, error))
	c.Handlers = append(c.Handlers, Handler{
		Name:     handler,
		Function: m,
	})
	return nil
}

func (c *Context) Run() {
	wsf, err := c.Slack.StartRTM("", "https://www.tumblr.com")
	if err != nil {
		log.Fatal("Unable to start realtime messaging websocket: %s\n", err.Error())
	}
	c.ws = wsf

	// send incoming events into the chIncomingEvents channel
	// and record when we started listening for events so we can ignore those which happened earlier
	var socketEstablished = time.Now().Unix()
	go c.ws.HandleIncomingEvents(c.chIncomingEvents)
	// keep the connection alive every 20s with a ping
	go c.ws.Keepalive(20 * time.Second)

	// process outgoing messages from chOutgoing
	go func() {
		for {
			select {
			case msg := <-c.chOutgoingMessages:
				log.Printf("Sending message %+v\n", msg)
				if err := c.ws.SendMessage(&msg); err != nil {
					log.Printf("Error: %s\n", err.Error())
				}
			}
		}
	}()

	// process incoming messages
	for {
		select {
		case msg := <-c.chIncomingEvents:
			//log.Printf("Received event:\n")
			switch msg.Data.(type) {
			case *slack.MessageEvent:
				msgevent := msg.Data.(*slack.MessageEvent)

				// if we didnt have trouble pulling the timestamp out, lets discard if it happened
				// before socketEstablished
				if ts, err := strconv.ParseInt(strings.Split(msgevent.Timestamp, ".")[0], 10, 64); err == nil {
					if socketEstablished > ts {
						log.Printf("Ignoring message %s at %d, which was sent before we started listening\n", msgevent.Msg.Text, ts)
						continue
					}
				} else {
					log.Printf("Unable to parse timestamp %s: %s\n", msgevent.Timestamp, err.Error())
				}
				c.Handle(msgevent)
			}
		}
	}
}

func (c *Context) Handle(m *slack.MessageEvent) {
	if m.Msg.Text == "" {
		return
	}

	log.Printf("Processing: %s\n", m.Msg.Text)
	for _, handler := range c.Handlers {
		log.Printf("Testing handler %s %v...\n", handler.Name, handler.Function)
		vals := handler.Function.Call([]reflect.Value{reflect.ValueOf(m)})
		log.Printf("Vals: %+v\n", vals)
		// try and pull out (bool, error) from the []reflect.Value
		handled := vals[0].Interface().(bool)
		//TODO why cant we just typeassert vals[1] as .(error)
		var err error = nil
		switch vals[1].Interface().(type) {
		case nil:
			err = nil
			break
		case error:
			err = vals[1].Interface().(error)
			break
		}
		log.Printf("Handled: %v err: %v\n", handled, err)
		if err != nil {
			log.Printf("Error handling message with %s: %s\n", handler.Name, err.Error())
			continue
		}
		if handled {
			log.Printf("%s handled message %s\n", handler.Name, m.Msg.Text)
			break
		}
	}
}

/*
  Helpers to construct messages for the bot
*/

// I dont know if there is a better way to do this, but scan the text of the message
// for anything mentioning our @userid or just user to say if someone mentioned us
func (c *Context) isBotMention(m *slack.MessageEvent) bool {
	//log.Printf("Looking for bot mention in %+v (@%s or %s)\n", m, botIdentity.UserId, botIdentity.User)
	mention := strings.Contains(m.Msg.Text, fmt.Sprintf("@%s", c.Self.UserId)) || strings.Contains(m.Msg.Text, c.Self.User)
	//log.Printf("Was mention: %v\n", mention)
	return mention
}

func (c *Context) lookupAssetsFromTags(tags []string) []collins.Asset {
	var assets []collins.Asset
	for _, t := range tags {
		log.Printf("Attempting to resolve %s to a collins asset\n", t)
		a, err := c.Collins.Get(t)
		if err != nil {
			log.Printf("Error resolving tag %s: %s", t, err.Error())
		} else {
			assets = append(assets, *a)
		}
	}
	return assets
}

func (c *Context) assetStringForSlack(asset collins.Asset) string {
	// this is crazy and hacky. There has to be a better way to format this
	var (
		nodeclass, nodeclass_ok           = asset.Attr("NODECLASS")
		pool, pool_ok                     = asset.Attr("POOL")
		primary_role, primary_role_ok     = asset.Attr("PRIMARY_ROLE")
		secondary_role, secondary_role_ok = asset.Attr("SECONDARY_ROLE")
	)

	return fmt.Sprintf("%s %s/%s/%s/%s %s",
		slackLinkIfSet(c.Collins.Link(asset), asset.Asset.Tag, true),
		slackLinkIfSet(c.Collins.LinkFromAttributeWeb("NODECLASS", nodeclass), nodeclass, nodeclass_ok),
		slackLinkIfSet(c.Collins.LinkFromAttributeWeb("POOL", pool), pool, pool_ok),
		slackLinkIfSet(c.Collins.LinkFromAttributeWeb("PRIMARY_ROLE", primary_role), primary_role, primary_role_ok),
		slackLinkIfSet(c.Collins.LinkFromAttributeWeb("SECONDARY_ROLE", secondary_role), secondary_role, secondary_role_ok),
		slackLinkIfSet(c.Collins.LinkFromAttributesWeb(map[string]string{
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

func (c *Context) assetAttachmentFields(asset collins.Asset) []slack.AttachmentField {
	var (
		nodeclass, nodeclass_ok           = asset.Attr("NODECLASS")
		pool, pool_ok                     = asset.Attr("POOL")
		primary_role, primary_role_ok     = asset.Attr("PRIMARY_ROLE")
		secondary_role, secondary_role_ok = asset.Attr("SECONDARY_ROLE")
	)
	return []slack.AttachmentField{
		slack.AttachmentField{
			Title: "Tag",
			Value: slackLinkIfSet(c.Collins.Link(asset), asset.Asset.Tag, true),
			Short: true,
		},
		slack.AttachmentField{
			Title: "Status:State",
			//Value: slackLinkIfSet(collins.LinkFromAttributeWeb("STATUS", asset.Asset.Status), asset.Asset.Status, true),
			Value: slackLinkIfSet(c.Collins.LinkFromAttributesWeb(map[string]string{
				"status": asset.Asset.Status,
				"state":  asset.Asset.State.Name,
			}, map[string]string{}), fmt.Sprintf("%s:%s", asset.Asset.Status, asset.Asset.State.Name), true),
			Short: true,
		},
		slack.AttachmentField{
			Title: "Nodeclass",
			Value: slackLinkIfSet(c.Collins.LinkFromAttributeWeb("NODECLASS", nodeclass), nodeclass, nodeclass_ok),
			Short: true,
		},
		slack.AttachmentField{
			Title: "Pool",
			Value: slackLinkIfSet(c.Collins.LinkFromAttributeWeb("POOL", pool), pool, pool_ok),
			Short: true,
		},
		slack.AttachmentField{
			Title: "Primary Role",
			Value: slackLinkIfSet(c.Collins.LinkFromAttributeWeb("PRIMARY_ROLE", primary_role), primary_role, primary_role_ok),
			Short: true,
		},
		slack.AttachmentField{
			Title: "Secondary Role",
			Value: slackLinkIfSet(c.Collins.LinkFromAttributeWeb("SECONDARY_ROLE", secondary_role), secondary_role, secondary_role_ok),
			Short: true,
		},
	}
}

func (c *Context) slackAssetAttachment(asset collins.Asset) slack.Attachment {
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
		TitleLink: c.Collins.Link(asset),
		Color:     color,
		Fallback:  c.Collins.Link(asset),
		Fields:    c.assetAttachmentFields(asset),
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

func isBotMessage(m *slack.MessageEvent) bool {
	return m.Msg.SubType == "bot_message"
}

// return a random string from an array of strings
func random(arr []string) string {
	rand.Seed(time.Now().Unix())
	return arr[rand.Intn(len(arr))]
}
