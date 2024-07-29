package main

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"

	//webs "golang.org/x/net/websocket"
	webs "nhooyr.io/websocket"
)

func isGotClose(err error) bool {
	return strings.Contains(
		err.Error(), "failed to get reader: received close frame",
	)
}

func wsServer(hub bool) {
	type hubMsg struct{ From, Msg string }
	var conns sync.Map
	var hubChan chan hubMsg

	hubHandler := func(ws *webs.Conn, r *http.Request) {
		defer ws.Close(webs.StatusNormalClosure, "")
		defer conns.Delete(r.RemoteAddr)
		conns.Store(r.RemoteAddr, ws)
		for {
			mt, msg, err := ws.Read(context.Background())
			if err != nil {
				if !errors.Is(err, io.EOF) && !isGotClose(err) {
					printErr(err, false)
				}
				return
			} else if mt != webs.MessageText {
				// TODO
				continue
			}
			hubChan <- hubMsg{
				r.RemoteAddr,
				strings.ReplaceAll(string(msg), "\n", ""),
			}
		}
	}
	echoHandler := func(ws *webs.Conn, r *http.Request) {
		defer ws.Close(webs.StatusNormalClosure, "")
		for {
			if mt, msg, err := ws.Read(context.Background()); err != nil {
				if !errors.Is(err, io.EOF) && !isGotClose(err) {
					printErr(err, false)
				}
				return
			} else if err := ws.Write(context.Background(), mt, msg); err != nil {
				printErr(err, false)
				return
			}
		}
	}
	if hub {
		hubChan = make(chan hubMsg, 5)
		go func() {
			for msg := range hubChan {
				bmsg, _ := json.Marshal(msg)
				bmsg = append(bmsg, '\n')
				conns.Range(func(iAddr, iConn interface{}) bool {
					a, ws := iAddr.(string), iConn.(*webs.Conn)
					if a != msg.From {
						ws.Write(context.Background(), webs.MessageText, bmsg)
					}
					return true
				})
			}
		}()
	}
	server := &http.Server{
		Handler: func() *http.ServeMux {
			r := http.NewServeMux()
			var handler func(*webs.Conn, *http.Request)
			if hub {
				handler = hubHandler
			} else {
				handler = echoHandler
			}
			opts := &webs.AcceptOptions{
				InsecureSkipVerify: true,
			}
			r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
				conn, err := webs.Accept(w, r, opts)
				if err == nil {
					handler(conn, r)
				}
			})
			return r
		}(),
		ErrorLog: log.New(cerr, "Error: ", 0),
	}
	defer server.Close()

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		printErr(err, true)
		return
	}
	if testOk {
		return
	}

	printErr(server.Serve(ln), true)
}

func wsClient(origin string) {
	/* TODO: Parse origin and addr better */
	if !strings.HasPrefix(addr, "ws") {
		addr = "ws://" + addr
	}
	if !strings.HasPrefix(origin, "http") {
		origin = "http://" + origin
	}
	ws, _, err := webs.Dial(context.Background(), addr, &webs.DialOptions{
		Host: origin,
	})
	if err != nil {
		printErr(err, true)
		return
	}
	defer ws.Close(webs.StatusNormalClosure, "")
	if testOk {
		return
	}

	// Get user input
	go func() {
		for {
			if input, err := cin.ReadBytes('\n'); err != nil {
				if !done {
					printErr(err, true)
				}
				return
			} else {
				err := ws.Write(context.Background(), webs.MessageText, input)
				if err != nil {
					printErr(err, true)
					return
				}
			}
		}
	}()
	// Get socket output
	go func() {
		for {
			if _, msg, err := ws.Read(context.Background()); err != nil {
				printErr(err, true)
				return
			} else {
				write(cout, "< "+string(msg), false)
			}
		}
	}()
	for {
	}
}
