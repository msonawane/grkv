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
	// s.logger.Info("set on", zap.String("node", s.mlNodeName), zap.Int("clients", len(s.grpcClients)))
	success, err := s.set(ctx, req)
	if err != nil || !success.Success {
		return success, err
	}
	for _, client := range s.grpcClients {
		// s.logger.Info("sending data to neighbours", zap.Any("node", client))
		// success, err := client.GRPCSet(ctx, req)
		go client.GRPCSet(ctx, req)
		// s.logger.Info("result of grpc call", zap.Any("success", success), zap.Error(err))
	}

	return success, err
}

// Delete keys.
func (s *Store) Delete(ctx context.Context, req *kvpb.DeleteRequest) (*kvpb.Success, error) {
	success, err := s.delete(ctx, req)
	if err != nil || !success.Success {
		return success, err
	}
	for _, client := range s.grpcClients {
		s.logger.Info("sending data to neighbours", zap.Any("node", client))
		client.GRPCDelete(ctx, req)
	}

	return success, err
}

// Get keys.
func (s *Store) Get(ctx context.Context, in *kvpb.GetRequest) (*kvpb.GetResponse, error) {
	gr, err := s.get(ctx, in)
	if len(gr.KeysNotFound) == 0 {
		return gr, err
	}
	// s.logger.Info("Getting from neighbour ")
	// req := &kvpb.GetRequest{
	// 	Keys: gr.KeysNotFound,
	// }
	// for _, client := range s.grpcClients {
	// 	resp, err := client.GRPCGet(ctx, req)
	// 	// fmt.Printf("resp: %#v, err: %#v\n", resp, err)
	// }
	return gr, err
}
