package collins

import (
	"encoding/json"
	"fmt"
)

/*
{
"ASSET":{},
"HARDWARE":{},
"LLDP":{},
"IPMI":{},
"ADDRESSES":[],
"POWER":[],
"ATTRIBS":{
"0":{
"DISK_STORAGE_TOTAL":"6121230925824",
"BASE_SERIAL":"0000000000",
"CHASSIS_TAG":"Testing this"
} } }
*/

type Asset struct {
	//TODO this is fucking stupid. I would rather replicate the ruby model instead of just the json representation straight from collins...
	Asset   AssetFields   `json:"ASSET"`
	Attribs AttributesMap `json:"ATTRIBS"`
}

// the map of dimension (default 0) -> map[string]string
type AttributesMap map[string]map[string]string
type AssetFields struct {
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

/*
{"status":"error","data":{"message":"Asset with tag 'tumblrtag400' does not exist"}}
*/
type ErrorResponse struct {
	Status string            `json:"status"`
	Data   map[string]string `json:"data"`
}

func (e *ErrorResponse) Error() error {
	return fmt.Errorf(e.Data["message"])
}
