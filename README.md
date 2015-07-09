# collinsbot
Slack collins bot

For great justice

## Features

Resolves asset tags and hostnames to useful data (hostname, link, pool, status)

## Build

Requires godep: `go get github.com/tools/godep`

```
godep go build
```

## Run

Populate your config.json, and pass a `-token` for your API key.

```
./collinsbot -config examples/config.json -token="$(cat slack.token)"
```

## TODO

I just randomly picked a go slack library (github.com/nlopes/slack). I am not super happy with it, and it seems pretty messy. Perhaps look into using:
* https://github.com/danryan/hal
* https://github.com/RobotsAndPencils/marvin
* https://github.com/ehazlett/phoenix

What kind of features should we support?
* implicitly pull out asset tags and hostnames from all messages
* subscribe to status changes, provisioning changes on an asset?
