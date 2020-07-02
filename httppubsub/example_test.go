package httppubsub

import (
	"bytes"
	"log"
	"net/http"
)

func ExampleNewServer() {
	ps, err := NewServer()
	if err != nil {
		log.Fatalf("error creating server %v", err)
	}

	http.ListenAndServe(":8080", ps)

	buf := bytes.NewBufferString(`{"msg": "Hello World!"`)
	http.Post("localhost:8080/broadcast", "application/json", buf)
}
