package extensions

import (
	"sync"
)

type LookerUpper struct {
	repo RetrieverRepo
	sync.Mutex
	db map[string]Extension
}

type RetrieverRepo interface {
	// WatchAll must send an immediate response after opening (for
	// initializiation), then send another response whenver the db has changed
	WatchAll() chan WatchResponse
}

func NewLookerUpper(repo RetrieverRepo) *LookerUpper {
	lu := &LookerUpper{
		repo: repo,
		db:   map[string]Extension{},
	}
	lu.initWatcher()
	return lu
}

func (lu *LookerUpper) Lookup(concept string) (*Extension, error) {
	lu.Lock()
	defer lu.Unlock()

	ext, ok := lu.db[concept]
	if !ok {
		return nil, nil
	}

	return &ext, nil
}

type WatchResponse []Extension

func (lu *LookerUpper) initWatcher() {
	updateCh := lu.repo.WatchAll()

	go func() {
		for res := range updateCh {

			lu.updateDB(res)
		}
	}()
}

func (lu *LookerUpper) updateDB(list []Extension) {
	lu.Lock()
	defer lu.Unlock()

	for _, ext := range list {
		lu.db[ext.Concept] = ext
	}

}
