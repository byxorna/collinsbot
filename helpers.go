package main

/*
  Helpers to construct messages for the bot
*/

import (
	"fmt"
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

//TODO this should return a slack Attachment instead
func assetStringForSlack(asset *c.Asset) string {
	// this is crazy and hacky. There has to be a better way to format this
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
	return fmt.Sprintf("<%s|%s> <http://%s|%s> [<%s|%s>/%s/%s/%s] <fixme|%s:%s>",
		collins.Link(*asset), asset.Asset.Tag,
		*hostname, *hostname,
		collins.LinkFromAttribute("NODECLASS", *nodeclass), *nodeclass,
		collins.LinkFromAttribute("POOL", *pool), *pool,
		collins.LinkFromAttribute("PRIMARY_ROLE", *primary_role), *primary_role,
		collins.LinkFromAttribute("SECONDARY_ROLE", *secondary_role), *secondary_role,
		status, state)
}
