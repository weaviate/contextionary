package compoundsplitting

import "C"
import (
	"bufio"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

// GenerateSplittingDictFile from
//  contextionaryIndexFile binary .idx file containing the words for the specific language
//  languageDictionaryFile a hunspell .dic file for the specific language
//  languageAffixesFile a hunspell .aff file for the specific language
//  to reduce file- and hunspell dependencies for the splitter
func GenerateSplittingDictFile(contextionaryIndexFile string, languageDictionaryFile string, languageAffixesFile string, outputFile string) error {
	dict := NewContextionaryDict(contextionaryIndexFile, languageDictionaryFile, languageAffixesFile)
	out, err := os.Create(outputFile)
	if err != nil {
		return err
	}
	defer out.Close()

	for word, occurrence := range dict.dict {
		line := fmt.Sprintf("%s,%v\n", word, occurrence)
		_, err := out.Write([]byte(line))
		if err != nil {
			return err
		}
	}
	return nil
}

// Dictionary filter for the splitting algorithm
// based on the words in the contextionary
type ContextionaryDict struct {
	dict map[string]int // storing the word and its occurrence
}

// NewContextionaryDict from
//  contextionaryIndexFile binary .idx file containing the words for the specific language
//  languageDictionaryFile a hunspell .dic file for the specific language
//  languageAffixesFile a hunspell .aff file for the specific language
func NewContextionaryDict(contextionaryIndexFile string, languageDictionaryFile string, languageAffixesFile string) *ContextionaryDict {
	dict := &ContextionaryDict{
		dict: make(map[string]int, 1200000),
	}
	hunspellFilter := Hunspell(languageAffixesFile, languageDictionaryFile)

	err := dict.loadContextionary(contextionaryIndexFile, hunspellFilter)
	if err != nil {
		panic(err.Error())
	}
	return dict
}

// NewContextionaryDictFromFile
// uses
func NewContextionaryDictFromFile(contextionaryDictFile string) (*ContextionaryDict, error) {
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
		lenScore +=len(word)
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

// passesFilter if the word is in the dictionary of the given language
func passesFilter(word string, filter *Hunhandle) bool {
	inDict := filter.Spell(word)
	if inDict {
		return true
	}
	// Check if upper case word
	inDict = filter.Spell(strings.Title(word))
	return inDict
}

// loadContextionary from binary file
func (cd *ContextionaryDict) loadContextionary(path string, filter *Hunhandle) error {
	data, readFileErr := ioutil.ReadFile(path)
	if readFileErr != nil {
		return readFileErr
	}

	// File format:
	// https://github.com/semi-technologies/weaviate-vector-generator#wordlist-file-format
	nrWordsBytes := data[0:8]
	//vectorLengthBytes := data[8:16]
	metaDataLengthBytes := data[16:24]

	nrWords := binary.LittleEndian.Uint64(nrWordsBytes)
	//vectorLength := binary.LittleEndian.Uint64(vectorLengthBytes)
	metaDataLength := binary.LittleEndian.Uint64(metaDataLengthBytes)

	// Read meta data
	metaDataBytes := data[24:24+metaDataLength]
	var metadata map[string]interface{}
	unMarshalErr := json.Unmarshal(metaDataBytes, &metadata)
	if unMarshalErr != nil {
		return unMarshalErr
	}

	var startOfTable uint64 = 24 + uint64(metaDataLength)
	var offset uint64 = 4 - (startOfTable % 4)
	startOfTable += offset

	for wordIndex := uint64(0); wordIndex < nrWords; wordIndex++ {
		// entryAddress is the index in the data where the pointer to
		// the word is located
		entryAddress := startOfTable + 8 * wordIndex
		pointerToWordByte := data[entryAddress: entryAddress+8]
		pointerToWord := binary.LittleEndian.Uint64(pointerToWordByte)
		word, occurence := getWordAndOccurence(data, pointerToWord)
		// Only add the word if it passes the filter
		if passesFilter(word, filter) {
			cd.dict[word] = int(occurence)
		}
	}

	return nil
}

// getWordAndOccurence from the data frame indecated by the pointer
func getWordAndOccurence(data []byte, pointer uint64) (string, uint64) {
	ocurrence := binary.LittleEndian.Uint64(data[pointer:pointer+8])

	pointer = pointer+8
	for i := uint64(0);;i++ {
		if data[pointer+i] == '\x00' {
			word := string(data[pointer:pointer+i])
			return word, ocurrence
		}
	}
}

type DictMock struct {
	scores map[string]float64
}

func (dm *DictMock) Contains(word string) bool {
	_, exists := dm.scores[word]
	return exists
}

func (dm *DictMock) Score(phrase []string) float64 {
	score := 0.0
	for _, word := range phrase {
		score += dm.scores[word]
	}
	return score
}