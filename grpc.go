package grkv

import (
	"context"
	"fmt"
	"net"

	"github.com/msonawane/grkv/kvpb"
	"google.golang.org/grpc"
)

// GRPCDelete keys.
func (s *Store) GRPCDelete(ctx context.Context, req *kvpb.DeleteRequest) (*kvpb.Success, error) {
	return s.delete(ctx, req)
}

// GRPCGet keys.
func (s *Store) GRPCGet(ctx context.Context, in *kvpb.GetRequest) (*kvpb.GetResponse, error) {
	return s.get(ctx, in)
}

// GRPCSet keys.
func (s *Store) GRPCSet(ctx context.Context, req *kvpb.SetRequest) (*kvpb.Success, error) {
	return s.set(ctx, req)
}

// startGRPC server.
func (s *Store) startGRPC() error {
	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.grpcIP, s.grpcPort))
	if err != nil {
		return err
	}
	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)
	kvpb.RegisterKeyValueStoreServer(grpcServer, s)
	s.grpcServer = grpcServer
	go grpcServer.Serve(lis)
	return nil
}

func (s *Store) stopGRPC() {
	s.logger.Info("stopping GRPC server")
	s.grpcServer.GracefulStop()
}
