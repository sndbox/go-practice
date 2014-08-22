package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"strings"
)

// Only accepts 7-bit ascii
type RunLengthWriter struct {
	w         io.Writer
	minLength int
}

func (w *RunLengthWriter) Write(b []byte) (int, error) {
	buf := make([]byte, 2)
	n := len(b)
	i := 0
	for i < n {
		x := b[i]
		if x&0x80 != 0 {
			return i, fmt.Errorf("Invalid 7-bit ascii: %c", x)
		}
		j := 1
		for ; i+j < n && j < 127; j++ {
			if b[i+j] != x {
				if b[i+j]&0x80 != 0 {
					return i, fmt.Errorf("Invalid 7-bit ascii: %c", b[i+j])
				}
				break
			}
		}
		var err error
		if j >= w.minLength {
			buf[0] = byte(0x80 | j)
			buf[1] = x
			_, err = w.w.Write(buf)
			i += j
		} else {
			buf[0] = x
			_, err = w.w.Write(buf[:1])
			i += 1
		}
		if err != nil {
			return i, err
		}
	}
	return i, nil
}

func NewRunLengthWriter(w io.Writer) *RunLengthWriter {
	return &RunLengthWriter{w, 5}
}

func main() {
	buf := new(bytes.Buffer)
	w := NewRunLengthWriter(buf)
	input := []byte("aaaaaaabbcccccd" + strings.Repeat("e", 255))
	n, err := w.Write(input)
	if err != nil || n != len(input) {
		log.Fatal(err)
	}

	fmt.Println(buf.Bytes())
}
