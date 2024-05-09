package main

import "net"

func socketClient() {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		printErr(err, true)
		return
	}
	defer conn.Close()
	// Get user input
	go func() {
		for {
			if input, err := cin.ReadString('\n'); err != nil {
				if !done {
					printErr(err, true)
				}
				return
			} else {
				if _, err = conn.Write([]byte(input)); err != nil {
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
