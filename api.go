package grkv

import (
	"context"

	"github.com/dgraph-io/badger/v2"
	"github.com/msonawane/grkv/kvpb"
)

// get keys.
func (s *Store) get(ctx context.Context, req *kvpb.GetRequest) (*kvpb.GetResponse, error) {
	resp := &kvpb.GetResponse{}
	err := s.db.View(func(txn *badger.Txn) error {
		for _, key := range req.Keys {
			item, err := txn.Get(key)
			if err == badger.ErrKeyNotFound || err != nil {
				resp.KeysNotFound = append(resp.KeysNotFound, key)
				continue
			}
			pair := kvpb.KeyValue{Key: key}
			err = item.Value(func(val []byte) error {
				pair.Value = val
				resp.Data = append(resp.Data, &pair)
				return nil
			})
			continue
		}
		return nil
	})
	return resp, err
}

// Set keys.
func (s *Store) set(ctx context.Context, req *kvpb.SetRequest) (*kvpb.Success, error) {
	success := kvpb.Success{Success: false}
	wb := s.db.NewWriteBatch()
	defer wb.Cancel()

	for _, pair := range req.Data {
		entry := badger.Entry{Key: pair.Key, Value: pair.Value, ExpiresAt: pair.ExpiresAt}
		err := wb.SetEntry(&entry)
		if err != nil {

			return &success, err
		}
	}
	err := wb.Flush()
	success.Success = true

	return &success, err
}

// Delete keys.
func (s *Store) delete(ctx context.Context, req *kvpb.DeleteRequest) (*kvpb.Success, error) {

	success := kvpb.Success{Success: false}
	wb := s.db.NewWriteBatch()
	defer wb.Cancel()

	for _, key := range req.Keys {
		err := wb.Delete(key)

		if err != nil {

			return &success, err
		}
	}
	err := wb.Flush()
	success.Success = true

	return &success, err

}

// Set keys.
func (s *Store) Set(ctx context.Context, req *kvpb.SetRequest) (*kvpb.Success, error) {
	return s.set(ctx, req)
}

// Delete keys.
func (s *Store) Delete(ctx context.Context, req *kvpb.DeleteRequest) (*kvpb.Success, error) {
	return s.delete(ctx, req)
}

// Get keys.
func (s *Store) Get(ctx context.Context, in *kvpb.GetRequest) (*kvpb.GetResponse, error) {
	return s.get(ctx, in)
}
