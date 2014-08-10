package main

import (
	"fmt"
)

var coins = []int{1, 5, 10, 50, 100, 500}

func printPattern(amount int, ptn []int, index int) {
	if amount == 0 {
		fmt.Println(ptn)
		return
	}
	if amount < 0 || index >= len(coins) {
		return
	}
	printPattern(amount-coins[index], append(ptn, coins[index]), index)
	printPattern(amount, ptn, index+1)
}

func main() {
	printPattern(18, []int{}, 0)
}
