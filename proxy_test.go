package main

import (
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func chooseAddr() string {
	lis, _ := net.Listen("tcp", "127.0.0.1:")
	lis.Close()
	return lis.Addr().String()
}

func TestHttpProxy(t *testing.T) {
	s := Server{Addr: chooseAddr()}
	go func() {
		err := s.ListenAndServe()
		if err != nil && !errors.Is(err, net.ErrClosed) {
			t.Error(err)
		}
	}()
	defer s.Close()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "ok")
	}))

	client := &http.Client{
		Transport: &http.Transport{
			Proxy: func(r *http.Request) (*url.URL, error) {
				return url.Parse("http://" + s.Addr)
			},
		},
	}
	resp, err := client.Get(ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	bs, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	if string(bs) != "ok" {
		t.Fatalf("want ok, got %s", bs)
	}
}

func TestHttpsProxy(t *testing.T) {
	s := Server{Addr: chooseAddr()}
	go func() {
		err := s.ListenAndServe()
		if err != nil && !errors.Is(err, net.ErrClosed) {
			t.Error(err)
		}
	}()
	defer s.Close()

	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "ok")
	}))

	transport := ts.Client().Transport.(*http.Transport)
	transport.Proxy = func(r *http.Request) (*url.URL, error) {
		return url.Parse("http://" + s.Addr)
	}
	client := &http.Client{
		Transport: transport,
	}
	resp, err := client.Get(ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	bs, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	if string(bs) != "ok" {
		t.Fatalf("want ok, got %s", bs)
	}
}
