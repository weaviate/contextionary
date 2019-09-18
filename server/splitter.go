package main

import (
	"strings"
	"unicode"
)

func NewSplitter() *Splitter {
	return &Splitter{}
}

type Splitter struct{}

func (s *Splitter) Split(corpus string) []string {
	return strings.FieldsFunc(corpus, func(c rune) bool {
		return !unicode.IsLetter(c) && !unicode.IsNumber(c)
	})
}
