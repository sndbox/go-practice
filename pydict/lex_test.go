package pydict

import (
	"fmt"
	"io/ioutil"
	"testing"
)

var _ = fmt.Println

func mustReadFile(p string) string {
	b, err := ioutil.ReadFile(p)
	if err != nil {
		panic(err)
	}
	return string(b)
}

func TestLex(t *testing.T) {
	expect := []string{
		"{", "foo", ":", "bar", ",",
		"[", "a", ",", "b", ",", "c", "]", ",", "}",
	}
	l := lex(mustReadFile("test1.input"))
	for _, v := range expect {
		item := l.nextItem()
		if item.iType == itemError {
			t.Error("error occurred")
		}
		if v != item.value {
			t.Errorf("got %s but want %s", item.value, v)
		}
	}
	if l.nextItem().iType != itemEOF {
		t.Error("expect EOF")
	}
}
