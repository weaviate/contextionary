package extensions

import (
	"sync"
)

type LookerUpper struct {
	repo Repo
	sync.Mutex
	db map[string]Extension
}

type Repo interface {
	// WatchAll must send an immediate response after opening (for
	// initializiation), then send another response whenver the db has changed
	WatchAll() chan WatchResponse
}

func NewLookerUpper(repo Repo) *LookerUpper {
	lu := &LookerUpper{
		repo: repo,
		db:   map[string]Extension{},
	}
	lu.initWatcher()
	return lu
}

func (lu *LookerUpper) Lookup(concept string) (*Extension, error) {
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
			// lu.Lock()
			// defer lu.Unlock()

			lu.updateDB(res)
		}
	}()
}

func (lu *LookerUpper) updateDB(list []Extension) {
	for _, ext := range list {
		lu.db[ext.Concept] = ext
	}

}
