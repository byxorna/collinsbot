package main

import (
	c "github.com/byxorna/collinsbot/collins"
)

type Settings struct {
	Token    string   `json:"token"`
	Channels []string `json:"channels"`
	Botname  string   `json:"botname"`
	Collins  c.Config `json:"collins"`
}
