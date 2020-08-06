package main

import (
	"fmt"
	"strings"
	"testing"

	"github.com/semi-technologies/contextionary/compoundsplitting"

	contextionary "github.com/semi-technologies/contextionary/contextionary/core"
	"github.com/semi-technologies/contextionary/extensions"
	"github.com/semi-technologies/contextionary/server/config"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_CorpusVectorizing_WithLogWeighting(t *testing.T) {
	t.Run("with a phrase", func(t *testing.T) {
		c11y := &fakeC11y{}
		swd := &fakeStopwordDetector{}
		config := &config.Config{
			OccurrenceWeightStrategy: OccurrenceStrategyLog,
			MaxCompoundWordLength:    1,
		}
		split := &primitiveSplitter{}
		extensions := &fakeExtensionLookerUpper{}
		logger := logrus.New()
		logger.SetLevel(logrus.DebugLevel)
		compoundSplitter := compoundsplitting.NewEmptyTestSplitter()
		v, err := NewVectorizer(c11y, swd, config, logger, split, extensions, compoundSplitter)
		require.Nil(t, err)

		vector, err := v.Corpi([]string{"car is mercedes"}, nil)
		require.Nil(t, err)
		assert.Equal(t, []float32{1, 0.15748154, 0, 3.685037}, vector.ToArray())
	})

	t.Run("with a and a weight override for all words", func(t *testing.T) {
		c11y := &fakeC11y{}
		swd := &fakeStopwordDetector{}
		config := &config.Config{
			OccurrenceWeightStrategy: OccurrenceStrategyLog,
			MaxCompoundWordLength:    1,
		}
		split := &primitiveSplitter{}
		extensions := &fakeExtensionLookerUpper{}
		logger := logrus.New()
		logger.SetLevel(logrus.DebugLevel)
		compoundSplitter := compoundsplitting.NewEmptyTestSplitter()
		v, err := NewVectorizer(c11y, swd, config, logger, split, extensions, compoundSplitter)
		require.Nil(t, err)

		overrides := map[string]string{
			"car":      "1",
			"mercedes": "3",
		}

		vector, err := v.Corpi([]string{"car is mercedes"}, overrides)
		require.Nil(t, err)
		assert.Equal(t, []float32{1, 0.5, 0, 3}, vector.ToArray())
	})

	t.Run("with a single word, no weighing should occurr", func(t *testing.T) {
		c11y := &fakeC11y{}
		swd := &fakeStopwordDetector{}
		config := &config.Config{
			OccurrenceWeightStrategy: OccurrenceStrategyLog,
			MaxCompoundWordLength:    1,
		}
		split := &primitiveSplitter{}
		extensions := &fakeExtensionLookerUpper{}
		logger := logrus.New()
		logger.SetLevel(logrus.DebugLevel)
		compoundSplitter := compoundsplitting.NewEmptyTestSplitter()
		v, err := NewVectorizer(c11y, swd, config, logger, split, extensions, compoundSplitter)
		require.Nil(t, err)

		vector, err := v.Corpi([]string{"mercedes"}, nil)
		require.Nil(t, err)
		assert.Equal(t, []float32{1, 0, 0, 4}, vector.ToArray())
	})
}

func Test_CorpusVectorizing_WithLinearWeighting(t *testing.T) {
	t.Run("with factor set to 0 - same as if there was no weighing", func(t *testing.T) {
		c11y := &fakeC11y{}
		swd := &fakeStopwordDetector{}
		config := &config.Config{
			OccurrenceWeightLinearFactor: 0,
			OccurrenceWeightStrategy:     OccurrenceStrategyLinear,
			MaxCompoundWordLength:        1,
		}
		split := &primitiveSplitter{}
		extensions := &fakeExtensionLookerUpper{}
		logger, _ := test.NewNullLogger()
		compoundSplitter := compoundsplitting.NewEmptyTestSplitter()
		v, err := NewVectorizer(c11y, swd, config, logger, split, extensions, compoundSplitter)
		require.Nil(t, err)

		vector, err := v.Corpi([]string{"car is mercedes"}, nil)
		require.Nil(t, err)
		assert.Equal(t, []float32{1, 1, 0, 2}, vector.ToArray())
	})

	t.Run("with factor set to 1 - most skewed to rare values", func(t *testing.T) {
		c11y := &fakeC11y{}
		swd := &fakeStopwordDetector{}
		config := &config.Config{
			OccurrenceWeightLinearFactor: 1,
			OccurrenceWeightStrategy:     OccurrenceStrategyLinear,
			MaxCompoundWordLength:        1,
		}
		split := &primitiveSplitter{}
		logger, _ := test.NewNullLogger()
		extensions := &fakeExtensionLookerUpper{}
		compoundSplitter := compoundsplitting.NewEmptyTestSplitter()
		v, err := NewVectorizer(c11y, swd, config, logger, split, extensions, compoundSplitter)
		require.Nil(t, err)

		vector, err := v.Corpi([]string{"car is mercedes"}, nil)
		require.Nil(t, err)
		assert.Equal(t, []float32{1, 0, 0, 4}, vector.ToArray())
	})

	t.Run("with factor set to 0,5 - some weighing takes place", func(t *testing.T) {
		c11y := &fakeC11y{}
		swd := &fakeStopwordDetector{}
		config := &config.Config{
			OccurrenceWeightLinearFactor: 0.5,
			OccurrenceWeightStrategy:     OccurrenceStrategyLinear,
			MaxCompoundWordLength:        1,
		}
		split := &primitiveSplitter{}
		logger := logrus.New()
		logger.SetLevel(logrus.DebugLevel)
		extensions := &fakeExtensionLookerUpper{}
		compoundSplitter := compoundsplitting.NewEmptyTestSplitter()
		v, err := NewVectorizer(c11y, swd, config, logger, split, extensions, compoundSplitter)
		require.Nil(t, err)

		vector, err := v.Corpi([]string{"car is mercedes"}, nil)
		require.Nil(t, err)
		assert.Equal(t, []float32{1, 0.6666667, 0, 2.6666667}, vector.ToArray())
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
			OccurrenceWeightLinearFactor: 0,
			OccurrenceWeightStrategy:     OccurrenceStrategyLinear,
			MaxCompoundWordLength:        4,
		}
		split := &primitiveSplitter{}
		logger := logrus.New()
		logger.SetLevel(logrus.DebugLevel)
		extensions := &fakeExtensionLookerUpper{}
		compoundSplitter := compoundsplitting.NewEmptyTestSplitter()
		v, err := NewVectorizer(c11y, swd, config, logger, split, extensions, compoundSplitter)
		require.Nil(t, err)

		vector, err := v.Corpi([]string{"the mercedes is a fast car"}, nil)
		require.Nil(t, err)
		assert.Equal(t, equalWeight(fastCarVector, mercedesVector), vector.ToArray(),
			"vector position is the centroid of 'mercedes' and 'fast_car'")
	})

	t.Run("with multi-word compound word", func(t *testing.T) {
		c11y := &fakeC11y{}
		swd := &fakeStopwordDetector{}
		config := &config.Config{
			OccurrenceWeightLinearFactor: 0,
			OccurrenceWeightStrategy:     OccurrenceStrategyLinear,
			MaxCompoundWordLength:        4,
		}
		split := &primitiveSplitter{}
		logger, _ := test.NewNullLogger()
		extensions := &fakeExtensionLookerUpper{}
		compoundSplitter := compoundsplitting.NewEmptyTestSplitter()
		v, err := NewVectorizer(c11y, swd, config, logger, split, extensions, compoundSplitter)
		require.Nil(t, err)

		vector, err := v.Corpi([]string{"the mercedes is like a formula 1 racing car"}, nil)
		require.Nil(t, err)
		assert.Equal(t, equalWeight(mercedesVector, formula1RacingCarVector), vector.ToArray(),
			"vector position is the centroid of 'mercedes' and 'formula_1_racing_car'")
	})

	t.Run("with a single word right after a compound word", func(t *testing.T) {
		c11y := &fakeC11y{}
		swd := &fakeStopwordDetector{}
		config := &config.Config{
			OccurrenceWeightLinearFactor: 0,
			OccurrenceWeightStrategy:     OccurrenceStrategyLinear,
			MaxCompoundWordLength:        4,
		}
		split := &primitiveSplitter{}
		logger, _ := test.NewNullLogger()
		extensions := &fakeExtensionLookerUpper{}
		compoundSplitter := compoundsplitting.NewEmptyTestSplitter()
		v, err := NewVectorizer(c11y, swd, config, logger, split, extensions, compoundSplitter)
		require.Nil(t, err)

		vector, err := v.Corpi([]string{"fast car mercedes"}, nil)
		require.Nil(t, err)
		assert.Equal(t, equalWeight(mercedesVector, fastCarVector), vector.ToArray(),
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
			OccurrenceWeightLinearFactor: 0,
			OccurrenceWeightStrategy:     OccurrenceStrategyLinear,
			MaxCompoundWordLength:        4,
		}
		split := &primitiveSplitter{}
		logger, _ := test.NewNullLogger()
		extensions := &fakeExtensionLookerUpper{}
		compoundSplitter := compoundsplitting.NewEmptyTestSplitter()
		v, err := NewVectorizer(c11y, swd, config, logger, split, extensions, compoundSplitter)
		require.Nil(t, err)

		vector, err := v.Corpi([]string{"the mercedes is a zebra"}, nil)
		require.Nil(t, err)
		assert.Equal(t, []float32{0.5, 2, 0, 2}, vector.ToArray(),
			"vector position is the centroid of 'mercedes' and custom word 'zebra'")
	})

	t.Run("with 2-word custom word 'zebra carrier'", func(t *testing.T) {
		c11y := &fakeC11y{}
		swd := &fakeStopwordDetector{}
		config := &config.Config{
			OccurrenceWeightLinearFactor: 0,
			OccurrenceWeightStrategy:     OccurrenceStrategyLinear,
			MaxCompoundWordLength:        4,
		}
		split := &primitiveSplitter{}
		logger, _ := test.NewNullLogger()
		extensions := &fakeExtensionLookerUpper{}
		compoundSplitter := compoundsplitting.NewEmptyTestSplitter()
		v, err := NewVectorizer(c11y, swd, config, logger, split, extensions, compoundSplitter)
		require.Nil(t, err)

		vector, err := v.Corpi([]string{"the mercedes is a zebra carrier"}, nil)
		require.Nil(t, err)
		assert.Equal(t, []float32{0.5, -2, 0, 2}, vector.ToArray(),
			"vector position is the centroid of 'mercedes' and custom word 'zebra carrier'")
	})
}

