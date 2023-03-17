package main

import (
  "flag"
  "fmt"
  "log"
  "net"
  "strconv"
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

type PrintFunc = func([]byte, string, string)

const (
  noPrint printStatus = iota
  doPrint
  bytesPrint
  lowerHexBytesPrint
  upperHexBytesPrint
  stopValPrint // Used for checking if values are in range
)

var (
  connectAddr *net.TCPAddr
  bufferLen uint64
  clientPrintFunc, serverPrintFunc PrintFunc = doPrintFunc, doPrintFunc
  printChan chan string
)

func main() {
  log.SetFlags(0)

  var clientPrint, serverPrint printStatus = 1, 1
  listenAddrStr := flag.String("listen", "", "Network address to listen on")
  connectAddrStr := flag.String("connect", "", "Network address to connect to")
  flag.Var(
    &clientPrint,
    "client-print",
    "Set the client data print (0 = off, 1 = as string, " +
    "2 = as bytes, 3 = as lower hex bytestring, 4 = as upper hex bytestring)",
  )
  flag.Var(
    &serverPrint,
    "server-print",
    "Set the server data print (0 = off, 1 = as string, " +
    "2 = as bytes, 3 = as lower hex bytestring, 4 = as upper hex bytestring)",
  )
  flag.Uint64Var(&bufferLen, "buffer", 1028, "Length of the buffer to read data from")
  flag.Parse()

  if *listenAddrStr == "" || *connectAddrStr == "" {
    log.Fatal("must provide listen and connect addrs")
  }

  switch clientPrint {
  case noPrint:
    clientPrintFunc = noPrintFunc
  case doPrint:
    clientPrintFunc = doPrintFunc
  case bytesPrint:
    clientPrintFunc = bytesPrintFunc
  case lowerHexBytesPrint:
    clientPrintFunc = lowerHexBytesPrintFunc
  case upperHexBytesPrint:
    clientPrintFunc = upperHexBytesPrintFunc
  default:
    log.Fatalln("unknown client-print value:", clientPrint)
  }
  switch serverPrint {
  case noPrint:
    serverPrintFunc = noPrintFunc
  case doPrint:
    serverPrintFunc = doPrintFunc
  case bytesPrint:
    serverPrintFunc = bytesPrintFunc
  case lowerHexBytesPrint:
    serverPrintFunc = lowerHexBytesPrintFunc
  case upperHexBytesPrint:
    serverPrintFunc = upperHexBytesPrintFunc
  default:
    log.Fatalln("unknown server-print value:", serverPrint)
  }

  listenAddr, err := net.ResolveTCPAddr("tcp", *listenAddrStr)
  if err != nil {
    log.Fatal(err)
  }
  connectAddr, err = net.ResolveTCPAddr("tcp", *connectAddrStr)
  if err != nil {
    log.Fatal(err)
  }

  ln, err := net.ListenTCP("tcp", listenAddr)
  if err != nil {
    log.Fatal(err)
  }
  defer ln.Close()

  if serverPrint + clientPrint != 0 {
    printChan = make(chan string, 50)
    go listenPrint()
  }

  fmt.Printf("Listening on %s and connecting to to %s...\n", listenAddr, connectAddr)
  for {
    c, err := ln.AcceptTCP()
    if err != nil {
      log.Fatal(err)
    }
    go handle(c)
  }
}

func handle(client *net.TCPConn) {
  clientAddrStr := client.RemoteAddr().String()
  logErr := func(errFmt string, args ...interface{}) {
    log.Printf("["+clientAddrStr+"] "+errFmt, args...)
  }

  server, err := net.DialTCP("tcp", nil, connectAddr)
  if err != nil {
    logErr("error connecting to server: %v", err)
    client.Close()
    return
  }
  go pipe(client, server, clientPrintFunc)
  pipe(server, client, serverPrintFunc)
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
      // TODO: Report error?
      return
    }
    pf(buf[:n], fromAddrStr, toAddrStr)
    if _, err := to.Write(buf[:n]); err != nil {
      // TODO: Report error?
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
    "%s => %s (%d bytes)\n" +
    "-------------------\n" +
    "%s\n" +
    "===================\n",
    from, to, len(b), string(b),
  )
}

func bytesPrintFunc(b []byte, from, to string) {
  printChan <- fmt.Sprintf(
    "%s => %s (%d bytes)\n" +
    "-------------------\n" +
    "%v\n" +
    "===================\n",
    from, to, len(b), string(b),
  )
}

func lowerHexBytesPrintFunc(b []byte, from, to string) {
  printChan <- fmt.Sprintf(
    "%s => %s (%d bytes)\n" +
    "-------------------\n" +
    "%x\n" +
    "===================\n",
    from, to, len(b), string(b),
  )
}

func upperHexBytesPrintFunc(b []byte, from, to string) {
  printChan <- fmt.Sprintf(
    "%s => %s (%d bytes)\n" +
    "-------------------\n" +
    "%X\n" +
    "===================\n",
    from, to, len(b), string(b),
  )
}
