package main

import (
	"fmt"
)

func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

func LevenshteinDistance(a, b string) int {
	n := len(a)
	m := len(b)
	dp := make([][]int, n+1)
	for i := 0; i < n+1; i++ {
		dp[i] = make([]int, m+1)
	}

	// init
	for i := 1; i < n+1; i++ {
		dp[i][0] = i
	}
	for i := 1; i < m+1; i++ {
		dp[0][i] = i
	}

	for i := 1; i < n+1; i++ {
		for j := 1; j < m+1; j++ {
			if a[i-1] == b[j-1] {
				dp[i][j] = dp[i-1][j-1]
			} else {
				dp[i][j] = min(min(dp[i-1][j], dp[i][j-1]), dp[i-1][j-1]) + 1
			}
		}
	}
	return dp[n][m]
}

func main() {
	display := func(a, b string) {
		fmt.Printf("%s, %s : %d\n", a, b, LevenshteinDistance(a, b))
	}

	display("kitten", "sitting")
	display("apple", "play")
}
