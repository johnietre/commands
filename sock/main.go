package main

/* TODO:
 * Use ANSI library to make it so user input is never interrupted by socket output
 */

import (
	"bufio"
	"flag"
	"os"
	"os/signal"
)

const maxBufferSize = 256

var (
	cout       = bufio.NewWriter(os.Stdout)
	cerr       = bufio.NewWriter(os.Stderr)
	cin        = bufio.NewReader(os.Stdin)
	addr       = ":8000"
	signalChan = make(chan os.Signal, 1)
	done       bool
)

func main() {
	var origin string
	var hub, ws, server bool
	flag.StringVar(&addr, "addr", ":8000", "Network address to connect/listen to")
	flag.StringVar(&origin, "origin", "", "(ws client only) origin of WS server (http/https, no port)")
	flag.BoolVar(&hub, "hub", false, "Run server as a chat hub")
	flag.BoolVar(&ws, "ws", false, "Connect using a web socket")
	flag.BoolVar(&server, "S", false, "Start server (default: echo)")
	flag.Parse()

	if server || hub {
		if ws {
			go wsServer(hub)
		} else {
			go socketServer(hub)
		}
	} else {
		if ws {
			go wsClient(origin)
		} else {
			go socketClient()
		}
	}

	signal.Notify(signalChan, os.Interrupt)
	<-signalChan
	done = true
}

func write(w *bufio.Writer, text string, newline bool) error {
	if newline {
		text += "\n"
	}
	if _, err := w.WriteString(text); err != nil {
		return err
	}
	return w.Flush()
}

func printErr(err error, fatal bool) {
	write(cerr, "Error: "+err.Error(), true)
	if fatal {
		signalChan <- os.Interrupt
	}
}
