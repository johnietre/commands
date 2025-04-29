// TODO: reuseaddr option
// TODO: monitor server optional password (and allow HTTPS)
package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/johnietre/go-jmux"
	utils "github.com/johnietre/utils/go"
	"github.com/spf13/cobra"
)

var (
	clientPrintFunc, serverPrintFunc PrintFunc = noPrintFunc, noPrintFunc
	clientPrintFile                            = os.Stdout
	serverPrintFile                            = os.Stdout

	connectAddr                    *net.TCPAddr
	clientListener, serverListener atomic.Pointer[net.TCPListener]

	printChan   chan PrintData
	tunnelChan  chan *BufferedConn
	waitingChan chan utils.Unit

	monitor Monitor

	shouldTunnel bool

	config   = Config{}
	password []byte
)

func main() {
	log.SetFlags(log.LstdFlags)

	cmd := makeCmd()
	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func run(cmd *cobra.Command, _ []string) {
	flags := cmd.Flags()
	cfgPath := utils.Must(flags.GetString("cfg"))
	if cfgPath != "" {
		f, err := os.Open(cfgPath)
		if err != nil {
			log.Fatal("error opening config file: ", err)
		}
		var newCfg ConfigPtrs
		if err := json.NewDecoder(f).Decode(&newCfg); err != nil {
			log.Fatal("error parsing config file: ", err)
		}
		f.Close()
		//config.FillEmptyFrom(&newCfg)
		config.PopulateCheckFlags(&newCfg, flags)
	}

	if config.Log != "" {
		logFile, err := utils.OpenAppend(config.Log)
		if err != nil {
			log.Fatalf("error opening log file: %v", err)
		}
		log.SetOutput(logFile)
	}

	if config.Connect != "" {
		addr, err := net.ResolveTCPAddr("tcp", config.Connect)
		if err != nil {
			log.Fatal("error resolving connect TCP address: ", err)
		}
		fmt.Printf("Connecting to servers at %s...\n", addr)
		connectAddr = addr
	}

	if config.Buffer == 0 {
		log.Fatal("must provide non-zero buffer size")
	}

	clientPrintFunc = config.ClientPrint.printFunc()
	serverPrintFunc = config.ServerPrint.printFunc()

	var err error
	if config.ClientPrintFile != "" && config.ClientPrint != noPrint {
		clientPrintFile, err = utils.OpenAppend(config.ClientPrintFile)
		if err != nil {
			log.Fatal("error opening client print file: ", err)
		}
	}
	if config.ServerPrintFile != "" && config.ServerPrint != noPrint {
		serverPrintFile, err = utils.OpenAppend(config.ServerPrintFile)
		if err != nil {
			log.Fatal("error opening server print file: ", err)
		}
	}

	startedServer := false

	if config.Listen != "" {
		if connectAddr == nil && config.ListenServers == "" {
			log.Fatal("must provide connect addr or listen-servers addr when proxying")
		}
		addr, err := net.ResolveTCPAddr("tcp", config.Listen)
		if err != nil {
			log.Fatal("error resolving listen TCP address: ", err)
		}
		startedServer = true
		monitor.wg.Add(1)
		go func() {
			runListenClients(addr)
			monitor.wg.Done()
		}()
	}

	if config.ListenServers != "" {
		if config.Listen == "" {
			log.Fatal("must provide listen addr with listen-servers addr")
		}
		getPassword()
		addr, err := net.ResolveTCPAddr("tcp", config.ListenServers)
		if err != nil {
			log.Fatal("error resolving listening (servers) TCP address: ", err)
		}
		if config.MaxAcceptedServers == 0 {
			config.MaxAcceptedServers = 10
		}
		shouldTunnel = true
		tunnelChan = make(chan *BufferedConn, config.MaxAcceptedServers)
		startedServer = true
		monitor.wg.Add(1)
		go func() {
			runListenServers(addr)
			monitor.wg.Done()
		}()
	}

	if config.Tunnel != "" {
		getPassword()
		if connectAddr == nil && config.ListenServers == "" {
			log.Fatal("must provide connect addr or listen-servers addr when tunneling")
		}
		addr, err := net.ResolveTCPAddr("tcp", config.Tunnel)
		if err != nil {
			log.Fatal("error resolving tunnel TCP address: ", err)
		}
		if config.MaxWaitingTunnels == 0 {
			config.MaxWaitingTunnels = 10
		}
		waitingChan = make(chan utils.Unit, config.MaxWaitingTunnels)
		startedServer = true
		monitor.wg.Add(1)
		go func() {
			runTunneler(addr)
			monitor.wg.Done()
		}()
	}

	if config.ServerPrint+config.ClientPrint != 0 {
		printChan = make(chan PrintData, 50)
		go listenPrint()
	}

	if !startedServer {
		cmd.Usage()
	}

	// TODO: start monitor server
	if config.MonitorServer != "" {
		monitor.Config = config
		go runMonitorServer()
	}

	sigCh := make(chan os.Signal, 5)
	signal.Notify(sigCh, os.Interrupt)
	go func() {
		for range sigCh {
			shutdown(true)
		}
	}()

	monitor.Wait()
	log.Print("finished running")
}

var (
	tunnelBytes      = []byte{0xff, 0xff, 0xff, 0xff}
	versionBytes     = []byte{0, 0, 0, 1}
	clientReadyBytes = []byte{0xfe, 0xfe, 0xfe, 0xfe}
	serverReadyBytes = []byte{0xfd, 0xfd, 0xfd, 0xfd}
	okBytes          = []byte{0x00, 0x00, 0x00, 0x01}
	errorBytes       = []byte{0x00, 0x00, 0x00, 0x02}
)

func runListenClients(addr *net.TCPAddr) {
	ln, err := net.ListenTCP("tcp", addr)
	if err != nil {
		log.Fatal("error listening: ", err)
	}
	clientListener.Store(ln)
	defer ln.Close()

	fmt.Printf("Listening for clients on %s...\n", addr)
	// NOTE: i for testing/logging purposes
	for i := 1; true; {
		c, err := ln.AcceptTCP()
		if err != nil {
			if monitor.ShuttingDown.Load() {
				break
			}
			log.Fatal("error accepting: ", err)
		}
		go func() {
			monitor.AddClient()
			defer monitor.RemoveClient()
			handle(NewBufferedConn(c), i)
		}()
	}
}

func runTunneler(addr *net.TCPAddr) {
	const retryTime = 2
	const maxErrCount = 5

	fmt.Printf("Tunneling to %s...\n", addr)
	errCount := 0
	// NOTE: i for testing/logging purposes
	for i := -1; !monitor.ShuttingDown.Load(); {
		if errCount == maxErrCount {
			// TODO: retry timer flag
			time.Sleep(time.Second * retryTime)
		}
		waitingChan <- utils.Unit{}
		conn, err := net.DialTCP("tcp", nil, addr)
		monitor.AddTotalTunnelConnectAttempts()
		if err == nil {
			if errCount >= maxErrCount {
				log.Print("tunneling reconnected")
			}
			errCount = 0
			monitor.AddTotalTunnelsConnected()
			go connectTunnel(NewBufferedConn(conn), i)
			continue
		}
		<-waitingChan
		if errCount == maxErrCount {
			continue
		}
		if !shouldIgnoreErr(err) {
			log.Printf("error tunneling: %v", err)
		}
		errCount++
		if errCount == maxErrCount {
			log.Printf(
				"%d tunnel connection errors encountered, "+
					"muting these errors and retrying every %d seconds...",
				maxErrCount, retryTime,
			)
		}
	}
}

// NOTE: num for logging/testing purposes
func connectTunnel(tunnel *BufferedConn, num int) {
	// TODO: timeout?
	var buf [4]byte
	// Send tunnel header
	if _, err := tunnel.Write(tunnelBytes); err != nil {
		if !shouldIgnoreErr(err) {
			log.Printf("error sending tunneling bytes: %v", err)
		}
		tunnel.Close()
		return
	}

	// Get and check version
	if _, err := io.ReadFull(tunnel, buf[:]); err != nil {
		if !shouldIgnoreErr(err) {
			log.Printf("error reading version: %v", err)
		}
		tunnel.Close()
		return
	} else if !bytes.Equal(buf[:], versionBytes) {
		log.Fatalf("can only handle version up to %v, got %v", versionBytes, buf)
		tunnel.Close()
		return
	}

	// Send password
	lenBytes := binary.BigEndian.AppendUint64(nil, uint64(len(password)))
	if _, err := utils.WriteAll(tunnel, lenBytes); err != nil {
		if !shouldIgnoreErr(err) {
			log.Printf("error sending password length bytes: %v", err)
		}
		tunnel.Close()
		return
	} else if _, err := utils.WriteAll(tunnel, password); err != nil {
		if !shouldIgnoreErr(err) {
			log.Printf("error sending password bytes: %v", err)
		}
		tunnel.Close()
		return
	}

	// Get response
	if _, err := io.ReadFull(tunnel, buf[:]); err != nil {
		// TODO: do a read full?
		if !shouldIgnoreErr(err) {
			log.Printf("error reading client password response bytes: %v", err)
		}
	} else if !bytes.Equal(buf[:], okBytes) {
		log.Fatal("invalid password")
	}

	// Get ready bytes from server
	if _, err := io.ReadFull(tunnel, buf[:]); err != nil {
		// TODO: do a read full?
		if !shouldIgnoreErr(err) {
			log.Printf("error reading client ready bytes: %v", err)
		}
	} else if !bytes.Equal(buf[:], clientReadyBytes) {
		log.Printf(
			"expected %v as client ready bytes, got %v",
			clientReadyBytes, buf,
		)
	}

	// Send ready bytes to server
	if _, err := utils.WriteAll(tunnel, serverReadyBytes); err != nil {
		if !shouldIgnoreErr(err) {
			log.Printf("error sending server ready bytes: %v", err)
		}
	} else {
		<-waitingChan
		func() {
			monitor.AddTunnel()
			defer monitor.RemoveTunnel()
			handle(tunnel, num)
		}()
		return
	}
	<-waitingChan
	tunnel.Close()
}

func handle(client *BufferedConn, num int) {
	clientAddrStr := client.RemoteAddr().String()
	logErr := func(errFmt string, args ...interface{}) {
		log.Printf("["+clientAddrStr+"] "+errFmt, args...)
	}

	var server *BufferedConn
	if shouldTunnel {
		if num < 0 {
		}
		// TODO: make flag
		timer := time.NewTimer(time.Second * 10)
		for {
			select {
			case server = <-tunnelChan:
			case <-timer.C:
				// NOTE: log something?
				monitor.AddTunnelWaitTimeouts()
			}
			if server == nil {
				// TODO: log something?
				break
			} else if checkTunnelReadiness(server, logErr) {
				server.Close()
				monitor.RemoveTunneled()
				monitor.AddTunneledFailedReady()
				break
			}
			server = nil
		}
		if !timer.Stop() {
			<-timer.C
		}
		if server == nil {
			client.Close()
			return
		}
		defer monitor.RemoveTunneled()
	} else {
		var err error
		srvr, err := net.DialTCP("tcp", nil, connectAddr)
		if err != nil {
			logErr("error connecting to server: %v", err)
			monitor.AddTotalConnectServerFails(err)
			client.Close()
			return
		}
		server = NewBufferedConn(srvr)
	}

	go pipe(client, server, clientPrintFunc, false)
	pipe(server, client, serverPrintFunc, true)
}

// Returns false if not ready (the passed conn will not be closed)
func checkTunnelReadiness(
	server *BufferedConn, logErr func(string, ...any),
) (ready bool) {
	// Send ready bytes to tunnel
	if _, err := utils.WriteAll(server, clientReadyBytes); err != nil {
		logErr("error sending client ready bytes: %v", err)
		return
	}

	// Get ready bytes from tunnel
	var buf [4]byte
	if _, err := io.ReadFull(server, buf[:]); err != nil {
		// TODO: do a read full?
		if !shouldIgnoreErr(err) {
			log.Printf("error reading server ready bytes: %v", err)
		}
		return
	} else if !bytes.Equal(buf[:], serverReadyBytes) {
		log.Printf(
			"expected %v as server ready bytes, got %v",
			serverReadyBytes, buf,
		)
	} else {
		ready = true
	}
	return
}

func runListenServers(addr *net.TCPAddr) {
	ln, err := net.ListenTCP("tcp", addr)
	if err != nil {
		log.Fatal("error listening (tunnels): ", err)
	}
	serverListener.Store(ln)
	defer ln.Close()

	fmt.Printf("Listening for servers on %s...\n", addr)
	for {
		c, err := ln.AcceptTCP()
		if err != nil {
			if monitor.ShuttingDown.Load() {
				break
			}
			log.Fatal("error accepting (tunnels): ", err)
		}
		//go handleServer(c)
		monitor.AddTotalAcceptedServers()
		handleServer(NewBufferedConn(c))
	}
}

func handleServer(server *BufferedConn) {
	// NOTE: monitor for specific errors/failures?
	shouldClose := utils.NewT(true)
	defer utils.DeferClose(shouldClose, server)

	var fullBuf [8]byte
	buf := fullBuf[:4]
	// TODO: timeout flag
	server.SetDeadline(time.Now().Add(time.Second * 2))

	// Get tunnel header
	if _, err := io.ReadFull(server, buf[:]); err != nil {
		if !shouldIgnoreErr(err) {
			log.Printf("error reading tunnel bytes: %v", err)
		}
		return
	} else if !bytes.Equal(buf[:], tunnelBytes) {
		log.Printf("expected %v as tunnel bytes, got %v", tunnelBytes, buf)
		return
	}

	// Send version
	if _, err := utils.WriteAll(server, versionBytes); err != nil {
		if !shouldIgnoreErr(err) {
			log.Printf("error sending version bytes: %v", err)
		}
		return
	}

	// Get and check password
	if _, err := io.ReadFull(server, fullBuf[:]); err != nil {
		if !shouldIgnoreErr(err) {
			log.Printf("error reading password length bytes: %v", err)
		}
		return
	}
	pwdLen := binary.BigEndian.Uint64(fullBuf[:])
	if pwdLen != uint64(len(password)) {
		server.Write(errorBytes)
		return
	}
	if ok, err := readCheckPassword(server); !ok {
		if !shouldIgnoreErr(err) {
			log.Printf("error reading password bytes: %v", err)
		}
		server.Write(errorBytes)
		return
	}

	// Send response
	if _, err := utils.WriteAll(server, okBytes); err != nil {
		if !shouldIgnoreErr(err) {
			log.Printf("error sending response bytes: %v", err)
		}
		return
	}

	*shouldClose = false
	server.SetDeadline(time.Time{})
	monitor.AddTunneled()
	tunnelChan <- server
}

// Only prints and closes "from"
func pipe(from, to *BufferedConn, pf PrintFunc, fromServer bool) {
	defer from.Close()
	fromAddrStr := from.RemoteAddr().String()
	toAddrStr := to.RemoteAddr().String()
	buf := make([]byte, config.Buffer)
	for {
		n, err := from.Read(buf[:])
		if err != nil {
			return
		}
		pf(buf[:n], fromAddrStr, toAddrStr, fromServer)
		if _, err := to.Write(buf[:n]); err != nil {
			return
		}
	}
}

const (
	// Maximum number of password bufs to keep in the pool. Any more are
	// discarded.
	maxPwdBufs = 1000
)

var (
	pwdCheckBufPool = utils.AlwaysNewPool(func() []byte {
		n := 1024
		if l := len(password); l < n {
			n = l
		}
		return make([]byte, n)
	})
	pwdBufsOut atomic.Int64
)

func readCheckPassword(r io.Reader) (bool, error) {
	if len(password) == 0 {
		return true, nil
	}
	buf, pwd := pwdCheckBufPool.Get(), password[:]
	bufNum := pwdBufsOut.Add(1)
	defer func() {
		if bufNum <= maxPwdBufs {
			pwdCheckBufPool.Put(buf)
		}
		pwdBufsOut.Add(-1)
	}()
	for len(pwd) != 0 {
		l := len(pwd)
		if l > len(buf) {
			l = len(buf)
		}
		buf := buf[:l]
		n, err := r.Read(buf)
		if err != nil {
			return false, err
		} else if !bytes.Equal(buf[:n], pwd[:n]) {
			return false, nil
		}
		pwd = pwd[n:]
	}
	return true, nil
}

func runMonitorServer() {
	srvr := &http.Server{
		Addr: config.MonitorServer,
		Handler: (func() http.Handler {
			r := jmux.NewRouter()
			r.GetFunc("/stats", func(c *jmux.Context) {
				c.RespHeader().Set("Content-Type", "application/json")
				c.WriteJSON(&monitor)
			})
			return r
		})(),
		// TODO: set error log?
	}
	fmt.Printf("Running monitoring server on %s...\n", srvr.Addr)
	if err := srvr.ListenAndServe(); err != nil {
		log.Fatal("error running monitor server: ", err)
	}
}

func listenPrint() {
	var err error
	for data := range printChan {
		if data.server {
			_, err = fmt.Fprint(serverPrintFile, data.msg)
			if err != nil {
				log.Fatal("error writing to server file: ", err)
			}
		} else {
			_, err = fmt.Fprint(clientPrintFile, data.msg)
			if err != nil {
				log.Fatal("error writing to client file: ", err)
			}
		}
	}
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

func getPassword() {
	envName, readFile := config.PwdEnvName, false
	if envName == "" {
		return
	}
	if strings.HasPrefix(envName, "file:") {
		readFile = true
		envName = envName[5:]
	}
	val, ok := os.LookupEnv(envName)
	if !ok {
		logFunc := log.Printf
		if config.RequirePwdEnvExists {
			logFunc = log.Fatalf
		}
		logFunc("password environment variable %s doesn't exist", envName)
		readFile = false
	}
	if !readFile {
		password = []byte(val)
		return
	}
	content, err := os.ReadFile(val)
	if err != nil {
		log.Fatalf(
			"error reading password from %s (gotten from ENVVAR %s): %v",
			val, envName, err,
		)
	}
	password = content
}

func shutdown(force bool) {
	if monitor.ShuttingDown.Swap(true) {
		if force {
			log.Print("forcing shutdown...")
			os.Exit(0)
		}
		return
	}
	log.Print("shutting down...")
	if ln := clientListener.Load(); ln != nil {
		ln.Close()
	}
	if ln := serverListener.Load(); ln != nil {
		ln.Close()
	}
}

type BufferedConn struct {
	net.Conn
	buf bytes.Buffer
	mtx sync.RWMutex
	/*
	  peeker bufio.Reader
	  mtx sync.Mutex
	*/
	peeker bufio.Reader
}

func NewBufferedConn(c net.Conn) *BufferedConn {
	return &BufferedConn{
		Conn: c,
	}
}

func (bc *BufferedConn) Peek(p []byte) (int, error) {
	l := len(p)
	bc.mtx.Lock()
	defer bc.mtx.Unlock()
	n := copy(p, bc.buf.Bytes())
	if n == l {
		return l, nil
	}
	nn, err := bc.Conn.Read(p[n:])
	n += nn
	bc.buf.Write(p[n:])
	return n, err
}

func (bc *BufferedConn) Read(p []byte) (n int, err error) {
	l := len(p)
	bc.mtx.Lock()
	defer bc.mtx.Unlock()
	n = copy(p, bc.buf.Next(l))
	if n != l {
		l, err = bc.Conn.Read(p[n:])
		n += l
	}
	return
}
