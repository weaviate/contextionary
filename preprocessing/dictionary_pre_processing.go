package preprocessing

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

// PreprocessDict temp storage for reading in the index file
type PreprocessDict struct {
	dict map[string]int
}

// GenerateSplittingDictFile from
//
//	contextionaryIndexFile binary .idx file containing the words for the specific language
//	languageDictionaryFile a hunspell .dic file for the specific language
//	languageAffixesFile a hunspell .aff file for the specific language
//	to reduce file- and hunspell dependencies for the splitter
func GenerateSplittingDictFile(contextionaryIndexFile string, languageDictionaryFile string, languageAffixesFile string, outputFile string) error {
	dict := NewPreprocessDict(contextionaryIndexFile, languageDictionaryFile, languageAffixesFile)
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

// NewPreprocessDict from
//
//	contextionaryIndexFile binary .idx file containing the words for the specific language
//	languageDictionaryFile a hunspell .dic file for the specific language
//	languageAffixesFile a hunspell .aff file for the specific language
func NewPreprocessDict(contextionaryIndexFile string, languageDictionaryFile string, languageAffixesFile string) *PreprocessDict {
	dict := &PreprocessDict{
		dict: make(map[string]int, 1200000),
	}
	hunspellFilter := Hunspell(languageAffixesFile, languageDictionaryFile)

	err := dict.loadContextionary(contextionaryIndexFile, hunspellFilter)
	if err != nil {
		panic(err.Error())
	}
	return dict
}

// loadContextionary from binary file
func (cd *PreprocessDict) loadContextionary(path string, filter *Hunhandle) error {
	data, readFileErr := ioutil.ReadFile(path)
	if readFileErr != nil {
		return readFileErr
	}

	// File format:
	// https://github.com/weaviate/weaviate-vector-generator#wordlist-file-format
	nrWordsBytes := data[0:8]
	//vectorLengthBytes := data[8:16]
	metaDataLengthBytes := data[16:24]

	nrWords := binary.LittleEndian.Uint64(nrWordsBytes)
	//vectorLength := binary.LittleEndian.Uint64(vectorLengthBytes)
	metaDataLength := binary.LittleEndian.Uint64(metaDataLengthBytes)

	// Read meta data
	metaDataBytes := data[24 : 24+metaDataLength]
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
		entryAddress := startOfTable + 8*wordIndex
		pointerToWordByte := data[entryAddress : entryAddress+8]
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
	ocurrence := binary.LittleEndian.Uint64(data[pointer : pointer+8])

	pointer = pointer + 8
	for i := uint64(0); ; i++ {
		if data[pointer+i] == '\x00' {
			word := string(data[pointer : pointer+i])
			return word, ocurrence
		}
	}
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
