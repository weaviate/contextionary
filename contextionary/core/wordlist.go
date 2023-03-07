/*                          _       _
 *__      _____  __ ___   ___  __ _| |_ ___
 *\ \ /\ / / _ \/ _` \ \ / / |/ _` | __/ _ \
 * \ V  V /  __/ (_| |\ V /| | (_| | ||  __/
 *  \_/\_/ \___|\__,_| \_/ |_|\__,_|\__\___|
 *
 * Copyright © 2016 - 2019 Weaviate. All rights reserved.
 * LICENSE: https://github.com/weaviate/weaviate/blob/master/LICENSE
 * DESIGN & CONCEPT: Bob van Luijt (@bobvanluijt)
 * CONTACT: hello@semi.technology
 */
package contextionary

// //// #include <string.h>
// //import "C"

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"syscall"
)

type Wordlist struct {
	vectorWidth           uint64
	numberOfWords         uint64
	metadata              map[string]interface{}
	occurrencePercentiles []uint64

	file         os.File
	startOfTable int
	mmap         []byte
}

func LoadWordlist(path string) (*Wordlist, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("Can't open the wordlist at %s: %+v", path, err)
	}

	file_info, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("Can't stat the wordlist at %s: %+v", path, err)
	}

	mmap, err := syscall.Mmap(int(file.Fd()), 0, int(file_info.Size()), syscall.PROT_READ, syscall.MAP_SHARED)
	if err != nil {
		return nil, fmt.Errorf("Can't mmap the file %s: %+v", path, err)
	}

	nrWordsBytes := mmap[0:8]
	vectorWidthBytes := mmap[8:16]
	metadataLengthBytes := mmap[16:24]

	nrWords := binary.LittleEndian.Uint64(nrWordsBytes)
	vectorWidth := binary.LittleEndian.Uint64(vectorWidthBytes)
	metadataLength := binary.LittleEndian.Uint64(metadataLengthBytes)

	metadataBytes := mmap[24 : 24+metadataLength]
	var metadata map[string]interface{}

	json.Unmarshal(metadataBytes, &metadata)

	// Compute beginning of word list lookup table.
	var start_of_table int = 24 + int(metadataLength)
	var offset int = 4 - (start_of_table % 4)
	start_of_table += offset

	wl := &Wordlist{
		vectorWidth:   vectorWidth,
		numberOfWords: nrWords,
		metadata:      metadata,
		startOfTable:  start_of_table,
		mmap:          mmap,
	}

	wl.initOccurrencePercentiles()

	return wl, nil
}

func (w *Wordlist) GetNumberOfWords() ItemIndex {
	return ItemIndex(w.numberOfWords)
}

func (w *Wordlist) OccurrencePercentile(percentile int) uint64 {
	if percentile < 0 || percentile > 100 {
		panic("incorrect usage of occurrence percentile, must be between 0 and 100")
	}

	return w.occurrencePercentiles[percentile]
}

func (w *Wordlist) FindIndexByWord(_needle string) ItemIndex {
	var needle = string([]byte(_needle))
	needle += "\x00"

	var bytes_needle = []byte(needle)

	var low ItemIndex = 0
	var high ItemIndex = ItemIndex(w.numberOfWords)

	for low <= high {
		var midpoint ItemIndex = (low + high) / 2

		ptr := w.getWordPtr(midpoint)

		// if the last word in the index is shorter than our needle, we would panic
		// by accessing a non-existing adress. To prevent this, the higher boundary
		// can never be higher than the len(index)-1
		endPos := 8 + len(bytes_needle)
		if endPos >= len(ptr) {
			endPos = len(ptr) - 1
		}

		// ignore the first 8 bytes as they are reserved for occurrence
		word := ptr[8:endPos]

		var cmp = bytes.Compare(bytes_needle, word)

		if cmp == 0 {
			return midpoint
		} else if cmp < 0 {
			high = midpoint - 1
		} else {
			low = midpoint + 1
		}
	}

	return -1
}

func (w *Wordlist) getWordPtr(index ItemIndex) []byte {
	entry_addr := ItemIndex(w.startOfTable) + index*8
	word_address_bytes := w.mmap[entry_addr : entry_addr+8]
	word_address := binary.LittleEndian.Uint64(word_address_bytes)
	return w.mmap[word_address:]
}

func (w *Wordlist) getWord(index ItemIndex) (string, uint64) {
	ptr := w.getWordPtr(index)
	occurrence := binary.LittleEndian.Uint64(ptr[0:8])
	for i := 8; i < len(ptr); i++ {
		if ptr[i] == '\x00' {
			return string(ptr[8:i]), occurrence
		}
	}

	return "", 0
}

func (w *Wordlist) initOccurrencePercentiles() {
	w.occurrencePercentiles = make([]uint64, 101) // make 101 elements longs, so both index 0 and 100 are included
	max := int(w.GetNumberOfWords())
	allOccs := make([]uint64, max)

	for i := ItemIndex(0); int(i) < max; i++ {
		_, occ := w.getWord(i)
		allOccs[i] = occ
	}

	sort.Slice(allOccs, func(a, b int) bool { return allOccs[a] < allOccs[b] })

	for i := 0; i <= 100; i++ { // note that this is 101 elements!
		if i == 0 {
			w.occurrencePercentiles[i] = 0
			continue
		}

		if i == 100 {
			w.occurrencePercentiles[i] = allOccs[len(allOccs)-1]
			continue
		}

		occ := uint64(float64(i) / 100 * float64(len(allOccs)))
		w.occurrencePercentiles[i] = occ
	}
}
