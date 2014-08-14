package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"strings"
	"sync"
)

var _ = fmt.Println

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

func contentLength(h HTTPHeader) (int, error) {
	cls, ok := h["content-length"]
	if !ok {
		return 0, fmt.Errorf("No Content-Length")
	}
	cl, err := strconv.Atoi(cls)
	if err != nil {
		return 0, fmt.Errorf("Invalid Content-Length")
	}
	return cl, nil
}

func isTransferEncodingChunked(h HTTPHeader) bool {
	te, ok := h["transfer-encoding"]
	if !ok {
		return false
	}
	return te == "chunked"
}

func createBodyReader(r io.Reader, h HTTPHeader) BodyReader {
	if cl, err := contentLength(h); err == nil {
		return NewFixedLengthBodyReader(r, cl)
	}
	if isTransferEncodingChunked(h) {
		return NewChunkedBodyReader(r)
	}
	return nil
}

type DialerFunc func(string) (net.Conn, error)

// Used to connect server. Can be mocked.
var serverDialer = func(addr string) (net.Conn, error) {
	return net.Dial("tcp", addr)
}

// Worker handles an HTTP request and serves as proxy
type Worker struct {
	clientConn   net.Conn
	serverConn   net.Conn
	clientReader *bufio.Reader
	serverReader *bufio.Reader
	req          *Request
	res          *Response
	done         chan struct{}
}

type stateFunc func(*Worker) stateFunc

func NewWorker() *Worker {
	return &Worker{
		clientConn:   nil,
		serverConn:   nil,
		clientReader: nil,
		serverReader: nil,
		req:          nil,
		res:          nil,
		done:         make(chan struct{}),
	}
}

func (w *Worker) Start(conn net.Conn) {
	log.Printf("I worker started")
	// defer conn.Close()
	w.clientConn = conn
	w.clientReader = bufio.NewReader(conn)

	for state := waitForRequest; state != nil; {
		state = state(w)
	}
	log.Printf("I worker finished")
}

func (w *Worker) Cancel() {
	w.done <- struct{}{}
}

func (w *Worker) dialToServer() error {
	host, ok := w.req.Headers["host"]
	if !ok {
		return fmt.Errorf("Missing host")
	}
	addr := appendPortIfNeeded(host)
	conn, err := serverDialer(addr)
	if err == nil {
		w.serverConn = conn
		w.serverReader = bufio.NewReader(conn)
	}
	return err
}

func (w *Worker) requestReceived(req *Request) stateFunc {
	w.req = req

	if req.Method != "GET" && req.Method != "HEAD" {
		log.Printf("E %s is not supported", req.Method)
		w.res = ResponseBadRequest // Should be appropriate response
		return sendErrorResponse
	}

	if err := w.dialToServer(); err != nil {
		log.Println(err)
		w.res = ResponseBadRequest
		return sendErrorResponse
	}

	log.Printf("I %s -> %s",
		w.clientConn.RemoteAddr().String(),
		w.serverConn.RemoteAddr().String())
	log.Printf("I %s %v", w.req.URI, w.req.Headers)

	RemoveHopByHopHeaders(w.req.Headers)
	WriteRequest(w.serverConn, req)
	return waitForResponse
}

func (w *Worker) responseReceived(res *Response) stateFunc {
	w.res = res
	log.Printf("I response: %d %v", w.res.Status, w.res.Headers)

	// TODO: call RemoveHopByHopHeaders()
	WriteResponse(w.clientConn, res)
	return receiveBody
}

func (w *Worker) transferBody(reader BodyReader, writer io.Writer) {
	log.Println("I transferBody")
	reader.Start()
	for {
		select {
		case b := <-reader.BodyReceived():
			if len(b) == 0 {
				log.Printf("I body received done")
				return
			}
			n, err := writer.Write(b)
			if n != len(b) || err != nil {
				log.Println("W write failed")
				reader.Cancel()
				return
			}
		case err := <-reader.ErrorOccurred():
			log.Printf("E reader error: %v", err)
			return
		case <-w.done:
			log.Println("W transferBody done")
			reader.Cancel()
			return
		}
	}
}

// state funcs

func waitForRequest(w *Worker) stateFunc {
	log.Printf("I waiting request\n")
	r := NewRequestReader(w.clientReader)
	r.Start()
	for {
		select {
		case req := <-r.RequestReceived():
			return w.requestReceived(req)
		case err := <-r.ErrorOccurred():
			log.Println(err)
			w.res = ResponseInternalError
			return sendErrorResponse
		case <-w.done:
			log.Println("W waitForRequest done")
			return finishWorker
		}
	}
	panic("not reached")
}

func waitForResponse(w *Worker) stateFunc {
	log.Printf("I waiting response\n")
	r := NewResponseReader(w.serverReader)
	r.Start()
	for {
		select {
		case res := <-r.ResponseReceived():
			return w.responseReceived(res)
		case err := <-r.ErrorOccurred():
			log.Println(err)
			w.res = ResponseInternalError
			return sendErrorResponse
		case <-w.done:
			log.Println("W waitForResponse done")
			return finishWorker
		}
	}
	panic("not reached")
}

func receiveBody(w *Worker) stateFunc {
	var wg sync.WaitGroup

	// client -> server
	// if shouldReadRequestBody(w.req) {
	// 	wg.Add(1)
	// 	go func() {
	// 		defer wg.Done()
	// 		w.requestBodyReadWrite()
	// 	}()
	// }

	// server -> client
	wg.Add(1)
	go func() {
		defer wg.Done()

		br := createBodyReader(w.serverReader, w.res.Headers)
		if br == nil {
			log.Printf("E unsupported transfer encoding")
			return
		}
		w.transferBody(br, w.clientConn)
	}()

	wg.Wait()
	return finishWorker
}

func sendErrorResponse(w *Worker) stateFunc {
	log.Printf("E sending error response: %v", w.res)
	WriteResponse(w.clientConn, w.res)
	return finishWorker
}

func finishWorker(w *Worker) stateFunc {
	if w.clientConn != nil {
		log.Printf("I client conn closing")
		w.clientConn.Close()
	}
	if w.serverConn != nil {
		log.Printf("I server conn closing")
		w.serverConn.Close()
	}
	close(w.done)
	return nil
}