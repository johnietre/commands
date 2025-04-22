// TODO: allow printing to files
// TODO: reuseaddr option
// TODO: clean shutdown with signal handling and connection tracking
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
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	utils "github.com/johnietre/utils/go"
	"github.com/spf13/cobra"
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

func (p printStatus) Type() string {
	return "PrintStatus"
}

func (p printStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(int(p))
}

func (p *printStatus) UnmarshalJSON(b []byte) error {
	i := 0
	if err := json.Unmarshal(b, &i); err != nil {
		return err
	}
	pi := printStatus(i)
	if pi < noPrint || pi >= stopValPrint {
		return fmt.Errorf("invalid value: %d", pi)
	}
	*p = pi
	return nil
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

type Config struct {
	Listen              string      `json:"listen,omitempty"`
	Connect             string      `json:"connect,omitempty"`
	Tunnel              string      `json:"tunnel,omitempty"`
	ListenServers       string      `json:"listenServers,omitempty"`
	ClientPrint         printStatus `json:"clientPrint,omitempty"`
	ServerPrint         printStatus `json:"serverPrint,omitempty"`
	Buffer              uint64      `json:"buffer,omitempty"`
	MaxWaitingTunnels   uint        `json:"maxOpenTunnels,omitempty"`
	MaxAcceptedServers  uint        `json:"maxAcceptedServers,omitempty"`
	PwdEnvName          string      `json:"pwdEnvName,omitempty"`
	RequirePwdEnvExists bool        `json:"requirePwdEnvExists,omitempty"`
	Log                 string      `json:"log,omitempty"`
}

func (c *Config) FillEmptyFrom(other *Config) {
	if c.Listen == "" {
		c.Listen = other.Listen
	}
	if c.Connect == "" {
		c.Connect = other.Connect
	}
	if c.Tunnel == "" {
		c.Tunnel = other.Tunnel
	}
	if c.ListenServers == "" {
		c.ListenServers = other.ListenServers
	}
	// NOTE: do something else?
	if c.ClientPrint == noPrint {
		c.ClientPrint = other.ClientPrint
	}
	if c.ServerPrint == noPrint {
		c.ServerPrint = other.ServerPrint
	}
	if c.Buffer == 0 {
		c.Buffer = other.Buffer
	}
	if c.MaxWaitingTunnels == 0 {
		c.MaxWaitingTunnels = other.MaxWaitingTunnels
	}
	if c.MaxAcceptedServers == 0 {
		c.MaxAcceptedServers = other.MaxAcceptedServers
	}
	// NOTE: do something else?
	if c.PwdEnvName == "" {
		c.PwdEnvName = other.PwdEnvName
	}
	if c.RequirePwdEnvExists == false {
		c.RequirePwdEnvExists = other.RequirePwdEnvExists
	}
	if c.Log == "" {
		c.Log = other.Log
	}
}

var (
	clientPrintFunc, serverPrintFunc PrintFunc = noPrintFunc, noPrintFunc

	connectAddr *net.TCPAddr

	printChan   chan string
	tunnelChan  chan *BufferedConn
	waitingChan chan utils.Unit
	wg          sync.WaitGroup

	shouldTunnel bool

	config   = Config{}
	password []byte
)

func main() {
	log.SetFlags(log.LstdFlags)

	cmd := &cobra.Command{
		Use:                   "proxyprint",
		Short:                 "Run a proxy which can print out communications in a variety of ways",
		Long:                  "Run a proxy which can print out communications in a variety of ways. A password can be specified using the PROXYPRINT_PWD environment variable (unless it is set otherwise).",
		Run:                   run,
		DisableFlagsInUseLine: true,
	}
	cmd.AddCommand(&cobra.Command{
		Use:   "generate",
		Short: "Generate a blank config file",
		Long:  "Generate a blank config file to populate. NOTE: the file is generated with mostly invalid values which must be changed or deleted.",
		Run: func(_ *cobra.Command, args []string) {
			path := "proxyprint.conf.json"
			if len(args) > 0 {
				info, err := os.Stat(args[0])
				if err != nil && !os.IsNotExist(err) {
					log.Fatal("error checking output path: ", err)
				} else if err == nil && info.IsDir() {
					path = filepath.Join(args[0], path)
				} else {
					path = args[0]
				}
			}
			f, err := os.Create(path)
			if err != nil {
				log.Fatal("error creating config file: ", err)
			}
			defer f.Close()
			enc := json.NewEncoder(f)
			enc.SetIndent("", "  ")
			config := Config{
				Listen:             "IP:PORT",
				Connect:            "IP:PORT",
				Tunnel:             "IP:PORT",
				ListenServers:      "IP:PORT",
				ClientPrint:        -1,
				ServerPrint:        -1,
				Buffer:             1 << 15,
				MaxAcceptedServers: 10,
				PwdEnvName:         "PROXYPRINT_PWD",
				Log:                "PATH",
			}
			if err := enc.Encode(config); err != nil {
				log.Fatal("error writing config file: ", err)
			}
		},
	})

	flags := cmd.Flags()

	flags.StringVar(&config.Listen, "listen", "", "Network address to listen on")
	flags.StringVar(&config.Connect, "connect", "", "Network address to connect to")
	flags.StringVar(
		&config.Tunnel,
		"tunnel",
		"",
		"Network address of proxyprint session to tunnel to",
	)
	flags.StringVar(
		&config.ListenServers,
		"listen-servers",
		"",
		"Network address to listen for tunneling servers on",
	)
	flags.Var(
		&config.ClientPrint,
		"client-print",
		"Set the client data print (0 = off*, 1 = as string, "+
			"2 = as bytes, 3 = as lower hex bytestring, 4 = as upper hex bytestring)",
	)
	flags.Var(
		&config.ServerPrint,
		"server-print",
		"Set the server data print (0 = off*, 1 = as string, "+
			"2 = as bytes, 3 = as lower hex bytestring, 4 = as upper hex bytestring)",
	)
	flags.Uint64Var(
		&config.Buffer,
		"buffer",
		1<<15,
		"Size of the buffer to use to copy data",
	)
	flags.UintVar(
		&config.MaxWaitingTunnels,
		"max-waiting-tunnels",
		10,
		"Set the number of tunnels (to servers) that can be waiting for servers at once."+
			"Used with --tunnel flag.",
	)
	flags.UintVar(
		&config.MaxAcceptedServers,
		"max-accepted-servers",
		10,
		"Set the number of tunneling servers that can be accepted/handled at once"+
			"Used with --listen-servers flag.",
	)
	flags.StringVar(
		&config.PwdEnvName,
		"pwd-env-name",
		"PROXYPRINT_PWD",
		"The environment variable for reading the tunneling password. "+
			"If the name of the variable starts with the string 'file:', the value "+
			"of the variable is treated as a file path and the pointed-to file is "+
			"read and its content used as the password. "+
			"An empty string means to not read any password "+
			"(will still use empty password). "+
			"An empty environment variable value means no (an empty) password.",
	)
	flags.BoolVar(
		&config.RequirePwdEnvExists,
		"require-pwd-env-exists",
		false,
		"Require environment variable value of pwd-env-name flag exists and, if "+
			"not, throw a fatal error. If false, only warn.",
	)
	flags.StringVar(
		&config.Log,
		"log", "", "File to output logs to (blank is command line)",
	)
	flags.String("cfg", "", "Path to config file")

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
		var newCfg Config
		if err := json.NewDecoder(f).Decode(&newCfg); err != nil {
			log.Fatal("error parsing config file: ", err)
		}
		f.Close()
		config.FillEmptyFrom(&newCfg)
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
			log.Fatal(err)
		}
		connectAddr = addr
	}

	if config.Buffer == 0 {
		log.Fatal("must provide non-zero buffer size")
	}

	clientPrintFunc = config.ClientPrint.printFunc()
	serverPrintFunc = config.ServerPrint.printFunc()

	startedServer := false

	if config.Listen != "" {
		if connectAddr == nil && config.ListenServers == "" {
			log.Fatal("must provide connect addr or listen-servers addr when proxying")
		}
		addr, err := net.ResolveTCPAddr("tcp", config.Listen)
		if err != nil {
			log.Fatal(err)
		}
		wg.Add(1)
		startedServer = true
		go runListenClients(addr)
	}

	if config.ListenServers != "" {
		if config.Listen == "" {
			log.Fatal("must provide listen addr with listen-servers addr")
		}
		getPassword()
		addr, err := net.ResolveTCPAddr("tcp", config.ListenServers)
		if err != nil {
			log.Fatal(err)
		}
		if config.MaxAcceptedServers == 0 {
			config.MaxAcceptedServers = 10
		}
		shouldTunnel = true
		tunnelChan = make(chan *BufferedConn, config.MaxAcceptedServers)
		wg.Add(1)
		startedServer = true
		go runListenServers(addr)
	}

	if config.Tunnel != "" {
		getPassword()
		if connectAddr == nil && config.ListenServers == "" {
			log.Fatal("must provide connect addr or listen-servers addr when tunneling")
		}
		addr, err := net.ResolveTCPAddr("tcp", config.Tunnel)
		if err != nil {
			log.Fatal(err)
		}
		if config.MaxWaitingTunnels == 0 {
			config.MaxWaitingTunnels = 10
		}
		waitingChan = make(chan utils.Unit, config.MaxWaitingTunnels)
		wg.Add(1)
		startedServer = true
		go runTunneler(addr)
	}

	if config.ServerPrint+config.ClientPrint != 0 {
		printChan = make(chan string, 50)
		go listenPrint()
	}

	if !startedServer {
		cmd.Usage()
	}
	wg.Wait()
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
	defer wg.Done()
	ln, err := net.ListenTCP("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}
	defer ln.Close()

	fmt.Printf("Listening for clients on %s...\n", addr)
	// NOTE: i for testing/logging purposes
	for i := 1; true; {
		c, err := ln.AcceptTCP()
		if err != nil {
			log.Fatal(err)
		}
		go handle(NewBufferedConn(c), i)
	}
}

