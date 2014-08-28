package main

import (
	"fmt"
	"math/rand"
	"testing"
)

func check(xs []int) error {
	for i := 0; i < len(xs)-1; i++ {
		if xs[i] > xs[i+1] {
			return fmt.Errorf("at %d: %d > %d", i, xs[i], xs[i+1])
		}
	}
	return nil
}

func TestInsertionSort(t *testing.T) {
	xs := rand.Perm(30)
	insertionSort(xs)
	if err := check(xs); err != nil {
		t.Error(err)
	}
}

func TestMergeSort(t *testing.T) {
	xs := rand.Perm(30)
	mergeSort(xs)
	if err := check(xs); err != nil {
		t.Error(err)
	}
}
