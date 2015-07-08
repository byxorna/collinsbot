package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/nlopes/slack"
	"log"
	"os"
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
		channelId, ts, err := api.PostMessage(ch, fmt.Sprintf("Hello, %s", ch), postParams)
		if err != nil {
			log.Printf("Unable to message %s: %s\n", ch, err.Error())
			continue
		}
		log.Printf("Sent message to %s at %s\n", channelId, ts)

	}

}
