package main

type Settings struct {
	Token    string   `json:"token"`
	Channels []string `json:"channels"`
	Botname  string   `json:"botname"`
}
