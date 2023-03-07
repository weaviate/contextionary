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

	contextionary "github.com/semi-technologies/contextionary/contextionary/core"
)

func fatal(err error) {
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

func main() {
	c11y, err := contextionary.LoadVectorFromDisk("./test/contextionary/example.knn", "./test/contextionary/example.idx")
	fatal(err)

	fmt.Println("results before building centroid based on keywords: ")
	kNN("city", c11y)

	// Combine contextionaries
	contextionaries := []contextionary.Contextionary{c11y}
	combined, err := contextionary.CombineVectorIndices(contextionaries)
	fatal(err)

	fmt.Println("results after building centroid based on keywords: ")
	kNN("ocean", combined)
}

func kNN(name string, contextionary contextionary.Contextionary) {
	itemIndex := contextionary.WordToItemIndex(name)
	if ok := itemIndex.IsPresent(); !ok {
		fatal(fmt.Errorf("item index for %s is not present", name))
	}

	list, distances, err := contextionary.GetNnsByItem(itemIndex, 20, 3)
	if err != nil {
		fatal(fmt.Errorf("get nns errored: %s", err))
	}

	for i := range list {
		w, err := contextionary.ItemIndexToWord(list[i])
		if err != nil {
			fmt.Printf("error: %s", err)
		}
		fmt.Printf("\n%d %f %s\n", list[i], distances[i], w)
	}

}
