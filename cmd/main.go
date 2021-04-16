package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/ardanlabs/conf"
	"github.com/msonawane/grkv"
	"github.com/msonawane/grkv/kvpb"
	"go.uber.org/zap"
)

type options struct {
	Path                string        `conf:"default:/tmp/rkv"`
	NoSync              bool          `conf:"default:false"`
	ValueLogGC          bool          `conf:"default:true"` // ValueLogGC enables a periodic goroutine that does a garbage collection of the value log while the underlying Badger is online.
	GCInterval          time.Duration `conf:"default:1m"`   // GCInterval is the interval between conditionally running the garbage collection process, based on the size of the vlog. By default, runs every 1m.
	MandatoryGCInterval time.Duration `conf:"default:10m"`  // MandatoryGCInterval is the interval between mandatory running the garbage collection process. By default, runs every 10m.
	GCThreshold         int64         `conf:"default:1000000"`
	GRPCIP              string        `conf:"default:127.0.0.1"` //GRPCIP is address where grpc listens on.
	GRPCPort            int           `conf:"default:9001"`
	MLBindAddr          string        `conf:"default:127.0.0.1"` //MLBindAddr where memberlist listens on
	MLBindPort          int           `conf:"default:8001"`      //MLBindPort where memberlist listens on.
	MLMembers           string        // comma seperated list of existing members.
}

var testData = []struct {
	key   string
	value string
}{
	{"a", "a"},
	{"b", "b"},
}

func main() {
	logger, _ := zap.NewDevelopment()
	options := options{}
	if parseError := conf.Parse(os.Args[1:], "RKV", &options); parseError != nil {
		if parseError == conf.ErrHelpWanted {
			usage, err := conf.Usage("RKV", &options)
			if err != nil {
				log.Fatal("generating usages", zap.Error(err))
			}
			fmt.Println(usage)
			return
		}
		logger.Fatal("error parsing configuration", zap.Error(parseError))
	}
	logger.Info("using options", zap.Any("opts", options))

	kvOpts := grkv.Options{
		Path:                options.Path,
		NoSync:              options.NoSync,
		ValueLogGC:          options.ValueLogGC,
		GCInterval:          options.GCInterval,
		MandatoryGCInterval: options.MandatoryGCInterval,
		GCThreshold:         options.GCThreshold,
		GRPCIP:              options.GRPCIP,
		GRPCPort:            options.GRPCPort,
		MLBindAddr:          options.MLBindAddr,
		MLBindPort:          options.MLBindPort,
		MLMembers:           options.MLMembers,
	}
	store, err := grkv.New(&kvOpts, logger)
	if err != nil {
		logger.Fatal("error creating store", zap.Error(err))
	}
	req := &kvpb.SetRequest{}
	for _, d := range testData {
		req.Data = append(req.Data, &kvpb.KeyValue{
			Key:   []byte(d.key),
			Value: []byte(d.value),
		})
	}

	time.Sleep(10 * time.Second)

	ctx := context.Background()

	success, err := store.Set(ctx, req)
	if err != nil {
		logger.Error("error Set", zap.Error(err))
	}
	if !success.Success {
		logger.Info("set failed")
	}

	time.Sleep(1 * time.Minute)
	store.Close()
}
