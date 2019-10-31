package extensions

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_LookerUpper(t *testing.T) {
	t.Run("looking up a non-existant concept", func(t *testing.T) {
		repo := newFakeRepo()
		lu := NewLookerUpper(repo)
		extension, err := lu.Lookup("non_existing_concept")
		require.Nil(t, err)
		assert.Nil(t, extension)
	})

	t.Run("looking up existing concepts", func(t *testing.T) {
		repo := newFakeRepo()
		lu := NewLookerUpper(repo)

		t.Run("with an initial concept", func(t *testing.T) {
			ext := Extension{
				Concept:    "flux_capacitor",
				Vector:     []float32{0, 1, 2},
				Occurrence: 1000,
			}
			repo.add(ext)
			time.Sleep(100 * time.Millisecond)
			actual, err := lu.Lookup("flux_capacitor")
			require.Nil(t, err)
			assert.Equal(t, &ext, actual)
		})

		t.Run("with second concept", func(t *testing.T) {
			ext := Extension{
				Concept:    "clux_fapacitor",
				Vector:     []float32{0, 1, 2},
				Occurrence: 1000,
			}
			repo.add(ext)
			time.Sleep(100 * time.Millisecond)

			t.Run("looking up the original concept", func(t *testing.T) {
				actual, err := lu.Lookup("flux_capacitor")
				require.Nil(t, err)
				require.NotNil(t, actual)
				assert.Equal(t, "flux_capacitor", actual.Concept)
			})

			t.Run("looking up the second concept concept", func(t *testing.T) {
				actual, err := lu.Lookup("clux_fapacitor")
				require.Nil(t, err)
				require.NotNil(t, actual)
				assert.Equal(t, "clux_fapacitor", actual.Concept)
			})
		})
	})
}

func newFakeRepo() *fakeRepo {
	repo := &fakeRepo{
		ch: make(chan WatchResponse),
	}

	return repo
}

type fakeRepo struct {
	ch         chan WatchResponse
	extensions []Extension
}

func (f *fakeRepo) WatchAll() chan WatchResponse {
	return f.ch
}

func (f *fakeRepo) add(ex Extension) {
	f.extensions = append(f.extensions, ex)
	f.ch <- f.extensions
}
