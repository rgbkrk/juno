package main

// This is a simple example that sends IOPub data to a remote server via POST
// It relies on zeromq/goczmq, but could just as easily use pebbe/zmq4

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"sync"

	"net/http"

	juno "github.com/rgbkrk/juno"
	zmq "github.com/zeromq/goczmq"
)

func main() {
	var connFile = flag.String("connection-file", "", "Path to connection file")
	var lampostServer = flag.String("lampost-server", "https://lampost.lambdaops.com", "URL to a lampost server")
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

	// TODO: sync waitgroup
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		for {
			select {
			case wireMessage := <-iopub.RecvChan:
				var message juno.Message
				err := message.ParseWireProtocol(wireMessage, connInfo)

				if err != nil {
					fmt.Fprintf(os.Stderr, "Unable to read message %v\n", err)
				}

				b, err := json.Marshal(message)
				fmt.Printf("%s\n", b)

				reader := bytes.NewReader(b)

				resp, err := http.Post(*lampostServer, "application/json", reader)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Unable to POST to %v with %v", *lampostServer, message)
					fmt.Fprintln(os.Stderr, err)
					if resp != nil {
						fmt.Fprintf(os.Stderr, "[%v] %v", resp.StatusCode, resp.Status)
					}
					return
				}
				fmt.Printf("[%v] %v", resp.StatusCode, resp.Status)

				if resp.StatusCode != 200 {
					continue
				}

				var body []byte
				resp.Body.Read(body)

				fmt.Printf("%v", body)

			}
		}
		// This code is unreachable, but you would normally
		//     defer wg.Done()
		// once the goroutine was finished
	}()

	wg.Wait()

}
