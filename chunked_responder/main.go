package main

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const defaultTotalSize = "100k"

var units = map[byte]int{
	'k': 1000,
	'm': 1000 * 1000,
	'g': 1000 * 1000 * 1000,
}

func sizeToInt(s string) (int, error) {
	if len(s) == 0 {
		return 0, fmt.Errorf("Invalid size")
	}
	var err error
	var m, sz int
	m, ok := units[s[len(s)-1:][0]]
	if ok {
		sz, err = strconv.Atoi(s[:len(s)-1])
	} else {
		m = 1
		sz, err = strconv.Atoi(s)
	}
	if err != nil {
		return 0, err
	}
	return sz * m, nil
}

type chunkResponder struct{}

func getSize(u *url.URL) (int, error) {
	query := u.Query()
	if size := query.Get("size"); size != "" {
		sz, err := sizeToInt(size)
		if err != nil {
			return 0, err
		}
		return sz, nil
	}
	return 0, fmt.Errorf("no size parameter")
}

func (c chunkResponder) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// What a horrble API! I need to rely on type casting when I want to make
	// http server be able to send chunked response. At least, this check should be done
	// at the startup, not each request handling...
	flusher, ok := w.(http.Flusher)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	sz, err := getSize(r.URL)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Header()["Content-Type"] = []string{"text/plain"}
	w.WriteHeader(http.StatusOK)
	for i := 0; i < sz-1; i++ {
		w.Write([]byte("a"))
		flusher.Flush()
		time.Sleep(time.Duration(10) * time.Millisecond)
	}
	w.Write([]byte("\n"))
	flusher.Flush()
}

func main() {
	http.Handle("/", &chunkResponder{})
	http.ListenAndServe(":9100", nil)
}
