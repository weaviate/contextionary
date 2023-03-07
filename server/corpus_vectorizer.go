package main

import (
	"errors"
	"fmt"
	"math"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/sirupsen/logrus"
	core "github.com/weaviate/contextionary/contextionary/core"
	errortypes "github.com/weaviate/contextionary/errors"
	"github.com/weaviate/contextionary/extensions"
	"github.com/weaviate/contextionary/server/config"
)

type Vectorizer struct {
	c11y                 core.Contextionary
	stopwordDetector     stopwordDetector
	config               *config.Config
	logger               logrus.FieldLogger
	splitter             splitter
	extensions           extensionLookerUpper
	cache                *sync.Map
	cacheCount           int32
	compoundWordSplitter compoundSplitter
}

const (
	OccurrenceStrategyLog    = "log"
	OccurrenceStrategyLinear = "linear"
)

type splitter interface {
	Split(corpus string) []string
}

type compoundSplitter interface {
	Split(word string) ([]string, error)
}

type extensionLookerUpper interface {
	Lookup(concept string) (*extensions.Extension, error)
}

func NewVectorizer(c11y core.Contextionary, sw stopwordDetector,
	config *config.Config, logger logrus.FieldLogger,
	splitter splitter, extensions extensionLookerUpper,
	compoundWordSplitter compoundSplitter) (*Vectorizer, error) {

	v := &Vectorizer{
		c11y:                 c11y,
		stopwordDetector:     sw,
		config:               config,
		splitter:             splitter,
		logger:               logger,
		extensions:           extensions,
		cache:                &sync.Map{},
		compoundWordSplitter: compoundWordSplitter,
	}

	if err := v.validateConfig(); err != nil {
		return nil, errortypes.NewInvalidUserInputf(err.Error())
	}

	return v, nil
}

func (cv *Vectorizer) validateConfig() error {
	s := cv.config.OccurrenceWeightStrategy
	switch s {
	case OccurrenceStrategyLinear, OccurrenceStrategyLog:
		// valid
	default:
		return fmt.Errorf("invalid config option: occurrence weight strategy: uncrecoginzed strategy '%s'", s)
	}

	return nil
}

var ErrNoUsableWords = errors.New("all words in corpus were either stopwords" +
	" or not present in the contextionary, cannot build vector")

func (cv *Vectorizer) Corpi(corpi []string, weightOverrides map[string]string) (*core.Vector, error) {
	var corpusVectors []core.Vector
	if weightOverrides == nil {
		// so we don't have to do no nil checks down the line
		weightOverrides = map[string]string{}
	}

	var source []core.InputElement

	for i, corpus := range corpi {
		parts := cv.splitter.Split(corpus)
		if len(parts) == 0 {
			continue
		}

		v, err := cv.vectorForWordOrWords(parts, weightOverrides)
		if err != nil {
			return nil, fmt.Errorf("at corpus %d: %v", i, err)
		}

		if v != nil {
			corpusVectors = append(corpusVectors, *v.vector)
			source = append(source, v.source...)
		}
	}

	if len(corpusVectors) == 0 {
		return nil, ErrNoUsableWords
	}

	vector, err := core.ComputeCentroid(corpusVectors)
	if err != nil {
		return nil, err
	}

	vector.Source = source
	return vector, nil
}

func (cv *Vectorizer) vectorForWordOrWords(parts []string, overrides map[string]string) (*vectorWithOccurrence, error) {
	if len(parts) > 1 {
		return cv.vectorForWords(parts, overrides)
	}

	return cv.VectorForWord(parts[0])
}

type vectorWithOccurrence struct {
	vector     *core.Vector
	occurrence uint64
	source     []core.InputElement
}

