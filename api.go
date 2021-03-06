package grkv

import (
	"context"

	"github.com/dgraph-io/badger/v2"
	"github.com/msonawane/grkv/kvpb"
	"go.uber.org/zap"
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
			_ = item.Value(func(val []byte) error {
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

	defer func() {
		if r := recover(); r != nil {
			s.logger.Error("Panic detected ", zap.Any("stacktrace", r))
		}
	}()

	success := kvpb.Success{Success: false}
	wb := s.db.NewWriteBatch()
	defer wb.Cancel()

	for _, pair := range req.Data {
		if pair != nil {
			if pair.Key != nil && pair.Value != nil {

				entry := badger.Entry{Key: pair.Key, Value: pair.Value, ExpiresAt: pair.ExpiresAt}
				err := wb.SetEntry(&entry)
				if err != nil {
					return &success, err
				}
			}
		}
	}
	err := wb.Flush()
	success.Success = true

	return &success, err
}

// Delete keys.
func (s *Store) delete(ctx context.Context, req *kvpb.DeleteRequest) (*kvpb.Success, error) {
	defer func() {
		if r := recover(); r != nil {
			s.logger.Error("Panic detected ", zap.Any("stacktrace", r))
		}
	}()
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
	// s.logger.Info("set on", zap.String("node", s.mlNodeName), zap.Int("clients", len(s.grpcClients)))
	defer func() {
		if r := recover(); r != nil {
			s.logger.Error("Panic detected ", zap.Any("stacktrace", r))
		}
	}()
	success, err := s.set(ctx, req)
	if err != nil || !success.Success {
		return success, err
	}
	rr := ReplicationRequest{
		Type:       SET,
		SetRequest: req,
	}
	go s.Replicate(rr)
	return success, err
}

// Delete keys.
func (s *Store) Delete(ctx context.Context, req *kvpb.DeleteRequest) (*kvpb.Success, error) {
	defer func() {
		if r := recover(); r != nil {
			s.logger.Error("Panic detected ", zap.Any("stacktrace", r))
		}
	}()
	success, err := s.delete(ctx, req)
	if err != nil || !success.Success {
		return success, err
	}
	rr := ReplicationRequest{
		Type:       DELETE,
		DelRequest: req,
	}
	go s.Replicate(rr)

	return success, err
}

// Get keys.
func (s *Store) Get(ctx context.Context, in *kvpb.GetRequest) (*kvpb.GetResponse, error) {
	defer func() {
		if r := recover(); r != nil {
			s.logger.Error("Panic detected ", zap.Any("stacktrace", r))
		}
	}()
	gr, err := s.get(ctx, in)
	if len(gr.KeysNotFound) == 0 && err == nil {
		return gr, err
	}

	if len(s.grpcClients) > 0 {

		// try getting from random neighbour.
		s.logger.Info("Getting from neighbour ")
		req := &kvpb.GetRequest{
			Keys: gr.KeysNotFound,
		}
		var resp *kvpb.GetResponse
		for _, client := range s.grpcClients {
			resp, err = client.GRPCGet(ctx, req)
			break
			// fmt.Printf("resp: %#v, err: %#v\n", resp, err)
		}
		// gr.KeysNotFound = resp.KeysNotFound
		gr.Data = append(gr.Data, resp.Data...)
		setRequest := &kvpb.SetRequest{
			Data: resp.Data,
		}
		go s.set(ctx, setRequest)
	}

	return gr, err
}

func (s *Store) GetWithStringKeys(ctx context.Context, keys ...string) (*kvpb.GetResponse, error) {
	gr := kvpb.GetRequest{}
	for _, key := range keys {
		gr.Keys = append(gr.Keys, []byte(key))
	}
	return s.Get(ctx, &gr)
}
