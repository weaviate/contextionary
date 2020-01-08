package extensions

import (
	"context"
	"fmt"
	"strings"
	"unicode"

	core "github.com/semi-technologies/contextionary/contextionary/core"
	"github.com/sirupsen/logrus"
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
	logger     logrus.FieldLogger
}

func NewStorer(vectorizer Vectorizer, repo StorerRepo, logger logrus.FieldLogger) *Storer {
	return &Storer{vectorizer, repo, logger}
}

func (s *Storer) Put(ctx context.Context, concept string, input ExtensionInput) error {
	s.logger.WithField("action", "extensions_put").
		WithField("concept", concept).
		WithField("extension", input).
		Debug("received request to add/replace custom extension")

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

	s.logger.WithField("action", "extensions_put_prestore").
		WithField("concept", ext.Concept).
		WithField("extension", ext).
		Debug("calculated vector, about to store in repo")

	err = s.repo.Put(ctx, ext)
	if err != nil {
		s.logger.WithField("action", "extensions_store_error").
			WithField("concept", ext.Concept).
			Errorf("repo put: %v", err)
		return fmt.Errorf("store extension: %v", err)
	}

	s.logger.WithField("action", "extensions_put_poststore").
		WithField("concept", ext.Concept).
		Debug("successfully stored extension in repo")

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
