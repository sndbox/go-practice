package main

// +build ignore

import (
	"fmt"
	"strings"
)

type Node struct {
	IsTail   bool
	Children map[byte]*Node
}

type Trie struct {
	Root *Node
}

func (t *Trie) Add(s string) {
	bs := []byte(s)
	current := t.Root
	for _, ch := range bs {
		if current.Children[ch] == nil {
			current.Children[ch] = NewNode()
		}
		current = current.Children[ch]
	}
	current.IsTail = true
}

func (t *Trie) Find(s string) bool {
	bs := []byte(s)
	current := t.Root
	for _, ch := range bs {
		if current.Children[ch] == nil {
			return false
		}
		current = current.Children[ch]
	}
	return current.IsTail
}

func NewNode() *Node {
	return &Node{false, make(map[byte]*Node)}
}

func NewTrie() *Trie {
	return &Trie{NewNode()}
}

// For Debug
func printTree(t *Trie) {
	var printFunc func(byte, *Node, int)
	printFunc = func(ch byte, node *Node, depth int) {
		indent := strings.Repeat(" ", depth)
		fmt.Printf("%s[%v]%t\n", indent, string(ch), node.IsTail)
		for k, child := range node.Children {
			printFunc(k, child, depth+2)
		}
	}
	printFunc(0, t.Root, 0)
}

func main() {
	t := NewTrie()
	t.Add("abr")
	t.Add("abra")
	t.Add("bbb")
	t.Add("abc")
	fmt.Println(t.Find("bb"))
	fmt.Println(t.Find("bbbb"))
	fmt.Println(t.Find("abr"))
	fmt.Println(t.Find("abra"))
	printTree(t)
}
