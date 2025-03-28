package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

func toBlankJoin(arr []int) string {
	strArr := make([]string, len(arr))
	for i, v := range arr {
		strArr[i] = strconv.Itoa(v)
	}
	return strings.Join(strArr, " ")
}

func query(c []int) [][2]int {
	fmt.Println("?", len(c), toBlankJoin(c))
	result := make([][2]int, 0, len(c)-1)

	for i := 0; i < len(c)-1; i++ {
		var a, b int
		fmt.Scan(&a, &b) // 2つの整数を入力
		result = append(result, [2]int{a, b})
	}
	return result
}

func answer(groups [][]int, edges [][][]int) {
	fmt.Println("!")
	for i := 0; i < len(groups); i++ {
		fmt.Println(toBlankJoin(groups[i]))
		for _, e := range edges[i] {
			fmt.Println(toBlankJoin(e))
		}
	}
}

func main() {
	var N, M, Q, L, W int
	fmt.Scan(&N, &M, &Q, &L, &W)
	G := make([]int, M)
	C := make([][]int, N)
	C_coord := make([][]int, N)
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

	// 各都市の仮の座標を算出する
	for i := 0; i < N; i++ {
		x := (C[i][0] + C[i][1]) / 2
		y := (C[i][2] + C[i][3]) / 2
		C_coord[i] = []int{x, y}
	}

	C_ids := make([]int, N)
	for i := 0; i < N; i++ {
		C_ids[i] = i
	}

	// 仮の座標準に都市をソート
	sort.Slice(C_ids, func(i, j int) bool {
		if C_coord[C_ids[i]][0] != C_coord[C_ids[j]][0] {
			return C_coord[C_ids[i]][0] < C_coord[C_ids[j]][0]
		}
		return C_coord[C_ids[i]][1] < C_coord[C_ids[j]][1]
	})

	// ソートした順に都市を分割
	groups := [][]int{}
	start_idx := 0
	for _, g := range G {
		slice := C_ids[start_idx : start_idx+g]
		groups = append(groups, slice)
		start_idx += g
	}

	edges := [][][]int{}
	for k := 0; k < M; k++ {
		edges = append(edges, [][]int{})
		for i := 0; i < G[k]-1; i += 2 {
			if i < G[k]-2 {
				ret := query(groups[k][i : i+3])
				for j := 0; j < len(ret); j++ {
					edges[k] = append(edges[k], ret[j][:])
				}
			} else {
				edges[k] = append(edges[k], groups[k][i:i+2])
			}
		}
	}
	answer(groups, edges)
}