func runTunneler(addr *net.TCPAddr) {
	const retryTime = 2
	const maxErrCount = 5

	defer wg.Done()
	fmt.Printf("Tunneling to %s...\n", addr)
	errCount := 0
	// NOTE: i for testing/logging purposes
	for i := -1; true; {
		if errCount == maxErrCount {
			// TODO: retry timer flag
			time.Sleep(time.Second * retryTime)
		}
		waitingChan <- utils.Unit{}
		conn, err := net.DialTCP("tcp", nil, addr)
		if err == nil {
			if errCount >= maxErrCount {
				log.Print("tunneling reconnected")
			}
			errCount = 0
			go handleTunnel(NewBufferedConn(conn), i)
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
func handleTunnel(tunnel *BufferedConn, num int) {
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
		handle(tunnel, num)
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
				break
			}
			if server == nil {
				break
			} else if checkTunnelReadiness(server, logErr) {
				break
			}
		}
		if !timer.Stop() {
			<-timer.C
		}
		if server == nil {
			client.Close()
			return
		}
	} else {
		var err error
		srvr, err := net.DialTCP("tcp", nil, connectAddr)
		if err != nil {
			logErr("error connecting to server: %v", err)
			client.Close()
			return
		}
		server = NewBufferedConn(srvr)
	}

	go pipe(client, server, clientPrintFunc)
	pipe(server, client, serverPrintFunc)
}

