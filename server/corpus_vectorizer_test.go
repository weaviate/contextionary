package main

import (
	"fmt"
	"testing"

	contextionary "github.com/semi-technologies/contextionary/contextionary/core"
	"github.com/semi-technologies/contextionary/server/config"
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
		}
		logger, _ := test.NewNullLogger()
		v := NewVectorizer(c11y, swd, config, logger)

		vector, err := v.Corpi([]string{"car is mercedes"})
		require.Nil(t, err)
		assert.Equal(t, []float32{1, 1, 0, 2}, vector.ToArray())
	})

	t.Run("with factor set to 1 - most skewed to rare values", func(t *testing.T) {
		c11y := &fakeC11y{}
		swd := &fakeStopwordDetector{}
		config := &config.Config{
			OccurenceWeightLinearFactor: 1,
		}
		logger, _ := test.NewNullLogger()
		v := NewVectorizer(c11y, swd, config, logger)

		vector, err := v.Corpi([]string{"car is mercedes"})
		require.Nil(t, err)
		assert.Equal(t, []float32{1, 0, 0, 4}, vector.ToArray())
	})

	t.Run("with factor set to 0,5 - some weighing takes place", func(t *testing.T) {
		c11y := &fakeC11y{}
		swd := &fakeStopwordDetector{}
		config := &config.Config{
			OccurenceWeightLinearFactor: 0.5,
		}
		logger, _ := test.NewNullLogger()
		v := NewVectorizer(c11y, swd, config, logger)

		vector, err := v.Corpi([]string{"car is mercedes"})
		require.Nil(t, err)
		assert.Equal(t, []float32{1, 0.6677797, 0, 2.6644409}, vector.ToArray())
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
	return word == "is"
}
