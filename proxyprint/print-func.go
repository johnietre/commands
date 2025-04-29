package main

import (
	"encoding/json"
	"fmt"
	"log"
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

type PrintData struct {
	msg    string
	server bool
}

type PrintFunc = func(b []byte, from, to string, server bool)

const (
	noPrint            printStatus = 0
	doPrint            printStatus = 1
	bytesPrint         printStatus = 2
	lowerHexBytesPrint printStatus = 3
	upperHexBytesPrint printStatus = 4
	stopValPrint       printStatus = 5 // Used for checking if values are in range
)

func noPrintFunc([]byte, string, string, bool) {}

func doPrintFunc(b []byte, from, to string, server bool) {
	printChan <- PrintData{
		msg: fmt.Sprintf(
			"%s => %s (%d bytes)\n"+
				"-------------------\n"+
				"%s\n"+
				"===================\n",
			from, to, len(b), b,
		),
		server: server,
	}
}

func bytesPrintFunc(b []byte, from, to string, server bool) {
	printChan <- PrintData{
		msg: fmt.Sprintf(
			"%s => %s (%d bytes)\n"+
				"-------------------\n"+
				"%v\n"+
				"===================\n",
			from, to, len(b), b,
		),
		server: server,
	}
}

func lowerHexBytesPrintFunc(b []byte, from, to string, server bool) {
	printChan <- PrintData{
		msg: fmt.Sprintf(
			"%s => %s (%d bytes)\n"+
				"-------------------\n"+
				"%x\n"+
				"===================\n",
			from, to, len(b), b,
		),
		server: server,
	}
}

func upperHexBytesPrintFunc(b []byte, from, to string, server bool) {
	printChan <- PrintData{
		msg: fmt.Sprintf(
			"%s => %s (%d bytes)\n"+
				"-------------------\n"+
				"%X\n"+
				"===================\n",
			from, to, len(b), b,
		),
		server: server,
	}
}
