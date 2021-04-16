package grkv

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/dgraph-io/badger/v2"
	"github.com/hashicorp/memberlist"
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
	GRPCIP              string          // GRPC IP
	GRPCPort            int             // GRPC Port
	MLBindAddr          string          // MLBindAddr where memberlist listens on
	MLBindPort          int             // MLBindPort where memberlist listens on.
	MLName              string          // name of memberlist node
	MLMembers           string          // List of members.

}

// Store holds database.
type Store struct {
	db                  *badger.DB
	vlogTicker          *time.Ticker // runs every 1m, check size of vlog and run GC conditionally.
	mandatoryVlogTicker *time.Ticker // runs every 10m, we always run vlog GC.
	logger              *zap.Logger
	grpcServer          *grpc.Server
	grpcIP              string
	grpcPort            int
	grpcClients         map[string]kvpb.KeyValueStoreClient
	grpcClientsLock     sync.RWMutex
	replicatorChans     map[string]chan ReplicationRequest
	replicatorChansLock sync.RWMutex
	ml                  *memberlist.Memberlist
	mlNodeName          string

	kvpb.UnimplementedKeyValueStoreServer
}

// New returns Store
func New(o *Options, logger *zap.Logger) (*Store, error) {
	s := &Store{
		logger:          logger,
		grpcIP:          o.GRPCIP,
		grpcPort:        o.GRPCPort,
		grpcClients:     make(map[string]kvpb.KeyValueStoreClient),
		replicatorChans: make(map[string]chan ReplicationRequest),
	}

	// buld memberlist
	hostname, _ := os.Hostname()
	s.mlNodeName = hostname + "_" + o.MLBindAddr + "_" + strconv.Itoa(o.MLBindPort)
	mlc := memberlist.DefaultLANConfig()
	mlc.Events = s
	mlc.BindPort = o.MLBindPort
	mlc.Name = s.mlNodeName
	mlc.BindAddr = o.MLBindAddr
	ml, err := memberlist.Create(mlc)
	if err != nil {
		return nil, fmt.Errorf("mkv: can't create memberlist: %w", err)
	}
	//  the only way memberlist would be empty here, following create is if
	// the current node suddenly died. Still, we check to be safe.
	if len(ml.Members()) == 0 {
		return nil, errors.New("memberlist can't find self")
	}

	if len(o.MLMembers) > 0 {
		parts := strings.Split(o.MLMembers, ",")
		_, err := ml.Join(parts)
		if err != nil {
			return nil, err
		}
	}

	s.ml = ml

	ml.LocalNode().Meta = []byte("grpc address")

	logger.Info("members", zap.Any("members", ml.Members()))

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
	err := s.ml.Leave(1 * time.Second)
	if err != nil {
		s.logger.Error("error leaving member list", zap.Error(err))
	}
	for k, v := range s.replicatorChans {
		close(v)
		delete(s.replicatorChans, k)
	}
	s.stopGRPC()
	if s.vlogTicker != nil {
		s.vlogTicker.Stop()
	}
	if s.mandatoryVlogTicker != nil {
		s.mandatoryVlogTicker.Stop()
	}

	err = s.db.Close()
	if err != nil {
		s.logger.Error("error closing DB", zap.Error(err))
	}
	return err
}
