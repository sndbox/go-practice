package main

import (
	"fmt"
)

type primeGenerator struct {
	ch   chan int
	done chan struct{}
}

func (p *primeGenerator) Close() {
	close(p.done)
}

func (p *primeGenerator) Next() int {
	return <-p.ch
}

func newPrimeGenerator() *primeGenerator {
	ch := make(chan int)
	done := make(chan struct{})
	go func() {
		defer close(ch)
		select {
		case ch <- 2:
		case <-done:
			return
		}
		primes := []int{2}
		i := 3
		for {
			for {
				isPrime := true
				for _, p := range primes {
					if i%p == 0 {
						isPrime = false
						break
					}
				}
				if isPrime {
					break
				}
				i += 2
			}
			select {
			case ch <- i:
			case <-done:
				return
			}
			primes = append(primes, i)
		}
	}()
	return &primeGenerator{ch, done}
}

func main() {
	g := newPrimeGenerator()
	defer func() {
		g.Close()
	}()
	for i := 0; i < 20; i++ {
		fmt.Println(g.Next())
	}
}
