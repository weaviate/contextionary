/*                          _       _
 *__      _____  __ ___   ___  __ _| |_ ___
 *\ \ /\ / / _ \/ _` \ \ / / |/ _` | __/ _ \
 * \ V  V /  __/ (_| |\ V /| | (_| | ||  __/
 *  \_/\_/ \___|\__,_| \_/ |_|\__,_|\__\___|
 *
 * Copyright © 2016 - 2019 Weaviate. All rights reserved.
 * LICENSE: https://github.com/weaviate/weaviate/blob/master/LICENSE
 * DESIGN & CONCEPT: Bob van Luijt (@bobvanluijt)
 * CONTACT: hello@weaviate.io
 */package contextionary

import (
	"regexp"
)

const simliarWordsLimit = 15

func safeGetSimilarWordsFromAny(c11y Contextionary, word string, n, k int) ([]string, []float32) {
	i := c11y.WordToItemIndex(word)
	if !i.IsPresent() {
		return []string{word}, []float32{1}
	}

	indices, newCertainties, err := c11y.GetNnsByItem(i, n, k)
	if err != nil {
		return []string{word}, []float32{1}
	}

	var words []string
	var certainties []float32
	for i, index := range indices {
		word, err := c11y.ItemIndexToWord(index)
		if err != nil {
			continue
		}

		if wordHasIllegalCharacters(word) {
			continue
		}

		words = append(words, word)
		certainties = append(certainties, newCertainties[i])
	}

	return words, certainties
}

func safeGetSimilarWordsWithCertaintyFromAny(c11y Contextionary, word string, certainty float32) []string {
	var matchingWords []string
	var matchtingCertainties []float32

	count := 0
	words, certainties := c11y.SafeGetSimilarWords(word, 100, 32)
	for i, word := range words {
		if count >= simliarWordsLimit {
			break
		}

		var dist float32
		if dist = DistanceToCertainty(certainties[i]); dist < certainty {
			continue
		}

		count++
		matchingWords = append(matchingWords, alphanumeric(word))
		matchtingCertainties = append(matchtingCertainties, dist)
	}

	return matchingWords
}

func wordHasIllegalCharacters(word string) bool {
	// we know that the schema based contextionary uses a leading dollar sign for
	// the class and property centroids, so we can easily filter them out
	return regexp.MustCompile("^\\$").MatchString(word)
}

func alphanumeric(word string) string {
	return regexp.MustCompile("[^a-zA-Z0-9_]+").ReplaceAllString(word, "")
}
