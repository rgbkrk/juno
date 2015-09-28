// Package juno implements the messaging spec of the Jupyter Messaging Protocol,
// just for you.
// http://jupyter-client.readthedocs.org/en/latest/messaging.html
package juno

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"hash"
	"os"

	zmq "github.com/zeromq/goczmq"
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
func (m *Message) ParseWireProtocol(wireMessage [][]byte, connInfo ConnectionInfo) (err error) {
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

	if i >= len(wireMessage) {
		return errors.New("Couldn't find delimeter")
	}

	// Extract the zmq identiti(es)
	//identities := wireMessage[:delimiterLocation]

	// If the message was signed
	if len(connInfo.Key) != 0 {
		var hasher func() hash.Hash

		if connInfo.SignatureScheme == "hmac-sha256" {
			hasher = sha256.New
		} else {
			return errors.New("juno only supports hmac-sha256 for signature scheme currently")
		}

		mac := hmac.New(hasher, []byte(connInfo.Key))
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

// NewConnectionInfo reads in a connection file and creates a ConnectionInfo
// struct
func NewConnectionInfo(filename string) (ConnectionInfo, error) {
	var connInfo ConnectionInfo
	connFile, err := os.Open(filename)

	if err != nil {
		return connInfo, fmt.Errorf("Couldn't open connection file: %v", err)
	}
	defer connFile.Close()

	jsonParser := json.NewDecoder(connFile)

	if err = jsonParser.Decode(&connInfo); err != nil {
		return connInfo, fmt.Errorf("Couldn't parse connection file: %v", err)
	}

	return connInfo, nil
}

// OpenConnectionFile wraps NewConnectionInfo for backwards compatiblity
func OpenConnectionFile(filename string) (ConnectionInfo, error) {
	return NewConnectionInfo(filename)
}

// JupyterSocket is a zmq.Socket coupled with connection information
type JupyterSocket struct {
	ZMQSocket *zmq.Sock
	ConnInfo  ConnectionInfo
}

// ReadMessage reads a Jupyter Protocol Message
func (s *JupyterSocket) ReadMessage() (Message, error) {
	var message Message
	wireMessage, err := s.ZMQSocket.RecvMessage()
	if err != nil {
		return message, fmt.Errorf("Error on receive: %v", err)
	}

	message, err = s.NewMessage(wireMessage)
	if err != nil {
		return message, fmt.Errorf("Error on parsing wire message: %v", err)
	}
	return message, nil
}

// NewMessage creates a new message from a wire message using the key from this
// socket's connection info
func (s *JupyterSocket) NewMessage(wireMessage [][]byte) (Message, error) {
	var message Message
	message.ParseWireProtocol(wireMessage, s.ConnInfo)
	return message, nil
}

// Destroy obliterates zmq sockets
func (s *JupyterSocket) Destroy() {
	s.ZMQSocket.Destroy()
}

// NewIOPubSocket creates a new IOPub socket (on SUB) with the connInfo given
// subscribe is a comma delimited list of topics to subscribe to
func NewIOPubSocket(connInfo ConnectionInfo, subscribe string) (*JupyterSocket, error) {
	connectionString := connInfo.IOPubConnectionString()

	rawIOPubSocket, err := zmq.NewSub(connectionString, subscribe)
	if err != nil {
		return nil, err
	}

	iopub := &JupyterSocket{
		ZMQSocket: rawIOPubSocket,
		ConnInfo:  connInfo,
	}

	return iopub, nil
}

// ConnectionString forms the string for zmq libraries to connect
func (connInfo *ConnectionInfo) ConnectionString(port int) string {
	connectionString := fmt.Sprintf("%s://%s:%d",
		connInfo.Transport,
		connInfo.IP,
		port)
	return connectionString
}

// IOPubConnectionString forms the connection string for the IOPub socket
func (connInfo *ConnectionInfo) IOPubConnectionString() string {
	return connInfo.ConnectionString(connInfo.IOPubPort)
}

// NewMessage creates a new message based only on the connection info
func (connInfo *ConnectionInfo) NewMessage(wireMessage [][]byte) (Message, error) {
	var message Message
	err := message.ParseWireProtocol(wireMessage, *connInfo)
	return message, err
}
