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

import contextionary "github.com/weaviate/contextionary/contextionary/core"

// Contextionary composes a regular contextionary with additional
// schema-related query methods
type Contextionary struct {
	contextionary.Contextionary
}

// New creates a new Contextionary from a contextionary.Contextionary which it
// extends with Schema-related search methods
func New(c contextionary.Contextionary) *Contextionary {
	return &Contextionary{
		Contextionary: c,
	}
}
