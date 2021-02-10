package grkv_test

import (
	"log"
	"os"
	"testing"

	"github.com/msonawane/grkv"
	"github.com/msonawane/grkv/kvpb"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

var store *grkv.Store
var store2 *grkv.Store

var err error
var lis *bufconn.Listener

const bufSize = 1024 * 1024

func TestMain(m *testing.M) {
	opts := grkv.Options{
		Path:       "/tmp/grkv-test1",
		GRPCIP:     "127.0.0.1",
		GRPCPort:   9001,
		MLBindAddr: "127.0.0.1",
		MLBindPort: 8001,
		MLMembers:  "127.0.0.1:8001",
	}
	logger, _ := zap.NewDevelopment()
	store, err = grkv.New(&opts, logger)
	if err != nil {
		logger.Fatal("error creating store", zap.Error(err))
	}

	lis = bufconn.Listen(bufSize)

	s := grpc.NewServer()
	kvpb.RegisterKeyValueStoreServer(s, store)
	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatalf("Server exited with error: %v", err)
		}
	}()

	opts2 := grkv.Options{
		Path:       "/tmp/grkv-test2",
		GRPCIP:     "127.0.0.2",
		GRPCPort:   9001,
		MLBindAddr: "127.0.0.2",
		MLBindPort: 8001,
		MLMembers:  "127.0.0.1:8001",
	}

	store2, err = grkv.New(&opts2, logger)
	if err != nil {
		logger.Fatal("error creating store2", zap.Error(err))
	}

	exitVal := m.Run()

	store.Close()
	store2.Close()

	os.Exit(exitVal)

}
