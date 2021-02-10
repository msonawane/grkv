package grkv

import (
	"context"
	"fmt"
	"net"

	"github.com/msonawane/grkv/kvpb"
	"go.uber.org/zap"
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
	s.logger.Info("GRPCSet on ", zap.String("node", s.mlNodeName))

	return s.set(ctx, req)
}

// startGRPC server.
func (s *Store) startGRPC() error {
	addr := fmt.Sprintf("%s:%d", s.grpcIP, s.grpcPort)
	s.logger.Info("starting grpc server on", zap.String("grpc_addr", addr))
	lis, err := net.Listen("tcp", addr)
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

func (s *Store) newGRPCClient(name, addr string) error {
	s.grpcClientsLock.Lock()
	defer s.grpcClientsLock.Unlock()

	if s.mlNodeName != name {
		addr = fmt.Sprintf("%s:%d", addr, s.grpcPort)
		s.logger.Info("creating grpc client for", zap.String("node", name), zap.String("addr", addr))
		ctx := context.Background()
		conn, err := grpc.DialContext(ctx, addr, grpc.WithInsecure())
		if err != nil {
			s.logger.Error("err creating grpc client", zap.Error(err))
			return err
		}
		client := kvpb.NewKeyValueStoreClient(conn)
		s.grpcClients[name] = client
		fmt.Printf("client: %#v\n", client)

	}
	return nil

}

func (s *Store) removeGRPCClient(nodeName string) {
	s.grpcClientsLock.Lock()
	defer s.grpcClientsLock.Unlock()
	delete(s.grpcClients, nodeName)
}
