package main

// This is a simple example that sends IOPub data as server sent events
// over HTTP

// It relies on zeromq/goczmq, but could just as easily use pebbe/zmq4

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"log"
	"net/http"

	juno "github.com/rgbkrk/juno"
	ssebroker "github.com/rgbkrk/juno/examples/iopub-sse/sse-broker"
	zmq "github.com/zeromq/goczmq"
)

func main() {
	var connFile = flag.String("connection-file", "", "Path to connection file")
	flag.Parse()

	if *connFile == "" {
		fmt.Fprint(os.Stderr, "Connection file is required\n")
		flag.Usage()
		os.Exit(2)
	}

	connInfo, err := juno.NewConnectionInfo(*connFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to open connection file\n")
		fmt.Fprintln(os.Stderr, err)
		os.Exit(3)
	}

	ioConnection := connInfo.IOPubConnectionString()
	iopub := zmq.NewSubChanneler(ioConnection, "")

	defer iopub.Destroy()

	broker := ssebroker.NewServer()

	go func() {
		for {
			select {
			case wireMessage := <-iopub.RecvChan:
				var message juno.Message
				err := message.ParseWireProtocol(wireMessage, connInfo)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Unable to read message %v\n", err)
				}

				b, err := json.Marshal(message)

				broker.Notifier <- b
			}
		}
	}()

	log.Fatal("HTTP server error: ", http.ListenAndServe("localhost:3000", broker))

}
