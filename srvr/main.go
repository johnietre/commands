package main

import (
	"flag"
	"net/http"
)

func main() {
	addrPtr := flag.String("a", "localhost:8000", "Network address to listen on")
	flag.Parse()
	println("Starting server on", *addrPtr)
	println("Error:", http.ListenAndServe(*addrPtr, http.FileServer(http.Dir("."))).Error())
}
