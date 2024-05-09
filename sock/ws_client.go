package main

import (
	"strings"

	webs "golang.org/x/net/websocket"
)

func wsClient(origin string) {
	/* TODO: Parse origin and addr better */
	if !strings.HasPrefix(addr, "ws") {
		addr = "ws://" + addr
	}
	if !strings.HasPrefix(origin, "http") {
		origin = "http://" + origin
	}
	ws, err := webs.Dial(addr, "", origin)
	if err != nil {
		printErr(err, true)
		return
	}
	defer ws.Close()
	// Get user input
	go func() {
		for {
			if input, err := cin.ReadString('\n'); err != nil {
				if !done {
					printErr(err, true)
				}
				return
			} else {
				if err = webs.Message.Send(ws, input); err != nil {
					printErr(err, true)
					return
				}
			}
		}
	}()
	// Get socket output
	go func() {
		var msg string
		for {
			if err := webs.Message.Receive(ws, &msg); err != nil {
				printErr(err, true)
				return
			} else {
				write(cout, "< "+msg, false)
			}
		}
	}()
	for {
	}
}