func (cv *Vectorizer) vectorForWords(words []string, overrides map[string]string) (*vectorWithOccurrence, error) {
	vectors, occurrences, words, err := cv.vectorsAndOccurrences(words)
	if err != nil {
		return nil, err
	}

	if len(vectors) == 0 {
		return nil, nil
	}

	weights, weightsDebug, err := cv.occurrencesToWeight(occurrences, words, overrides)
	if err != nil {
		return nil, err
	}
	cv.debugOccurrenceWeighing(occurrences, weights, words, weightsDebug, overrides)
	weights32 := float64SliceTofloat32(weights)
	centroid, err := core.ComputeWeightedCentroid(vectors, weights32)
	if err != nil {
		return nil, err
	}

	return &vectorWithOccurrence{
		vector: centroid,
		source: buildVectorInputElements(words, weights, occurrences),
	}, nil
}

func buildVectorInputElements(words []string, weights []float64, occurrences []uint64) []core.InputElement {
	out := make([]core.InputElement, len(words))
	for i := range words {
		out[i].Concept = words[i]
		out[i].Weight = weights[i]
		out[i].Occurrence = occurrences[i]
	}

	return out
}

func float64SliceTofloat32(in []float64) []float32 {
	out := make([]float32, len(in), len(in))
	for i, v := range in {
		out[i] = float32(v)
	}
	return out
}

func (cv *Vectorizer) vectorsAndOccurrences(words []string) ([]core.Vector, []uint64, []string, error) {
	var vectors []core.Vector
	var occurrences []uint64
	var debugOutput []string

	for wordPos := 0; wordPos < len(words); wordPos++ {
	additionalWordLoop:
		for additionalWords := cv.config.MaxCompoundWordLength - 1; additionalWords >= 0; additionalWords-- {
			if (wordPos + additionalWords) < len(words) {
				// we haven't reached the end of the corpus yet, so this words plus the
				// next n additional words could still form a compound word, we need to
				// check.
				// Note that n goes all the way down to zero, so once we didn't find
				// any compound words, we're checking the individual word.
				compound := cv.compound(cv.nextWords(words, wordPos, additionalWords)...)
				vector, err := cv.VectorForWord(compound)
				if err != nil {
					return nil, nil, nil, err
				}

				if vector != nil {
					// this compound word exists, use its vector and occurrence
					vectors = append(vectors, *vector.vector)
					occurrences = append(occurrences, vector.occurrence)
					if len(vector.source) > 0 {
						compound = vector.source[0].Concept
					}
					debugOutput = append(debugOutput, compound)

					// however, now we must make sure to skip the additionalWords
					wordPos += additionalWords
					break additionalWordLoop
				}
			}
		}
	}

	cv.logger.WithField("action", "vectorize_corpus").
		WithField("input", strings.Join(words, " ")).
		WithField("interpreted_as", strings.Join(debugOutput, " ")).
		Debug()

	return vectors, occurrences, debugOutput, nil
}

func (cv *Vectorizer) nextWords(words []string, startPos int, additional int) []string {
	endPos := startPos + 1 + additional
	return words[startPos:endPos]
}

func (cv *Vectorizer) compound(words ...string) string {
	return strings.Join(words, "_")
}

func (cv *Vectorizer) VectorForWord(word string) (*vectorWithOccurrence, error) {
	ext, err := cv.extensions.Lookup(word)
	if err != nil {
		return nil, fmt.Errorf("lookup custom word: %s", err)
	}

	if ext == nil {
		return cv.vectorForLibraryWord(word)
	}

	return cv.vectorFromExtension(ext)
}

