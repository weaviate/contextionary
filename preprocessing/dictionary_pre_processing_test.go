package preprocessing

import (
	"bufio"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/weaviate/contextionary/compoundsplitting"
)

func TestPreprocessorSplitterDictFile(t *testing.T) {
	// Create the file
	outputFile := "test_dict.splitdict"
	GenerateSplittingDictFile("../test/compoundsplitting/contextionary.idx", "../test/compoundsplitting/nl_NL.dic", "../test/compoundsplitting/nl_NL.aff", outputFile)

	// Validate the output file
	file, err := os.Open(outputFile)
	if err != nil {
		t.Fail()
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	found := false
	for scanner.Scan() {
		line := scanner.Text()
		split := strings.Split(line, ",")
		if split[0] == "appellantes" {
			found = true
			break
		}
	}
	assert.True(t, found)

	if err := scanner.Err(); err != nil {
		t.Fail()
	}

	err = file.Close()
	if err != nil {
		t.Fail()
	}

	// Load from output file
	dict, err := compoundsplitting.NewContextionaryDict(outputFile)
	if err != nil {
		t.Fail()
	}

	assert.True(t, dict.Contains("amsterdam"))
	assert.True(t, dict.Contains("appellante"))
	assert.True(t, dict.Contains("appellantes"))

	// Remove test file
	err = os.Remove(outputFile)
	if err != nil {
		t.Fail()
	}
}
