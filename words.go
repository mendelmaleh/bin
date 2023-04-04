package main

import (
	"bufio"
	"math/rand"
	"os"
	"strings"
	"time"
)

func lower(s string) bool {
	for _, r := range s {
		if 'a' > r || r > 'z' {
			return false
		}
	}

	return true
}

type Words struct {
	words []string
	rand  *rand.Rand

	num, min, max int
}

func (w *Words) Code() string {
	code := make([]string, w.num)

	for i := range code {
		code[i] = w.words[w.rand.Intn(len(w.words))]
	}

	return strings.Join(code, "-")
}

func NewWords(path string) (*Words, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	w := Words{
		words: []string{},
		rand:  rand.New(rand.NewSource(time.Now().UnixNano())),

		num: 3, min: 3, max: 5,
	}

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if t := scanner.Text(); w.min <= len(t) && len(t) <= w.max && lower(t) {
			w.words = append(w.words, t)
		}
	}

	return &w, nil
}
