/*                          _       _
 *__      _____  __ ___   ___  __ _| |_ ___
 *\ \ /\ / / _ \/ _` \ \ / / |/ _` | __/ _ \
 * \ V  V /  __/ (_| |\ V /| | (_| | ||  __/
 *  \_/\_/ \___|\__,_| \_/ |_|\__,_|\__\___|
 *
 * Copyright © 2016 - 2019 Weaviate. All rights reserved.
 * LICENSE: https://github.com/creativesoftwarefdn/weaviate/blob/develop/LICENSE.md
 * DESIGN & CONCEPT: Bob van Luijt (@bobvanluijt)
 * CONTACT: hello@creativesoftwarefdn.org
 */package schema

import (
	"fmt"

	pb "github.com/creativesoftwarefdn/contextionary/contextionary"
	"github.com/fatih/camelcase"
)

// SearchType to search for either class names or property names
type SearchType string

const (
	// SearchTypeClass to search the contextionary for class names
	SearchTypeClass SearchType = "class"
	// SearchTypeProperty to search the contextionary for property names
	SearchTypeProperty SearchType = "property"
)

// SearchParams to be used for a SchemaSearch. See individual properties for
// additional documentation on what they do
type SearchParams struct {
	*pb.SchemaSearchParams
}

// Validate the feasibility of the specified arguments
func (p SearchParams) Validate() error {
	if p.Name == "" {
		return fmt.Errorf("Name cannot be empty")
	}

	if err := p.validateCertaintyOrWeight(p.Certainty); err != nil {
		return fmt.Errorf("invalid Certainty: %s", err)
	}

	if p.SearchType != pb.SearchType_CLASS && p.SearchType != pb.SearchType_PROPERTY {
		return fmt.Errorf(
			"SearchType must be SearchType_CLASS or SearchType_PROPERTY, but got '%s'", p.SearchType)
	}

	for i, keyword := range p.Keywords {
		if err := p.validateKeyword(keyword); err != nil {
			return fmt.Errorf("invalid keyword at position %d: %s", i, err)
		}
	}

	return nil
}

func (p SearchParams) validateKeyword(kw *pb.Keyword) error {
	if kw.Keyword == "" {
		return fmt.Errorf("Keyword cannot be empty")
	}

	if len(camelcase.Split(kw.Keyword)) > 1 {
		return fmt.Errorf("invalid Keyword: keywords cannot be camelCased - "+
			"instead split your keyword up into several keywords, this way each word "+
			"of your camelCased string can have its own weight, got '%s'", kw.Keyword)
	}

	if err := p.validateCertaintyOrWeight(kw.Weight); err != nil {
		return fmt.Errorf("invalid Weight: %s", err)
	}

	return nil
}

func (p SearchParams) validateCertaintyOrWeight(c float32) error {
	if c >= 0 && c <= 1 {
		return nil
	}

	return fmt.Errorf("must be between 0 and 1, but got '%f'", c)
}
