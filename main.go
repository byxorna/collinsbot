package main

import (
	"encoding/json"
	"flag"
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
	flag.StringVar(&cli.botname, "botname", "", "Bot name")
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
	//postParams.LinkNames = 1
	//postParams.Parse = "full"

	api = slack.New(settings.Token)
	api.SetDebug(cli.debug)
	resp, err := api.AuthTest()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Authed with Slack successfully: %+v\n", resp)

	// handlers are a set of functions that process a messsage and either handle them (true)
	// or skip them (move on to next handler, and possibly blow up)
	messagehandlers := map[string]func(*slack.MessageEvent) (bool, error){
		"YouAliveHandler":      YouAliveHandler,
		"AssetTagHandler":      AssetTagHandler,
		"AssetHostnameHandler": AssetHostnameHandler,
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
	//TODO this isnt super useful, because slack only does simple formatting over websockets
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

				for name, handler := range messagehandlers {
					handled, err := handler(msgevent)
					if err != nil {
						log.Printf("Error handling message with %s: %s\n", name, handler, err.Error())
						continue
					}
					if handled {
						log.Printf("%s handled message %s\n", name, msgevent.Msg.Text)
						break
					}
				}

			}
		}
	}

}
