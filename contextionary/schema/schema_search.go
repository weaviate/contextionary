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
 */package schema

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/fatih/camelcase"
	pb "github.com/weaviate/contextionary/contextionary"
	contextionary "github.com/weaviate/contextionary/contextionary/core"
	"github.com/weaviate/contextionary/errors"
)

// SearchResult is a single search result. See wrapping Search Results for the Type
type SearchResult struct {
	Name      string
	Certainty float32
}

// SearchResults is grouping of SearchResults for a SchemaSearch
type SearchResults struct {
	Type    SearchType
	Results []SearchResult
}

// Len of the result set
func (r SearchResults) Len() int {
	return len(r.Results)
}

// SchemaSearch can be used to search for related classes and properties, see
// documentation of SearchParams for more details on how to use it and
// documentation on *pb.SchemaSearchResults for more details on how to use the return
// value
func (con *Contextionary) SchemaSearch(params *pb.SchemaSearchParams) (*pb.SchemaSearchResults, error) {
	p := SearchParams{params}
	if err := p.Validate(); err != nil {
		return nil, errors.NewInvalidUserInputf("invalid search params: %s", err)
	}

	centroid, err := con.centroidFromNameAndKeywords(p)
	if err != nil {
		return nil, errors.NewInvalidUserInputf("could not build centroid from name and keywords: %s", err)
	}

	rawResults, err := con.knnSearch(*centroid)
	if err != nil {
		return nil, errors.NewInternalf("could not perform knn search: %s", err)
	}

	if p.SearchType == pb.SearchType_CLASS {
		return con.handleClassSearch(p, rawResults)
	}

	// since we have passed validation we know that anything that's not a class
	// search must be a property search
	return con.handlePropertySearch(p, rawResults)
}

func (con *Contextionary) centroidFromNameAndKeywords(p SearchParams) (*contextionary.Vector, error) {
	nameVector, err := con.camelCaseWordToVector(p.Name)
	if err != nil {
		return nil, fmt.Errorf("invalid name in search: %s", err)
	}

	if len(p.Keywords) == 0 {
		return nameVector, nil
	}

	vectors := make([]contextionary.Vector, len(p.Keywords)+1, len(p.Keywords)+1)
	weights := make([]float32, len(p.Keywords)+1, len(p.Keywords)+1)
	// set last vector to className which always has weight=1
	vectors[len(vectors)-1] = *nameVector
	weights[len(vectors)-1] = 1

	for i, keyword := range p.Keywords {
		kwVector, err := con.wordToVector(keyword.Keyword)
		if err != nil {
			return nil, fmt.Errorf("invalid keyword in search: %s", err)
		}
		vectors[i] = *kwVector
		weights[i] = keyword.Weight
	}

	return contextionary.ComputeWeightedCentroid(vectors, weights)
}

func (con *Contextionary) camelCaseWordToVector(w string) (*contextionary.Vector, error) {
	parts := camelcase.Split(w)
	if len(parts) == 1 {
		// no camelcasing, no need to build a centroid
		return con.wordToVector(w)
	}

	vectors := make([]contextionary.Vector, len(parts), len(parts))
	weights := make([]float32, len(parts), len(parts))
	for i, part := range parts {
		v, err := con.wordToVector(part)
		if err != nil {
			return nil, fmt.Errorf("invalid camelCased compound word: %s", err)
		}

		vectors[i] = *v
		weights[i] = 1 // on camel-casing all parts are weighted equally
	}

	return contextionary.ComputeWeightedCentroid(vectors, weights)
}

func (con *Contextionary) wordToVector(w string) (*contextionary.Vector, error) {
	w = strings.ToLower(w)
	itemIndex := con.WordToItemIndex(w)
	if ok := itemIndex.IsPresent(); !ok {
		return nil, fmt.Errorf(
			"the word '%s' is not present in the contextionary and therefore not a valid search term", w)
	}

	vector, err := con.GetVectorForItemIndex(itemIndex)
	if err != nil {
		return nil, fmt.Errorf("could not get vector for word '%s' with itemIndex '%d': %s",
			w, itemIndex, err)
	}

	return vector, nil
}

