package agent

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/pabotesu/kurohabaki-client/internal/wg"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type Agent struct {
	wgIf       *wg.WireGuardInterface
	etcdClient *clientv3.Client
	selfPubKey string
	cancel     context.CancelFunc
}

func New(wgIf *wg.WireGuardInterface, etcdClient *clientv3.Client, selfPubKey string) *Agent {
	return &Agent{
		wgIf:       wgIf,
		etcdClient: etcdClient,
		selfPubKey: selfPubKey,
	}
}

func (a *Agent) Run(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	a.cancel = cancel

	// Signal handling for graceful shutdown
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		sig := <-sigCh
		log.Printf("Received signal: %v, shutting down agent...", sig)
		cancel()
	}()

	// Start peer watcher
	go StartPeerWatcher(ctx, a.etcdClient, a.wgIf, a.selfPubKey)

	// Block until cancelled
	<-ctx.Done()
	log.Println("Agent shutting down...")

	// Clean up
	a.wgIf.Close()
}
