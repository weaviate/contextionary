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
 */package schema

import (
	"errors"
	"testing"

	"github.com/semi-technologies/contextionary/contextionary"
	"github.com/stretchr/testify/assert"
)

func Test__SchemaSearch_Validation(t *testing.T) {
	tests := schemaSearchTests{
		{
			name: "valid params",
			searchParams: SearchParams{
				SchemaSearchParams: &contextionary.SchemaSearchParams{
					SearchType: contextionary.SearchType_CLASS,
					Name:       "foo",
					Certainty:  1.0,
				},
			},
			expectedError: nil,
		},
		{
			name: "missing search name",
			searchParams: SearchParams{
				SchemaSearchParams: &contextionary.SchemaSearchParams{
					SearchType: contextionary.SearchType_CLASS,
					Name:       "",
					Certainty:  0.0,
				},
			},
			expectedError: errors.New("Name cannot be empty"),
		},
		{
			name: "certainty too low",
			searchParams: SearchParams{
				SchemaSearchParams: &contextionary.SchemaSearchParams{
					SearchType: contextionary.SearchType_CLASS,
					Name:       "bestName",
					Certainty:  -4,
				},
			},
			expectedError: errors.New("invalid Certainty: must be between 0 and 1, but got '-4.000000'"),
		},
		{
			name: "certainty too high",
			searchParams: SearchParams{
				SchemaSearchParams: &contextionary.SchemaSearchParams{
					SearchType: contextionary.SearchType_CLASS,
					Name:       "bestName",
					Certainty:  4,
				},
			},
			expectedError: errors.New("invalid Certainty: must be between 0 and 1, but got '4.000000'"),
		},
		{
			name: "missing kind on class search",
			searchParams: SearchParams{
				SchemaSearchParams: &contextionary.SchemaSearchParams{
					SearchType: contextionary.SearchType_CLASS,
					Name:       "bestName",
					Certainty:  0.5,
				},
			},
			expectedError: errors.New("Kind cannot be empty"),
		},
		{
			name: "valid keywords",
			searchParams: SearchParams{
				SchemaSearchParams: &contextionary.SchemaSearchParams{
					SearchType: contextionary.SearchType_CLASS,
					Name:       "foo",
					Certainty:  1.0,
					Keywords: []*contextionary.Keyword{
						{
							Keyword: "foobar",
							Weight:  1.0,
						},
					},
				},
			},
			expectedError: nil,
		},
		{
			name: "keywords with empty names",
			searchParams: SearchParams{
				SchemaSearchParams: &contextionary.SchemaSearchParams{
					SearchType: contextionary.SearchType_CLASS,
					Name:       "foo",
					Certainty:  1.0,
					Keywords: []*contextionary.Keyword{
						{
							Keyword: "",
							Weight:  1.0,
						},
					},
				},
			},
			expectedError: errors.New("invalid keyword at position 0: Keyword cannot be empty"),
		},
		{
			name: "keywords with invalid weights",
			searchParams: SearchParams{
				SchemaSearchParams: &contextionary.SchemaSearchParams{
					SearchType: contextionary.SearchType_CLASS,
					Name:       "foo",
					Certainty:  1.0,
					Keywords: []*contextionary.Keyword{{
						Keyword: "bestkeyword",
						Weight:  1.3,
					}},
				},
			},
			expectedError: errors.New("invalid keyword at position 0: invalid Weight: " +
				"must be between 0 and 1, but got '1.300000'"),
		},
		{
			name: "CamelCased keywords",
			searchParams: SearchParams{
				SchemaSearchParams: &contextionary.SchemaSearchParams{
					SearchType: contextionary.SearchType_CLASS,
					Name:       "foo",
					Certainty:  1.0,
					Keywords: []*contextionary.Keyword{{
						Keyword: "worstKeyword",
						Weight:  0.8,
					}},
				},
			},
			expectedError: errors.New("invalid keyword at position 0: invalid Keyword: " +
				"keywords cannot be camelCased - instead split your keyword up into several keywords, " +
				"this way each word of your camelCased string can have its own weight, got 'worstKeyword'"),
		},
	}

	tests.AssertValidation(t)
}

func (s schemaSearchTests) AssertValidation(t *testing.T) {
	for _, test := range s {
		t.Run(test.name, func(t *testing.T) {
			err := test.searchParams.Validate()

			// assert error
			assert.Equal(t, test.expectedError, err, "should match the expected error")

		})
	}
}
