package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"sync"

	webs "golang.org/x/net/websocket"
)

func wsServer(hub bool) {
	type hubMsg struct{ From, Msg string }
	var conns sync.Map
	var hubChan chan hubMsg

	hubHandler := func(ws *webs.Conn) {
		defer ws.Close()
		defer conns.Delete(ws.RemoteAddr().String())
		conns.Store(ws.RemoteAddr().String(), ws)
		var msg string
		for {
			if err := webs.Message.Receive(ws, &msg); err != nil {
				if err.Error() != "EOF" {
					printErr(err, false)
				}
				return
			}
			hubChan <- hubMsg{
				ws.RemoteAddr().String(),
				strings.ReplaceAll(msg, "\n", ""),
			}
		}
	}
	echoHandler := func(ws *webs.Conn) {
		defer ws.Close()
    log.Print(ws.Request().Header.Values("Sec-Websocket-Protocol"))
		var msg string
		for {
			if err := webs.Message.Receive(ws, &msg); err != nil {
				if err.Error() != "EOR" {
					printErr(err, false)
				}
				return
			} else if err = webs.Message.Send(ws, msg); err != nil {
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
				smsg := string(append(bmsg, '\n'))
				conns.Range(func(iAddr, iConn interface{}) bool {
					a, ws := iAddr.(string), iConn.(*webs.Conn)
					if a != msg.From {
						webs.Message.Send(ws, smsg)
					}
					return true
				})
			}
		}()
	}
	server := &http.Server{
		Addr: addr,
		Handler: func() *http.ServeMux {
			r := http.NewServeMux()
      wsSrvr := &webs.Server{
        Handshake: func(config *webs.Config, r *http.Request) error {
          config.Protocol = []string{}
          return nil
        },
      }
			if hub {
				wsSrvr.Handler = webs.Handler(hubHandler)
			} else {
				wsSrvr.Handler = webs.Handler(echoHandler)
			}
      r.Handle("/", wsSrvr)
			return r
		}(),
		ErrorLog: log.New(cerr, "Error: ", 0),
	}
	printErr(server.ListenAndServe(), true)
}
