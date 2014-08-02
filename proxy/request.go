package main

import (
	"bufio"
	"fmt"
	"io"
	"strings"
	"unicode"
)

type Request struct {
	Method  string
	URI     string
	Version string
	Headers HTTPHeader
	Body    []byte
}

func capitalizeHeader(h string) string {
	ret := make([]rune, len(h))
	cap := true
	for i, c := range h {
		r := rune(c)
		if cap && unicode.IsLetter(r) {
			ret[i] = unicode.ToUpper(r)
			cap = false
		} else {
			ret[i] = r
		}
		if c == '-' {
			cap = true
		}
	}
	return string(ret)
}

type requestParser struct {
	r   BaseParser
	req *Request
}

func (r *requestParser) readRequestLine() error {
	rl, err := r.r.ReadLine()
	if err != nil {
		return fmt.Errorf("Failed to read request line: %v", err)
	}
	fields := strings.Split(rl, " ")
	if len(fields) != 3 {
		return fmt.Errorf("Invalid request line")
	}
	r.req.Method = fields[0]
	r.req.URI = fields[1]
	r.req.Version = fields[2]
	return nil
}

func (r *requestParser) readHeaders() error {
	headers, err := r.r.ReadHeaders()
	if err == nil {
		r.req.Headers = headers
	}
	return err
}

func (r *requestParser) readBody() error {
	// TODO: Implement.
	return nil
}

func ParseRequest(r io.Reader) (*Request, error) {
	fmt.Printf("Reading request from %v\n", r)
	p := requestParser{BaseParser{bufio.NewReader(r)}, &Request{}}
	if err := p.readRequestLine(); err != nil {
		return nil, err
	}
	if err := p.readHeaders(); err != nil {
		return nil, err
	}
	if err := p.readBody(); err != nil {
		return nil, err
	}
	return p.req, nil
}

func WriteRequest(w io.Writer, req *Request) error {
	fmt.Fprintf(w, "%s %s %s\r\n", req.Method, req.URI, req.Version)
	for k, v := range req.Headers {
		fmt.Fprintf(w, "%s: %s\r\n", capitalizeHeader(k), v)
	}
	fmt.Fprintf(w, "\r\n")
	for n := 0; n < len(req.Body); {
		m, err := w.Write(req.Body[n:])
		if err != nil {
			return err
		}
		n += m
	}
	return nil
}