func Test_CorpusVectorizing_UnknownCompoundWords(t *testing.T) {
	// these tests use weight factor 0, this makes the vector position
	// calculation a bit easier to understand, weighting itself is already
	// tested separately

	t.Run("word not in contextionary and word not splittable", func(t *testing.T) {
		c11y := &fakeC11y{}
		swd := &fakeStopwordDetector{}
		config := &config.Config{
			OccurrenceWeightStrategy: OccurrenceStrategyLog,
			MaxCompoundWordLength:    1,
		}
		split := &primitiveSplitter{}
		extensions := &fakeExtensionLookerUpper{}
		logger := logrus.New()
		logger.SetLevel(logrus.DebugLevel)
		compoundSplitter := compoundsplitting.NewEmptyTestSplitter()
		v, err := NewVectorizer(c11y, swd, config, logger, split, extensions, compoundSplitter)
		require.Nil(t, err)
		_, err = v.Corpi([]string{"steammachine"}, nil)
		require.NotNil(t, err)

	})

	t.Run("word not in contextionary but compounds are", func(t *testing.T) {
		c11y := &fakeC11y{}
		swd := &fakeStopwordDetector{}
		config := &config.Config{
			OccurrenceWeightStrategy: OccurrenceStrategyLog,
			MaxCompoundWordLength:    1,
		}
		split := &primitiveSplitter{}
		extensions := &fakeExtensionLookerUpper{}
		logger := logrus.New()
		logger.SetLevel(logrus.DebugLevel)
		compoundSplitter := compoundsplitting.NewTestSplitter(map[string]float64{
			"steam":   1.0,
			"machine": 1.0,
		})
		v, err := NewVectorizer(c11y, swd, config, logger, split, extensions, compoundSplitter)
		require.Nil(t, err)
		vec, err := v.Corpi([]string{"steammachine"}, nil)
		require.Nil(t, err)
		require.NotNil(t, vec)
		assert.Equal(t, []float32{1, 0.5, 0, 0}, vec.ToArray())
	})

	t.Run("word in compounds but not in contextionary", func(t *testing.T) {
		// This should never be the case and shows that the contextionary is incompatible with the word count splitter
		// Maybe the contextionary index file got updated without the compound splitter gettinng updated?
		c11y := &fakeC11y{}
		swd := &fakeStopwordDetector{}
		config := &config.Config{
			OccurrenceWeightStrategy: OccurrenceStrategyLog,
			MaxCompoundWordLength:    1,
		}
		split := &primitiveSplitter{}
		extensions := &fakeExtensionLookerUpper{}
		logger := logrus.New()
		logger.SetLevel(logrus.DebugLevel)
		compoundSplitter := compoundsplitting.NewTestSplitter(map[string]float64{
			"roller": 1.0,
			"blade":  1.0,
		})
		v, err := NewVectorizer(c11y, swd, config, logger, split, extensions, compoundSplitter)
		require.Nil(t, err)
		_, err = v.Corpi([]string{"rollerblade"}, nil)
		require.NotNil(t, err)
	})

	// TODO add test that runs into splitting a compound word into two known words and builds a vector out of them

	// TODO test no compound splitting found
}

