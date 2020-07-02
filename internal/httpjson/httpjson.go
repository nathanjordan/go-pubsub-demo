package httpjson

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type errorObj struct {
	Err string `json:"error"`
}

// JSONError returns a jsonified error message to the client.
func JSONError(rw http.ResponseWriter, status int, format string, v ...interface{}) {
	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(status)

	obj := errorObj{
		Err: fmt.Sprintf(format, v...),
	}
	enc := json.NewEncoder(rw)
	if err := enc.Encode(obj); err != nil {
		// if there's an error here there's not much we can do
		_, _ = io.WriteString(rw, "error encoding json error message")
	}
}
