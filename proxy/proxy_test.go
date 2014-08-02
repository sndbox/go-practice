package main

import (
	"testing"
)

func TestAppendPortIfNeeded(t *testing.T) {
	ExpectEqual(t, "www.google.com:80", appendPortIfNeeded("www.google.com"))
	ExpectEqual(t, "www.google.com:443",
		appendPortIfNeeded("www.google.com:443"))
}
