// Tinkering with go, zmq, and Jupyter
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	zmq "github.com/pebbe/zmq4"
)

// MessageHeader is a Jupyter message header
// See: http://jupyter-client.readthedocs.org/en/latest/messaging.html
type MessageHeader struct {
	MessageID   string `json:"msg_id"`
	Username    string `json:"username"`
	Session     string `json:"session"`
	MessageType string `json:"msg_type"`
	Version     string `json:"version"`
}

// Message is a generic Jupyter message
type Message struct {
	Header       MessageHeader          `json:"header"`
	ParentHeader MessageHeader          `json:"parent_header"`
	Metadata     map[string]interface{} `json:"metadata"`
	Content      interface{}            `json:"content"`
}

// ConnectionInfo represents the runtime connection data used by Jupyter kernels
type ConnectionInfo struct {
	IOPubPort       int    `json:"iopub_port"`
	StdinPort       int    `json:"stdin_port"`
	IP              string `json:"ip"`
	Transport       string `json:"transport"`
	HBPort          int    `json:"hb_port"`
	SignatureScheme string `json:"signature_scheme"`
	ShellPort       int    `json:"shell_port"`
	ControlPort     int    `json:"control_port"`
	Key             string `json:"key"`
}

func main() {
	flag.Parse()
	if flag.NArg() < 1 {
		log.Fatalln("Need a command line argument for the connection file.")
	}

	connFile, err := os.Open(flag.Arg(0))
	if err != nil {
		fmt.Errorf("Couldn't open connection file: %v", err)
		os.Exit(1)
	}

	jsonParser := json.NewDecoder(connFile)

	var connInfo ConnectionInfo

	if err = jsonParser.Decode(&connInfo); err != nil {
		fmt.Errorf("Couldn't parse connection file: %v", err)
		os.Exit(2)
	}

	iopubSocket, err := zmq.NewSocket(zmq.SUB)
	if err != nil {
		fmt.Errorf("Couldn't start the iopub socket: %v", err)
	}

	defer iopubSocket.Close()

	connectionString := fmt.Sprintf("%s://%s:%d",
		connInfo.Transport,
		connInfo.IP,
		connInfo.IOPubPort)

	iopubSocket.Connect(connectionString)
	iopubSocket.SetSubscribe("")

	fmt.Println("Connected to")
	fmt.Println(connInfo)

	for {
		msg, err := iopubSocket.Recv(0)
		if err != nil {
			fmt.Printf("Error on receive: %v\n", err)
		}

		fmt.Println(msg)
	}

}
