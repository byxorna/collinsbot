package main

/*
  Helpers to construct messages for the bot
*/

import (
	c "github.com/byxorna/collinsbot/collins"
	"log"
)

func lookupAssetsFromTags(tags []string) []*c.Asset {
	var assets []*c.Asset
	for _, t := range tags {
		log.Printf("Attempting to resolve %s to a collins asset\n", t)
		a, err := collins.Get(t)
		if err != nil {
			log.Printf("Error resolving tag %s: %s", t, err.Error())
		} else {
			assets = append(assets, a)
		}
	}
	return assets
}
