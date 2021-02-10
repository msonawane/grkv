package grkv

import (
	"time"

	"github.com/dgraph-io/badger/v2"
	"github.com/msonawane/grkv/kvpb"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// Options for kv.
type Options struct {
	Path                string          // Path is the directory path to the Badger db to use.
	BadgerOptions       *badger.Options // BadgerOptions contains any specific Badger options you might want to specify.
	NoSync              bool            // NoSync causes the database to skip fsync calls after each write to the log. This is unsafe, so it should be used with caution.
	ValueLogGC          bool            // ValueLogGC enables a periodic goroutine that does a garbage collection of the value log while the underlying Badger is online.
	GCInterval          time.Duration   // GCInterval is the interval between conditionally running the garbage collection process, based on the size of the vlog. By default, runs every 1m.
	MandatoryGCInterval time.Duration   // MandatoryGCInterval is the interval between mandatory running the garbage collection process. By default, runs every 10m.
	GCThreshold         int64           // GCThreshold sets threshold in bytes for the vlog size to be included in the garbage collection cycle. By default, 1GB.
	GRPCIP              string
	GRPCPort            int
}

// Store holds database.
type Store struct {
	db                  *badger.DB
	dbPath              string
	vlogTicker          *time.Ticker // runs every 1m, check size of vlog and run GC conditionally.
	mandatoryVlogTicker *time.Ticker // runs every 10m, we always run vlog GC.
	logger              *zap.Logger
	grpcServer          *grpc.Server
	grpcIP              string
	grpcPort            int
	kvpb.UnimplementedKeyValueStoreServer
}

// New returns Store
func New(o *Options, logger *zap.Logger) (*Store, error) {
	s := &Store{logger: logger}
	// build badger options
	if o.BadgerOptions == nil {
		defaultOpts := badger.DefaultOptions(o.Path)
		o.BadgerOptions = &defaultOpts
	}
	o.BadgerOptions.SyncWrites = !o.NoSync

	// Open badgerdb
	db, err := badger.Open(*o.BadgerOptions)
	if err != nil {
		return nil, err
	}
	s.db = db

	if o.ValueLogGC {
		var gcInterval time.Duration
		var mandatoryGCInterval time.Duration
		var threshold int64

		if gcInterval = 1 * time.Minute; o.GCInterval != 0 {
			gcInterval = o.GCInterval
		}
		if mandatoryGCInterval = 10 * time.Minute; o.MandatoryGCInterval != 0 {
			mandatoryGCInterval = o.MandatoryGCInterval
		}
		if threshold = int64(1 << 30); o.GCThreshold != 0 {
			threshold = o.GCThreshold
		}

		s.vlogTicker = time.NewTicker(gcInterval)
		s.mandatoryVlogTicker = time.NewTicker(mandatoryGCInterval)
		go s.runVlogGC(threshold)
	}
	err = s.startGRPC()
	return s, err
}

func (s *Store) runVlogGC(threshold int64) {
	s.logger.Info("starting runVlogGC")
	// Get initial size on start.
	_, lastVlogSize := s.db.Size()

	runGC := func() {
		var err error
		for err == nil {
			// If a GC is successful, immediately run it again.
			err = s.db.RunValueLogGC(0.7)
		}
		_, lastVlogSize = s.db.Size()
	}

	for {
		select {
		case <-s.vlogTicker.C:
			_, currentVlogSize := s.db.Size()
			if currentVlogSize < lastVlogSize+threshold {
				continue
			}
			runGC()
		case <-s.mandatoryVlogTicker.C:
			runGC()
		}
	}
}

// Close is used to gracefully close the Store.
func (s *Store) Close() error {
	s.logger.Info("closing store")
	s.stopGRPC()
	if s.vlogTicker != nil {
		s.vlogTicker.Stop()
	}
	if s.mandatoryVlogTicker != nil {
		s.mandatoryVlogTicker.Stop()
	}

	dbCloseErr := s.db.Close()
	if dbCloseErr != nil {
		s.logger.Error("error closing DB", zap.Error(dbCloseErr))
	}
	return dbCloseErr
}
