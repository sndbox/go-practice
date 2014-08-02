package main

import (
	"bytes"
	"strings"
	"testing"
)

func ExpectEqual(t *testing.T, expect, actual string) {
	if expect != actual {
		t.Errorf("Got %s, want %s", actual, expect)
	}
}

func TestCapitalizeHeader(t *testing.T) {
	ExpectEqual(t, "Content-Length", capitalizeHeader("content-length"))
	ExpectEqual(t, "X-Foo", capitalizeHeader("x-foo"))
}

func TestParseRequest(t *testing.T) {
	r := strings.NewReader("GET / HTTP/1.1\r\nHost: www.google.com\r\n\r\n")
	req, err := ParseRequest(r)
	if err != nil {
		t.Error("error", err)
	}
	ExpectEqual(t, "GET", req.Method)
	ExpectEqual(t, "/", req.URI)
	ExpectEqual(t, "HTTP/1.1", req.Version)
	ExpectEqual(t, "www.google.com", req.Headers["host"])
}

func TestWriteRequest(t *testing.T) {
	req := &Request{
		Method:  "GET",
		URI:     "/",
		Version: "HTTP/1.1",
		Headers: map[string]string{
			"Host": "localhost",
		},
	}

	buf := new(bytes.Buffer)
	if err := WriteRequest(buf, req); err != nil {
		t.Error("error", err)
	}

	expect := "GET / HTTP/1.1\r\nHost: localhost\r\n\r\n"
	actual := buf.String()
	ExpectEqual(t, expect, actual)
}
