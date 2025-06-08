package wg

import (
	"log"
	"net"

	"github.com/pabotesu/kurohabaki-client/config"
	"golang.zx2c4.com/wireguard/device"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

type WGConfig struct {
	PrivateKey   *device.NoisePrivateKey
	ListenPort   *int
	ReplacePeers bool
	Peers        []WGPeerConfig
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
	publicKey := mustParseKey(cfg.ServerPeer.PublicKey)
	devicePrivateKey := device.NoisePrivateKey{}
	copy(devicePrivateKey[:], privateKey[:])
	devicePublicKey := device.NoisePublicKey{}
	copy(devicePublicKey[:], publicKey[:])
	endpoint := resolveUDPAddr(cfg.ServerPeer.Endpoint)
	allowedIP := parseCIDR(cfg.ServerPeer.AllowedIPs)

	return &WGConfig{
		PrivateKey:   &devicePrivateKey,
		ListenPort:   nil, // optional: set to nil for auto
		ReplacePeers: true,
		Peers: []WGPeerConfig{
			{
				PublicKey:                   devicePublicKey,
				Endpoint:                    endpoint,
				PersistentKeepaliveInterval: uint16Ptr(cfg.ServerPeer.PersistentKeepalive),
				ReplaceAllowedIPs:           true,
				AllowedIPs: []net.IPNet{
					allowedIP,
				},
			},
		},
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
