package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"
	"unicode"
)

var _ = log.Println

// Not map[string][]string, unlike http.Header
type HTTPHeader map[string]string

type Request struct {
	Method  string
	URI     string
	Version string
	Headers HTTPHeader
	//Body    []byte
}

type Response struct {
	Version string
	Status  int
	Phrase  string
	Headers HTTPHeader
	//Body    []byte
}

var ResponseInternalError = &Response{
	Version: "HTTP/1.1",
	Status:  500,
	Phrase:  "Internal Server Error",
}

var ResponseBadRequest = &Response{
	Version: "HTTP/1.1",
	Status:  400,
	Phrase:  "Bad Request",
}

type BaseHandler struct {
	r *bufio.Reader
}

// similar to readLineSlice() in net/textproto/reader.go
func (h *BaseHandler) ReadLine() (string, error) {
	var line []byte
	for {
		l, more, err := h.r.ReadLine()
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

func (h *BaseHandler) ReadHeaders() (HTTPHeader, error) {
	headers := make(map[string]string)
	for {
		line, err := h.ReadLine()
		if err != nil {
			return nil, fmt.Errorf("Failed to read headers")
		}
		if len(line) == 0 {
			break
		}
		fs := strings.SplitN(line, ":", 2)
		if len(fs) != 2 {
			return nil, fmt.Errorf("Invalid header format")
		}
		hdr := strings.ToLower(strings.TrimSpace(fs[0]))
		headers[hdr] = strings.TrimSpace(fs[1])
	}
	return headers, nil
}

// Messages

type ErrorOccurred struct {
	Error error
}

type RequestHeaderReceived struct {
	Req *Request
}

type RequestBodyReceived struct {
	Body  []byte
	IsEnd bool
}

type ResponseHeaderReceived struct {
	Res *Response
}

type ResponseBodyReceived struct {
	Body  []byte
	IsEnd bool
}

type ClientDone struct{}

// Client Handler

type ClientHandler struct {
	h   BaseHandler
	w   io.Writer
	req *Request
}

func NewClientHandler(r io.Reader, w io.Writer) *ClientHandler {
	return &ClientHandler{
		h:   BaseHandler{bufio.NewReader(r)},
		w:   w,
		req: &Request{},
	}
}

func (h *ClientHandler) readRequestLine() error {
	rl, err := h.h.ReadLine()
	if err != nil {
		return fmt.Errorf("Failed to read request line: %v", err)
	}
	fields := strings.Split(rl, " ")
	if len(fields) != 3 {
		return fmt.Errorf("Invalid request line")
	}
	h.req.Method = fields[0]
	h.req.URI = fields[1]
	h.req.Version = fields[2]
	return nil
}

func (h *ClientHandler) readHeaders() error {
	headers, err := h.h.ReadHeaders()
	if err == nil {
		h.req.Headers = headers
	}
	return err
}

func (h *ClientHandler) readBodyIfNeeded() chan interface{} {
	ch := make(chan interface{})
	go func() {
		// TODO: Implement
	}()
	return ch
}

func (h *ClientHandler) writeResponseHeader(res *Response) {
	fmt.Fprintf(h.w, "%s %d %s\r\n", res.Version, res.Status, res.Phrase)
	for k, v := range res.Headers {
		fmt.Fprintf(h.w, "%s: %s\r\n", k, v)
	}
	fmt.Fprintf(h.w, "\r\n")
}

func (h *ClientHandler) writeBody(b []byte) error {
	for n := 0; n < len(b); {
		m, err := h.w.Write(b[n:])
		if err != nil {
			return err
		}
		n += m
	}
	return nil
}

func (h *ClientHandler) handleMessage(m interface{}) (bool, error) {
	done := false
	switch msg := m.(type) {
	case *ResponseHeaderReceived:
		//log.Printf("sending res hdr to client: %v\n", msg.Res)
		h.writeResponseHeader(msg.Res)
	case *ResponseBodyReceived:
		done = msg.IsEnd
		//log.Printf("sending res body to client: n=%v, done=%t\n",
		//	len(msg.Body), done)
		if err := h.writeBody(msg.Body); err != nil {
			//log.Printf("failed to write client: %v\n", err)
			return true, err
		}
		return done, nil
	// TODO: Implement all message handling
	default:
		panic("Invalid message received")
	}
	return done, nil
}

func (h *ClientHandler) loop(ch chan interface{}) {
	readch := h.readBodyIfNeeded()
	for {
		select {
		case msg := <-ch:
			done, err := h.handleMessage(msg)
			if err != nil {
				ch <- &ErrorOccurred{err}
			}
			if done {
				return
			}
		case msg := <-readch:
			ch <- msg
		}
	}
}

func (h *ClientHandler) Start() chan interface{} {
	ch := make(chan interface{})
	go func() {
		defer func() {
			ch <- &ClientDone{}
		}()
		if err := h.readRequestLine(); err != nil {
			ch <- &ErrorOccurred{err}
		}
		if err := h.readHeaders(); err != nil {
			ch <- &ErrorOccurred{err}
		}
		ch <- &RequestHeaderReceived{h.req}

		h.loop(ch)
	}()
	return ch
}

// Server Handler

type ServerHandler struct {
	h   BaseHandler
	w   io.Writer
	res *Response
}

func NewServerHandler(r io.Reader, w io.Writer) *ServerHandler {
	return &ServerHandler{
		h:   BaseHandler{bufio.NewReader(r)},
		w:   w,
		res: &Response{},
	}
}

func parseStatusCode(ss string) (int, error) {
	status, err := strconv.Atoi(ss)
	first := status / 100
	if err != nil || (first < 1 || first > 5) {
		return 0, fmt.Errorf("Invalid status code")
	}
	return status, nil
}

func (h *ServerHandler) readStatusLine() error {
	sl, err := h.h.ReadLine()
	if err != nil {
		return fmt.Errorf("Failed to read status line: %v", err)
	}
	// TODO: Not an ideal
	fields := strings.Split(sl, " ")
	if len(fields) < 3 {
		return fmt.Errorf("Invalid status line: %s", sl)
	}
	h.res.Version = fields[0]
	h.res.Status, err = parseStatusCode(fields[1])
	if err != nil {
		return err
	}
	h.res.Phrase = strings.Join(fields[2:], " ")
	return nil
}

func (h *ServerHandler) readHeaders() error {
	headers, err := h.h.ReadHeaders()
	if err == nil {
		h.res.Headers = headers
	}
	return err
}

func (h *ServerHandler) contentLength() (int, error) {
	cls, ok := h.res.Headers["content-length"]
	if !ok {
		// TODO: this should be a 411 error
		return 0, fmt.Errorf(
			"No Content-Length, chunked encoding isn't supported")
	}
	cl, err := strconv.Atoi(cls)
	if err != nil {
		return 0, fmt.Errorf("Invalid Content-Length")
	}
	return cl, nil
}

// TODO: maybe I can move this to BaseHandler
// TODO: don't rely on "Content-Length". This header doesn't always exist.
func (h *ServerHandler) readBodyIfNeeded() chan interface{} {
	ch := make(chan interface{})
	go func() {
		b := make([]byte, 4096)
		contentLength, _ := h.contentLength()
		for n := 0; n < contentLength; {
			m, err := h.h.r.Read(b)
			if err != nil {
				nerr := fmt.Errorf("Failed to read body: %v", err)
				ch <- &ErrorOccurred{nerr}
				return
			}
			n += m
			ch <- &ResponseBodyReceived{b[:m], n == contentLength}
		}
	}()
	return ch
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

func (h *ServerHandler) writeRequest(req *Request) {
	fmt.Fprintf(h.w, "%s %s %s\r\n", req.Method, req.URI, req.Version)
	for k, v := range req.Headers {
		fmt.Fprintf(h.w, "%s: %s\r\n", capitalizeHeader(k), v)
	}
	fmt.Fprintf(h.w, "\r\n")
}

func (h *ServerHandler) loop(ch chan interface{}) {
	readch := h.readBodyIfNeeded()
	for msg := range readch {
		ch <- msg
	}
}

func (h *ServerHandler) Start(req *Request) chan interface{} {
	ch := make(chan interface{})
	go func() {
		h.writeRequest(req)
		// TODO: request body sending

		if err := h.readStatusLine(); err != nil {
			ch <- &ErrorOccurred{err}
			return
		}
		if err := h.readHeaders(); err != nil {
			ch <- &ErrorOccurred{err}
			return
		}

		// TODO: remove this
		if _, err := h.contentLength(); err != nil {
			ch <- &ErrorOccurred{err}
			return
		}

		ch <- &ResponseHeaderReceived{h.res}

		h.loop(ch)
	}()
	return ch
}