type fakeC11y struct{}

func (f *fakeC11y) GetNumberOfItems() int {
	panic("not implemented")
}

func (f *fakeC11y) GetVectorLength() int {
	panic("not implemented")
}

func (f *fakeC11y) OccurrencePercentile(foo int) uint64 {
	panic("not implemented")
}

const (
	notInContextionary     = -1
	machineIndex           = 10
	steamIndex             = 9
	formula1RacingCarIndex = 8
	fastCarIndex           = 7
	mercedesIndex          = 6
)

var (
	fastCarVector           = []float32{0, 2, 2, 0.5}
	mercedesVector          = []float32{1, 0, 0, 4}
	formula1RacingCarVector = []float32{-3, 0, -3, 0}
	steamVector             = []float32{1, 0, 0, 0}
	machineVector           = []float32{1, 1, 0, 0}
)

func (f *fakeC11y) WordToItemIndex(word string) contextionary.ItemIndex {
	if strings.Contains(word, "_") {
		// this is a compound word
		if word == "fast_car" {
			return fastCarIndex
		}
		if word == "formula_1_racing_car" {
			return formula1RacingCarIndex
		}
		return -1
	}

	switch word {
	case "car":
		return 5
	case "mercedes":
		return mercedesIndex
	case "steam":
		return steamIndex
	case "machine":
		return machineIndex
	case "steammachine", "rollerblade", "roller", "blade":
		return notInContextionary
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
	case mercedesIndex:
		return 100, nil
	case fastCarIndex:
		return 300, nil
	case formula1RacingCarIndex:
		return 50, nil
	case steamIndex:
		return 100, nil
	case machineIndex:
		return 60, nil
	default:
		return 0, fmt.Errorf("no behavior for item %v in fake", item)
	}
}

func (f *fakeC11y) GetVectorForItemIndex(item contextionary.ItemIndex) (*contextionary.Vector, error) {
	switch item {
	case 5:
		v := contextionary.NewVector([]float32{1, 2, 0, 0})
		return &v, nil
	case mercedesIndex:
		v := contextionary.NewVector(mercedesVector)
		return &v, nil
	case fastCarIndex:
		v := contextionary.NewVector(fastCarVector)
		return &v, nil
	case formula1RacingCarIndex:
		v := contextionary.NewVector(formula1RacingCarVector)
		return &v, nil
	case steamIndex:
		v := contextionary.NewVector(steamVector)
		return &v, nil
	case machineIndex:
		v := contextionary.NewVector(machineVector)
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

func equalWeight(vectors ...[]float32) []float32 {
	// no sanity checks as this will only be used in tests, we'll notice panics
	// then
	sums := make([]float32, len(vectors[0]), len(vectors[0]))
	for _, vector := range vectors {
		for i, element := range vector {
			sums[i] = sums[i] + element
		}
	}

	mean := make([]float32, len(sums), len(sums))
	for i := range sums {
		mean[i] = sums[i] / float32(len(vectors))
	}

	return mean
}
