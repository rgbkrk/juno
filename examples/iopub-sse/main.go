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
	zmq "github.com/zeromq/goczmq"
)

// Broker distributes messages to clients
type Broker struct {

	// Events are pushed to this channel by the main events-gathering routine
	Notifier chan []byte

	// New client connections
	newClients chan chan []byte

	// Closed client connections
	closingClients chan chan []byte

	// Client connections registry
	clients map[chan []byte]bool
}

// NewServer sets up a server relying on broker for us
func NewServer() (broker *Broker) {
	// Instantiate a broker
	broker = &Broker{
		Notifier:       make(chan []byte, 1),
		newClients:     make(chan chan []byte),
		closingClients: make(chan chan []byte),
		clients:        make(map[chan []byte]bool),
	}

	// Set it running - listening and broadcasting events
	go broker.listen()

	return
}

func (broker *Broker) ServeHTTP(rw http.ResponseWriter, req *http.Request) {

	// Make sure that the writer supports flushing.
	//
	flusher, ok := rw.(http.Flusher)

	if !ok {
		http.Error(rw, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}

	rw.Header().Set("Content-Type", "text/event-stream")
	rw.Header().Set("Cache-Control", "no-cache")
	rw.Header().Set("Connection", "keep-alive")
	rw.Header().Set("Access-Control-Allow-Origin", "*")

	// Each connection registers its own message channel with the Broker's connections registry
	messageChan := make(chan []byte)

	// Signal the broker that we have a new connection
	broker.newClients <- messageChan

	// Remove this client from the map of connected clients
	// when this handler exits.
	defer func() {
		broker.closingClients <- messageChan
	}()

	// Listen to connection close and un-register messageChan
	notify := rw.(http.CloseNotifier).CloseNotify()

	go func() {
		<-notify
		broker.closingClients <- messageChan
	}()

	for {

		// Write to the ResponseWriter
		// Server Sent Events compatible
		fmt.Fprintf(rw, "data: %s\n\n", <-messageChan)

		// Flush the data immediatly instead of buffering it for later.
		flusher.Flush()
	}

}

func (broker *Broker) listen() {
	for {
		select {
		case s := <-broker.newClients:

			// A new client has connected.
			// Register their message channel
			broker.clients[s] = true
			log.Printf("Client added. %d registered clients", len(broker.clients))
		case s := <-broker.closingClients:

			// A client has dettached and we want to
			// stop sending them messages.
			delete(broker.clients, s)
			log.Printf("Removed client. %d registered clients", len(broker.clients))
		case event := <-broker.Notifier:

			// We got a new event from the outside!
			// Send event to all connected clients
			for clientMessageChan := range broker.clients {
				clientMessageChan <- event
			}
		}
	}

}

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

	broker := NewServer()

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
