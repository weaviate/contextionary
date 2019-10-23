package main

import (
	"fmt"
	"strings"

	core "github.com/semi-technologies/contextionary/contextionary/core"
	"github.com/semi-technologies/contextionary/server/config"
	"github.com/sirupsen/logrus"
)

type Vectorizer struct {
	c11y             core.Contextionary
	stopwordDetector stopwordDetector
	config           *config.Config
	logger           logrus.FieldLogger
	splitter         splitter
}

type splitter interface {
	Split(corpus string) []string
}

func NewVectorizer(c11y core.Contextionary, sw stopwordDetector,
	config *config.Config, logger logrus.FieldLogger, splitter splitter) *Vectorizer {
	return &Vectorizer{
		c11y:             c11y,
		stopwordDetector: sw,
		config:           config,
		splitter:         splitter,
		logger:           logger,
	}
}

func (cv *Vectorizer) Corpi(corpi []string) (*core.Vector, error) {
	var corpusVectors []core.Vector
	for i, corpus := range corpi {
		parts := cv.splitter.Split(corpus)
		if len(parts) == 0 {
			continue
		}

		v, err := cv.vectorForWordOrWords(parts)
		if err != nil {
			return nil, fmt.Errorf("at corpus %d: %v", i, err)
		}

		if v != nil {
			corpusVectors = append(corpusVectors, *v.vector)
		}
	}

	if len(corpusVectors) == 0 {
		return nil, fmt.Errorf("all words in corpus were either stopwords" +
			" or not present in the contextionary, cannot build vector")
	}

	vector, err := core.ComputeCentroid(corpusVectors)
	if err != nil {
		return nil, err
	}

	return vector, nil
}

func (cv *Vectorizer) vectorForWordOrWords(parts []string) (*vectorWithOccurrence, error) {
	if len(parts) > 1 {
		return cv.vectorForWords(parts)
	}

	return cv.vectorForWord(parts[0])
}

type vectorWithOccurrence struct {
	vector     *core.Vector
	occurrence uint64
}

func (cv *Vectorizer) vectorForWords(words []string) (*vectorWithOccurrence, error) {
	vectors, occurrences, err := cv.vectorsAndOccurrences(words)
	if err != nil {
		return nil, err
	}

	if len(vectors) == 0 {
		return nil, nil
	}

	weights := cv.occurencesToWeight(occurrences)
	centroid, err := core.ComputeWeightedCentroid(vectors, weights)
	if err != nil {
		return nil, err
	}

	return &vectorWithOccurrence{
		vector: centroid,
	}, nil
}

func (cv *Vectorizer) vectorsAndOccurrences(words []string) ([]core.Vector, []uint64, error) {
	var vectors []core.Vector
	var occurrences []uint64
	var debugOutput []string

	for i := 0; i < len(words); i++ {
		if (i + 1) < len(words) {
			// there is another (following) word in the corpus, so this could be a compound word
			compound := cv.compound(words[i], words[i+1])
			vector, err := cv.vectorForWord(compound)
			if err != nil {
				return nil, nil, err
			}

			if vector != nil {
				// this compound word exists, use it's vector and occurence
				vectors = append(vectors, *vector.vector)
				occurrences = append(occurrences, vector.occurrence)
				debugOutput = append(debugOutput, compound)

				// however, now we must make sure to skip the next word (right half)
				i++

				// and break the loop, so we don't index the left (half) again
				continue
			}

		}

		vector, err := cv.vectorForWord(words[i])
		if err != nil {
			return nil, nil, err
		}

		if vector == nil {
			continue
		}

		vectors = append(vectors, *vector.vector)
		occurrences = append(occurrences, vector.occurrence)
		debugOutput = append(debugOutput, words[i])
	}

	cv.logger.WithField("action", "vectorize_corpus").
		WithField("input", strings.Join(words, " ")).
		WithField("interpreted_as", strings.Join(debugOutput, " ")).
		Debug()

	return vectors, occurrences, nil
}

func (cv *Vectorizer) compound(words ...string) string {
	return strings.Join(words, "_")
}

func (cv *Vectorizer) vectorForWord(word string) (*vectorWithOccurrence, error) {
	if cv.stopwordDetector.IsStopWord(word) {
		return nil, nil
	}

	wi := cv.c11y.WordToItemIndex(word)
	if !wi.IsPresent() {
		return nil, nil
	}

	v, err := cv.c11y.GetVectorForItemIndex(wi)
	if err != nil {
		return nil, err
	}

	o, err := cv.c11y.ItemIndexToOccurrence(wi)
	if err != nil {
		return nil, err
	}

	return &vectorWithOccurrence{
		vector:     v,
		occurrence: o,
	}, nil
}

func (cv *Vectorizer) occurencesToWeight(occs []uint64) []float32 {
	factor := cv.config.OccurenceWeightLinearFactor
	max, min := maxMin(occs)

	weights := make([]float32, len(occs), len(occs))
	for i, occ := range occs {
		// w = 1 - ( (O - Omin) / (Omax - Omin) * s )
		weights[i] = 1 - ((float32(occ) - float32(min)) / float32(max-min) * factor)
	}

	return weights
}

func maxMin(input []uint64) (max uint64, min uint64) {
	for _, curr := range input {
		if curr < min {
			min = curr
		} else if curr > max {
			max = curr
		}
	}

	return
}
