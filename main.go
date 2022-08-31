package main

import (
	"errors"
	"flag"
	"log"
	"net"
)

var addr = flag.String("listen", "127.0.0.1:8080", "listen address")

func main() {
	flag.Parse()
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	s := Server{Addr: *addr}
	log.Printf("http proxy server listening %s\n", *addr)
	err := s.ListenAndServe()
	if err != nil && !errors.Is(err, net.ErrClosed) {
		log.Print(err)
	}
}
