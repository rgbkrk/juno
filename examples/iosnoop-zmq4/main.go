package iosnoop_zmq4

// This is a simple example that opens up an IOPub connection to a Jupyter
// kernel using pebbe/zmq4 and displays each message.

// For an alternative example that uses zeromq/goczmq, see ../iosnoop/main.go

import (
	"flag"
	"fmt"
	"os"

	zmq "github.com/pebbe/zmq4"
	juno "github.com/rgbkrk/juno"
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

	// In this example, we omit error checking for zmq calls for simplicity,
	// but you should check for errors in your code.
	iopub, _ := zmq.NewSocket(zmq.SUB)
	defer iopub.Close()
	_ = iopub.Connect(ioConnection)
	// Subscribe to all messages, see http://api.zeromq.org/4-1:zmq-setsockopt#toc41
	_ = iopub.SetSubscribe("")

	// Listen for messages... forever!
	for {
		wireMessage, _ := iopub.RecvMessageBytes(0)
		var message juno.Message
		err = message.ParseWireProtocol(wireMessage, connInfo)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to read message %v\n", err)
		}
		fmt.Println(message)

	}

}
