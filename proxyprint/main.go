package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"sync"
	"syscall"
	"time"
)

type printStatus int

func (p printStatus) String() string {
	return fmt.Sprint(int(p))
}

func (p *printStatus) Set(s string) error {
	n, err := strconv.Atoi(s)
	*p = printStatus(n)
	if *p < noPrint || *p >= stopValPrint {
		return fmt.Errorf("invalid value: %d", *p)
	}
	return err
}

func (p printStatus) printFunc() PrintFunc {
	switch p {
	case noPrint:
		return noPrintFunc
	case doPrint:
		return doPrintFunc
	case bytesPrint:
		return bytesPrintFunc
	case lowerHexBytesPrint:
		return lowerHexBytesPrintFunc
	case upperHexBytesPrint:
		return upperHexBytesPrintFunc
	}
	log.Fatalln("unknown printStatus value:", p)
	return nil
}

type PrintFunc = func(b []byte, from, to string)

const (
	noPrint printStatus = iota
	doPrint
	bytesPrint
	lowerHexBytesPrint
	upperHexBytesPrint
	stopValPrint // Used for checking if values are in range
)

var (
	bufferLen                        uint64
	clientPrintFunc, serverPrintFunc PrintFunc = noPrintFunc, noPrintFunc

	connectAddr *net.TCPAddr

	printChan  chan string
	tunnelChan chan *net.TCPConn
	wg         sync.WaitGroup

	shouldTunnel bool
)

func main() {
	log.SetFlags(0)

	var clientPrint, serverPrint printStatus = 0, 0
	listenAddrStr := flag.String("listen", "", "Network address to listen on")
	connectAddrStr := flag.String("connect", "", "Network address to connect to")
	tunnelAddrStr := flag.String(
		"tunnel",
		"",
		"Network address of proxyprint session to tunnel to",
	)
	listenSrvrsAddrStr := flag.String(
		"listen-servers",
		"",
		"Network address to listen for tunneling servers on",
	)
	flag.Var(
		&clientPrint,
		"client-print",
		"Set the client data print (0 = off, 1 = as string, "+
			"2 = as bytes, 3 = as lower hex bytestring, 4 = as upper hex bytestring)",
	)
	flag.Var(
		&serverPrint,
		"server-print",
		"Set the server data print (0 = off, 1 = as string, "+
			"2 = as bytes, 3 = as lower hex bytestring, 4 = as upper hex bytestring)",
	)
	flag.Uint64Var(
		&bufferLen,
		"buffer",
		1<<15,
		"Size of the buffer to use to copy data",
	)
	logFilePath := flag.String(
		"log", "", "File to output logs to (blank is command line",
	)
	flag.Parse()

	if *logFilePath != "" {
		logFile, err := os.OpenFile(
			*logFilePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644,
		)
		if err != nil {
			log.Fatalf("error opening log file: %v", err)
		}
		log.SetOutput(logFile)
	}

	if *connectAddrStr != "" {
		addr, err := net.ResolveTCPAddr("tcp", *connectAddrStr)
		if err != nil {
			log.Fatal(err)
		}
		connectAddr = addr
	}

	if bufferLen == 0 {
		log.Fatal("must provide non-zero buffer size")
	}

	clientPrintFunc = clientPrint.printFunc()
	serverPrintFunc = serverPrint.printFunc()

	if *listenAddrStr != "" {
		if connectAddr == nil && *listenSrvrsAddrStr == "" {
			log.Fatal("must provide connect addr or listen-servers addr when proxying")
		}
		addr, err := net.ResolveTCPAddr("tcp", *listenAddrStr)
		if err != nil {
			log.Fatal(err)
		}
		wg.Add(1)
		go runListenClients(addr)
	}

	if *listenSrvrsAddrStr != "" {
		if *listenAddrStr == "" {
			log.Fatal("must provide listen addr with listen-servers addr")
		}
		addr, err := net.ResolveTCPAddr("tcp", *listenSrvrsAddrStr)
		if err != nil {
			log.Fatal(err)
		}
		shouldTunnel = true
		tunnelChan = make(chan *net.TCPConn, 10)
		wg.Add(1)
		go runListenServers(addr)
	}

	if *tunnelAddrStr != "" {
		if connectAddr == nil && *listenSrvrsAddrStr == "" {
			log.Fatal("must provide connect addr or listen-servers addr when tunneling")
		}
		addr, err := net.ResolveTCPAddr("tcp", *tunnelAddrStr)
		if err != nil {
			log.Fatal(err)
		}
		wg.Add(1)
		go runTunneler(addr)
	}

	if serverPrint+clientPrint != 0 {
		printChan = make(chan string, 50)
		go listenPrint()
	}

	wg.Wait()
}

var (
	tunnelBytes      = []byte{0xff, 0xff, 0xff, 0xff}
	clientReadyBytes = []byte{0xfe, 0xfe, 0xfe, 0xfe}
)

