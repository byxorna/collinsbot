package main

import (
	"log"
	"regexp"
)

var (
	tagRegex, _ = regexp.Compile(`\b(vm-[0-9a-f]{8}|\d{6,})\b`)
	// captures hosts that have prefix-8chhash[.dc1[.tumblr.net]]
	hostRegex, _ = regexp.Compile(`\b([\w\-]{2,}-[a-f0-9]{8}(?:\.\w{3,})?(?:\.tumblr\.net)?)\b`)
)

// given a message, extract the list of what we think are asset tags
func extractAssetTags(txt string) []string {
	//try to detect hostnames or asset tags
	tagres := tagRegex.FindAllStringSubmatch(txt, -1)
	if len(tagres) > 0 {
		// we abuse a map of string to empty interface as a set for uniq'ing tags
		var seentags = make(map[string]interface{}, len(tagres))
		var tags = []string{}
		log.Printf("Found some asset tags: %+v", tagres)
		for _, m := range tagres {
			// strip out duplicate tags
			if _, ok := seentags[m[1]]; ok {
				continue
			}
			// otherwise we havent seen this tag yet, append it to the list
			seentags[m[1]] = struct{}{}
			tags = append(tags, m[1])
		}
		return tags
	}
	return []string{}
}

func extractHostnames(txt string) []string {
	//try to detect hostnames or asset tags
	hostres := hostRegex.FindAllStringSubmatch(txt, -1)
	if len(hostres) > 0 {
		var hosts = make([]string, len(hostres))
		log.Printf("Found some asset hostnames: %+v", hostres)
		for i, m := range hostres {
			hosts[i] = m[1]
		}
		return hosts
	}
	return []string{}
}