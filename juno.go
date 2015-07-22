// Package juno implements light amounts of the Jupyter Messaging Protocol
// http://jupyter-client.readthedocs.org/en/latest/messaging.html
package juno

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	zmq "github.com/pebbe/zmq4"
)

// MessageHeader is a Jupyter message header
// http://jupyter-client.readthedocs.org/en/latest/messaging.html
type MessageHeader struct {
	MessageID   string `json:"msg_id"`
	Username    string `json:"username"`
	Session     string `json:"session"`
	MessageType string `json:"msg_type"`
	Version     string `json:"version"`
}

// Message is a generic Jupyter message (not a wire message)
// http://jupyter-client.readthedocs.org/en/latest/messaging.html#general-message-format
type Message struct {
	Header       MessageHeader          `json:"header"`
	ParentHeader MessageHeader          `json:"parent_header"`
	Metadata     map[string]interface{} `json:"metadata"`
	Content      map[string]interface{} `json:"content"`
}

// MimeBundle is a collection of `mimetypes -> data`
// Example:
//     'text/html' -> '<h1>Hey!</h1>'
//     'image/png' -> 'R0lGODlhAQABAIAAAAAAAP///yH5BAEAAAAALAAAAAABAAEAAAIBRAA7'
type MimeBundle map[string]string

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

// DELIMITER denotes the Jupyter multipart message
const DELIMITER = "<IDS|MSG>"

// ParseWireProtocol fills a Message with all the juicy Jupyter bits
func (m *Message) ParseWireProtocol(wireMessage [][]byte, key []byte) (err error) {
	var i int
	var el []byte

	// Wire protocol
	// http://jupyter-client.readthedocs.org/en/latest/messaging.html#the-wire-protocol
	/**
		[
	  		b'u-u-i-d',         # zmq identity(ies)
	  		b'<IDS|MSG>',       # delimiter
	  		b'baddad42',        # HMAC signature
			b'{header}',        # serialized header dict
			b'{parent_header}', # serialized parent header dict
			b'{metadata}',      # serialized metadata dict
			b'{content}',       # serialized content dict
			b'blob',            # extra raw data buffer(s)
	  		...
		]
	*/
	// Determine where the delimiter is
	for i, el = range wireMessage {
		if string(el) == DELIMITER {
			break // Found our delimiter
		}
	}

	if string(wireMessage[i]) != DELIMITER {
		return errors.New("Couldn't find delimeter")
	}

	// Extract the zmq identiti(es)
	//identities := wireMessage[:delimiterLocation]

	// If the message was signed
	if len(key) != 0 {
		// TODO: Programmatic selection of scheme
		mac := hmac.New(sha256.New, key)
		for _, msgpart := range wireMessage[i+2 : i+6] {
			mac.Write(msgpart)
		}
		signature := make([]byte, hex.DecodedLen(len(wireMessage[i+1])))
		hex.Decode(signature, wireMessage[i+1])
		if !hmac.Equal(mac.Sum(nil), signature) {
			return errors.New("Invalid signature")
		}
	}

	json.Unmarshal(wireMessage[i+2], &m.Header)
	json.Unmarshal(wireMessage[i+3], &m.ParentHeader)
	json.Unmarshal(wireMessage[i+4], &m.Metadata)
	json.Unmarshal(wireMessage[i+5], &m.Content)

	return nil
}

// OpenConnectionFile is a helper method that opens a connection file and reads
// it into a ConnectionInfo struct
func OpenConnectionFile(filename string) (ConnectionInfo, error) {
	var connInfo ConnectionInfo
	connFile, err := os.Open(filename)
	if err != nil {
		return connInfo, fmt.Errorf("Couldn't open connection file: %v", err)
	}

	jsonParser := json.NewDecoder(connFile)

	if err = jsonParser.Decode(&connInfo); err != nil {
		return connInfo, fmt.Errorf("Couldn't parse connection file: %v", err)
	}

	return connInfo, nil
}

// JupyterSocket is a zmq.Socket coupled with connection information
type JupyterSocket struct {
	ZMQSocket *zmq.Socket
	ConnInfo  ConnectionInfo
}

// ReadMessage reads a Jupyter Protocol Message
func (s *JupyterSocket) ReadMessage() (Message, error) {
	var message Message
	wireMessage, err := s.ZMQSocket.RecvMessageBytes(0)
	if err != nil {
		return message, fmt.Errorf("Error on receive: %v", err)
	}

	message.ParseWireProtocol(wireMessage, []byte(s.ConnInfo.Key))

	return message, nil
}

// Close shutdowns zmq sockets
func (s *JupyterSocket) Close() {
	s.ZMQSocket.Close()
}

// NewIOPubSocket creates a new IOPub socket (on SUB) with the connInfo given
func NewIOPubSocket(connInfo ConnectionInfo, subscribe string) (*JupyterSocket, error) {
	rawIOPubSocket, err := zmq.NewSocket(zmq.SUB)
	if err != nil {
		return nil, err
	}

	connectionString := fmt.Sprintf("%s://%s:%d",
		connInfo.Transport,
		connInfo.IP,
		connInfo.IOPubPort)

	rawIOPubSocket.Connect(connectionString)
	rawIOPubSocket.SetSubscribe("")

	iopub := &JupyterSocket{
		ZMQSocket: rawIOPubSocket,
		ConnInfo:  connInfo,
	}

	return iopub, nil
}
