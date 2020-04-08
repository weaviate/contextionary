package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	pb "github.com/semi-technologies/contextionary/contextionary"
)

type Classifier struct {
	docIndex      int
	ranked        []wordWithDistance
	inputCorpus   string
	oldPrediction string
	control       string
	splitter      *Splitter
	tfidf         tfidfScorer
}

type tfidfScorer interface {
	Get(term string, doc int) float32
}

func NewClassifier(docIndex int, doc doc, tfidf tfidfScorer) *Classifier {
	splitter := NewSplitter()
	return &Classifier{
		splitter:      splitter,
		docIndex:      docIndex,
		inputCorpus:   doc.corpus,
		oldPrediction: doc.oldPrediction,
		control:       doc.control,
		tfidf:         tfidf,
	}
}

func (c *Classifier) Run() []percentile {
	words := c.splitter.Split(c.inputCorpus)
	distances := make([]wordWithDistance, len(words))

	for i, word := range words {
		word = strings.ToLower(word)
		dist, avgDist, prediction := minimumDistance(word, mainCategories)
		distances[i] = wordWithDistance{
			Distance:        dist,
			AverageDistance: avgDist,
			Prediction:      prediction,
			Word:            word,
			InformationGain: avgDist - dist,
		}
	}

	c.ranked = rank(distances)
	return c.makeNewPredictions(words)
}

func minimumDistance(word string, cats []mainCategory) (float32, float32, string) {
	var all []float32
	minimum := float32(1000000.00)
	var prediction string

	vec, err := c11y.VectorForWord(context.TODO(), &pb.Word{Word: word})
	if err != nil {
		return -10, -10, ""
	}

	wordVec := extractVector(vec)
	for _, cat := range cats {
		dist, err := cosineDist(wordVec, cat.Vector)
		if err != nil {
			log.Fatal(err)
		}

		all = append(all, dist)

		if dist < minimum {
			minimum = dist
			prediction = cat.Name
		}
	}

	return minimum, avg(all), prediction
}

func avg(in []float32) float32 {
	var sum float32
	for _, curr := range in {
		sum += curr
	}

	return sum / float32(len(in))
}

func (c *Classifier) makeNewPredictions(words []string) []percentile {
	out := make([]percentile, 10)
	for perc := 10; perc <= 10; perc += 10 {
		var newCorpus []string
		for _, word := range words {
			word = strings.ToLower(word)

			if score := c.tfidf.Get(word, c.docIndex); score < 0.0 {
				continue
			}

			if c.isInPercentile(perc, word) {
				newCorpus = append(newCorpus, word)
			}
		}

		if len(newCorpus) == 0 {
			// if we end up with 0 words, take the topmost single word instead
			newCorpus = words[0:3]
		}

		corpus := strings.ToLower(strings.Join(newCorpus, " "))
		vec, err := c11y.VectorForCorpi(context.Background(), &pb.Corpi{Corpi: []string{corpus}})
		if err != nil {
			log.Fatal(fmt.Errorf("%s with corpus: %s", err, corpus))
		}

		vector := extractVector(vec)

		var minimum = float32(100000)
		var prediction string
		for _, cat := range mainCategories {
			dist, err := cosineDist(vector, cat.Vector)
			if err != nil {
				log.Fatal(err)
			}

			if dist < minimum {
				minimum = dist
				prediction = cat.Name
			}
		}

		out[perc/10-1] = percentile{Percentile: perc, Prediction: prediction, Match: prediction == c.control}
	}

	return out
}

type percentile struct {
	Percentile int
	Prediction string
	Match      bool
}

func (c *Classifier) isInPercentile(percentage int, needle string) bool {
	cutoff := int(float32(percentage) / float32(100) * float32(len(c.ranked)))
	selection := c.ranked[:cutoff]

	for _, hay := range selection {
		if needle == hay.Word {
			return true
		}
	}

	return false
}
