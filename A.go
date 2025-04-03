package main

import (
	"bufio"
	"container/heap"
	"fmt"
	"math"
	"math/rand"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

var writer *bufio.Writer
var scanner *bufio.Scanner
var start_time time.Time

func toBlankJoin(arr *[]int) string {
	strArr := make([]string, len(*arr))
	for i, v := range *arr {
		strArr[i] = strconv.Itoa(v)
	}
	return strings.Join(strArr, " ")
}

func answer(groups *[][]int, edges *[][][]int) {
	writer.WriteString("!\n")
	for i := 0; i < len(*groups); i++ {
		writer.WriteString(toBlankJoin(&(*groups)[i]) + "\n")
		for _, e := range (*edges)[i] {
			writer.WriteString(toBlankJoin(&e) + "\n")
		}
	}
	writer.Flush()
}

func query(c *[]int) [][2]int {
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

func distSquared(a *[]float32, b *[]float32) float32 {
	return float32(math.Sqrt(float64(((*a)[0]-(*b)[0])*((*a)[0]-(*b)[0]) + ((*a)[1]-(*b)[1])*((*a)[1]-(*b)[1]))))
}

func project_coords(coords *[][]float32, bounds *[][]int) {
	for i := 0; i < len(*coords); i++ {
		x := (*coords)[i][0]
		y := (*coords)[i][1]
		if x < float32((*bounds)[i][0]) {
			x = float32((*bounds)[i][0])
		} else if x > float32((*bounds)[i][1]) {
			x = float32((*bounds)[i][1])
		}
		if y < float32((*bounds)[i][2]) {
			y = float32((*bounds)[i][2])
		} else if y > float32((*bounds)[i][3]) {
			y = float32((*bounds)[i][3])
		}
		(*coords)[i][0] = x
		(*coords)[i][1] = y
	}
}

func compute_gradients(coords *[][]float32, queries *[][]int, query_edges *[][][]int, anchor_set *map[int]struct{}, margin float32, anchor_weight float32, grads *[][]float32) {
	for q_idx, q := range *queries {
		edges := (*query_edges)[q_idx]
		for _, e := range edges {
			i := e[0]
			j := e[1]
			d_ij := distSquared(&(*coords)[i], &(*coords)[j])
			for _, k := range q {
				if k == i || k == j {
					continue
				}
				d_ik := distSquared(&(*coords)[i], &(*coords)[k])
				gap := d_ij + margin - d_ik
				if gap > 0 {
					_, ok_i := (*anchor_set)[i]
					_, ok_j := (*anchor_set)[j]
					ok := ok_i || ok_j
					var weight float32
					grad_ij := make([]float32, 2)
					grad_ik := make([]float32, 2)
					if ok {
						weight = anchor_weight
					} else {
						weight = 1.0
					}

					if d_ij > 1e-6 {
						grad_ij[0] = ((*coords)[i][0] - (*coords)[j][0]) / d_ij
						grad_ij[1] = ((*coords)[i][1] - (*coords)[j][1]) / d_ij
					} else {
						grad_ij[0] = 0.0
						grad_ij[1] = 0.0
					}

					if d_ik > 1e-6 {
						grad_ik[0] = ((*coords)[i][0] - (*coords)[k][0]) / d_ik
						grad_ik[1] = ((*coords)[i][1] - (*coords)[k][1]) / d_ik
					} else {
						grad_ik[0] = 0.0
						grad_ik[1] = 0.0
					}

					(*grads)[i][0] += weight * (grad_ij[0] - grad_ik[0])
					(*grads)[i][1] += weight * (grad_ij[1] - grad_ik[1])
					(*grads)[j][0] -= weight * grad_ij[0]
					(*grads)[j][1] -= weight * grad_ij[1]
					(*grads)[k][0] -= weight * grad_ik[0]
					(*grads)[k][1] -= weight * grad_ik[1]
				}
			}
		}
	}
}

func optimize_coords(coords *[][]float32, bounds *[][]int, queries *[][]int, query_edges *[][][]int, anchor_set *map[int]struct{}, lr float32, margin float32, anchor_weight float32, grads *[][]float32) {
	for time.Since(start_time) < 1000 * time.Millisecond {
		for i := 0; i < len(*grads); i++ {
			(*grads)[i][0] = 0.0
			(*grads)[i][1] = 0.0
		}
		compute_gradients(coords, queries, query_edges, anchor_set, margin, anchor_weight, grads)
		for i := 0; i < len(*coords); i++ {
			(*coords)[i][0] -= lr * (*grads)[i][0]
			(*coords)[i][1] -= lr * (*grads)[i][1]
		}
		project_coords(coords, bounds)
	}
}

func getRandomAnchor(anchor_set map[int]struct{}) int {
	anchor_ids := make([]int, 0, len(anchor_set))
	for id := range anchor_set {
		anchor_ids = append(anchor_ids, id)
	}
	return anchor_ids[rand.Intn(len(anchor_ids))]
}

func randomSample(slice []int, count int) []int {
	if count > len(slice) {
		panic("count is greater than the length of the slice")
	}

	// スライスをシャッフル
	rand.Seed(time.Now().UnixNano()) // シードを設定
	shuffled := make([]int, len(slice))
	copy(shuffled, slice)
	rand.Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})

	// 指定された数の要素を選択
	return shuffled[:count]
}

