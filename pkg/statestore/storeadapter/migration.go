// Copyright 2023 The Swarm Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package storeadapter

import (
	"strings"

	"github.com/ethersphere/bee/pkg/storage"
	"github.com/ethersphere/bee/pkg/storage/migration"
)

func allSteps() migration.Steps {
	return map[uint64]migration.StepFn{
		1: epochMigration,
		2: deletePrefix("sync_interval"),
		3: deletePrefix("sync_interval"),
	}
}

func deletePrefix(prefix string) migration.StepFn {
	return func(s storage.BatchedStore) error {
		store := &StateStorerAdapter{s}
		return store.Iterate(prefix, func(key, val []byte) (stop bool, err error) {
			return false, store.Delete(string(key))
		})
	}
}

func epochMigration(s storage.BatchedStore) error {

	var deleteEntries = []string{
		"statestore_schema",
		"tags",
		"sync_interval",
		"kademlia-counters",
		"addressbook",
		"batch",
	}

	return s.Iterate(storage.Query{
		Factory: func() storage.Item { return &rawItem{&proxyItem{obj: []byte(nil)}} },
	}, func(res storage.Result) (stop bool, err error) {
		if strings.HasPrefix(res.ID, stateStoreNamespace) {
			return false, nil
		}
		for _, e := range deleteEntries {
			if strings.HasPrefix(res.ID, e) {
				_ = s.Delete(&rawItem{&proxyItem{key: res.ID}})
				return false, nil
			}
		}

		item := res.Entry.(*rawItem)
		item.key = res.ID
		item.ns = stateStoreNamespace
		if err := s.Put(item); err != nil {
			return true, err
		}
		_ = s.Delete(&rawItem{&proxyItem{key: res.ID}})
		return false, nil
	})
}