func (con *Contextionary) handleClassSearch(p SearchParams, search rawResults) (*pb.SchemaSearchResults, error) {
	return &pb.SchemaSearchResults{
		Type:    p.SearchType,
		Results: search.extractClassNames(p),
	}, nil
}

func (con *Contextionary) handlePropertySearch(p SearchParams, search rawResults) (*pb.SchemaSearchResults, error) {
	return &pb.SchemaSearchResults{
		Type:    p.SearchType,
		Results: search.extractPropertyNames(p),
	}, nil
}

func (con *Contextionary) knnSearch(vector contextionary.Vector) (rawResults, error) {
	list, distances, err := con.GetNnsByVector(vector, 10000, 3)
	if err != nil {
		return nil, fmt.Errorf("could not get nearest neighbors for vector '%v': %s", vector, err)
	}

	results := make(rawResults, len(list), len(list))
	for i := range list {
		word, err := con.ItemIndexToWord(list[i])
		if err != nil {
			return results, fmt.Errorf("got a result from kNN search, but don't have a word for this index: %s", err)
		}

		results[i] = rawResult{
			name:     word,
			distance: distances[i],
		}
	}

	return results, nil
}

// rawResult is a helper struct to contain the results of the kNN-search. It
// does not yet contain the desired output. This means the names can be both
// classes/properties and arbitrary words. Furthermore the certainty has not
// yet been normalized , so it is merely the raw kNN distance
type rawResult struct {
	name     string
	distance float32
}

type rawResults []rawResult

func (r rawResults) extractClassNames(p SearchParams) []*pb.SchemaSearchResult {
	var results []*pb.SchemaSearchResult
	regex := regexp.MustCompile(fmt.Sprintf("^\\$%s\\[([A-Za-z]+)\\]$", "OBJECT"))

	for _, rawRes := range r {
		if regex.MatchString(rawRes.name) {
			certainty := distanceToCertainty(rawRes.distance)
			if certainty < p.Certainty {
				continue
			}

			results = append(results, &pb.SchemaSearchResult{
				Name:      regex.FindStringSubmatch(rawRes.name)[1], //safe because we ran .MatchString before
				Certainty: certainty,
			})
		}
	}

	return results
}

func (r rawResults) extractPropertyNames(p SearchParams) []*pb.SchemaSearchResult {
	var results []*pb.SchemaSearchResult
	regex := regexp.MustCompile("^\\$[A-Za-z]+\\[[A-Za-z]+\\]\\[([A-Za-z]+)\\]$")

	propsMap := map[string][]*pb.SchemaSearchResult{}

	for _, rawRes := range r {
		if regex.MatchString(rawRes.name) {
			name := regex.FindStringSubmatch(rawRes.name)[1] //safe because we ran .MatchString before
			certainty := distanceToCertainty(rawRes.distance)
			if certainty < p.Certainty {
				continue
			}

			res := &pb.SchemaSearchResult{
				Name:      name,
				Certainty: certainty,
			}
			if _, ok := propsMap[name]; !ok {
				propsMap[name] = []*pb.SchemaSearchResult{res}
			} else {
				propsMap[name] = append(propsMap[name], res)
			}
		}
	}

	// now calculate mean of duplicate results
	for _, resultsPerName := range propsMap {
		results = append(results, &pb.SchemaSearchResult{
			Name:      resultsPerName[0].Name,
			Certainty: meanCertainty(resultsPerName),
		})
	}

	return results
}

func meanCertainty(rs []*pb.SchemaSearchResult) float32 {
	var compound float32
	for _, r := range rs {
		compound += r.Certainty
	}

	return compound / float32(len(rs))
}

func distanceToCertainty(d float32) float32 {
	return 1 - d/12
}
