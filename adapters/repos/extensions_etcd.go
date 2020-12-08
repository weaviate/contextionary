package repos

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/coreos/etcd/clientv3"
	"github.com/semi-technologies/contextionary/extensions"
	"github.com/semi-technologies/contextionary/server/config"
	"github.com/sirupsen/logrus"
)

type EtcdExtensionRepo struct {
	client *clientv3.Client
	logger logrus.FieldLogger
	prefix string
}

func NewEtcdExtensionRepo(client *clientv3.Client, logger logrus.FieldLogger,
	config *config.Config) *EtcdExtensionRepo {
	return &EtcdExtensionRepo{client: client, logger: logger, prefix: config.ExtensionsPrefix}
}

func (r *EtcdExtensionRepo) WatchAll() chan extensions.WatchResponse {
	returnCh := make(chan extensions.WatchResponse)

	go func() {
		rch := r.client.Watch(context.Background(), r.prefix, clientv3.WithPrefix())
		for wresp := range rch {
			for range wresp.Events {
				r.updateConsumers(returnCh)
			}
		}
	}()

	return returnCh
}

func (r *EtcdExtensionRepo) updateConsumers(returnCh chan extensions.WatchResponse) {
	res, err := r.client.Get(context.Background(), r.prefix, clientv3.WithPrefix())
	if err != nil {
		r.logger.WithField("action", "extensions_retrieve_all").
			WithError(err).Error()
		return
	}

	var exts []extensions.Extension
	for _, kv := range res.Kvs {
		var ext extensions.Extension
		err := json.Unmarshal(kv.Value, &ext)
		if err != nil {
			r.logger.WithField("action", "extensions_retrieve_all").
				WithError(err).Error()
			return
		}

		exts = append(exts, ext)
	}

	returnCh <- exts
}

func (r *EtcdExtensionRepo) Put(ctx context.Context, ext extensions.Extension) error {
	extBytes, err := json.Marshal(ext)
	if err != nil {
		return fmt.Errorf("marshal extension to json: %v", err)
	}

	_, err = r.client.Put(ctx, fmt.Sprintf("%s%s", r.prefix, ext.Concept), string(extBytes))
	if err != nil {
		return fmt.Errorf("etcd put: %v", err)
	}

	return nil
}
