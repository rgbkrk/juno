# juno

Go library for interacting with the Jupyter Messaging Protocol.

```go
import "github.com/rgbkrk/juno"

// Read in the connection information from a runtime
// For example, this would accept ~/Library/Jupyter/runtime/kernel-123.json
// and parse the connection information
connInfo, err := juno.NewConnectionInfo(pathToKernelRuntime)
// Test the err...

// Create an iopub socket that provides go channels
ioConnection := connInfo.IOPubConnectionString()
// Can use any zmq library, goczmq used here
iopub := zmq.NewSubChanneler(ioConnection, "")
defer iopub.Destroy()

// Listen for messages... forever!
for {
  select {
  case wireMessage := <-iopub.RecvChan:
    // Get a wire message
    // Then parse it
    var message juno.Message
    err := message.ParseWireProtocol(wireMessage, connInfo)
    // Do something with the message
  }
}
```
