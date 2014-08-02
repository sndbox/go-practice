package main

import (
	"bufio"
	"fmt"
	"strings"
)

// Not map[string][]string, unlike http.Header
type HTTPHeader map[string]string

type BaseParser struct {
	r *bufio.Reader
}

// similar to readLineSlice() in net/textproto/reader.go
func (r *BaseParser) ReadLine() (string, error) {
	var line []byte
	for {
		l, more, err := r.r.ReadLine()
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

func (r *BaseParser) ReadHeaders() (HTTPHeader, error) {
	headers := make(map[string]string)
	for {
		line, err := r.ReadLine()
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
		h := strings.ToLower(strings.TrimSpace(fs[0]))
		headers[h] = strings.TrimSpace(fs[1])
	}
	return headers, nil
}
