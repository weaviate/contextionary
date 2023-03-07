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
 */package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	contextionary "github.com/semi-technologies/contextionary/contextionary/core"
)

func fatal(err error) {
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

var cleanupRegexp = regexp.MustCompile("[^a-zA-Z0-9 ]+")

func clean(input string) string {
	return cleanupRegexp.ReplaceAllString(input, "")
}

func main() {
	setMapping()

	c11yRoot := os.Args[1]

	c11y, err := contextionary.LoadVectorFromDisk(c11yRoot+"/contextionary-en.knn", c11yRoot+"/contextionary-en.idx")
	fatal(err)

	for _, text := range sampleTexts {
		vector := vectorForText(text.Content, c11y)
		err := put(vectorDocument{text, convertArrayToBase64(vector.ToArray())})
		fatal(err)
	}

	searchString(strings.Join(os.Args[2:], " "), c11y)
}

func vectorForText(input string, c11y contextionary.Contextionary) *contextionary.Vector {
	words := strings.Split(clean(input), " ")

	var total int
	// var stopWords int
	var vectors []contextionary.Vector
	var occurrences []uint64
	var stopWords int
	var maxOcc uint64
	var minOcc uint64 = 1e15
	var presentWords []string

	for _, word := range words {
		total++
		if isStopWord(word) {
			stopWords++
			continue
		}

		itemIndex := c11y.WordToItemIndex(word)
		if ok := itemIndex.IsPresent(); ok {
			vector, err := c11y.GetVectorForItemIndex(itemIndex)
			fatal(err)

			occurrence, err := c11y.ItemIndexToOccurrence(itemIndex)
			fatal(err)

			vectors = append(vectors, *vector)
			if occurrence < minOcc {
				minOcc = occurrence
			}
			if occurrence > maxOcc {
				maxOcc = occurrence
			}

			occurrences = append(occurrences, occurrence)
			presentWords = append(presentWords, word)
		}

	}

	// calculate weights by normalizing the occurrences to 0..1
	weights := make([]float32, len(occurrences), len(occurrences))
	for i, occ := range occurrences {
		// _ = occ
		// weights[i] = 1
		weight := 1 - float32(occ-minOcc)/float32(maxOcc-minOcc)
		weights[i] = weight

		// fmt.Printf("%s: %f\n", presentWords[i], weight)
	}

	centroid, err := contextionary.ComputeWeightedCentroid(vectors, weights)
	fatal(err)

	// fmt.Printf("%d stop words out of %d removed. %d of the remainder contained\n", stopWords, total, len(vectors))

	return centroid

}
