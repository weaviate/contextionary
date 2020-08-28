package compoundsplitting

import (
	"bufio"
	"os"
	"strconv"
	"strings"
)


// Dictionary filter for the splitting algorithm
// based on the words in the contextionary
type ContextionaryDict struct {
	dict map[string]int // storing the word and its occurrence
}

// NewContextionaryDict
// uses a dictionary file that was created using the preprocessing procedures
func NewContextionaryDict(contextionaryDictFile string) (*ContextionaryDict, error) {
	file, err := os.Open(contextionaryDictFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	dict := &ContextionaryDict{
		dict: make(map[string]int, 400000),
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		split := strings.Split(line, ",")
		occurrence, err := strconv.Atoi(split[1])
		if err != nil {
			return nil, err
		}
		dict.dict[split[0]] = occurrence
	}

	return dict, nil
}

// Contains true if word is in contextionary
func (cd *ContextionaryDict) Contains(word string) bool {
	_, exists := cd.dict[word]
	return exists
}

//Score prefers long and few words
func (cd *ContextionaryDict) Score(phrase []string) float64 {
	// Prefer longer words as scoring
	// Assumption is that the compound words are on average more similar to splittings that
	// share most of the characters with the compound.
	lenScore := 0
	for _, word := range phrase {
		lenScore += len(word)
	}

	// Give a boost for less words
	if len(phrase) == 2 {
		lenScore += 3
	}
	if len(phrase) == 3 {
		lenScore += 1
	}

	return float64(lenScore)
}


// DictMock used for unit testing
type DictMock struct {
	scores map[string]float64
}

// Contains
func (dm *DictMock) Contains(word string) bool {
	_, exists := dm.scores[word]
	return exists
}

// Score
func (dm *DictMock) Score(phrase []string) float64 {
	score := 0.0
	for _, word := range phrase {
		score += dm.scores[word]
	}
	return score
}