// Returns false if not ready (the passed conn will be closed)
func checkTunnelReadiness(
	server *BufferedConn, logErr func(string, ...any),
) (ready bool) {
	// Send ready bytes to tunnel
	if _, err := utils.WriteAll(server, clientReadyBytes); err != nil {
		logErr("error sending client ready bytes: %v", err)
		server.Close()
		server = nil
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
		handleServer(NewBufferedConn(c))
	}
}

func handleServer(server *BufferedConn) {
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
	tunnelChan <- server
}

// Only prints and closes "from"
func pipe(from, to *BufferedConn, pf PrintFunc) {
	defer from.Close()
	fromAddrStr := from.RemoteAddr().String()
	toAddrStr := to.RemoteAddr().String()
	buf := make([]byte, config.Buffer)
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
		from, to, len(b), b,
	)
}

func bytesPrintFunc(b []byte, from, to string) {
	printChan <- fmt.Sprintf(
		"%s => %s (%d bytes)\n"+
			"-------------------\n"+
			"%v\n"+
			"===================\n",
		from, to, len(b), b,
	)
}

func lowerHexBytesPrintFunc(b []byte, from, to string) {
	printChan <- fmt.Sprintf(
		"%s => %s (%d bytes)\n"+
			"-------------------\n"+
			"%x\n"+
			"===================\n",
		from, to, len(b), b,
	)
}

func upperHexBytesPrintFunc(b []byte, from, to string) {
	printChan <- fmt.Sprintf(
		"%s => %s (%d bytes)\n"+
			"-------------------\n"+
			"%X\n"+
			"===================\n",
		from, to, len(b), b,
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
