package repos

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/semi-technologies/contextionary/extensions"
	"github.com/semi-technologies/contextionary/server/config"
	"github.com/sirupsen/logrus"
)

type ModuleExtensionRepo struct {
	client        *http.Client
	logger        logrus.FieldLogger
	origin        string
	watchInterval time.Duration
}

func NewExtensionsRepo(logger logrus.FieldLogger,
	config *config.Config, watchInterval time.Duration) *ModuleExtensionRepo {
	client := &http.Client{}
	return &ModuleExtensionRepo{
		client:        client,
		logger:        logger,
		origin:        config.ExtensionsStorageOrigin,
		watchInterval: watchInterval,
	}
}

func (r *ModuleExtensionRepo) WatchAll() chan extensions.WatchResponse {
	returnCh := make(chan extensions.WatchResponse)

	go func() {
		t := time.Tick(r.watchInterval)
		for {
			r.updateConsumers(returnCh)
			<-t
		}
	}()

	return returnCh
}

func (f *ModuleExtensionRepo) uri(path string) string {
	return fmt.Sprintf("%s%s", f.origin, path)
}

func (r *ModuleExtensionRepo) updateConsumers(returnCh chan extensions.WatchResponse) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET",
		r.uri("/v1/modules/text2vec-contextionary/extensions-storage/"), nil)
	if err != nil {
		r.logger.WithField("action", "extensions_retrieve_all").
			WithError(err).Error()
		return
	}

	res, err := r.client.Do(req)
	if err != nil {
		r.logger.WithField("action", "extensions_retrieve_all").
			WithError(err).Error()
		return
	}

	defer res.Body.Close()
	if res.StatusCode > 399 {
		r.logger.WithField("action", "extensions_retrieve_all").
			WithError(fmt.Errorf("expected status < 399, got %d", res.StatusCode)).
			Error()
		return
	}

	var exts []extensions.Extension
	scanner := bufio.NewScanner(res.Body)
	for scanner.Scan() {
		if err := scanner.Err(); err != nil {
			r.logger.WithField("action", "extensions_retrieve_all").
				WithError(err).Error()
			return
		}

		rawExt := scanner.Bytes()
		var ext extensions.Extension
		err := json.Unmarshal(rawExt, &ext)
		if err != nil {
			r.logger.WithField("action", "extensions_retrieve_all").
				WithError(err).Error()
			return
		}

		exts = append(exts, ext)
	}

	returnCh <- exts
}

func (r *ModuleExtensionRepo) Put(ctx context.Context, ext extensions.Extension) error {
	extBytes, err := json.Marshal(ext)
	if err != nil {
		return fmt.Errorf("marshal extension to json: %v", err)
	}

	req, err := http.NewRequestWithContext(ctx, "PUT", r.uri(fmt.Sprintf(
		"/v1/modules/text2vec-contextionary/extensions-storage/%s", ext.Concept)), bytes.NewReader(extBytes))

	res, err := r.client.Do(req)
	if err != nil {
		return fmt.Errorf("put: %v", err)
	}

	defer res.Body.Close()
	if res.StatusCode > 399 {
		return fmt.Errorf("expected status < 399, got %d", res.StatusCode)
	}

	return nil
}
