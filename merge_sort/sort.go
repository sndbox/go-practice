package main

import (
	"fmt"
)

var _ = fmt.Println

const minPartitionSize = 5

func insertionSort(xs []int) {
	n := len(xs)
	for i := 1; i < n; i++ {
		tmp := xs[i]
		j := i - 1
		for j >= 0 && xs[j] > tmp {
			xs[j+1] = xs[j]
			j--
		}
		xs[j+1] = tmp
	}
}

func mergeSort(xs []int) {
	if len(xs) < minPartitionSize {
		insertionSort(xs)
		return
	}
	m := len(xs) / 2
	a, b := xs[:m], xs[m:]
	mergeSort(a)
	mergeSort(b)
	// TODO: avoid copy
	t := make([]int, len(xs))
	i, j, k := 0, 0, 0
	for i < len(a) && j < len(b) {
		if a[i] < b[j] {
			t[k] = a[i]
			i++
		} else {
			t[k] = b[j]
			j++
		}
		k++
	}
	for i < len(a) {
		t[k] = a[i]
		i++
		k++
	}
	for j < len(b) {
		t[k] = b[j]
		j++
		k++
	}
	for i := 0; i < len(xs); i++ {
		xs[i] = t[i]
	}
}
