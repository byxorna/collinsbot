package main

import (
	"encoding/json"
	"flag"
	"github.com/nlopes/slack"
	"log"
	"os"
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
	go ws.HandleIncomingEvents(chIncoming)
	// keep the connection alive every 20s with a ping
	go ws.Keepalive(20 * time.Second)
	// process outgoing messages from chOutgoing
	go func() {
		for {
			select {
			case msg := <-chOutgoing:
				log.Printf("Sending message %+v\n", msg)
				ws.SendMessage(&msg)
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
				a := msg.Data.(*slack.MessageEvent)
				log.Printf("Message: %+v\n", a)
				//TODO: look for indicators that this is a request to us, and resolve shit?
			}
		}
	}

}
