package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
)

var port = flag.String("port", "8080", "port number")

type HTTPHeader map[string]string

type Request struct {
	Method  string
	URI     string
	Version string
	Headers HTTPHeader
}

type requestParser struct {
	r   *bufio.Reader
	req *Request
}

func (p *requestParser) readLine() (string, error) {
	var line []byte
	for {
		l, more, err := p.r.ReadLine()
		if err != nil {
			return "", err
		}
		if line == nil && !more {
			return string(l), nil
		}
		line = append(line, l...)
		if !more {
			break
		}
	}
	return string(line), nil
}

func (p *requestParser) readRequestLine() error {
	rl, err := p.readLine()
	if err != nil {
		return fmt.Errorf("Failed to read request line: %v", err)
	}
	fields := strings.Split(rl, " ")
	if len(fields) != 3 {
		return fmt.Errorf("Invalid request line")
	}
	p.req.Method = fields[0]
	p.req.URI = fields[1]
	p.req.Version = fields[2]
	return nil
}

func (p *requestParser) readHeaders() error {
	headers := make(map[string]string)
	for {
		line, err := p.readLine()
		if err != nil {
			return fmt.Errorf("Failed to read headers")
		}
		if len(line) == 0 {
			break
		}
		fs := strings.SplitN(line, ":", 2)
		if len(fs) != 2 {
			return fmt.Errorf("Invalid header format")
		}
		hdr := strings.ToLower(strings.TrimSpace(fs[0]))
		headers[hdr] = strings.TrimSpace(fs[1])
	}
	p.req.Headers = headers
	return nil
}

func (p *requestParser) parse() (*Request, error) {
	if err := p.readRequestLine(); err != nil {
		return nil, err
	}
	if err := p.readHeaders(); err != nil {
		return nil, err
	}
	return p.req, nil
}

func newRequestParser(r io.Reader) *requestParser {
	return &requestParser{bufio.NewReader(r), &Request{}}
}

func readConn(conn net.Conn) <-chan []byte {
	ch := make(chan []byte)
	go func() {
		defer func() {
			close(ch)
			log.Printf("readConn done\n")
		}()
		b := make([]byte, 4096)
		for {
			n, err := conn.Read(b)
			if n > 0 {
				res := make([]byte, n)
				copy(res, b[:n])
				ch <- res
			}
			if n == 0 {
				return
			}
			if err != nil {
				return
			}
		}
	}()
	return ch
}

// Reads bytes from c1 and writes them to c2, and vice versa.
// Repeat until either c1 or c2 get closed.
// TODO: Writes are synchronous and no error handling. Make them async.
func pipe(c1, c2 net.Conn) {
	ch1 := readConn(c1)
	ch2 := readConn(c2)

	for {
		select {
		case b := <-ch1:
			if len(b) == 0 {
				return
			}
			c2.Write(b)
		case b := <-ch2:
			if len(b) == 0 {
				return
			}
			c1.Write(b)
		}
	}
}

func handle(conn net.Conn) {
	defer conn.Close()
	req, err := newRequestParser(conn).parse()
	if err != nil || req.Method != "CONNECT" {
		conn.Write([]byte("HTTP/1.1 400 Bad Request\r\n\r\n"))
		return
	}
	log.Printf("R %v\n", req)

	servConn, err := net.Dial("tcp", req.URI)
	if err != nil {
		conn.Write([]byte("HTTP/1.1 500 Internal Server Error\r\n\r\n"))
		return
	}
	defer servConn.Close()

	_, err = conn.Write([]byte("HTTP/1.1 200 Connection established\r\n\r\n"))
	if err != nil {
		return
	}

	log.Printf("C %s -> %s\n",
		conn.RemoteAddr().String(), servConn.RemoteAddr().String())

	pipe(conn, servConn)
}

func serve() {
	ln, err := net.Listen("tcp", ":"+*port)
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

func main() { serve() }
