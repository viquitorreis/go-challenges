package main

// heap para as ordens de compra
// max-heap -> maior preço sempre no topo
type bidsHeap []int

func (m *bidsHeap) Push(x any) {
	*m = append(*m, x.(int))
}

func (m *bidsHeap) Pop() any {
	old := *m
	size := len(*m)
	x := old[size-1]
	*m = old[:size-1]
	return x
}

func (m *bidsHeap) Less(i, j int) bool {
	return (*m)[i] > (*m)[j]
}

func (m *bidsHeap) Swap(i, j int) {
	(*m)[i], (*m)[j] = (*m)[j], (*m)[i]
}

func (m *bidsHeap) Len() int {
	return len(*m)
}

// heap para as ordens de venda
// min-heap -> menor preço sempre no topo
type asksHeap []int

func (m *asksHeap) Push(x any) {
	*m = append(*m, x.(int))
}

func (m *asksHeap) Pop() any {
	old := *m
	size := len(*m)
	x := old[size-1]
	*m = old[:size-1]
	return x
}

func (m *asksHeap) Less(i, j int) bool {
	return (*m)[i] < (*m)[j]
}

func (m *asksHeap) Swap(i, j int) {
	(*m)[i], (*m)[j] = (*m)[j], (*m)[i]
}

func (m *asksHeap) Len() int {
	return len(*m)
}
