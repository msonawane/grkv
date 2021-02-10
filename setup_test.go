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
var err error
var lis *bufconn.Listener

const bufSize = 1024 * 1024

func TestMain(m *testing.M) {
	opts := grkv.Options{
		Path: "/tmp/grkv-test",
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

	exitVal := m.Run()

	store.Close()
	os.Exit(exitVal)

}
