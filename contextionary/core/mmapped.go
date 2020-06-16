/*                          _       _
 *__      _____  __ ___   ___  __ _| |_ ___
 *\ \ /\ / / _ \/ _` \ \ / / |/ _` | __/ _ \
 * \ V  V /  __/ (_| |\ V /| | (_| | ||  __/
 *  \_/\_/ \___|\__,_| \_/ |_|\__,_|\__\___|
 *
 * Copyright © 2016 - 2019 Weaviate. All rights reserved.
 * LICENSE: https://github.com/semi-technologies/weaviate/blob/develop/LICENSE.md
 * DESIGN & CONCEPT: Bob van Luijt (@bobvanluijt)
 * CONTACT: hello@semi.technology
 */
package contextionary

import (
	"encoding/binary"
	"fmt"
	"log"
	"math"
	"os"
	"syscall"

	annoy "github.com/semi-technologies/contextionary/contextionary/core/annoyindex"
)

type mmappedIndex struct {
	word_index *Wordlist
	knn        annoy.AnnoyIndex
	knnRaw     []byte
	dimensions int
}

func (m *mmappedIndex) GetNumberOfItems() int {
	return int(m.word_index.numberOfWords)
}

// Returns the length of the used vectors.
func (m *mmappedIndex) GetVectorLength() int {
	return int(m.word_index.vectorWidth)
}

func (m *mmappedIndex) WordToItemIndex(word string) ItemIndex {
	return m.word_index.FindIndexByWord(word)
}

func (m *mmappedIndex) ItemIndexToWord(item ItemIndex) (string, error) {
	if item >= 0 && item <= m.word_index.GetNumberOfWords() {
		w, _ := m.word_index.getWord(item)
		return w, nil
	} else {
		return "", fmt.Errorf("Index out of bounds")
	}
}

func (m *mmappedIndex) ItemIndexToOccurrence(item ItemIndex) (uint64, error) {
	if item >= 0 && item <= m.word_index.GetNumberOfWords() {
		_, occ := m.word_index.getWord(item)
		return occ, nil
	} else {
		return 0, fmt.Errorf("Index out of bounds")
	}
}

func (m *mmappedIndex) GetVectorForItemIndex(item ItemIndex) (*Vector, error) {
	if item < 0 && item > m.word_index.GetNumberOfWords() {
		return nil, fmt.Errorf("Index out of bounds")
	}

	var floats []float32
	floats = m.getItem(int(item))

	return &Vector{vector: floats}, nil
}

func (m *mmappedIndex) getItem(index int) []float32 {
	offset := 16
	vectorSize := m.dimensions * 4
	begin := index*(offset+vectorSize) + offset
	end := begin + vectorSize
	return vectorFromBytes(m.knnRaw[begin:end])
}

func vectorFromBytes(in []byte) []float32 {
	out := make([]float32, len(in)/4)
	for offset := 0; offset < len(in); offset += 4 {
		bits := binary.LittleEndian.Uint32(in[offset : offset+4])
		float := math.Float32frombits(bits)
		out[offset/4] = float
	}

	return out
}

// Compute the distance between two items.
func (m *mmappedIndex) GetDistance(a ItemIndex, b ItemIndex) (float32, error) {
	if a >= 0 && b >= 0 && a <= m.word_index.GetNumberOfWords() && b <= m.word_index.GetNumberOfWords() {
		return m.knn.GetDistance(int(a), int(b)), nil
	} else {
		return 0, fmt.Errorf("Index out of bounds")
	}
}

func (m *mmappedIndex) GetNnsByItem(item ItemIndex, n int, k int) ([]ItemIndex, []float32, error) {
	if item >= 0 && item <= m.word_index.GetNumberOfWords() {
		var items []int
		var distances []float32

		m.knn.GetNnsByItem(int(item), n, k, &items, &distances)

		var indices []ItemIndex = make([]ItemIndex, len(items))
		for i, x := range items {
			indices[i] = ItemIndex(x)
		}

		return indices, distances, nil
	} else {
		return nil, nil, fmt.Errorf("Index out of bounds")
	}
}

func (m *mmappedIndex) GetNnsByVector(vector Vector, n int, k int) ([]ItemIndex, []float32, error) {
	if len(vector.vector) == m.GetVectorLength() {
		var items []int
		var distances []float32

		m.knn.GetNnsByVector(vector.vector, n, k, &items, &distances)

		var indices []ItemIndex = make([]ItemIndex, len(items))
		for i, x := range items {
			indices[i] = ItemIndex(x)
		}

		return indices, distances, nil
	} else {
		return nil, nil, fmt.Errorf("Wrong vector length provided")
	}
}

// SafeGetSimilarWords returns n similar words in the contextionary,
// examining k trees. It is guaratueed to have results, even if the word is
// not in the contextionary. In this case the list only contains the word
// itself. It can then still be used for exact match or levensthein-based
// searches against db backends.
func (m *mmappedIndex) SafeGetSimilarWords(word string, n, k int) ([]string, []float32) {
	return safeGetSimilarWordsFromAny(m, word, n, k)
}

// SafeGetSimilarWordsWithCertainty returns  similar words in the
// contextionary, if they are close enough to match the required certainty.
// It is guaratueed to have results, even if the word is not in the
// contextionary. In this case the list only contains the word itself. It can
// then still be used for exact match or levensthein-based searches against
// db backends.
func (m *mmappedIndex) SafeGetSimilarWordsWithCertainty(word string, certainty float32) []string {
	return safeGetSimilarWordsWithCertaintyFromAny(m, word, certainty)
}

func LoadVectorFromDisk(annoy_index string, word_index_file_name string) (Contextionary, error) {
	word_index, err := LoadWordlist(word_index_file_name)

	if err != nil {
		return nil, fmt.Errorf("Could not load vector: %+v", err)
	}

	knn := annoy.NewAnnoyIndexEuclidean(int(word_index.vectorWidth))
	knn.Load(annoy_index)

	knnRaw, err := loadAnnoyIndexDirectly(annoy_index)
	if err != nil {
		return nil, fmt.Errorf("load raw index: %v", err)
	}

	idx := &mmappedIndex{
		word_index: word_index,
		knn:        knn,
		knnRaw:     knnRaw,
		dimensions: int(word_index.vectorWidth),
	}

	return idx, nil
}

// directly load the annoy index file to avoid memory leaks in the annoy
// go-port of the C library, see #26
func loadAnnoyIndexDirectly(path string) ([]byte, error) {
	fmt.Println("loading index directly")
	file, err := os.Open(path)
	if err != nil {
		log.Fatalf("Can't open the knn file at %s: %+v", path, err)
	}

	file_info, err := file.Stat()
	if err != nil {
		log.Fatalf("Can't stat the knn file at %s: %+v", path, err)
	}

	mmap, err := syscall.Mmap(int(file.Fd()), 0, int(file_info.Size()), syscall.PROT_READ, syscall.MAP_SHARED)
	if err != nil {
		log.Fatalf("Can't mmap the knn file %s: %+v", path, err)
	}

	return mmap, nil
}
