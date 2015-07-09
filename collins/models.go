package collins

import (
	"encoding/json"
	"fmt"
)

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
