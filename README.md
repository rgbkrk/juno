# juno

Go library for interacting with the Jupyter Messaging Protocol.

```go
import "github.com/rgbkrk/juno"

// Read in the connection information from a runtime
// For example, this would accept ~/Library/Jupyter/runtime/kernel-123.json
// and parse the connection information
connInfo, err := juno.OpenConnectionFile(pathToKernelRuntime)
...

// Create an iopub socket
iopub, err := juno.NewIOPubSocket(connInfo, "")
...

// Read in a Jupyter Message
message, err := iopub.ReadMessage()
```

