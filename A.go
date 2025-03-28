package main

import "fmt"

func main() {
	var N int
	var M int
	var Q int
	var L int
	var W int
	fmt.Scan(&N, &M, &Q, &L, &W)
	G := make([]int, M)
	C := make([][]int, N)
	for i := 0; i < M; i++ {
		fmt.Scan(&G[i])
	}

	for i := 0; i < N; i++ {
		l := make([]int, 4)
		for j := 0; j < 4; j++ {
			fmt.Scan(&l[j])
		}
		C[i] = l
	}
}
