package wg

import (
	"log"
	"net"

	"github.com/pabotesu/kurohabaki-client/config"
	"github.com/pabotesu/kurohabaki-client/internal/etcd"
	"golang.zx2c4.com/wireguard/device"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

type WGConfig struct {
	PrivateKey   *device.NoisePrivateKey
	ListenPort   *int
	ReplacePeers bool
	Peers        []WGPeerConfig
	Routes       []string
}

type WGPeerConfig struct {
	PublicKey                   device.NoisePublicKey
	Endpoint                    *net.UDPAddr
	PersistentKeepaliveInterval *uint16
	ReplaceAllowedIPs           bool
	AllowedIPs                  []net.IPNet
}

func BuildWGConfig(cfg *config.Config) *WGConfig {
	privateKey := mustParseKey(cfg.Interface.PrivateKey)
	// Replace 'ServerPeer' with the correct field name from config.Config, e.g., 'Peer' or 'Server'
	publicKey := mustParseKey(cfg.ServerConfig.PublicKey)

	devicePrivateKey := device.NoisePrivateKey{}
	copy(devicePrivateKey[:], privateKey[:])
	devicePublicKey := device.NoisePublicKey{}
	copy(devicePublicKey[:], publicKey[:])
	endpoint := resolveUDPAddr(cfg.ServerConfig.Endpoint)
	allowedIP := parseCIDR(cfg.ServerConfig.AllowedIPs)

	return &WGConfig{
		PrivateKey:   &devicePrivateKey,
		ListenPort:   nil, // optional: set to nil for auto
		ReplacePeers: true,
		Peers: []WGPeerConfig{
			{
				PublicKey:                   devicePublicKey,
				Endpoint:                    endpoint,
				PersistentKeepaliveInterval: uint16Ptr(cfg.ServerConfig.PersistentKeepalive),
				ReplaceAllowedIPs:           true,
				AllowedIPs: []net.IPNet{
					allowedIP,
				},
			},
		},
		Routes: cfg.Interface.Routes,
	}
}

// mustParseKey converts a base64 WireGuard key string to wgtypes.Key or panics if invalid.
func mustParseKey(b64 string) wgtypes.Key {
	key, err := wgtypes.ParseKey(b64)
	if err != nil {
		log.Fatalf("invalid WireGuard key: %v", err)
	}
	return key
}

func resolveUDPAddr(endpoint string) *net.UDPAddr {
	addr, err := net.ResolveUDPAddr("udp", endpoint)
	if err != nil {
		log.Fatalf("invalid endpoint format: %v", err)
	}
	return addr
}

func parseCIDR(cidr string) net.IPNet {
	_, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		log.Fatalf("invalid CIDR: %v", err)
	}
	return *ipnet
}

func uint16Ptr(v int) *uint16 {
	u := uint16(v)
	return &u
}

func ConvertNodesToPeers(nodes []etcd.Node) ([]WGPeerConfig, error) {
	var peers []WGPeerConfig
	for _, n := range nodes {
		pubKey := mustParseDevicePublicKey(n.PublicKey)
		endpoint, err := net.ResolveUDPAddr("udp", n.Endpoint)
		if err != nil {
			return nil, err
		}
		_, ipnet, err := net.ParseCIDR(n.IP + "/32")
		if err != nil {
			return nil, err
		}

		peers = append(peers, WGPeerConfig{
			PublicKey:                   pubKey,
			Endpoint:                    endpoint,
			AllowedIPs:                  []net.IPNet{*ipnet},
			ReplaceAllowedIPs:           true,
			PersistentKeepaliveInterval: uint16Ptr(5),
		})
	}
	return peers, nil
}

// mustParseDevicePublicKey converts a base64 WireGuard key string to wgtypes.Key or panics if invalid.
func mustParseDevicePublicKey(b64 string) device.NoisePublicKey {
	key := mustParseKey(b64)
	var npk device.NoisePublicKey
	copy(npk[:], key[:])
	return npk
}

func SamePeers(a, b []WGPeerConfig) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].PublicKey != b[i].PublicKey {
			return false
		}
		if a[i].Endpoint == nil || b[i].Endpoint == nil {
			if a[i].Endpoint != b[i].Endpoint {
				return false
			}
		} else if a[i].Endpoint.String() != b[i].Endpoint.String() {
			return false
		}
		// 他のフィールドも必要なら追加
	}
	return true
}
