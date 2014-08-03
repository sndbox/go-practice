package main

import (
	"bytes"
	"strconv"
	"strings"
	"testing"
)

func ExpectEqual(t *testing.T, expect, actual string) {
	if expect != actual {
		t.Errorf("Got %s, want %s", actual, expect)
	}
}

func TestClientHandlerStart(t *testing.T) {
	r := strings.NewReader("GET / HTTP/1.1\r\nHost: www.google.com\r\n\r\n")
	w := new(bytes.Buffer)
	h := NewClientHandler(r, w)

	ch := h.Start()

	msg, ok := (<-ch).(*RequestHeaderReceived)
	if !ok {
		t.Errorf("Failed to receive message")
	}
	ExpectEqual(t, "GET", msg.Req.Method)
	ExpectEqual(t, "/", msg.Req.URI)
	ExpectEqual(t, "HTTP/1.1", msg.Req.Version)
	ExpectEqual(t, "www.google.com", msg.Req.Headers["host"])

	res := &Response{
		Version: "HTTP/1.1",
		Status:  200,
		Phrase:  "OK",
	}
	ch <- &ResponseHeaderReceived{res}
	ch <- &ResponseBodyReceived{[]byte("FooBar"), true}
	output := w.String()
	ExpectEqual(t, "HTTP/1.1 200 OK\r\n\r\nFooBar", output)
}

func TestServerHandler(t *testing.T) {
	ss := []string{
		"HTTP/1.1 200 OK\r\n",
		"Content-Length: 6\r\n",
		"\r\n",
		"FooBar",
	}
	r := strings.NewReader(strings.Join(ss, ""))
	w := new(bytes.Buffer)
	h := NewServerHandler(r, w)

	req := &Request{
		Method:  "GET",
		URI:     "/",
		Version: "HTTP/1.1",
		Headers: map[string]string{
			"Host": "localhost",
		},
	}
	ch := h.Start(req)

	hdrmsg, ok := (<-ch).(*ResponseHeaderReceived)
	if !ok {
		t.Errorf("Failed to receive response header")
	}
	ExpectEqual(t, "HTTP/1.1", hdrmsg.Res.Version)
	ExpectEqual(t, "200", strconv.Itoa(hdrmsg.Res.Status))
	ExpectEqual(t, "OK", hdrmsg.Res.Phrase)
	ExpectEqual(t, "6", hdrmsg.Res.Headers["content-length"])

	bodymsg, ok := (<-ch).(*ResponseBodyReceived)
	if !ok {
		t.Errorf("Failed to receive response body")
	}
	ExpectEqual(t, "FooBar", string(bodymsg.Body))
}
