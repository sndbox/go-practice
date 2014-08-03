package main

import (
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
)

func waitForRequestHeader(ch chan interface{}) (*Request, error) {
	m := <-ch
	// TODO: maybe handle EOF?
	switch msg := m.(type) {
	case *RequestHeaderReceived:
		return msg.Req, nil
	case *ErrorOccurred:
		return nil, msg.Error
	}
	return nil, fmt.Errorf("Failed to read request header")
}

func appendPortIfNeeded(h string) string {
	// TODO: support https
	pos := strings.LastIndex(h, ":")
	if pos == -1 {
		return h + ":80"
	}
	p, err := strconv.Atoi(h[pos+1:])
	if err != nil || p == 0 {
		return h + ":80"
	}
	return h
}

func dialForRequest(req *Request) (net.Conn, error) {
	host, ok := req.Headers["host"]
	if !ok {
		return nil, fmt.Errorf("No Host header")
	}
	addr := appendPortIfNeeded(host)
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func handleClientMessage(m interface{}, clChan, svChan chan interface{}) bool {
	done := false
	switch msg := m.(type) {
	case *ClientDone:
		log.Println("client done")
		done = true
	case *ErrorOccurred:
		log.Println(msg.Error)
		// TODO: helper function
		clChan <- &ResponseHeaderReceived{ResponseBadRequest}
		clChan <- &ResponseBodyReceived{nil, true}
	default:
		log.Printf("Unexpected message from client: %v\n", msg)
		panic("")
	}
	return done
}

func handleServerMessage(m interface{}, clChan, svChan chan interface{}) bool {
	switch msg := m.(type) {
	case *ResponseHeaderReceived:
		log.Printf("response header received: status=%d\n", msg.Res.Status)
		clChan <- msg
	case *ResponseBodyReceived:
		log.Printf("response body received: n=%d\n", len(msg.Body))
		clChan <- msg
	case *ErrorOccurred:
		log.Println(msg.Error)
		// TODO: helper function
		clChan <- &ResponseHeaderReceived{ResponseInternalError}
		clChan <- &ResponseBodyReceived{nil, true}
	default:
		log.Printf("Unexpected message from server: %v\n", msg)
		panic("")
	}
	return false
}

// TODO: make this testable
func handle(conn net.Conn) {
	log.Printf("client connected: %s\n", conn.RemoteAddr().String())
	defer conn.Close()
	cl := NewClientHandler(conn, conn)
	clChan := cl.Start()
	// TODO: Do I need to stop cl?

	req, err := waitForRequestHeader(clChan)
	if err != nil {
		log.Println(err)
		return
	}

	svConn, err := dialForRequest(req)
	if err != nil {
		log.Println(err)
		return
	}
	defer svConn.Close()

	sv := NewServerHandler(svConn, svConn)

	svChan := sv.Start(req)

	// TODO: support timeout
	// TODO: use a channel to break the loop
	done := false
	for !done {
		select {
		case msg := <-clChan:
			done = handleClientMessage(msg, clChan, svChan)
		case msg := <-svChan:
			done = handleServerMessage(msg, clChan, svChan)
		}
	}
}

func Serve() {
	port := "8080"
	ln, err := net.Listen("tcp", ":"+port)
	if err != nil {
		panic(err)
	}
	defer ln.Close()
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		go handle(conn)
	}
}

func main() {
	Serve()
}
