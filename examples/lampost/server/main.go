package main

import (
	"encoding/json"
	"net/http"

	juno "github.com/rgbkrk/juno"
	broker "github.com/rgbkrk/juno/examples/iopub-sse/sse-broker"
)

func main() {
	b := broker.NewServer()

	http.HandleFunc("/ioju", func(w http.ResponseWriter, r *http.Request) {
		decoder := json.NewDecoder(r.Body)
		var message juno.Message
		err := decoder.Decode(&message)
		if err != nil {
			panic(err)
		}
		bMessage, err := json.Marshal(message)
		b.Notifier <- bMessage
	})
	http.HandleFunc("/events", b)

	http.ListenAndServe(":8080")
}