func (cv *Vectorizer) vectorForLibraryWord(word string) (*vectorWithOccurrence, error) {
	if cv.stopwordDetector.IsStopWord(word) {
		cv.logger.WithField("action", "vectorize_library_word").
			WithField("word", word).
			WithField("stopword", true).
			Debug("is stopword - skipping")

		return nil, nil
	}

	if int(cv.cacheCount) > cv.config.MaximumVectorCacheSize {
		before := time.Now()
		cv.logger.WithField("action", "vectorize_start_purge_cache").
			Debug("start purging vectorization cache")

		cv.cache.Range(func(key, value interface{}) bool {
			cv.cache.Delete(key)
			atomic.AddInt32(&cv.cacheCount, -1)
			return true
		})

		cv.logger.WithField("action", "vectorize_complete_purge_cache").
			WithField("took", time.Since(before)).
			Debug("complete purging vectorization cache")
	}
	cached, ok := cv.cache.Load(word)
	if ok {
		return cached.(*vectorWithOccurrence), nil
	}

	wi := cv.c11y.WordToItemIndex(word)
	if wi.IsPresent() {
		// create vector out of it
		v, o, err := cv.itemIndexToVectorAndOccurence(wi)
		if err != nil {
			return nil, err
		}

		cv.logger.WithField("action", "vectorize_library_word").
			WithField("word", word).
			WithField("stopword", false).
			WithField("present", true).
			WithField("compound", false).
			WithField("occurence", o).
			Debug("present including")

		return cv.newCachedVectorWithOccurence(word, v, o), nil
	}

	// Word not in contextioanry try to compound split
	compoundWords, splitterErr := cv.compoundWordSplitter.Split(word)
	if splitterErr != nil {
		return nil, splitterErr
	}
	if len(compoundWords) > 0 {
		compoundVector, err := cv.compoundToVectorWithOccurence(compoundWords)
		switch err.(type) {
		case nil:
			break
		case errortypes.NotFound:
			// Just don't return a vector
			return nil, nil
		default:
			return nil, err
		}
		cv.logger.WithField("action", "vectorize_library_word").
			WithField("word", word).
			WithField("stopword", false).
			WithField("present", true).
			WithField("compound", true).
			WithField("occurence", compoundVector.occurrence).
			Debug("present including")

		return compoundVector, nil
	}

	cv.logger.WithField("action", "vectorize_library_word").
		WithField("word", word).
		WithField("stopword", false).
		WithField("present", false).
		WithField("compound", false).
		Debug("not present - skipping")
	return nil, nil
}

func (cv *Vectorizer) compoundToVectorWithOccurence(words []string) (*vectorWithOccurrence, error) {

	vectors := []core.Vector{}
	occurenceSum := uint64(0)
	for _, word := range words {
		wi := cv.c11y.WordToItemIndex(word)
		if !wi.IsPresent() {
			cv.logger.WithFields(logrus.Fields{
				"compounds": words,
				"missing":   word,
			}).Error("compounds of compound word splitting where not available in the contextionary there seems to be a compatiblity issue - skippinng")
			return nil, errortypes.NewNotFoundf("compound missing")
		}
		v, o, err := cv.itemIndexToVectorAndOccurence(wi)
		if err != nil {
			return nil, err
		}
		vectors = append(vectors, *v)
		occurenceSum += o
	}

	// new occurence is average occurence
	occurenceAvg := occurenceSum / uint64(len(vectors))

	centroid, err := core.ComputeCentroid(vectors)
	if err != nil {
		return nil, err
	}
	return cv.newCachedVectorWithOccurence(strings.Join(words, ""), centroid, occurenceAvg, words...), nil
}

func (cv *Vectorizer) itemIndexToVectorAndOccurence(wi core.ItemIndex) (*core.Vector, uint64, error) {

	v, err := cv.c11y.GetVectorForItemIndex(wi)
	if err != nil {
		return nil, 0, err
	}

	o, err := cv.c11y.ItemIndexToOccurrence(wi)
	if err != nil {
		return nil, 0, err
	}

	return v, o, nil
}

func (cv *Vectorizer) newCachedVectorWithOccurence(word string, vector *core.Vector, occurence uint64, parts ...string) *vectorWithOccurrence {
	inputWord := word
	if len(parts) > 0 {
		allParts := strings.Join(parts, ", ")
		inputWord += " (" + allParts + ")"
	}

	vo := &vectorWithOccurrence{
		vector:     vector,
		occurrence: occurence,
		source: []core.InputElement{
			{
				Concept:    inputWord,
				Occurrence: occurence,
				Weight:     1,
			},
		},
	}

	cv.cache.Store(word, vo)
	atomic.AddInt32(&cv.cacheCount, 1)
	return vo
}

