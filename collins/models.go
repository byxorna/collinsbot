package collins

import (
	"encoding/json"
	"fmt"
)

/*
find response
{
"status": "success:ok",
"data": {
    "Pagination": {
      "PreviousPage": 0,
      "CurrentPage": 0,
      "NextPage": 0,
      "TotalResults": 3
    },
    "Data": [...]
*/

type PaginationData struct {
	PreviousPage int
	CurrentPage  int
	NextPage     int
	TotalResults int
}

type PagedAssetResponse struct {
	Pagination PaginationData
	Data       []Asset
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
