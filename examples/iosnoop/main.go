package main

// This is a simple example that opens up an IOPub connection to a Jupyter
// kernel using zeromq/goczmq and displays each message.

// For an alternative example that uses pebbe/zmq4, see ../iosnoop-zmq4/main.go

import (
	"flag"
	"fmt"
	"os"

	juno "github.com/rgbkrk/juno"
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

	// Listen for messages... forever!
	for {
		select {
		case wireMessage := <-iopub.RecvChan:
			var message juno.Message
			err := message.ParseWireProtocol(wireMessage, connInfo)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Unable to read message %v\n", err)
			}
			fmt.Println(message)
		}
	}

}
