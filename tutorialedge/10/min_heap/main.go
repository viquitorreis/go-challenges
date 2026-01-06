package main

import (
	"bufio"
	"container/heap"
	"fmt"
	"strings"
)

func main() {
	fmt.Println("Word Frequency Test")

	text := `Lorem Ipsum is simply dummy text of the printing and typesetting industry. Lorem Ipsum has been the industry's standard dummy text ever since the 1500s, when an unknown printer took a galley of type and scrambled it to make a type specimen book. It has survived not only five centuries, but also the leap into electronic typesetting, remaining essentially unchanged. It was popularised in the 1960s with the release of Letraset sheets containing Lorem Ipsum passages, and more recently with desktop publishing software like Aldus PageMaker including versions of Lorem Ipsum.`

	results := CountWords(text)
	MostCommon := Top5Words(results)

	fmt.Println(MostCommon)
}

func CountWords(text string) map[string]int {
	count := make(map[string]int, 0)

	scanner := bufio.NewScanner(strings.NewReader(text))
	scanner.Split(bufio.ScanWords)
	for scanner.Scan() {
		count[scanner.Text()]++
	}

	return count
}

func Top5Words(wordmap map[string]int) []word {
	maxHeap := &Word{}
	heap.Init(maxHeap)

	for k, v := range wordmap {
		heap.Push(maxHeap, word{
			w:    k,
			freq: v,
		})
	}

	top5 := []word{}

	for i := 0; i < 5; i++ {
		popped := heap.Pop(maxHeap).(word)
		// fmt.Printf("%s: %d\n", popped.w, popped.freq)
		top5 = append(top5, popped)
	}

	return top5
}

type word struct {
	w    string
	freq int
}

type Word []word

func (w *Word) Push(x any) {
	*w = append(*w, x.(word))
}

func (w *Word) Pop() any {
	old := *w
	n := len(old)
	x := (*w)[n-1]
	*w = (*w)[:n-1]
	return x
}

func (w *Word) Less(i, j int) bool {
	return (*w)[i].freq > (*w)[j].freq
}

func (w *Word) Len() int {
	return len(*w)
}

func (w *Word) Swap(i, j int) {
	(*w)[i], (*w)[j] = (*w)[j], (*w)[i]
}