type Edge struct {
	from   int
	to     int
	weight float32
}

type EdgeHeap []Edge

func (h EdgeHeap) Len() int           { return len(h) }
func (h EdgeHeap) Less(i, j int) bool { return h[i].weight < h[j].weight }
func (h EdgeHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *EdgeHeap) Push(x interface{}) {
	*h = append(*h, x.(Edge))
}

func (h *EdgeHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

// プリム法による最小全域木の構築
func Prim(nodes *[]int, D *[][]float32) [][]int {
	// 隣接リストを構築
	adj := make([][]Edge, len(*nodes))
	for i := 0; i < len(*nodes); i++ {
		for j := 0; j < len(*nodes); j++ {
			if i != j {
				adj[i] = append(adj[i], Edge{from: i, to: j, weight: (*D)[(*nodes)[i]][(*nodes)[j]]})
			}
		}
	}

	marked := make([]bool, len(*nodes))
	for i := 0; i < len(*nodes); i++ {
		marked[i] = false
	}
	marked_cnt := 0
	heapq := &EdgeHeap{}
	heap.Init(heapq)
	marked[0] = true
	marked_cnt++
	for _, e := range adj[0] {
		heap.Push(heapq, e)
	}

	out_edge := make([][]int, 0)
	for marked_cnt < len(*nodes) {
		e := heap.Pop(heapq).(Edge)
		if marked[e.to] {
			continue
		}
		marked[e.to] = true
		marked_cnt++
		edge := []int{(*nodes)[e.from], (*nodes)[e.to]}
		out_edge = append(out_edge, edge)
		for _, e2 := range adj[e.to] {
			if !marked[e2.to] {
				heap.Push(heapq, e2)
			}
		}
	}
	return out_edge
}

func main() {
	start_time = time.Now()
	scanner = bufio.NewScanner(os.Stdin)
	writer = bufio.NewWriter(os.Stdout)

	scanner.Scan()
	line := scanner.Text()
	parts := strings.Split(line, " ")
	N, _ := strconv.Atoi(parts[0])
	M, _ := strconv.Atoi(parts[1])
	Q, _ := strconv.Atoi(parts[2])
	L, _ := strconv.Atoi(parts[3])
	_, _ = strconv.Atoi(parts[4]) // Wは使わない

	G := make([]int, M)
	C := make([][]int, N)
	C_coord := make([][]float32, N)

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
		x := (float32(C[i][0]) + float32(C[i][1])) / 2
		y := (float32(C[i][2]) + float32(C[i][3])) / 2
		C_coord[i] = []float32{x, y}
	}

	uncertainties := make([]int, N)
	uncertainties_id := make([]int, N)
	for i := 0; i < N; i++ {
		uncertainties[i] = (C[i][1] - C[i][0]) + (C[i][3] - C[i][2])
		uncertainties_id[i] = i
	}

	sort.Slice(uncertainties_id, func(i, j int) bool {
		return uncertainties[uncertainties_id[i]] < uncertainties[uncertainties_id[j]]
	})

	anchor_set := make(map[int]struct{})
	for i := 0; i < N/10; i++ {
		anchor_set[uncertainties_id[i]] = struct{}{}
	}

	id_without_anchor := make([]int, 0)
	for i := 0; i < N; i++ {
		if _, ok := anchor_set[i]; !ok {
			id_without_anchor = append(id_without_anchor, i)
		}
	}

	queries := make([][]int, Q)
	for i := 0; i < Q; i++ {
		query := make([]int, L)
		anchor := getRandomAnchor(anchor_set)
		others := randomSample(id_without_anchor, L-1)
		query[0] = anchor
		for idx := 0; idx < L-1; idx++ {
			query[idx+1] = others[idx]
		}
		queries[i] = query
	}
	query_edges := make([][][]int, Q)
	for i, q := range queries {
		edges := query(&q)
		query_edges[i] = make([][]int, len(edges))
		for j, e := range edges {
			query_edges[i][j] = []int{e[0], e[1]}
		}
	}

	grads := make([][]float32, N)
	for i := 0; i < N; i++ {
		grads[i] = make([]float32, 2)
	}
	optimize_coords(&C_coord, &C, &queries, &query_edges, &anchor_set, 0.001, 1.0, 2.0, &grads)

	// 各都市間の距離を算出する
	D := make([][]float32, N)
	for i := 0; i < N; i++ {
		D[i] = make([]float32, N)
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

	for _, g_id := range G_ids {
		g := G[g_id]
		slice := make([]int, g)
		// 最初の都市を選ぶ
		// 最初の都市は現状選べる最も近い都市との距離が最小である都市を選ぶ
		first_city_id := -1
		second_city_id := -1
		first_d_min := float32(math.MaxFloat32)
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
			slice[0] = first_city_id
			C_selected[first_city_id] = true
		} else {
			slice[0] = first_city_id
			C_selected[first_city_id] = true
			if g > 1 {
				slice[1] = second_city_id
				C_selected[second_city_id] = true
			}
		}

		for idx := 2; idx < g; idx++ {
			d_min := float32(math.MaxFloat32)
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

	edges := make([][][]int, M)
	for k := 0; k < M; k++ {
		e := Prim(&groups[k], &D)
		edges[k] = e
		edges = append(edges, [][]int{})
	}
	answer(&groups, &edges)
}
