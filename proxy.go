package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"time"
)

type Server struct {
	Addr string
	lis  net.Listener
}

func (s *Server) ListenAndServe() error {
	lis, err := net.Listen("tcp", s.Addr)
	if err != nil {
		return err
	}
	s.lis = lis

	for {
		conn, err := lis.Accept()
		if err != nil {
			return err
		}

		go handleConn(conn)
	}
}

const handshakeTimeout = 5 * time.Second

func handleConn(conn net.Conn) {
	defer conn.Close()
	conn.SetReadDeadline(time.Now().Add(handshakeTimeout))
	req, err := http.ReadRequest(bufio.NewReader(conn))
	if err != nil {
		log.Print("read request:", err)
		return
	}
	conn.SetReadDeadline(time.Time{})

	if req.Method == "CONNECT" {
		handleHTTPS := func(conn net.Conn, req *http.Request) error {
			port := req.URL.Port()
			if port == "" {
				port = "443"
			}
			remoteConn, e := net.DialTimeout("tcp", net.JoinHostPort(req.URL.Hostname(), port), handshakeTimeout)
			if e != nil {
				return e
			}
			defer remoteConn.Close()
			_, e = fmt.Fprintf(conn, "%s 200 Connection established\r\n\r\n", req.Proto)
			if e != nil {
				return e
			}
			return relay(remoteConn, conn)
		}
		err = handleHTTPS(conn, req)
		if err != nil {
			log.Print(err)
		}
		return
	}

	handleHTTP := func(conn net.Conn, req *http.Request) error {
		port := req.URL.Port()
		if port == "" {
			port = "80"
		}
		remoteConn, e := net.DialTimeout("tcp", net.JoinHostPort(req.URL.Hostname(), port), 5*time.Second)
		if e != nil {
			return e
		}
		defer remoteConn.Close()
		e = req.Write(remoteConn)
		if e != nil {
			return e
		}
		_, e = io.Copy(conn, remoteConn)
		return e
	}
	err = handleHTTP(conn, req)
	if err != nil {
		log.Print(err)
	}
}

func (s *Server) Close() error {
	if s.lis == nil {
		return nil
	}
	return s.lis.Close()
}

func relay(dst, src net.Conn) error {
	errChan := make(chan error, 1)
	go func() {
		_, err := io.Copy(dst, src)
		errChan <- err
		dst.SetReadDeadline(time.Now().Add(5 * time.Second))
	}()
	_, err := io.Copy(src, dst)
	src.SetReadDeadline(time.Now().Add(5 * time.Second))
	err2 := <-errChan
	if err != nil && !errors.Is(err, os.ErrDeadlineExceeded) {
		return err
	}
	if err2 != nil && !errors.Is(err2, os.ErrDeadlineExceeded) {
		return err2
	}
	return nil
}
