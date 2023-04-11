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

	c WordsConfig
}

type WordsConfig struct {
	Dict string // dictonary file for the word generator

	Num, Min, Max int
}

func (w *Words) Code() string {
	code := make([]string, w.c.Num)

	for i := range code {
		code[i] = w.words[w.rand.Intn(len(w.words))]
	}

	return strings.Join(code, "-")
}

func NewWords(config WordsConfig) (*Words, error) {
	f, err := os.Open(config.Dict)
	if err != nil {
		return nil, err
	}

	w := Words{
		words: []string{},
		rand:  rand.New(rand.NewSource(time.Now().UnixNano())),

		c: config,
	}

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if t := scanner.Text(); w.c.Min <= len(t) && len(t) <= w.c.Max && lower(t) {
			w.words = append(w.words, t)
		}
	}

	return &w, nil
}
