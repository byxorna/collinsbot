package main

import (
	"encoding/json"
	"flag"
	c "github.com/byxorna/collinsbot/collins"
	"github.com/byxorna/collinsbot/handlers"
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
	settings Settings
	//ws         *slack.SlackWS
	postParams slack.PostMessageParameters
	//collins    *c.Client

	handlerContext *handlers.Context

	// message handlers are functions that process a message event
	// similar to http route handlers. The first to return true stops processing
	messagehandlers = []string{
		"Help",
		"YouAlive",
		"AssetTag",
		"AssetHostname",
		//"WTF",
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

	collins := c.New(settings.Collins.Username, settings.Collins.Password, settings.Collins.Host)

	// set up posting params
	postParams = slack.NewPostMessageParameters()
	//postParams.Username = settings.Botname
	//postParams.LinkNames = 1
	// we will perform proper formatting per https://api.slack.com/docs/formatting, do make the server do no processing
	postParams.Parse = "none"

	api := slack.New(settings.Token)
	api.SetDebug(cli.debug)
	authresp, err := api.AuthTest()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Authed with Slack successfully as %s (%s)\n", authresp.User, authresp.UserId)

	log.Printf("Creating new handler context with message handlers:\n")
	for _, v := range messagehandlers {
		log.Printf("  %s\n", v)
	}
	handlerContext, err = handlers.New(collins, authresp, api, messagehandlers)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Starting up message handler")
	handlerContext.Run()

}
