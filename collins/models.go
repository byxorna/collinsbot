package collins

import "encoding/json"

type Asset struct {
	Id     int    `json:"ID"`
	Tag    string `json:"TAG"`
	Status string `json:"STATUS"`
	Type   string `json:"TYPE"`
	State  State  `json:"STATE"`
}

type State struct {
	Id          int    `json:"ID"`
	Status      string `json:"STATUS"`
	Name        string `json:"NAME"`
	Label       string `json:"LABEL"`
	Description string `json:"DESCRIPTION"`
}

type GenericResponse struct {
	Status string          `json:"status"`
	Data   json.RawMessage `json:"data"`
}
