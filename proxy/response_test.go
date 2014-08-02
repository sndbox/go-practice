package main

import (
	"bytes"
	"strconv"
	"strings"
	"testing"
)

func TestParseResponse(t *testing.T) {
	ss := [...]string{
		"HTTP/1.1 200 OK\r\n",
		"Content-Type: text/plain\r\n",
		"Content-Length: 6\r\n",
		"\r\n",
		"FooBar",
	}
	r := strings.NewReader(strings.Join(ss[:], ""))
	res, err := ParseResponse(r)
	if err != nil {
		t.Errorf("error", err)
	}

	ExpectEqual(t, "HTTP/1.1", res.Version)
	ExpectEqual(t, "200", strconv.Itoa(res.Status))
	ExpectEqual(t, "OK", res.Phrase)
	ExpectEqual(t, "FooBar", string(res.Body))
}

func TestWriteResponse(t *testing.T) {
	res := ResponseInternalError

	buf := new(bytes.Buffer)
	if err := WriteResponse(buf, res); err != nil {
		t.Error("error", err)
	}

	expect := "HTTP/1.1 500 Internal Server Error\r\n\r\n"
	actual := buf.String()
	if expect != actual {
		t.Error("mismatch", expect, actual)
	}
}
