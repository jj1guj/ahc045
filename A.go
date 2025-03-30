package main

import (
	"bufio"
	"fmt"
	"math"
	"os"
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

func query(c []int, writer *bufio.Writer, scanner *bufio.Scanner) [][2]int {
	writer.WriteString(fmt.Sprintf("? %d %s\n", len(c), toBlankJoin(c)))
	writer.Flush()

	result := make([][2]int, 0, len(c)-1)
	for i := 0; i < len(c)-1; i++ {
		scanner.Scan()
		line := scanner.Text()
		parts := strings.Split(line, " ")
		a, _ := strconv.Atoi(parts[0])
		b, _ := strconv.Atoi(parts[1])
		result = append(result, [2]int{a, b})
	}
	return result
}

func distSquared(a []int, b []int) int {
	return (a[0]-b[0])*(a[0]-b[0]) + (a[1]-b[1])*(a[1]-b[1])
}

func answer(groups [][]int, edges [][][]int, writer *bufio.Writer) {
	writer.WriteString("!\n")
	for i := 0; i < len(groups); i++ {
		writer.WriteString(toBlankJoin(groups[i]) + "\n")
		for _, e := range edges[i] {
			writer.WriteString(toBlankJoin(e) + "\n")
		}
	}
	writer.Flush()
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	writer := bufio.NewWriter(os.Stdout)

	scanner.Scan()
	line := scanner.Text()
	parts := strings.Split(line, " ")
	N, _ := strconv.Atoi(parts[0])
	M, _ := strconv.Atoi(parts[1])
	_, _ = strconv.Atoi(parts[2]) // Qは使わない
	L, _ := strconv.Atoi(parts[3])
	_, _ = strconv.Atoi(parts[4]) // Wは使わない

	G := make([]int, M)
	C := make([][]int, N)
	C_coord := make([][]int, N)

	scanner.Scan()
	line = scanner.Text()
	parts = strings.Split(line, " ")
	for i := 0; i < M; i++ {
		G[i], _ = strconv.Atoi(parts[i])
	}

	for i := 0; i < N; i++ {
		C[i] = make([]int, 4)
		scanner.Scan()
		line = scanner.Text()
		parts = strings.Split(line, " ")
		for j := 0; j < 4; j++ {
			C[i][j], _ = strconv.Atoi(parts[j])
		}
	}

	// 各都市の仮の座標を算出する
	for i := 0; i < N; i++ {
		x := (C[i][0] + C[i][1]) / 2
		y := (C[i][2] + C[i][3]) / 2
		C_coord[i] = []int{x, y}
	}

	// 各都市間のあり得る最大の距離を算出する
	// TODO: 総当たりの4パターンを計算しているが高速化の余地あり
	D := make([][]int, N)
	for i := 0; i < N; i++ {
		D[i] = make([]int, N)
		for j := 0; j < N; j++ {
			if i == j {
				D[i][j] = 0
			} else if i < j {
				coord_i := [][]int{
					{C[i][0], C[i][2]},
					{C[i][0], C[i][3]},
					{C[i][1], C[i][2]},
					{C[i][1], C[i][3]},
				}
				coord_j := [][]int{
					{C[j][0], C[j][2]},
					{C[j][0], C[j][3]},
					{C[j][1], C[j][2]},
					{C[j][1], C[j][3]},
				}
				d_max := 0
				for ci := range coord_i {
					for cj := range coord_j {
						d_max = max(d_max, distSquared(coord_i[ci], coord_j[cj]))
					}
				}
				D[i][j] = d_max
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

	// 都市を選び、直前までに選ばれた各都市から最も近い都市を選ぶ
	C_selected := make([]bool, N)
	for i := range C_selected {
		C_selected[i] = false
	}
	groups := make([][]int, M)

	for _, g_id := range G_ids {
		g := G[g_id]
		slice := make([]int, g)
		// 最初の都市を選ぶ
		// 最初の都市は現状選べる最も近い都市との距離が最小である都市を選ぶ
		first_city_id := -1
		first_d_min := math.MaxInt
		ref := -1
		for i := 0; i < N; i++ {
			if !C_selected[i] {
				if ref == -1 {
					ref = i
				}
				for _, id := range D_ids[i] {
					if !C_selected[id] {
						if D[i][id] < first_d_min {
							first_d_min = D[i][id]
							first_city_id = i
							break
						}
					}
				}
			}
		}
		if first_city_id == -1 {
			first_city_id = ref
		}
		slice[0] = first_city_id
		C_selected[first_city_id] = true
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
		groups[g_id] = slice
	}

	edges := [][][]int{}
	for k := 0; k < M; k++ {
		edges = append(edges, [][]int{})
		for i := 0; i < G[k]-1; i += L - 1 {
			if i+L <= G[k] {
				ret := query(groups[k][i:i+L], writer, scanner)
				for j := 0; j < len(ret); j++ {
					edges[k] = append(edges[k], ret[j][:])
				}
			} else if G[k]-i >= 2 {
				ret := query(groups[k][i:G[k]], writer, scanner)
				for j := 0; j < len(ret); j++ {
					edges[k] = append(edges[k], ret[j][:])
				}
			} else {
				edges[k] = append(edges[k], groups[k][i:i+2])
			}
		}
	}
	answer(groups, edges, writer)
}
