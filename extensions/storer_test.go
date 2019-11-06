package extensions

import (
	"context"
	"fmt"
	"testing"

	core "github.com/semi-technologies/contextionary/contextionary/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func Test_Storer(t *testing.T) {
	t.Run("with invalid inputs", func(t *testing.T) {
		repo := &fakeStorerRepo{}
		s := NewStorer(&fakeVectorizer{}, repo)
		inp := ExtensionInput{
			Definition: "an electrical device to store energy in the short term",
			Weight:     1,
		}

		type testCase struct {
			concept     string
			inp         ExtensionInput
			expectedErr error
		}

		tests := []testCase{
			testCase{
				concept:     "lowerAndUpperCase",
				expectedErr: fmt.Errorf("invalid extension: concept must be made up of all lowercase letters and/or numbers, for custom compund words use spaces, e.g. 'flux capacitor'"),
				inp:         inp,
			},
			testCase{
				concept:     "a",
				expectedErr: fmt.Errorf("invalid extension: concept must have at least two characters"),
				inp:         inp,
			},
			testCase{
				concept:     "foo",
				expectedErr: fmt.Errorf("invalid extension: definition cannot be empty"),
				inp:         ExtensionInput{Weight: 1},
			},
			testCase{
				concept:     "foo",
				expectedErr: fmt.Errorf("invalid extension: weight must be between 0 and 1"),
				inp:         ExtensionInput{Weight: -1, Definition: "foo bar"},
			},
			testCase{
				concept:     "foo",
				expectedErr: fmt.Errorf("invalid extension: weight must be between 0 and 1"),
				inp:         ExtensionInput{Weight: 3, Definition: "foo bar"},
			},
			testCase{ // TODO: add feature, then remove limitation
				concept:     "foo",
				expectedErr: fmt.Errorf("invalid extension: weights below 1 (extending an existing concept) not supported yet - coming soon"),
				inp:         ExtensionInput{Weight: 0.7, Definition: "foo bar"},
			},
		}

		for _, test := range tests {
			t.Run(test.concept, func(t *testing.T) {
				err := s.Put(context.Background(), test.concept, test.inp)
				assert.Equal(t, test.expectedErr, err)
			})
		}
	})

	t.Run("with valid input (single word)", func(t *testing.T) {
		repo := &fakeStorerRepo{}
		s := NewStorer(&fakeVectorizer{}, repo)
		concept := "capacitor"
		inp := ExtensionInput{
			Definition: "an electrical device to store energy in the short term",
			Weight:     1,
		}

		expectedExtension := Extension{
			Input:      inp,
			Concept:    concept,
			Vector:     []float32{1, 2, 3},
			Occurrence: 1000,
		}
		repo.On("Put", expectedExtension).Return(nil)
		err := s.Put(context.Background(), concept, inp)
		require.Nil(t, err)
		repo.AssertExpectations(t)

	})

	t.Run("with valid input (compound word)", func(t *testing.T) {
		// this is a special case because users will input their words using
		// spaces, but we store them using snake_case
		repo := &fakeStorerRepo{}
		s := NewStorer(&fakeVectorizer{}, repo)
		concept := "flux capacitor"
		inp := ExtensionInput{
			Definition: "an energy source for cars to travel through time",
			Weight:     1,
		}

		expectedExtension := Extension{
			Input:      inp,
			Concept:    "flux_capacitor",
			Vector:     []float32{1, 2, 3},
			Occurrence: 1000,
		}
		repo.On("Put", expectedExtension).Return(nil)
		err := s.Put(context.Background(), concept, inp)
		require.Nil(t, err)
		repo.AssertExpectations(t)
	})
}

type fakeVectorizer struct{}

func (f *fakeVectorizer) Corpi(corpi []string) (*core.Vector, error) {
	v := core.NewVector([]float32{1, 2, 3})
	return &v, nil
}

type fakeStorerRepo struct {
	mock.Mock
}

func (f *fakeStorerRepo) Put(ctx context.Context, ext Extension) error {
	args := f.Called(ext)
	return args.Error(0)
}
