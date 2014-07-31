// A part of go-tour

package main

// +build ignore

import (
	"fmt"
)

type Fetcher interface {
	// Fetch returns the body of URL and a slice of URLs fond on that page.
	Fetch(url string) (body string, urls []string, err error)
}

// Crawl uses fetcher to recursively crawl pages starting with url,
// to a maximum of depth.
func CrawlImpl(
	url string, depth int, fetcher Fetcher, m map[string]bool,
	finish chan int) {
	if depth <= 0 {
		finish <- 0
		return
	}
	body, urls, err := fetcher.Fetch(url)
	if err != nil {
		fmt.Println(err)
		finish <- 0
		return
	}
	fmt.Printf("found: %s %q\n", url, body)
	channels := []chan int{}
	for _, u := range urls {
		if !m[u] {
			m[u] = true
			ch := make(chan int)
			go CrawlImpl(u, depth-1, fetcher, m, ch)
			channels = append(channels, ch)
		}
	}
	for _, ch := range channels {
		<-ch
	}
	finish <- 0
}

func Crawl(url string, depth int, fetcher Fetcher) {
	m := make(map[string]bool)
	ch := make(chan int)
	go CrawlImpl(url, depth, fetcher, m, ch)
	<-ch
}

func main() {
	Crawl("http://golang.org/", 4, fetcher)
}

// fakeFetcher is Fetcher that returns canned results.
type fakeFetcher map[string]*fakeResult

type fakeResult struct {
	body string
	urls []string
}

func (f fakeFetcher) Fetch(url string) (string, []string, error) {
	if res, ok := f[url]; ok {
		return res.body, res.urls, nil
	}
	return "", nil, fmt.Errorf("not found: %s", url)
}

var fetcher = fakeFetcher{
	"http://golang.org/": &fakeResult{
		"The Go programming Language",
		[]string{
			"http://golang.org/pkg/",
			"http://golang.org/cmd/",
		},
	},
	"http://golang.org/pkg/": &fakeResult{
		"Packages",
		[]string{
			"http://golang.org/",
			"http://golang.org/cmd/",
			"http://golang.org/pkg/fmt/",
			"http://golang.org/pkg/os/",
		},
	},
	"http://golang.org/pkg/fmt/": &fakeResult{
		"Package fmt",
		[]string{
			"http://golang.org/",
			"http://golang.org/pkg/",
		},
	},
	"http://golang.org/pkg/os/": &fakeResult{
		"Package os",
		[]string{
			"http://golang.org/",
			"http://golang.org/pkg/",
		},
	},
}
