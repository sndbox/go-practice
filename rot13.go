// A part of go-tour

package main

// +build ignore

import (
	"fmt"
	"io"
	"os"
	"strings"
)

type rot13Reader struct {
	r io.Reader
}

func (r *rot13Reader) Read(b []byte) (int, error) {
	n, err := r.r.Read(b)
	if err != nil {
		return 0, err
	}
	for i := 0; i < n; i++ {
		diff := byte(0)
		if b[i] >= byte('A') && b[i] <= byte('Z') {
			diff = b[i] - byte('A')
		} else {
			diff = b[i] - byte('a')
		}
		if diff >= 13 {
			b[i] -= 13
		} else {
			b[i] += 13
		}
	}
	return n, nil
}

func main() {
	s := strings.NewReader("Lbh penpxrq gur pbqr!")
	r := rot13Reader{s}
	io.Copy(os.Stdout, &r)
	fmt.Println()
}
