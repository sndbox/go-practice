package main

import (
	"fmt"
	"net"
	"strconv"
	"strings"
)

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

func sendRequest(req *Request) (*Response, error) {
	fmt.Printf("Trying to handle %v\n", req)
	// TODO: Return appropriate response when an error occurred
	// TODO: Add "fetcher" interface to improve testability
	host, ok := req.Headers["host"]
	if !ok {
		return ResponseInvalidRequest, nil
	}
	addr := appendPortIfNeeded(host)
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		fmt.Println(err)
		return ResponseInternalError, err
	}
	fmt.Printf("dialing %s...\n", conn.RemoteAddr().String())
	defer conn.Close()

	if err := WriteRequest(conn, req); err != nil {
		fmt.Println(err)
		return ResponseInternalError, err
	}
	fmt.Printf("sent the request %v\n", req)

	res, err := ParseResponse(conn)
	if err != nil {
		fmt.Println(err)
		return ResponseInternalError, err
	}
	return res, nil
}

func handle2(conn net.Conn) {
	res := ResponseInternalError
	defer func() {
		if err := WriteResponse(conn, res); err != nil {
			fmt.Println(err)
		}
		fmt.Printf("response: %d\n", res.Status)
		conn.Close()
	}()

	request, err := ParseRequest(conn)
	if err != nil {
		fmt.Println(err)
		return
	}

	res, err = sendRequest(request)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("request", request)
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
			fmt.Println(err)
			continue
		}
		go handle2(conn)
	}
}

func main() {
	Serve()
}
