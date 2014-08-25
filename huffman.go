package main

import (
	"container/heap"
	"fmt"
)

type node struct {
	value rune
	freq  int
	left  *node
	right *node
	index int
}

type priorityQueue []*node

func (pq priorityQueue) Len() int { return len(pq) }

func (pq priorityQueue) Less(i, j int) bool {
	return pq[i].freq < pq[j].freq
}

func (pq priorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *priorityQueue) Push(x interface{}) {
	n := len(*pq)
	node := x.(*node)
	node.index = n
	*pq = append(*pq, node)
}

func (pq *priorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	node := old[n-1]
	node.index = -1
	*pq = old[:n-1]
	return node
}

func (pq *priorityQueue) update(nd *node, freq int) {
	nd.freq = freq
	heap.Fix(pq, nd.index)
}

func freqCount(s string) map[rune]int {
	counts := make(map[rune]int)
	for _, r := range s {
		counts[r] += 1
	}
	return counts
}

func createPriorityQueue(m map[rune]int) priorityQueue {
	pq := make(priorityQueue, len(m))
	i := 0
	for value, freq := range m {
		pq[i] = &node{value, freq, nil, nil, i}
		i++
	}
	heap.Init(&pq)
	return pq
}

func buildTree(pq priorityQueue) *node {
	for pq.Len() > 1 {
		nd1 := heap.Pop(&pq).(*node)
		nd2 := heap.Pop(&pq).(*node)
		nn := &node{-1, nd1.freq + nd2.freq, nd1, nd2, 0}
		heap.Push(&pq, nn)
	}
	return heap.Pop(&pq).(*node)
}

func dumpCode(root *node) {
	var iter func(*node, string)
	iter = func(nd *node, code string) {
		if nd == nil {
			return
		}
		if nd.value != -1 {
			fmt.Printf("%c: %s\n", nd.value, code)
		}
		iter(nd.left, code+"0")
		iter(nd.right, code+"1")
	}
	iter(root, "")
}

func main() {
	pq := createPriorityQueue(freqCount("aaaaabccddd"))
	root := buildTree(pq)
	dumpCode(root)
}
