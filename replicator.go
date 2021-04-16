package grkv

import (
	"context"
	"time"

	"github.com/msonawane/grkv/kvpb"
	"go.uber.org/zap"
)

type ReplicationRequest struct {
	Type       string
	SetRequest *kvpb.SetRequest
	DelRequest *kvpb.DeleteRequest
}

const deadLine = 5 * time.Second

func (s *Store) replicator(name, addr string, reqChan chan ReplicationRequest, client kvpb.KeyValueStoreClient) {
	s.logger.Info("Starting replicator ", zap.String("name", name), zap.String("addr", addr))
	errorCounter := 0
	for v := range reqChan {
		if errorCounter > 10 {
			s.logger.Info("more than 10 errors in replication. exit.")
			return
		}
		ctx, cancelFunc := context.WithDeadline(context.Background(), time.Now().Add(deadLine))

		if v.SetRequest != nil {
			_, err := client.GRPCSet(ctx, v.SetRequest)
			if err != nil {
				errorCounter = errorCounter + 1
				s.logger.Error("error in replication", zap.Error(err))
			}
		}
		if v.DelRequest != nil {
			_, err := client.GRPCDelete(ctx, v.DelRequest)
			if err != nil {
				errorCounter = errorCounter + 1
				s.logger.Error("error in replication", zap.Error(err))
			}
		}
		cancelFunc()
	}

	s.logger.Info("shutting down replicator ", zap.String("name", name), zap.String("addr", addr))
}

func (s *Store) Replicate(rr ReplicationRequest) {
	for _, v := range s.replicatorChans {
		v <- rr
	}
}
