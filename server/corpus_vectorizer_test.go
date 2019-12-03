package main

import (
	"fmt"
	"strings"
	"testing"

	contextionary "github.com/semi-technologies/contextionary/contextionary/core"
	"github.com/semi-technologies/contextionary/extensions"
	"github.com/semi-technologies/contextionary/server/config"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_CorpusVectorizing_WithLinearWeighting(t *testing.T) {
	t.Run("with factor set to 0 - same as if there was no weighing", func(t *testing.T) {
		c11y := &fakeC11y{}
		swd := &fakeStopwordDetector{}
		config := &config.Config{
			OccurenceWeightLinearFactor: 0,
			MaxCompoundWordLength:       1,
		}
		split := &primitiveSplitter{}
		extensions := &fakeExtensionLookerUpper{}
		logger, _ := test.NewNullLogger()
		v := NewVectorizer(c11y, swd, config, logger, split, extensions)

		vector, err := v.Corpi([]string{"car is mercedes"})
		require.Nil(t, err)
		assert.Equal(t, []float32{1, 1, 0, 2}, vector.ToArray())
	})

	t.Run("with factor set to 1 - most skewed to rare values", func(t *testing.T) {
		c11y := &fakeC11y{}
		swd := &fakeStopwordDetector{}
		config := &config.Config{
			OccurenceWeightLinearFactor: 1,
			MaxCompoundWordLength:       1,
		}
		split := &primitiveSplitter{}
		logger, _ := test.NewNullLogger()
		extensions := &fakeExtensionLookerUpper{}
		v := NewVectorizer(c11y, swd, config, logger, split, extensions)

		vector, err := v.Corpi([]string{"car is mercedes"})
		require.Nil(t, err)
		assert.Equal(t, []float32{1, 0, 0, 4}, vector.ToArray())
	})

	t.Run("with factor set to 0,5 - some weighing takes place", func(t *testing.T) {
		c11y := &fakeC11y{}
		swd := &fakeStopwordDetector{}
		config := &config.Config{
			OccurenceWeightLinearFactor: 0.5,
			MaxCompoundWordLength:       1,
		}
		split := &primitiveSplitter{}
		logger := logrus.New()
		logger.SetLevel(logrus.DebugLevel)
		extensions := &fakeExtensionLookerUpper{}
		v := NewVectorizer(c11y, swd, config, logger, split, extensions)

		vector, err := v.Corpi([]string{"car is mercedes"})
		require.Nil(t, err)
		assert.Equal(t, []float32{1, 0.6677797, 0, 2.6644409}, vector.ToArray())
	})
}

func Test_CorpusVectorizing_WithCompoundWords(t *testing.T) {
	// these tests use weight factor 0, this makes the vector position
	// calculation a bit easier to understand, weighting itself is already
	// tested separately

	t.Run("with 2-word compound word 'fast car'", func(t *testing.T) {
		c11y := &fakeC11y{}
		swd := &fakeStopwordDetector{}
		config := &config.Config{
			OccurenceWeightLinearFactor: 0,
			MaxCompoundWordLength:       4,
		}
		split := &primitiveSplitter{}
		logger, _ := test.NewNullLogger()
		extensions := &fakeExtensionLookerUpper{}
		v := NewVectorizer(c11y, swd, config, logger, split, extensions)

		vector, err := v.Corpi([]string{"the mercedes is a fast car"})
		require.Nil(t, err)
		assert.Equal(t, []float32{0.5, 1, 1, 2.25}, vector.ToArray(),
			"vector position is the centroid of 'mercedes' and 'fast_car'")
	})

	t.Run("with multi-word compound word", func(t *testing.T) {
		c11y := &fakeC11y{}
		swd := &fakeStopwordDetector{}
		config := &config.Config{
			OccurenceWeightLinearFactor: 0,
			MaxCompoundWordLength:       4,
		}
		split := &primitiveSplitter{}
		logger, _ := test.NewNullLogger()
		extensions := &fakeExtensionLookerUpper{}
		v := NewVectorizer(c11y, swd, config, logger, split, extensions)

		vector, err := v.Corpi([]string{"the mercedes is like a formula 1 racing car"})
		require.Nil(t, err)
		assert.Equal(t, []float32{-1, 0, -1.5, 2}, vector.ToArray(),
			"vector position is the centroid of 'mercedes' and 'formula_1_racing_car'")
	})

	t.Run("with a single word right after a compound word", func(t *testing.T) {
		c11y := &fakeC11y{}
		swd := &fakeStopwordDetector{}
		config := &config.Config{
			OccurenceWeightLinearFactor: 0,
			MaxCompoundWordLength:       4,
		}
		split := &primitiveSplitter{}
		logger, _ := test.NewNullLogger()
		extensions := &fakeExtensionLookerUpper{}
		v := NewVectorizer(c11y, swd, config, logger, split, extensions)

		vector, err := v.Corpi([]string{"fast car mercedes"})
		require.Nil(t, err)
		assert.Equal(t, []float32{0.5, 1, 1, 2.25}, vector.ToArray(),
			"vector position is the centroid of 'mercedes' and 'fast_car'")
	})
}

func Test_CorpusVectorizing_WithCustomWords(t *testing.T) {
	// these tests use weight factor 0, this makes the vector position
	// calculation a bit easier to understand, weighting itself is already
	// tested separately

	t.Run("with single custom word 'zebra'", func(t *testing.T) {
		c11y := &fakeC11y{}
		swd := &fakeStopwordDetector{}
		config := &config.Config{
			OccurenceWeightLinearFactor: 0,
			MaxCompoundWordLength:       4,
		}
		split := &primitiveSplitter{}
		logger, _ := test.NewNullLogger()
		extensions := &fakeExtensionLookerUpper{}
		v := NewVectorizer(c11y, swd, config, logger, split, extensions)

		vector, err := v.Corpi([]string{"the mercedes is a zebra"})
		require.Nil(t, err)
		assert.Equal(t, []float32{0.5, 2, 0, 2}, vector.ToArray(),
			"vector position is the centroid of 'mercedes' and custom word 'zebra'")
	})

	t.Run("with 2-word custom word 'zebra carrier'", func(t *testing.T) {
		c11y := &fakeC11y{}
		swd := &fakeStopwordDetector{}
		config := &config.Config{
			OccurenceWeightLinearFactor: 0,
			MaxCompoundWordLength:       4,
		}
		split := &primitiveSplitter{}
		logger, _ := test.NewNullLogger()
		extensions := &fakeExtensionLookerUpper{}
		v := NewVectorizer(c11y, swd, config, logger, split, extensions)

		vector, err := v.Corpi([]string{"the mercedes is a zebra carrier"})
		require.Nil(t, err)
		assert.Equal(t, []float32{0.5, -2, 0, 2}, vector.ToArray(),
			"vector position is the centroid of 'mercedes' and custom word 'zebra carrier'")
	})
}

type fakeC11y struct{}

func (f *fakeC11y) GetNumberOfItems() int {
	panic("not implemented")
}

func (f *fakeC11y) GetVectorLength() int {
	panic("not implemented")
}

func (f *fakeC11y) WordToItemIndex(word string) contextionary.ItemIndex {
	if strings.Contains(word, "_") {
		// this is a compound word
		if word == "fast_car" {
			return 7
		}
		if word == "formula_1_racing_car" {
			return 8
		}
		return -1
	}

	switch word {
	case "car":
		return 5
	case "mercedes":
		return 6
	default:
		panic(fmt.Sprintf("no behavior for word '%s' in fake", word))
	}
}

func (f *fakeC11y) ItemIndexToWord(item contextionary.ItemIndex) (string, error) {
	panic("not implemented")
}

func (f *fakeC11y) ItemIndexToOccurrence(item contextionary.ItemIndex) (uint64, error) {
	switch item {
	case 5:
		return 20000, nil
	case 6:
		return 100, nil
	case 7:
		return 300, nil
	case 8:
		return 50, nil
	default:
		return 0, fmt.Errorf("no behavior for item %v in fake", item)
	}
}

func (f *fakeC11y) GetVectorForItemIndex(item contextionary.ItemIndex) (*contextionary.Vector, error) {
	switch item {
	case 5:
		v := contextionary.NewVector([]float32{1, 2, 0, 0})
		return &v, nil
	case 6:
		v := contextionary.NewVector([]float32{1, 0, 0, 4})
		return &v, nil
	case 7:
		v := contextionary.NewVector([]float32{0, 2, 2, 0.5})
		return &v, nil
	case 8:
		v := contextionary.NewVector([]float32{-3, 0, -3, 0})
		return &v, nil
	default:
		return nil, fmt.Errorf("no vector for item %v in fake", item)
	}
}

func (f *fakeC11y) GetDistance(a contextionary.ItemIndex, b contextionary.ItemIndex) (float32, error) {
	panic("not implemented")
}

func (f *fakeC11y) GetNnsByItem(item contextionary.ItemIndex, n int, k int) ([]contextionary.ItemIndex, []float32, error) {
	panic("not implemented")
}

func (f *fakeC11y) GetNnsByVector(vector contextionary.Vector, n int, k int) ([]contextionary.ItemIndex, []float32, error) {
	panic("not implemented")
}

func (f *fakeC11y) SafeGetSimilarWords(word string, n int, k int) ([]string, []float32) {
	panic("not implemented")
}

func (f *fakeC11y) SafeGetSimilarWordsWithCertainty(word string, certainty float32) []string {
	panic("not implemented")
}

type fakeStopwordDetector struct{}

func (f *fakeStopwordDetector) IsStopWord(word string) bool {
	return word == "is" || word == "the" || word == "a" || word == "like"
}

type primitiveSplitter struct{}

func (s *primitiveSplitter) Split(corpus string) []string {
	return strings.Split(corpus, " ")
}

type fakeExtensionLookerUpper struct{}

func (f *fakeExtensionLookerUpper) Lookup(word string) (*extensions.Extension, error) {
	switch word {
	case "zebra":
		return &extensions.Extension{
			Concept:    "zebra",
			Occurrence: 1000,
			Vector:     []float32{0, 4, 0, 0},
		}, nil
	case "zebra_carrier":
		return &extensions.Extension{
			Concept:    "zebra",
			Occurrence: 1000,
			Vector:     []float32{0, -4, 0, 0},
		}, nil
	default:
		return nil, nil
	}
}
