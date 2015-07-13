package handlers

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
		var tags = make([]string, len(tagres))
		for i, m := range tagres {
			tags[i] = m[1]
		}
		uniq(&tags)
		log.Printf("Found some asset tags: %+v", tags)
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
		uniq(&hosts)
		log.Printf("now: %+v", hosts)
		return hosts
	}
	return []string{}
}

func uniq(arr *[]string) {
	h := map[string]interface{}{}
	j := 0
	for i, v := range *arr {
		if _, ok := h[v]; !ok {
			h[v] = struct{}{}
			(*arr)[j] = (*arr)[i]
			j++
		}
	}
	*arr = (*arr)[:j]
}
