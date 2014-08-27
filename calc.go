package main

import (
	"bufio"
	"fmt"
	"os"
)

// expr := term (('+' | '-') term)*
// term := primary (('*' | '/') primary)*
// primary := NUMBER | '(' expr ')'

func expr(s string) (n int, out string) {
	x, s := term(s)
	for len(s) > 0 && (s[0] == '+' || s[0] == '-') {
		y, s1 := term(s[1:])
		if s[0] == '+' {
			x = x + y
		} else {
			x = x - y
		}
		s = s1
	}
	return x, s
}

func term(s string) (n int, out string) {
	x, s := primary(s)
	for len(s) > 0 && (s[0] == '*' || s[0] == '/') {
		y, s1 := primary(s[1:])
		if s[0] == '*' {
			x = x * y
		} else {
			x = x / y
		}
		s = s1
	}
	return x, s
}

func primary(s string) (n int, out string) {
	if len(s) == 0 {
		panic("expect number or '('")
	}
	if s[0] == '(' {
		n, t := expr(s[1:])
		if len(t) == 0 || t[0] != ')' {
			panic("expect ')'")
		}
		return n, t[1:]
	}
	i := 0
	m := 0
	for ; i < len(s); i++ {
		if s[i] >= '0' && s[i] <= '9' {
			m = m*10 + int(s[i]-'0')
		} else {
			break
		}
	}
	if i == 0 {
		panic("invalid number")
	}
	return m, s[i:]
}

func calc(s string) int {
	n, _ := expr(s)
	return n
}

func repl() {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
			repl()
		}
	}()

	r := bufio.NewReader(os.Stdin)
	for {
		l, err := r.ReadString('\n')
		if err != nil {
			break
		}
		fmt.Println(calc(l))
	}
}

func main() {
	repl()
}
