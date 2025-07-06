package agent

import (
	"context"
	"time"

	"github.com/pabotesu/kurohabaki-client/internal/etcd"
	"github.com/pabotesu/kurohabaki-client/internal/logger"
	"github.com/pabotesu/kurohabaki-client/internal/wg"
	clientv3 "go.etcd.io/etcd/client/v3"
)

func StartPeerWatcher(ctx context.Context, cli *clientv3.Client, wgIf *wg.WireGuardInterface, selfPubKey string) {
	logger.Println("StartPeerWatcher: launched") // debug mode only
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	var prevPeers []wg.WGPeerConfig

	for {
		select {
		case <-ctx.Done():
			// shutdown signal received
			logger.Println("Peer watcher shutting down...")
			return

		case <-ticker.C:
			// get current peers from etcd
			logger.Println("FetchPeers: start fetching from etcd...")

			peers, err := etcd.FetchPeers(cli, selfPubKey)
			if err != nil {
				// エラーは常に表示（重要な問題なので）
				logger.Printf("Failed to fetch peers: %s", err.Error())
				continue
			}

			// debug mode only
			logger.Printf("FetchPeers: %d node(s) fetched", len(peers))
			if logger.IsDebugMode() {
				for _, n := range peers {
					logger.Printf("Node details: %+v", n)
				}
			}

			currentPeers, err := wg.ConvertNodesToPeers(peers)
			if err != nil {
				logger.Printf("Failed to convert nodes to peers: %v", err)
				continue
			}

			// debug mode only
			logger.Printf("Peers converted: %d", len(currentPeers))

			if !wg.SamePeers(prevPeers, currentPeers) {
				logger.Println("Peer list updated, applying to interface...")
				if err := wgIf.UpdatePeers(currentPeers); err != nil {
					logger.Printf("Failed to update WireGuard peers: %v", err)
				} else {
					prevPeers = currentPeers
					logger.Println("Peers updated successfully")
				}
			} else {
				logger.Println("No peer changes detected")
			}
		}
	}
}
