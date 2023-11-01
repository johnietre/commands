package main

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
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
		Addr: addr,
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
	printErr(server.ListenAndServe(), true)
}
