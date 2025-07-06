package agent

import (
	"context"

	"github.com/pabotesu/kurohabaki-client/internal/logger"
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

// Run should block until context is cancelled
func (a *Agent) Run(ctx context.Context) {
	// Log agent startup (debug mode only)
	logger.Println("🟢 Agent.Run started")

	ctx, cancel := context.WithCancel(ctx)
	a.cancel = cancel

	// Note: Signal handling is managed in the up.go command,
	// removing duplicate signal handling here

	// Start peer watcher (debug mode only)
	logger.Println("🟢 Launching StartPeerWatcher goroutine")
	go StartPeerWatcher(ctx, a.etcdClient, a.wgIf, a.selfPubKey)

	// Block until context is done - THIS IS CRUCIAL
	<-ctx.Done()

	// Always log shutdown as it's important operational info
	logger.Println("Agent shutting down...")

	// Clean up resources
	a.wgIf.Close()
}

// Stop cancels the agent's context, triggering shutdown
func (a *Agent) Stop() {
	if a.cancel != nil {
		a.cancel()
	}
}
