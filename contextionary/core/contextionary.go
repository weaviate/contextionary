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
 */

// Package contextionary provides the toolset to add context to words.
package contextionary

// ItemIndex is an opaque type that models an index number used to identify a
// word.
type ItemIndex int

// IsPresent can be used after retrieving a word index (which does not error on
// its own), to see if the word was actually present in the contextionary.
func (i *ItemIndex) IsPresent() bool {
	return *i >= 0
}

// Contextionary is the API to decouple the K-nn interface that is needed for
// Weaviate from a concrete implementation.
type Contextionary interface {

	// Return the number of items that is stored in the index.
	GetNumberOfItems() int

	// Returns the length of the used vectors.
	GetVectorLength() int

	// Look up a word, return an index.
	// Check for presence of the index with index.IsPresent()
	WordToItemIndex(word string) ItemIndex

	// Based on an index, return the assosiated word.
	ItemIndexToWord(item ItemIndex) (string, error)

	// Based on an index, return the assosiated word.
	ItemIndexToOccurrence(item ItemIndex) (uint64, error)

	//OccurrencePercentile shows the occurrence of the mentioned percentile in ascending order
	OccurrencePercentile(perc int) uint64

	// Get the vector of an item index.
	GetVectorForItemIndex(item ItemIndex) (*Vector, error)

	// Compute the distance between two items.
	GetDistance(a ItemIndex, b ItemIndex) (float32, error)

	// Get the n nearest neighbours of item, examining k trees.
	// Returns an array of indices, and of distances between item and the n-nearest neighbors.
	GetNnsByItem(item ItemIndex, n, k int) ([]ItemIndex, []float32, error)

	// Get the n nearest neighbours of item, examining k trees.
	// Returns an array of indices, and of distances between item and the n-nearest neighbors.
	GetNnsByVector(vector Vector, n, k int) ([]ItemIndex, []float32, error)

	// SafeGetSimilarWords returns n similar words in the contextionary,
	// examining k trees. It is guaratueed to have results, even if the word is
	// not in the contextionary. In this case the list only contains the word
	// itself. It can then still be used for exact match or levensthein-based
	// searches against db backends.
	SafeGetSimilarWords(word string, n, k int) ([]string, []float32)

	// SafeGetSimilarWordsWithCertainty returns  similar words in the
	// contextionary, if they are close enough to match the required certainty.
	// It is guaratueed to have results, even if the word is not in the
	// contextionary. In this case the list only contains the word itself. It can
	// then still be used for exact match or levensthein-based searches against
	// db backends.
	SafeGetSimilarWordsWithCertainty(word string, certainty float32) []string
}
