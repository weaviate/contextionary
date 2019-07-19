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
}

func NewVectorizer(c11y core.Contextionary, sw stopwordDetector,
	config *config.Config, logger logrus.FieldLogger) *Vectorizer {
	return &Vectorizer{
		c11y:             c11y,
		stopwordDetector: sw,
		config:           config,
		logger:           logger,
	}
}

func (cv *Vectorizer) Corpi(corpi []string) (*core.Vector, error) {
	var corpusVectors []core.Vector
	for i, corpus := range corpi {
		parts := strings.Split(corpus, " ")
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
	var vectors []core.Vector
	var occurrences []uint64

	for _, word := range words {
		vector, err := cv.vectorForWord(word)
		if err != nil {
			return nil, err
		}

		if vector == nil {
			continue
		}

		vectors = append(vectors, *vector.vector)
		occurrences = append(occurrences, vector.occurrence)
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
