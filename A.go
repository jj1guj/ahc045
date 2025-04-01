package main

import (
	"bufio"
	"fmt"
	"math"
	"math/rand"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

func toBlankJoin(arr *[]int) string {
	strArr := make([]string, len(*arr))
	for i, v := range *arr {
		strArr[i] = strconv.Itoa(v)
	}
	return strings.Join(strArr, " ")
}

func query(c *[]int, writer *bufio.Writer, scanner *bufio.Scanner) [][2]int {
	writer.WriteString(fmt.Sprintf("? %d %s\n", len(*c), toBlankJoin(c)))
	writer.Flush()

	result := make([][2]int, 0, len(*c)-1)
	for i := 0; i < len(*c)-1; i++ {
		scanner.Scan()
		line := scanner.Text()
		parts := strings.Split(line, " ")
		a, _ := strconv.Atoi(parts[0])
		b, _ := strconv.Atoi(parts[1])
		result = append(result, [2]int{a, b})
	}
	return result
}

func distSquared(a *[]int, b *[]int) int {
	return ((*a)[0]-(*b)[0])*((*a)[0]-(*b)[0]) + ((*a)[1]-(*b)[1])*((*a)[1]-(*b)[1])
}

func sort_group(group *[]int, D *[][]int, d_sum *int) []int {
	g := len(*group)
	ret_group := make([]int, g)
	first_city_id := -1
	second_city_id := -1
	d_min := math.MaxInt
	marked := map[int]bool{}
	for _, i := range *group {
		marked[i] = false
		for _, j := range *group {
			if i != j {
				if (*D)[i][j] < d_min {
					d_min = (*D)[i][j]
					first_city_id = i
					second_city_id = j
				}
			}
		}
	}

	ret_group[0] = first_city_id
	ret_group[1] = second_city_id
	(*d_sum) += d_min
	marked[first_city_id] = true
	marked[second_city_id] = true
	for i := 2; i < g; i++ {
		d_min = math.MaxInt
		city_id := -1
		for j := 0; j < i; j++ {
			for _, id := range *group {
				if !marked[id] {
					if (*D)[ret_group[j]][id] < d_min {
						d_min = (*D)[ret_group[j]][id]
						city_id = id
					}
				}
			}
		}
		ret_group[i] = city_id
		marked[city_id] = true
		(*d_sum) += d_min
	}
	return ret_group
}

func answer(groups *[][]int, edges *[][][]int, writer *bufio.Writer) {
	writer.WriteString("!\n")
	for i := 0; i < len(*groups); i++ {
		writer.WriteString(toBlankJoin(&(*groups)[i]) + "\n")
		for _, e := range (*edges)[i] {
			writer.WriteString(toBlankJoin(&e) + "\n")
		}
	}
	writer.Flush()
}

func main() {
	start := time.Now()
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
				D[i][j] = distSquared(&C_coord[i], &C_coord[j])
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

	// 考えられるコストを記録する
	d_sum := 0
	d_sum_list := make([]int, M)

	for _, g_id := range G_ids {
		g := G[g_id]
		slice := make([]int, g)
		// 最初の都市を選ぶ
		// 最初の都市は現状選べる最も近い都市との距離が最小である都市を選ぶ
		first_city_id := -1
		second_city_id := -1
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
							second_city_id = id
							break
						}
					}
				}
			}
		}
		if first_city_id == -1 {
			first_city_id = ref
			d_sum_list[g_id] = 0
			slice[0] = first_city_id
			C_selected[first_city_id] = true
		} else {
			slice[0] = first_city_id
			d_sum_list[g_id] = 0
			C_selected[first_city_id] = true
			if g > 1 {
				d_sum_list[g_id] += first_d_min
				slice[1] = second_city_id
				C_selected[second_city_id] = true
				d_sum += first_d_min
			}
		}

		for idx := 2; idx < g; idx++ {
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
			d_sum_list[g_id] += d_min
			slice[idx] = city_id
			d_sum += d_min
			C_selected[city_id] = true
		}
		groups[g_id] = slice
	}

	if M > 1 {
		cnt := 0
		for time.Since(start) <= 1500*time.Millisecond {
			cnt += 1
			g1 := rand.Intn(M)
			g2 := rand.Intn(M)
			for g1 == g2 {
				g2 = rand.Intn(M)
			}
			g1_idx := rand.Intn(G[g1])
			g2_idx := rand.Intn(G[g2])

			tmp := groups[g1][g1_idx]
			groups[g1][g1_idx] = groups[g2][g2_idx]
			groups[g2][g2_idx] = tmp

			rec := d_sum
			var g1_new, g2_new []int
			var rec_g1, rec_g2 int
			if G[g1] > 1 {
				rec -= d_sum_list[g1]
				rec_g1 = 0
				g1_new = sort_group(&groups[g1], &D, &rec_g1)
				rec += rec_g1
			}

			if G[g2] > 1 {
				rec -= d_sum_list[g2]
				rec_g2 = 0
				g2_new = sort_group(&groups[g2], &D, &rec_g2)
				rec += rec_g2
			}

			if rec < d_sum {
				d_sum = rec
				if G[g1] > 1 {
					groups[g1] = g1_new
					d_sum_list[g1] = rec_g1
				}

				if G[g2] > 1 {
					groups[g2] = g2_new
					d_sum_list[g2] = rec_g2
				}
			} else {
				tmp := groups[g1][g1_idx]
				groups[g1][g1_idx] = groups[g2][g2_idx]
				groups[g2][g2_idx] = tmp
			}
		}
	}

	edges := [][][]int{}
	for k := 0; k < M; k++ {
		edges = append(edges, [][]int{})
		for i := 0; i < G[k]-1; i += L - 1 {
			if i+L <= G[k] {
				subSlice := groups[k][i : i+L]
				ret := query(&subSlice, writer, scanner)
				for j := 0; j < len(ret); j++ {
					edges[k] = append(edges[k], ret[j][:])
				}
			} else if G[k]-i >= 2 {
				subSlice := groups[k][i:G[k]]
				ret := query(&subSlice, writer, scanner)
				for j := 0; j < len(ret); j++ {
					edges[k] = append(edges[k], ret[j][:])
				}
			} else {
				edges[k] = append(edges[k], groups[k][i:i+2])
			}
		}
	}
	answer(&groups, &edges, writer)
}
