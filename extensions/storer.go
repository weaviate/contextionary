package extensions

import (
	"context"
	"fmt"
	"strings"
	"unicode"

	core "github.com/semi-technologies/contextionary/contextionary/core"
)

type Vectorizer interface {
	Corpi(corpi []string) (*core.Vector, error)
}

type StorerRepo interface {
	Put(ctx context.Context, ext Extension) error
}

type Storer struct {
	vectorizer Vectorizer
	repo       StorerRepo
}

func NewStorer(vectorizer Vectorizer, repo StorerRepo) *Storer {
	return &Storer{vectorizer, repo}
}

func (s *Storer) Put(ctx context.Context, concept string, input ExtensionInput) error {
	err := s.validate(concept, input)
	if err != nil {
		return fmt.Errorf("invalid extension: %v", err)
	}

	vector, err := s.vectorizer.Corpi([]string{input.Definition})
	if err != nil {
		return fmt.Errorf("vectorize definition: %v", err)
	}

	concept = s.compound(concept)

	ext := Extension{
		Concept:    concept,
		Input:      input,
		Vector:     vector.ToArray(), // nil-check can be omitted as vectorizer will return non-nil if err==nil
		Occurrence: 1000,             // TODO: Improve!
	}

	err = s.repo.Put(ctx, ext)
	if err != nil {
		return fmt.Errorf("store extension: %v", err)
	}

	return nil
}

func (s *Storer) compound(inp string) string {
	parts := strings.Split(inp, " ")
	return strings.Join(parts, "_")
}

func (s *Storer) validate(concept string, input ExtensionInput) error {
	if len(concept) < 2 {
		return fmt.Errorf("concept must have at least two characters")
	}

	for _, r := range concept {
		if !unicode.IsLower(r) && !unicode.IsSpace(r) && !unicode.IsNumber(r) {
			return fmt.Errorf("concept must be made up of all lowercase letters and/or numbers, " +
				"for custom compund words use spaces, e.g. 'flux capacitor'")
		}
	}

	if len(input.Definition) == 0 {
		return fmt.Errorf("definition cannot be empty")
	}

	if input.Weight > 1 || input.Weight < 0 {
		return fmt.Errorf("weight must be between 0 and 1")
	}

	if input.Weight < 1 {
		return fmt.Errorf("weights below 1 (extending an existing concept) not supported yet - coming soon")
	}

	return nil
}
