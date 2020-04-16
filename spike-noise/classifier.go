package main

import (
	"context"
	"fmt"
	"log"
	"math"
	"sort"
	"strings"
	"sync"

	"github.com/davecgh/go-spew/spew"
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
	sync.Mutex
}

type tfidfScorer interface {
	// Get(term string, doc int) float32
	GetAllTerms(docIndex int) []TermWithTfIdf
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

	concurrent := 20
	for i := 0; i < len(words); i += concurrent {
		end := i + concurrent
		if end > len(words) {
			end = len(words)
		}

		batch := words[i:end]

		var wg = &sync.WaitGroup{}
		for j, elem := range batch {
			wg.Add(1)
			go func(i, j int, elem string) {

				word := strings.ToLower(elem)
				dist, avgDist, prediction := minimumDistance(word, targets)

				c.Lock()
				distances[i+j] = wordWithDistance{
					Distance:        dist,
					AverageDistance: avgDist,
					Prediction:      prediction,
					Word:            word,
					InformationGain: avgDist - dist,
				}
				c.Unlock()

				wg.Done()
			}(i, j, elem)
		}

		wg.Wait()
	}

	c.ranked = rank(distances)
	return c.makeNewPredictions(words)
}

func rank(in []wordWithDistance) []wordWithDistance {
	i := 0
	filtered := make([]wordWithDistance, len(in))
	for _, w := range in {
		if w.Distance < -9 {
			continue
		}

		filtered[i] = w
		i++
	}
	out := filtered[:i]
	sort.Slice(out, func(a, b int) bool { return out[a].InformationGain > out[b].InformationGain })

	// simple dedup since it's already ordered, we only need to check the previous element
	indexOut := 0
	dedupped := make([]wordWithDistance, len(out))
	for i, elem := range out {
		if i == 0 {
			dedupped[indexOut] = elem
			indexOut++
			continue
		}

		if elem.Word == out[i-1].Word {
			continue
		}

		dedupped[indexOut] = elem
		indexOut++
	}

	return dedupped[:indexOut]
}

func minimumDistance(word string, cats []target) (float32, float32, string) {
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
	perc := 8
	out := make([]percentile, 1)
	var newCorpus []string
	tfscores := c.tfidf.GetAllTerms(c.docIndex)
	for _, word := range words {
		word = strings.ToLower(word)

		tfscore := findTfScore(tfscores, word)
		if tfscore < -1 {
			continue
		}

		if c.isInPercentile(perc, word) && c.isInTfPercentile(tfscores, 80, word) {
			newCorpus = append(newCorpus, word)
		}
	}

	if len(newCorpus) == 0 {
		// if we end up with 0 words, take the topmost single word instead
		limit := 3
		if len(c.ranked) < limit {
			limit = len(c.ranked)
		}

		newCorpus = make([]string, limit+1)
		for i := 0; i <= limit; i++ {
			newCorpus[i] = c.ranked[i].Word
		}
	}

	tfTopWords := 2
	if len(tfscores) < tfTopWords {
		tfTopWords = len(c.ranked)
	}
	for i := 0; i < tfTopWords; i++ {
		newCorpus = append(newCorpus, tfscores[i].Term)
	}

	corpus := strings.ToLower(strings.Join(newCorpus, " "))
	vec, err := c11y.VectorForCorpi(context.Background(), &pb.Corpi{
		Corpi:     []string{corpus},
		Overrides: c.boostByInformationGain(perc),
	})
	if err != nil {
		spew.Dump(c.ranked)
		log.Fatal(fmt.Errorf("%s with corpus: %s and doc %s", err, corpus, c.inputCorpus))
	}

	vector := extractVector(vec)

	var minimum = float32(100000)
	var prediction string
	for _, cat := range targets {
		dist, err := cosineDist(vector, cat.Vector)
		if err != nil {
			log.Fatal(err)
		}

		if dist < minimum {
			minimum = dist
			prediction = cat.Name
		}
	}

	out[0] = percentile{Percentile: perc, Prediction: prediction, Match: prediction == c.control}

	return out
}

func (c *Classifier) boostByInformationGain(percentage int) []*pb.Override {
	maxBoost := float32(3)
	cutoff := int(float32(percentage) / float32(100) * float32(len(c.ranked)))
	out := make([]*pb.Override, cutoff)
	for i, word := range c.ranked[:cutoff] {
		boost := 1 - float32(math.Log(float64(i)/float64(cutoff)))*float32(1)
		if math.IsInf(float64(boost), 1) || boost > maxBoost {
			boost = maxBoost
		}
		if i == 0 {
			boost = boost / 2
		}
		out[i] = &pb.Override{
			Expression: fmt.Sprintf("%f * w", boost),
			Word:       word.Word,
		}
	}

	return out
}

func findTfScore(list []TermWithTfIdf, term string) float32 {
	for _, item := range list {
		if item.Term == term {
			return item.RelativeScore
		}
	}

	panic(fmt.Sprintf("term not in tf idf list: %s", term))
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

func (c *Classifier) isInTfPercentile(tf []TermWithTfIdf, percentage int, needle string) bool {
	cutoff := int(float32(percentage) / float32(100) * float32(len(tf)))
	selection := tf[:cutoff]

	for _, hay := range selection {
		if needle == hay.Term {
			return true
		}
	}

	return false
}
