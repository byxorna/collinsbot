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
	ws         *slack.SlackWS
	postParams slack.PostMessageParameters
	collins    *c.Client

	botIdentity *slack.AuthTestResponse

	// message handlers are functions that process a message event
	// similar to http route handlers. The first to return true stops processing
	messagehandlers = []Handler{
		Handler{"Help", HelpHandler},
		Handler{"YouAlive", YouAliveHandler},
		Handler{"AssetTag", AssetTagHandler},
		Handler{"AssetHostname", AssetHostnameHandler},
		//		Handler{"WTF", WTFHandler},
	}

	helpinfo = map[string]string{
		"help": "show this help output",
		"yt?":  "see if I am still alive",
		"mention any asset tag or hostname": "get a link to the asset",
	}
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
	// we will perform proper formatting per https://api.slack.com/docs/formatting, do make the server do no processing
	postParams.Parse = "none"

	api = slack.New(settings.Token)
	api.SetDebug(cli.debug)
	authresp, err := api.AuthTest()
	if err != nil {
		log.Fatal(err)
	}
	botIdentity = authresp
	log.Printf("Authed with Slack successfully as %s (%s)\n", botIdentity.User, botIdentity.UserId)

	chIncomingEvents := make(chan slack.SlackEvent)
	chOutgoingMessages := make(chan slack.OutgoingMessage)
	ws, err = api.StartRTM("", "https://www.tumblr.com")
	if err != nil {
		log.Fatal("Unable to start realtime messaging websocket: %s\n", err.Error())
	}

	// send incoming events into the chIncomingEvents channel
	// and record when we started listening for events so we can ignore those which happened earlier
	var socketEstablished = time.Now().Unix()
	go ws.HandleIncomingEvents(chIncomingEvents)
	// keep the connection alive every 20s with a ping
	go ws.Keepalive(20 * time.Second)

	// process outgoing messages from chOutgoing
	go func() {
		for {
			select {
			case msg := <-chOutgoingMessages:
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
		case msg := <-chIncomingEvents:
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
				if msgevent.Msg.Text == "" {
					continue
				}

				log.Printf("Processing: %s\n", msgevent.Msg.Text)
				for _, handler := range messagehandlers {
					//log.Printf("Testing handler %s...\n", handler.Name)
					handled, err := handler.Function(msgevent, chOutgoingMessages)
					if err != nil {
						log.Printf("Error handling message with %s: %s\n", handler.Name, err.Error())
						continue
					}
					if handled {
						log.Printf("%s handled message %s\n", handler.Name, msgevent.Msg.Text)
						break
					}
				}
			}
		}
	}

}