func (cv *Vectorizer) vectorFromExtension(ext *extensions.Extension) (*vectorWithOccurrence, error) {
	v := core.NewVector(ext.Vector)
	return &vectorWithOccurrence{
		vector:     &v,
		occurrence: uint64(ext.Occurrence),
		source: []core.InputElement{
			{
				Concept:    ext.Concept,
				Weight:     1,
				Occurrence: uint64(ext.Occurrence),
			},
		},
	}, nil
}

type weighingDebugInfo struct {
	Max uint64 `json:"max"`
	Min uint64 `json:"min"`
}

func (cv *Vectorizer) occurrencesToWeight(occs []uint64, words []string,
	overrides map[string]string) ([]float64, weighingDebugInfo, error) {
	max, min := maxMin(occs)
	var weigher func(uint64) float64

	switch cv.config.OccurrenceWeightStrategy {
	case OccurrenceStrategyLog:
		weigher = makeLogWeigher(min, max)
	case OccurrenceStrategyLinear:
		linFactor := cv.config.OccurrenceWeightLinearFactor
		weigher = makeLinWeigher(min, max, linFactor)
	default:
		panic(fmt.Sprintf("vectorizer config validation is broken, impossible option '%s'",
			cv.config.OccurrenceWeightStrategy))
	}

	weights := make([]float64, len(occs), len(occs))
	for i, occ := range occs {
		res := weigher(occ)
		if expr, ok := overrides[words[i]]; ok {
			calc, err := NewEvaluator(expr, res).Do()
			if err != nil {
				return nil, weighingDebugInfo{}, fmt.Errorf("override expression for '%s': '%s': %v", words[i], expr, err)
			}
			res = calc
		}

		weights[i] = res
	}

	return weights, weighingDebugInfo{max, min}, nil
}

func maxMin(input []uint64) (max uint64, min uint64) {
	if len(input) >= 1 {
		min = input[0]
	}

	for _, curr := range input {
		if curr < min {
			min = curr
		} else if curr > max {
			max = curr
		}
	}

	return
}

func makeLinWeigher(min, max uint64, factor float32) func(uint64) float64 {
	return func(occ uint64) float64 {
		// w = 1 - ( (O - Omin) / (Omax - Omin) * s )
		return 1 - ((float64(occ) - float64(min)) / float64(max-min) * float64(factor))
	}
}

func makeLogWeigher(min, max uint64) func(uint64) float64 {
	return func(occ uint64) float64 {
		// Note the 1.05 that's 1 + minimal weight of 0.05. This way, the most common
		// word is not removed entirely, but still weighted somewhat
		return 2 * (1.05 - (math.Log(float64(occ)) / math.Log(float64(max))))
	}
}

func (cv *Vectorizer) debugOccurrenceWeighing(occurrences []uint64, weights []float64,
	words []string, weightsDebug weighingDebugInfo, overrides map[string]string) {
	if !(len(occurrences) == len(weights) && len(weights) == len(words)) {
		cv.logger.
			WithField("action", "weigh_vectorized_occurrences").
			WithFields(logrus.Fields{
				"lenOccurrences": len(occurrences),
				"lenWeights":     len(weights),
				"lenWords":       len(words),
			}).Debug("sizes don't match")
	}

	type word struct {
		Occurrence         uint64  `json:"occurrence"`
		Weight             float64 `json:"weight"`
		Word               string  `json:"word"`
		Overriden          bool    `json:"overriden"`
		OverrideExpression string  `json:"overrideExpression"`
	}

	out := make([]word, len(occurrences), len(occurrences))
	for i := range words {
		expr, overr := overrides[words[i]]
		out[i] = word{
			Word:               words[i],
			Occurrence:         occurrences[i],
			Weight:             weights[i],
			Overriden:          overr,
			OverrideExpression: expr,
		}
	}

	cv.logger.
		WithField("action", "debug_vector_weights").
		WithField("parameters", weightsDebug).WithField("words", out).Debug()
}
