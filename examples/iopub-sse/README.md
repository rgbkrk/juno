This is an example of taking all IOPub messages and sending them as Server Sent Events over HTTP.

To run this example, run a jupyter kernel and find out where that kernel's connection file is:

```bash
$ jupyter console
Jupyter Console 4.1.0.dev

[ZMQTerminalIPythonApp] Loading IPython extension: storemagic

In [1]: %connect_info
{
  "stdin_port": 54490,
  "ip": "127.0.0.1",
  "control_port": 54491,
  "hb_port": 54492,
  "signature_scheme": "hmac-sha256",
  "key": "44644cf2-1f06-44c1-a9dd-b087ff9c2d84",
  "shell_port": 54488,
  "transport": "tcp",
  "iopub_port": 54489
}

Paste the above JSON into a file, and connect with:
    $> ipython <app> --existing <file>
or, if you are local, you can connect with just:
    $> ipython <app> --existing /Users/rgbkrk/Library/Jupyter/runtime/kernel-27804.json
or even just:
    $> ipython <app> --existing
if this is the most recent IPython session you have started.
```

Now run this example with your own connection file:

```bash
$ go run main.go -connection-file /Users/rgbkrk/Library/Jupyter/runtime/kernel-27804.json
2015/09/28 11:41:56 Client added. 1 registered clients
```

After that, open index.html in the same folder, type some commands in your jupyter console and watch it go!

```bash
$ open index.html
```

<img width="1324" alt="screenshot 2015-09-28 11 43 05" src="https://cloud.githubusercontent.com/assets/836375/10142075/000dc8ca-65d6-11e5-86df-c45c04bd2cbc.png">
