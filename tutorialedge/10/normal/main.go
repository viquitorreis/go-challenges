package main

import (
	"bufio"
	"fmt"
	"sort"
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
	freq := make(map[string]int)

	scanner := bufio.NewScanner(strings.NewReader(text))
	scanner.Split(bufio.ScanWords)
	for scanner.Scan() {
		fmt.Printf("word is: %s\n", scanner.Text())
		freq[scanner.Text()]++
	}

	return freq
}

type Word struct {
	str  string
	freq int
	len  int
}

func Top5Words(wordmap map[string]int) []Word {
	// 1. converter o map em slice
	words := make([]Word, 0, len(wordmap))

	for word, freq := range wordmap {
		words = append(words, Word{
			str:  word,
			freq: freq,
			len:  len(word),
		})
	}

	// 2. ordenar por frequencia
	sort.Slice(words, func(i, j int) bool {
		return words[i].freq > words[j].freq // asc
	})

	// 3. retornar as top 5
	if len(words) == 5 {
		return words
	}

	return words[:5]
}