func runListenClients(addr *net.TCPAddr) {
	defer wg.Done()
	ln, err := net.ListenTCP("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}
	defer ln.Close()

	fmt.Printf("Listening for clients on %s...\n", addr)
	for {
		c, err := ln.AcceptTCP()
		if err != nil {
			log.Fatal(err)
		}
		go handle(c)
	}
}

func runTunneler(addr *net.TCPAddr) {
	defer wg.Done()
	fmt.Printf("Tunneling to %s...\n", addr)
	errCount := 0
	for {
		if errCount == 5 {
			time.Sleep(time.Second * 5)
		}
		conn, err := net.DialTCP("tcp", nil, addr)
		if err == nil {
			errCount = 0
			go handleTunnel(conn)
			continue
		}
		if errCount == 5 {
			continue
		}
		if !shouldIgnoreErr(err) {
			log.Printf("error tunneling: %v", err)
		}
		errCount++
		if errCount == 5 {
			log.Println(
				"5 tunnel connection errors encountered, " +
					"muting these errors and retrying every 5 seconds...",
			)
		}
	}
}

func handleTunnel(tunnel *net.TCPConn) {
	if _, err := tunnel.Write(tunnelBytes); err != nil {
		if !shouldIgnoreErr(err) {
			log.Printf("error sending tunneling bytes: %v", err)
		}
		tunnel.Close()
		return
	}
	var buf [4]byte
	if n, err := tunnel.Read(buf[:]); err != nil {
		if !shouldIgnoreErr(err) {
			log.Printf("error reading ready bytes: %v", err)
		}
	} else if n != 4 {
		log.Printf("expected 4 ready bytes, got %d bytes", n)
	} else if !bytes.Equal(buf[:], clientReadyBytes) {
		log.Printf("expected %v as ready bytes, got %v", clientReadyBytes, buf)
	} else {
		handle(tunnel)
		return
	}
	tunnel.Close()
}

func handle(client *net.TCPConn) {
	clientAddrStr := client.RemoteAddr().String()
	logErr := func(errFmt string, args ...interface{}) {
		log.Printf("["+clientAddrStr+"] "+errFmt, args...)
	}

	var server *net.TCPConn
	if shouldTunnel {
		server = <-tunnelChan
		if _, err := server.Write(clientReadyBytes); err != nil {
			logErr("error sending ready bytes: %v", err)
			server.Close()
			client.Close()
			return
		}
	} else {
		var err error
		server, err = net.DialTCP("tcp", nil, connectAddr)
		if err != nil {
			logErr("error connecting to server: %v", err)
			client.Close()
			return
		}
	}

	go pipe(client, server, clientPrintFunc)
	pipe(server, client, serverPrintFunc)
}

func runListenServers(addr *net.TCPAddr) {
	defer wg.Done()
	ln, err := net.ListenTCP("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}
	defer ln.Close()

	fmt.Printf("Listening for servers on %s...\n", addr)
	for {
		c, err := ln.AcceptTCP()
		if err != nil {
			log.Fatal(err)
		}
		//go handleServer(c)
		handleServer(c)
	}
}

func handleServer(server *net.TCPConn) {
	var buf [4]byte
	server.SetReadDeadline(time.Now().Add(time.Second))
	if n, err := server.Read(buf[:]); err != nil {
		if !shouldIgnoreErr(err) {
			log.Printf("error reading tunnel bytes: %v", err)
		}
	} else if n != 4 {
		log.Printf("expected 4 tunnel bytes, got %d bytes", n)
	} else if !bytes.Equal(buf[:], tunnelBytes) {
		log.Printf("expected %v as tunnel bytes, got %v", tunnelBytes, buf)
	} else {
		server.SetReadDeadline(time.Time{})
		tunnelChan <- server
		return
	}
	server.Close()
}

// Only prints and closes "from"
func pipe(from, to *net.TCPConn, pf PrintFunc) {
	defer from.Close()
	fromAddrStr := from.RemoteAddr().String()
	toAddrStr := to.RemoteAddr().String()
	buf := make([]byte, bufferLen)
	for {
		n, err := from.Read(buf[:])
		if err != nil {
			return
		}
		pf(buf[:n], fromAddrStr, toAddrStr)
		if _, err := to.Write(buf[:n]); err != nil {
			return
		}
	}
}

func listenPrint() {
	for s := range printChan {
		fmt.Print(s)
	}
}

func noPrintFunc([]byte, string, string) {}

func doPrintFunc(b []byte, from, to string) {
	printChan <- fmt.Sprintf(
		"%s => %s (%d bytes)\n"+
			"-------------------\n"+
			"%s\n"+
			"===================\n",
		from, to, len(b), string(b),
	)
}

func bytesPrintFunc(b []byte, from, to string) {
	printChan <- fmt.Sprintf(
		"%s => %s (%d bytes)\n"+
			"-------------------\n"+
			"%v\n"+
			"===================\n",
		from, to, len(b), string(b),
	)
}

func lowerHexBytesPrintFunc(b []byte, from, to string) {
	printChan <- fmt.Sprintf(
		"%s => %s (%d bytes)\n"+
			"-------------------\n"+
			"%x\n"+
			"===================\n",
		from, to, len(b), string(b),
	)
}

func upperHexBytesPrintFunc(b []byte, from, to string) {
	printChan <- fmt.Sprintf(
		"%s => %s (%d bytes)\n"+
			"-------------------\n"+
			"%X\n"+
			"===================\n",
		from, to, len(b), string(b),
	)
}

func shouldIgnoreErr(err error) bool {
	ret := err == nil ||
		errors.Is(err, io.EOF) ||
		errors.Is(err, syscall.EPIPE) ||
		errors.Is(err, syscall.ECONNRESET)
	if !ret {
		opErr := &net.OpError{}
		if errors.As(err, &opErr) {
			ret = opErr.Timeout()
		}
	}
	return ret
}
