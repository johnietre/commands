package main

import (
	//"encoding/json"
	"fmt"
	"net"
	"os"
	//"strings"
	//"sync"
)

func udpServer(hub bool) {
	laddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		printErr(err, true)
		return
	}
	conn, err := net.ListenUDP("udp", laddr)
	if err != nil {
		printErr(err, true)
		return
	}
	defer conn.Close()
	if testOk {
		conn.Close()
		os.Exit(0)
	}
	printErr(fmt.Errorf("Not implemented yet"), true)
	return

	/*
			type hubMsg struct{ From, Msg string }
			var conns sync.Map
			var hubChan chan hubMsg
			if hub {
				hubChan = make(chan hubMsg, 5)
				go func() {
					for msg := range hubChan {
						bmsg, _ := json.Marshal(msg)
						bmsg = append(bmsg, '\n')
						conns.Range(func(iAddr, iConn interface{}) bool {
							a, c := iAddr.(string), iConn.(net.Conn)
							if a != msg.From {
								c.Write(bmsg)
							}
							return true
						})
					}
				}()
			}
		  for !done {
		    var buf [maxBufferSize]byte
		    //
		  }
			for !done {
				conn, err := ln.Accept()
				if err != nil {
					continue
				}
				if !hub {
					go func(c net.Conn) {
						defer c.Close()
						var buf [maxBufferSize]byte
						for {
							if l, err := conn.Read(buf[:]); err != nil {
								if err.Error() != "EOF" {
									printErr(err, false)
								}
								return
							} else {
								if _, err := conn.Write(buf[:l]); err != nil {
									printErr(err, false)
									return
								}
							}
						}
					}(conn)
				} else {
					go func(c net.Conn) {
						defer c.Close()
						defer conns.Delete(c.RemoteAddr().String())
						conns.Store(c.RemoteAddr().String(), c)
						var buf [maxBufferSize]byte
						for {
							if l, err := conn.Read(buf[:]); err != nil {
								if err.Error() != "EOF" {
									printErr(err, false)
								}
								return
							} else {
								hubChan <- hubMsg{
									c.RemoteAddr().String(),
									strings.ReplaceAll(string(buf[:l]), "\n", ""),
								}
							}
						}
					}(conn)
				}
			}
	*/
}

func udpClient() {
	conn, err := net.Dial("udp", addr)
	if err != nil {
		printErr(err, true)
		return
	}
	defer conn.Close()
	if testOk {
		conn.Close()
		os.Exit(0)
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
				if _, err = conn.Write(input); err != nil {
					printErr(err, true)
					return
				}
			}
		}
	}()
	// Get socket output
	go func() {
		var buf [maxBufferSize]byte
		for {
			if l, err := conn.Read(buf[:]); err != nil {
				printErr(err, true)
				return
			} else {
				write(cout, "< "+string(buf[:l]), false)
			}
		}
	}()
	<-make(chan struct{})
}
