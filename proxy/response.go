package main

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

type Response struct {
	Version string
	Status  int
	Phrase  string
	Headers HTTPHeader
	Body    []byte
}

var ResponseInternalError = &Response{
	Version: "HTTP/1.1",
	Status:  500,
	Phrase:  "Internal Server Error",
}

var ResponseInvalidRequest = &Response{
	Version: "HTTP/1.1",
	Status:  400,
	Phrase:  "Invalid Request",
}

func parseStatusCode(ss string) (int, error) {
	status, err := strconv.Atoi(ss)
	first := status / 100
	if err != nil || (first < 1 || first > 5) {
		return 0, fmt.Errorf("Invalid status code")
	}
	return status, nil
}

type responseParser struct {
	r   BaseParser
	res *Response
}

func (r *responseParser) Read(b []byte) (int, error) {
	return r.r.r.Read(b)
}

func (r *responseParser) readStatusLine() error {
	sl, err := r.r.ReadLine()
	if err != nil {
		return fmt.Errorf("Failed to read status line: %v", err)
	}
	// TODO: Not an ideal
	fields := strings.Split(sl, " ")
	if len(fields) < 3 {
		return fmt.Errorf("Invalid status line: %s", sl)
	}
	r.res.Version = fields[0]
	r.res.Status, err = parseStatusCode(fields[1])
	if err != nil {
		return err
	}
	r.res.Phrase = strings.Join(fields[2:], " ")
	return nil
}

func (r *responseParser) readHeaders() error {
	headers, err := r.r.ReadHeaders()
	if err == nil {
		r.res.Headers = headers
	}
	return err
}

func (r *responseParser) readBody() error {
	// TODO: Don't rely on "Content-Length". This header don't always exist.
	cls, ok := r.res.Headers["content-length"]
	if !ok {
		return fmt.Errorf("Response don't contain Content-Length")
	}
	cl, err := strconv.Atoi(cls)
	if err != nil {
		return fmt.Errorf("Invalid Content-Length")
	}
	// TODO: make this logic allow partial read
	b := make([]byte, cl)
	for n := 0; n < cl; {
		m, err := r.Read(b[n:])
		if err != nil {
			return err
		}
		n += m
	}
	r.res.Body = b
	return nil
}

func ParseResponse(r io.Reader) (*Response, error) {
	p := responseParser{BaseParser{bufio.NewReader(r)}, &Response{}}
	if err := p.readStatusLine(); err != nil {
		return nil, err
	}
	if err := p.readHeaders(); err != nil {
		return nil, err
	}
	if err := p.readBody(); err != nil {
		return nil, err
	}
	return p.res, nil
}

func WriteResponse(w io.Writer, r *Response) error {
	fmt.Fprintf(w, "%s %d %s\r\n", r.Version, r.Status, r.Phrase)
	for k, v := range r.Headers {
		fmt.Fprintf(w, "%s: %s\r\n", k, v)
	}
	fmt.Fprintf(w, "\r\n")
	for n := 0; n < len(r.Body); {
		m, err := w.Write(r.Body[n:])
		if err != nil {
			return err
		}
		n += m
	}
	return nil
}
