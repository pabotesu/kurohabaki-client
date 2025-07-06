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
	logger.Println("StartPeerWatcher: launched") // デバッグモードでのみ表示

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	var prevPeers []wg.WGPeerConfig

	for {
		select {
		case <-ctx.Done():
			// シャットダウン時のログはデバッグモードでのみ表示
			logger.Println("Peer watcher shutting down...")
			return

		case <-ticker.C:
			// 定期的なフェッチ開始のログはデバッグモードでのみ表示
			logger.Println("FetchPeers: start fetching from etcd...")

			peers, err := etcd.FetchPeers(cli, selfPubKey)
			if err != nil {
				// エラーは常に表示（重要な問題なので）
				logger.Printf("Failed to fetch peers: %s", err.Error())
				continue
			}

			// 以下のログはデバッグモードでのみ表示
			logger.Printf("FetchPeers: %d node(s) fetched", len(peers))
			if logger.IsDebugMode() {
				for _, n := range peers {
					logger.Printf("Node details: %+v", n)
				}
			}

			currentPeers, err := wg.ConvertNodesToPeers(peers)
			if err != nil {
				// 変換エラーは重要なので常に表示
				logger.Printf("Failed to convert nodes to peers: %v", err)
				continue
			}

			// デバッグ情報
			logger.Printf("Peers converted: %d", len(currentPeers))

			if !wg.SamePeers(prevPeers, currentPeers) {
				// ピア変更は重要なので非デバッグモードでも表示
				logger.Println("Peer list updated, applying to interface...")
				if err := wgIf.UpdatePeers(currentPeers); err != nil {
					logger.Printf("Failed to update WireGuard peers: %v", err)
				} else {
					prevPeers = currentPeers
					logger.Println("Peers updated successfully")
				}
			} else {
				// 変更なしはデバッグモードでのみ表示
				logger.Println("No peer changes detected")
			}
		}
	}
}
