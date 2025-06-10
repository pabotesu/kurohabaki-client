package agent

import (
	"context"
	"log"
	"time"

	"github.com/pabotesu/kurohabaki-client/internal/etcd"
	"github.com/pabotesu/kurohabaki-client/internal/wg"
	clientv3 "go.etcd.io/etcd/client/v3"
)

func StartPeerWatcher(ctx context.Context, cli *clientv3.Client, wgIf *wg.WireGuardInterface, selfPubKey string) {
	log.Println("Starting peer watcher...")

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	var prevPeers []wg.WGPeerConfig

	for {
		select {
		case <-ctx.Done():
			log.Println("Peer watcher shutting down...")
			return

		case <-ticker.C:
			nodes, err := etcd.FetchPeers(cli, selfPubKey)
			if err != nil {
				log.Printf("Failed to fetch peers from etcd: %v", err)
				continue
			}

			log.Println("[PeerWatcher] Fetched nodes from etcd:")
			for _, node := range nodes {
				log.Printf("[PeerWatcher] etcd node: PubKey=%s, IP=%s, Endpoint=%s", node.PublicKey, node.IP, node.Endpoint)
			}

			currentPeers, err := wg.ConvertNodesToPeers(nodes)
			if err != nil {
				log.Printf("Failed to convert nodes to peers: %v", err)
				continue
			}

			if !wg.SamePeers(prevPeers, currentPeers) {
				log.Println("Peer list updated, applying to interface...")
				if err := wgIf.UpdatePeers(currentPeers); err != nil {
					log.Printf("Failed to update WireGuard peers: %v", err)
				} else {
					prevPeers = currentPeers
				}
			}
		}
	}

}
