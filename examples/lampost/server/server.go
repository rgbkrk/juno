package lampost-server

import (
	"encoding/json"
	"fmt"
	"net/http"

	juno "github.com/rgbkrk/juno"
	broker "github.com/rgbkrk/juno/examples/iopub-sse/sse-broker"
)

func main() {
	b := broker.NewServer()

	fs := http.FileServer(http.Dir("static"))
	http.Handle("/", fs)

	http.HandleFunc("/api/ioju", func(w http.ResponseWriter, r *http.Request) {
		decoder := json.NewDecoder(r.Body)
		var message juno.Message
		err := decoder.Decode(&message)
		if err != nil {
			panic(err)
		}
		bMessage, err := json.Marshal(message)
		b.Notifier <- bMessage
		fmt.Println(string(bMessage[:40]))
	})
	http.Handle("/events", b)

	fmt.Println("About to serve on 127.0.0.1:8080")

	http.ListenAndServe(":8080", nil)
}
