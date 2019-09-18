package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Splitter(t *testing.T) {
	type testcase struct {
		name   string
		input  string
		output []string
	}

	tests := []testcase{
		testcase{
			name:   "single word",
			input:  "single",
			output: []string{"single"},
		},
		testcase{
			name:   "words separated by space",
			input:  "hello my name is John",
			output: []string{"hello", "my", "name", "is", "John"},
		},
		testcase{
			name:   "multiple spaces in between words",
			input:  "hello     John",
			output: []string{"hello", "John"},
		},

		testcase{
			name:   "words with numbers",
			input:  "foo1 foo2",
			output: []string{"foo1", "foo2"},
		},

		testcase{
			name:   "hyphenated words",
			input:  "r2-d2",
			output: []string{"r2", "d2"},
		},

		testcase{
			name:   "on commas (with and without spaces)",
			input:  "jane, john,anna",
			output: []string{"jane", "john", "anna"},
		},

		testcase{
			name:   "on other characters",
			input:  "foobar baz#(*@@baq",
			output: []string{"foobar", "baz", "baq"},
		},

		testcase{
			name:   "words containing umlauts (upper and lower)",
			input:  "Ölpreis über 80 dollar!",
			output: []string{"Ölpreis", "über", "80", "dollar"},
		},

		testcase{
			name:   "words containing turkish characters",
			input:  "Ölpreis über 80 dollar!",
			output: []string{"Ölpreis", "über", "80", "dollar"},
		},

		testcase{
			name:   "words containing turkish characters",
			input:  "Weaviate ayrıca Türkçe konuşabilir",
			output: []string{"Weaviate", "ayrıca", "Türkçe", "konuşabilir"},
		},

		testcase{
			name:   "mixed characters including a '<'",
			input:  "car, car#of,,,,brand<mercedes, color!!blue",
			output: []string{"car", "car", "of", "brand", "mercedes", "color", "blue"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			out := NewSplitter().Split(test.input)
			assert.Equal(t, test.output, out, "output matches expected output")
		})

	}
}
