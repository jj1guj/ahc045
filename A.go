package main

import (
	"fmt"
	"math"
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

func distSquared(a []int, b []int) int {
	return (a[0]-b[0])*(a[0]-b[0]) + (a[1]-b[1])*(a[1]-b[1])
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

	// 各都市間の距離を算出する
	D := make([][]int, N)
	for i := 0; i < N; i++ {
		D[i] = make([]int, N)
		for j := 0; j < N; j++ {
			if i <= j {
				D[i][j] = distSquared(C_coord[i], C_coord[j])
			} else {
				D[i][j] = D[j][i]
			}
		}
	}

	// 各都市について距離が近い順にソートする
	D_ids := make([][]int, N)
	for i := 0; i < N; i++ {
		ids := make([]int, N-1)
		idx := 0
		for j := 0; j < N; j++ {
			if i != j {
				ids[idx] = j
				idx++
			}
		}

		sort.Slice(ids, func(a, b int) bool {
			return D[i][ids[a]] < D[i][ids[b]]
		})
		D_ids[i] = ids
	}

	// グループ内の都市数が多い順にソートする
	G_ids := make([]int, M)
	for i := 0; i < M; i++ {
		G_ids[i] = i
	}
	sort.Slice(G_ids, func(a, b int) bool {
		return G[G_ids[a]] > G[G_ids[b]]
	})

	// 入力順に都市を選び、直前までに選ばれた各都市から最も近い都市を選ぶ
	C_selected := make([]bool, N)
	for i := range C_selected {
		C_selected[i] = false
	}
	groups := make([][]int, M)

	for _, g_id := range G_ids {
		g := G[g_id]
		slice := make([]int, g)
		for i := 0; i < N; i++ {
			if !C_selected[i] {
				slice[0] = i
				C_selected[i] = true
				for idx := 1; idx < g; idx++ {
					d_min := math.MaxInt
					city_id := -1
					for j := 0; j < idx; j++ {
						for _, id := range D_ids[slice[j]] {
							if !C_selected[id] {
								if D[slice[j]][id] < d_min {
									d_min = D[slice[j]][id]
									city_id = id
								}
							}
						}
					}
					slice[idx] = city_id
					C_selected[city_id] = true
				}
				break
			}
		}
		groups[g_id] = slice
	}

	edges := [][][]int{}
	for k := 0; k < M; k++ {
		edges = append(edges, [][]int{})
		for i := 0; i < G[k]-1; i += L - 1 {
			if i+L <= G[k] {
				ret := query(groups[k][i : i+L])
				for j := 0; j < len(ret); j++ {
					edges[k] = append(edges[k], ret[j][:])
				}
			} else if G[k]-i >= 2 {
				ret := query(groups[k][i:G[k]])
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
