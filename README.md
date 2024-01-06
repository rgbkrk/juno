# juno

Go library for interacting with the Jupyter Messaging Protocol.

```go
package main

import (
	"log"
	"os"

	"github.com/rgbkrk/juno"
	"github.com/zeromq/goczmq"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Please provide a path to a kernel runtime")
	}

	pathToKernelRuntime := os.Args[1]

	connInfo, err := juno.NewConnectionInfo(pathToKernelRuntime)

	if err != nil {
		log.Fatal(err)
	}

	// Create an iopub socket that provides go channels
	ioConnection := connInfo.IOPubConnectionString()
	iopub := goczmq.NewSubChanneler(ioConnection, "")
	defer iopub.Destroy()

	// Listen for messages... forever!
	for {
		select {
		case wireMessage := <-iopub.RecvChan:
			var message juno.Message
			err := message.ParseWireProtocol(wireMessage, connInfo)
			// Do something with the message

			if err != nil {
				// Handle the error
				log.Println("Error parsing message: ", err)
				continue
			}

			log.Println("Message: ", message)

		}
	}
}
```