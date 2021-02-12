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
package schema

// This file contains the logic to build an in-memory contextionary from the actions & things classes and properties.

import (
	"fmt"
	"strings"

	"github.com/fatih/camelcase"

	"github.com/semi-technologies/weaviate/entities/models"
	"github.com/semi-technologies/weaviate/entities/schema"

	libcontextionary "github.com/semi-technologies/contextionary/contextionary/core"
)

type stopwordDetector interface {
	IsStopWord(word string) bool
}

func BuildInMemoryContextionaryFromSchema(schema schema.Schema, context *libcontextionary.Contextionary, stopwordDetector stopwordDetector) (*libcontextionary.Contextionary, error) {
	in_memory_builder := libcontextionary.InMemoryBuilder((*context).GetVectorLength())

	err := add_names_from_schema_properties(context, stopwordDetector, in_memory_builder, schema.SemanticSchemaFor())
	if err != nil {
		return nil, err
	}

	in_memory_contextionary := in_memory_builder.Build(10)
	x := libcontextionary.Contextionary(in_memory_contextionary)
	return &x, nil
}

// This function adds words in the form of $THING[Blurp]
func add_names_from_schema_properties(context *libcontextionary.Contextionary, stopwordDetector stopwordDetector,
	in_memory_builder *libcontextionary.MemoryIndexBuilder, schema *models.Schema) error {
	for _, class := range schema.Classes {
		class_centroid_name := fmt.Sprintf("$%v[%v]", "OBJECT", class.Class)
		// Split name on camel case, and add each word part to a equally weighted word vector.
		camel_parts := camelcase.Split(class.Class)
		var vectors []libcontextionary.Vector = make([]libcontextionary.Vector, 0)
		for _, part := range camel_parts {
			part = strings.ToLower(part)
			if stopwordDetector.IsStopWord(part) {
				// we don't need to check if every single word is a stop word,
				// because UC validation has already prevented this from happening
				continue
			}

			// Lookup vector for the keyword.
			idx := (*context).WordToItemIndex(part)
			if idx.IsPresent() {
				vector, err := (*context).GetVectorForItemIndex(idx)

				if err != nil {
					return fmt.Errorf("Could not fetch vector for a found index. Data corruption?")
				} else {
					vectors = append(vectors, *vector)
				}
			} else {
				return fmt.Errorf("Could not find camel cased name part '%v' for class '%v' in the contextionary", part, class.Class)
			}

			centroid, err := libcontextionary.ComputeCentroid(vectors)
			if err != nil {
				return fmt.Errorf("Could not compute centroid")
			} else {
				in_memory_builder.AddWord(class_centroid_name, *centroid)
			}
		}

		// NOW FOR THE PROPERTIES;
		// basically the same code as above.
		for _, property := range class.Properties {
			property_centroid_name := fmt.Sprintf("%v[%v]", class_centroid_name, property.Name)
			// Split name on camel case, and add each word part to a equally weighted word vector.
			camel_parts := camelcase.Split(property.Name)
			var vectors []libcontextionary.Vector = make([]libcontextionary.Vector, 0)
			for _, part := range camel_parts {
				part = strings.ToLower(part)
				if stopwordDetector.IsStopWord(part) {
					// we don't need to check if every single word is a stop word,
					// because UC validation has already prevented this from happening
					continue
				}
				// Lookup vector for the keyword.
				idx := (*context).WordToItemIndex(part)
				if idx.IsPresent() {
					vector, err := (*context).GetVectorForItemIndex(idx)

					if err != nil {
						return fmt.Errorf("Could not fetch vector for a found index. Data corruption?")
					} else {
						vectors = append(vectors, *vector)
					}
				} else {
					return fmt.Errorf("Could not find camel cased part of name '%v' for property %v in class '%v' in the contextionary, consider adding some keywords instead.", part, property.Name, class.Class)
				}

				centroid, err := libcontextionary.ComputeCentroid(vectors)
				if err != nil {
					return fmt.Errorf("Could not compute centroid: %v", err)
				} else {
					in_memory_builder.AddWord(property_centroid_name, *centroid)
				}
			}
		}
	}

	return nil
}
