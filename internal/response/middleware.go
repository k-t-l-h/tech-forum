package response

import (
	"encoding/json"
	"net/http"
)

type Response struct {
	Status int         `json:"status"`
	Body   interface{} `json:"body,omitempty"`
}

func Respond(writer http.ResponseWriter, status int, body interface{}) {

	writer.Header().Set("Content-Type", "application/json")

	writer.WriteHeader(status)
	json.NewEncoder(writer).Encode(body)
}
