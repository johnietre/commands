package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"sync"
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
	Listen        string      `json:"listen,omitempty"`
	Connect       string      `json:"connect,omitempty"`
	Tunnel        string      `json:"tunnel,omitempty"`
	ListenServers string      `json:"listenServers,omitempty"`
	ClientPrint   printStatus `json:"clientPrint,omitempty"`
	ServerPrint   printStatus `json:"serverPrint,omitempty"`
	Buffer        uint64      `json:"buffer,omitempty"`
	Log           string      `json:"log,omitempty"`
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
	if c.Log == "" {
		c.Log = other.Log
	}
}

var (
	clientPrintFunc, serverPrintFunc PrintFunc = noPrintFunc, noPrintFunc

	connectAddr *net.TCPAddr

	printChan  chan string
	tunnelChan chan *net.TCPConn
	wg         sync.WaitGroup

	shouldTunnel bool

	config = Config{}
)

func main() {
	log.SetFlags(0)

	cmd := &cobra.Command{
		Use:                   "proxyprint",
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
				Listen:        "IP:PORT",
				Connect:       "IP:PORT",
				Tunnel:        "IP:PORT",
				ListenServers: "IP:PORT",
				ClientPrint:   -1,
				ServerPrint:   -1,
				Buffer:        1 << 15,
				Log:           "PATH",
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
	flag.StringVar(
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
		addr, err := net.ResolveTCPAddr("tcp", config.ListenServers)
		if err != nil {
			log.Fatal(err)
		}
		shouldTunnel = true
		tunnelChan = make(chan *net.TCPConn, 10)
		wg.Add(1)
		startedServer = true
		go runListenServers(addr)
	}

	if config.Tunnel != "" {
		if connectAddr == nil && config.ListenServers == "" {
			log.Fatal("must provide connect addr or listen-servers addr when tunneling")
		}
		addr, err := net.ResolveTCPAddr("tcp", config.Tunnel)
		if err != nil {
			log.Fatal(err)
		}
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
