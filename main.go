package main

import (
	"encoding/json"
	"flag"
	"fmt"
	c "github.com/byxorna/collinsbot/collins"
	"github.com/nlopes/slack"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

var (
	cli struct {
		file    string
		token   string
		botname string
		debug   bool
	}
	settings   Settings
	api        *slack.Slack
	postParams slack.PostMessageParameters
	collins    *c.Client
)

func init() {
	flag.StringVar(&cli.token, "token", "", "Slack API token")
	flag.StringVar(&cli.botname, "botname", "collinsbot", "Bot name")
	flag.StringVar(&cli.file, "config", "", "File containing Slack API token")
	flag.BoolVar(&cli.debug, "debug", false, "Turn on Slack API debugging")
	flag.Parse()
}

func main() {
	if cli.file != "" {
		log.Printf("Loading config from %s\n", cli.file)
		f, err := os.Open(cli.file)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		err = json.NewDecoder(f).Decode(&settings)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		log.Fatal("You need to pass a json file to -config")
	}

	// override whats in the settings with whats on the cli
	if cli.token != "" {
		log.Printf("Slack token passed via CLI: %s\n", cli.token)
		settings.Token = cli.token
	}
	if settings.Botname == "" || cli.botname != "" {
		settings.Botname = cli.botname
	}

	if settings.Token == "" {
		log.Fatal("You need to give me an API token!")
	}

	collins = c.New(settings.Collins.Username, settings.Collins.Password, settings.Collins.Host)

	// set up posting params
	postParams = slack.NewPostMessageParameters()
	postParams.Username = settings.Botname

	api = slack.New(settings.Token)
	api.SetDebug(cli.debug)
	resp, err := api.AuthTest()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Authed with Slack successfully: %+v\n", resp)

	for _, ch := range settings.Channels {
		log.Printf("Joining channel %s\n", ch)
		channel, err := api.JoinChannel(ch)
		if err != nil {
			log.Printf("Unable to join %s: %s\n", ch, err.Error())
			continue
		}
		log.Printf("Joined channel %s; Topic is %s, created by %s\n", channel.Name, channel.Topic.Value, channel.Topic.Creator)

		// leave the channel when we exit
		/* TODO: uncomment when using real bot API key, not user key
		defer func() {
			log.Printf("Leaving channel %s\n", channel.Name)
			left, err := api.LeaveChannel(channel.Id)
			if err != nil {
				log.Printf("Unable to leave channel %s (%s): %s\n", channel.Name, channel.Id, err.Error())
			}
			log.Printf("Left channel %s: %b", channel.Name, left)
		}()
		*/

		/*
			channelId, ts, err := api.PostMessage(channel.Id, fmt.Sprintf("Hello, %s", ch), postParams)
			if err != nil {
				log.Printf("Unable to message %s: %s\n", ch, err.Error())
				continue
			}
			log.Printf("Sent message to %s at %s\n", channelId, ts)
		*/

	}

	chIncoming := make(chan slack.SlackEvent)
	chOutgoing := make(chan slack.OutgoingMessage)
	ws, err := api.StartRTM("", "https://www.tumblr.com")
	if err != nil {
		log.Fatal("Unable to start realtime messaging websocket: %s\n", err.Error())
	}
	// send incoming events into the chIncoming channel
	// and record when we started listening for events so we can ignore those which happened earlier
	var socketEstablished = time.Now().Unix()
	go ws.HandleIncomingEvents(chIncoming)
	// keep the connection alive every 20s with a ping
	go ws.Keepalive(20 * time.Second)
	// process outgoing messages from chOutgoing
	go func() {
		for {
			select {
			case msg := <-chOutgoing:
				log.Printf("Sending message %+v\n", msg)
				if err := ws.SendMessage(&msg); err != nil {
					log.Printf("Error: %s\n", err.Error())
				}
			}
		}
	}()
	// process incoming messages
	for {
		select {
		case msg := <-chIncoming:
			//log.Printf("Received event:\n")
			switch msg.Data.(type) {
			case *slack.MessageEvent:
				msgevent := msg.Data.(*slack.MessageEvent)
				if ts, err := strconv.ParseInt(strings.Split(msgevent.Timestamp, ".")[0], 10, 64); err == nil {
					// if we didnt have trouble pulling the timestamp out, lets discard if it happened
					// before socketEstablished
					if socketEstablished > ts {
						log.Printf("Ignoring message %s at %d, which was sent before we started listening\n", msgevent.Msg.Text, ts)
						continue
					}
				} else {
					log.Printf("Unable to parse timestamp %s: %s\n", msgevent.Timestamp, err.Error())
				}

				// handle messages with any asset tags present - we will turn them into collins links
				tags := extractAssetTags(msgevent)
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
					msg := ws.NewOutgoingMessage(strings.Join(items, "\n"), msgevent.ChannelId)
					log.Printf("Sending %+v\n", msg)
					chOutgoing <- *msg
				}

				// handle messages with any hostnames present - if assets, link them
				hosts := extractHostnames(msgevent)
				for _, host := range hosts {
					//TODO!
					log.Printf("Found host %s in %s\n", host, msgevent.Msg.Text)
				}
			}
		}
	}

}
