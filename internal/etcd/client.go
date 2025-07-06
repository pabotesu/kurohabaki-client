package etcd

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/pabotesu/kurohabaki-client/internal/logger"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
)

// ConfigureEtcdLogger sets up etcd client logging based on debug mode
func ConfigureEtcdLogger(debug bool) {
	var zapLogConfig zap.Config
	if debug {
		// In debug mode, use development config with more verbose output
		zapLogConfig = zap.NewDevelopmentConfig()
	} else {
		// In production mode, only show critical errors
		zapLogConfig = zap.NewProductionConfig()
		zapLogConfig.Level = zap.NewAtomicLevelAt(zap.ErrorLevel)
	}

	zapLogger, _ := zapLogConfig.Build()

	// Set the global zap logger which etcd client will use
	zap.ReplaceGlobals(zapLogger)
}

type Node struct {
	PublicKey string
	IP        string
	Endpoint  string
	LastSeen  time.Time
}

func FetchPeers(cli *clientv3.Client, selfPubKey string) ([]Node, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := cli.Get(ctx, "/kurohabaki/nodes/", clientv3.WithPrefix())
	if err != nil {
		return nil, fmt.Errorf("failed to fetch peers from etcd: %w", err)
	}

	nodesMap := make(map[string]*Node)

	for _, kv := range resp.Kvs {
		parts := strings.Split(string(kv.Key), "/")
		if len(parts) < 5 {
			continue
		}
		pubKey := parts[3]
		field := parts[4]

		if pubKey == selfPubKey {
			logger.Printf("🚫 Skipping self pubKey: %s", pubKey)
			continue
		}

		node := nodesMap[pubKey]
		if node == nil {
			node = &Node{PublicKey: pubKey}
			nodesMap[pubKey] = node
		}

		switch field {
		case "ip":
			node.IP = string(kv.Value)
		case "endpoint":
			node.Endpoint = string(kv.Value)
		case "last_seen":
			t, _ := time.Parse(time.RFC3339, string(kv.Value))
			node.LastSeen = t
		}
	}

	var peers []Node
	for _, n := range nodesMap {
		if n.IP != "" && n.Endpoint != "" {
			peers = append(peers, *n)
		}
	}

	return peers, nil
}
