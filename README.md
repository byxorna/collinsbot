# collinsbot
Slack collins bot

[![Docker Container Status](http://dockeri.co/image/byxorna/collinsbot)](https://registry.hub.docker.com/u/byxorna/collinsbot/)

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

### Run in docker

An automated build is maintained on the hub at https://registry.hub.docker.com/u/byxorna/collinsbot/. Inject your config into the container and give it a whirl.

```
docker run -v $(pwd)/my/config.json:/etc/config.json byxorna/collinsbot -config /etc/config.json
```

## TODO

I just randomly picked a go slack library (github.com/nlopes/slack). I am not super happy with it, and it seems pretty messy. Perhaps look into using:
* https://github.com/danryan/hal
* https://github.com/RobotsAndPencils/marvin
* https://github.com/ehazlett/phoenix

What kind of features should we support?
* implicitly pull out asset tags and hostnames from all messages
* subscribe to status changes, provisioning changes on an asset?
