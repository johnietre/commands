package main

/* TODO:
 * Use ANSI library to make it so user input is never interrupted by socket output
 * Ignore all EOF errors
 * Allow option to just test if address is accessible
 * Allow option to test if multiple ports are accessible
 * Flag for specifying laddr
 * Flag for timeout
 */

import (
	"bufio"
	"log"
	"os"
	"os/signal"

	"github.com/spf13/cobra"
)

const maxBufferSize = 256

var (
	cout       = bufio.NewWriter(os.Stdout)
	cerr       = bufio.NewWriter(os.Stderr)
	cin        = bufio.NewReader(os.Stdin)
	addr       = "127.0.0.1:8000"
	signalChan = make(chan os.Signal, 1)
	done       bool
	testOk     bool
)

func main() {
	log.SetFlags(0)

	var origin string
	var hub, ws, udp, server bool

	rootCmd := &cobra.Command{
		Use:                   "sock [FLAGS] [address to connect/listen (default: 127.0.0.1:8000)]",
		Short:                 "Connect to a socket or create a socket server",
		Long:                  "Connect to a socket or create a socket server. The default is to use TCP.",
		Args:                  cobra.MaximumNArgs(1),
		DisableFlagsInUseLine: true,
		Run: func(_ *cobra.Command, args []string) {
			addr = "127.0.0.1:8000"
			if len(args) != 0 {
				addr = args[0]
			}

			if server || hub {
				log.Printf("Running on %s", addr)
				if ws {
					go wsServer(hub)
				} else if udp {
					go udpServer(hub)
				} else {
					go tcpServer(hub)
				}
			} else {
				if ws {
					go wsClient(origin)
				} else if udp {
					go udpClient()
				} else {
					go tcpClient()
				}
			}

			signal.Notify(signalChan, os.Interrupt)
			<-signalChan
			done = true
		},
	}

	flags := rootCmd.Flags()
	flags.StringVar(
		&origin, "origin", "127.0.0.1",
		"(ws client only) origin of WS server (http/https, no port)",
	)
	flags.BoolVar(&hub, "hub", false, "Run server as a chat hub")
	flags.BoolVar(&ws, "ws", false, "Connect using a web socket")
	flags.BoolVar(&udp, "udp", false, "Connect using UDP")
	flags.BoolVarP(&server, "server", "S", false, "Start server (default: echo)")
	flags.BoolVar(
		&testOk, "test", false,
		"Only test if address is available, immediately exit on success/failure",
	)
	rootCmd.MarkFlagsMutuallyExclusive("ws", "udp")

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
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
