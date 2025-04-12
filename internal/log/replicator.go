package log

import (
	"context"
	"sync"

	api "github.com/nitishsharma2825/proglog/api/v1"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type Replicator struct {
	mu sync.Mutex

	DialOptions []grpc.DialOption
	LocalServer api.LogClient // local client to log the messages from remote servers

	logger  *zap.Logger
	servers map[string]chan struct{} // servers -> chan for replication
	closed  bool                     // indicates that replicator is closed, no new servers to replicate
	close   chan struct{}            // if this is closed, replication is closed
}

func (r *Replicator) Join(name, addr string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.init()
	if r.closed {
		return nil
	}

	if _, ok := r.servers[name]; ok {
		// already replicating so skip
		return nil
	}
	// add this server to list of currently replicating servers
	r.servers[name] = make(chan struct{})
	go r.replicate(addr, r.servers[name])
	return nil
}

func (r *Replicator) replicate(addr string, leave chan struct{}) {
	cc, err := grpc.NewClient(addr, r.DialOptions...)
	if err != nil {
		r.logError(err, "failed to dial", addr)
	}
	defer cc.Close()

	client := api.NewLogClient(cc)
	ctx := context.Background()
	stream, err := client.ConsumeStream(
		ctx,
		&api.ConsumeRequest{
			Offset: 0,
		},
	)
	if err != nil {
		r.logError(err, "failed to consume", addr)
		return
	}
	records := make(chan *api.Record)
	go func() {
		for {
			recv, err := stream.Recv()
			if err != nil {
				r.logError(err, "failed to receive", addr)
				return
			}
			records <- recv.Record
		}
	}()

	for {
		select {
		case <-r.close: // global replication is closed
			return
		case <-leave: // replication for this server is closed
			return
		case record := <-records:
			_, err = r.LocalServer.Produce(
				ctx,
				&api.ProduceRequest{
					Record: record,
				},
			)
			if err != nil {
				r.logError(err, "failed to produce", addr)
				return
			}
		}
	}
}

func (r *Replicator) Leave(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.init()
	if _, ok := r.servers[name]; !ok {
		return nil
	}
	close(r.servers[name])
	delete(r.servers, name)
	return nil
}

func (r *Replicator) init() {
	if r.logger == nil {
		r.logger = zap.L().Named("replicator")
	}
	if r.servers == nil {
		r.servers = make(map[string]chan struct{})
	}
	if r.close == nil {
		r.close = make(chan struct{})
	}
}

func (r *Replicator) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.init()

	if r.closed {
		return nil
	}
	r.closed = true
	close(r.close)
	return nil
}

func (r *Replicator) logError(err error, msg, addr string) {
	r.logger.Error(
		msg,
		zap.String("addr", addr),
		zap.Error(err),
	)
}
