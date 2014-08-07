package main

import (
	"fmt"
	"math/rand"
	"strings"
)

type Node struct {
	Value int
	Left  *Node
	Right *Node
}

func insert(node *Node, value int) *Node {
	if node == nil {
		return &Node{value, nil, nil}
	}
	if node.Value > value {
		node.Left = insert(node.Left, value)
	} else {
		node.Right = insert(node.Right, value)
	}
	return node
}

func find(node *Node, value int) *Node {
	if node == nil || node.Value == value {
		return node
	}
	if value > node.Value {
		return find(node.Right, value)
	}
	return find(node.Left, value)
}

func randomN(n int) *Node {
	var root *Node
	for _, i := range rand.Perm(n) {
		root = insert(root, i)
	}
	return root
}

// 4.7 Create an algorithm to decide if n2 is a subtree of n1
func isSubtree(n1, n2 *Node) bool {
	tn := find(n1, n2.Value)
	if tn == nil {
		return false
	}

	var iter func(*Node, *Node) bool
	iter = func(t1, t2 *Node) bool {
		if t1 == nil && t2 == nil {
			return true
		}
		if t1 == nil || t2 == nil {
			return false
		}
		if t1.Value != t2.Value {
			return false
		}
		return iter(t1.Left, t2.Left) && iter(t1.Right, t2.Right)
	}
	return iter(tn, n2)
}

// 4.8 Design an algorithm to print all paths which sum up to the given value.
// TODO: This prints a path multiple times. This shouldn't print a path more
// than once.
func printPathsSumUpTo(n *Node, sum int) {
	vs := make([]int, 0)
	var iter func(n *Node)
	iter = func(n *Node) {
		if n == nil {
			return
		}
		vs = append(vs, n.Value)
		tmp := 0
		for i := len(vs) - 1; i >= 0; i-- {
			tmp += vs[i]
			if tmp == sum {
				fmt.Println(vs[i:])
			}
		}
		iter(n.Left)
		iter(n.Right)
		vs = vs[:len(vs)-1]
	}
	iter(n)
}

func godump(n *Node) {
	ch := make(chan string)
	var iter func(*Node, int)
	iter = func(n *Node, depth int) {
		if n == nil {
			return
		}
		ch <- fmt.Sprintf("%s%d", strings.Repeat(" ", depth), n.Value)
		iter(n.Left, depth+1)
		iter(n.Right, depth+1)
	}
	go func() {
		iter(n, 0)
		close(ch)
	}()
	for s := range ch {
		fmt.Println(s)
	}
}

func dump(n *Node) {
	var iter func(*Node, int)
	iter = func(n *Node, depth int) {
		if n == nil {
			return
		}
		fmt.Printf("%s%d\n", strings.Repeat(" ", depth), n.Value)
		iter(n.Left, depth+1)
		iter(n.Right, depth+1)
	}
	iter(n, 0)
}

func main() {
	tmp := randomN(10)
	godump(tmp)

	fmt.Println("--- 4.7 ---")
	t2 := find(tmp, 4)
	t3 := randomN(3)
	fmt.Println(isSubtree(tmp, t2))
	fmt.Println(isSubtree(tmp, t3))

	fmt.Println("--- 4.8 ---")
	printPathsSumUpTo(tmp, 5)
}
