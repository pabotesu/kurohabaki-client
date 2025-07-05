package wg

import (
	"net"
	"testing"

	"golang.zx2c4.com/wireguard/device"
)

func TestSamePeers(t *testing.T) {
	// Helper function to create a test public key
	createPubKey := func(byte byte) device.NoisePublicKey {
		var key device.NoisePublicKey
		for i := range key {
			key[i] = byte
		}
		return key
	}

	// Helper function to create a test UDP address
	createUDPAddr := func(ip string, port int) *net.UDPAddr {
		return &net.UDPAddr{
			IP:   net.ParseIP(ip),
			Port: port,
		}
	}

	// Helper function to create test allowed IPs
	createAllowedIP := func(cidr string) net.IPNet {
		_, ipnet, _ := net.ParseCIDR(cidr)
		return *ipnet
	}

	// Helper function to create a uint16 pointer
	uint16Ptr := func(v int) *uint16 {
		u := uint16(v)
		return &u
	}

	tests := []struct {
		name     string
		peersA   []WGPeerConfig
		peersB   []WGPeerConfig
		expected bool
	}{
		{
			name:     "Empty lists",
			peersA:   []WGPeerConfig{},
			peersB:   []WGPeerConfig{},
			expected: true,
		},
		{
			name: "Same single peer",
			peersA: []WGPeerConfig{
				{
					PublicKey:                   createPubKey(1),
					Endpoint:                    createUDPAddr("192.168.1.1", 51820),
					PersistentKeepaliveInterval: uint16Ptr(25),
					ReplaceAllowedIPs:           true,
					AllowedIPs:                  []net.IPNet{createAllowedIP("10.0.0.1/32")},
				},
			},
			peersB: []WGPeerConfig{
				{
					PublicKey:                   createPubKey(1),
					Endpoint:                    createUDPAddr("192.168.1.1", 51820),
					PersistentKeepaliveInterval: uint16Ptr(25),
					ReplaceAllowedIPs:           true,
					AllowedIPs:                  []net.IPNet{createAllowedIP("10.0.0.1/32")},
				},
			},
			expected: true,
		},
		{
			name: "Different public keys",
			peersA: []WGPeerConfig{
				{
					PublicKey: createPubKey(1),
					Endpoint:  createUDPAddr("192.168.1.1", 51820),
				},
			},
			peersB: []WGPeerConfig{
				{
					PublicKey: createPubKey(2),
					Endpoint:  createUDPAddr("192.168.1.1", 51820),
				},
			},
			expected: false,
		},
		{
			name: "Different endpoints",
			peersA: []WGPeerConfig{
				{
					PublicKey: createPubKey(1),
					Endpoint:  createUDPAddr("192.168.1.1", 51820),
				},
			},
			peersB: []WGPeerConfig{
				{
					PublicKey: createPubKey(1),
					Endpoint:  createUDPAddr("192.168.1.2", 51820),
				},
			},
			expected: false,
		},
		{
			name: "Different lengths",
			peersA: []WGPeerConfig{
				{
					PublicKey: createPubKey(1),
					Endpoint:  createUDPAddr("192.168.1.1", 51820),
				},
			},
			peersB: []WGPeerConfig{
				{
					PublicKey: createPubKey(1),
					Endpoint:  createUDPAddr("192.168.1.1", 51820),
				},
				{
					PublicKey: createPubKey(2),
					Endpoint:  createUDPAddr("192.168.1.2", 51820),
				},
			},
			expected: false,
		},
		{
			name: "Nil endpoint in one peer",
			peersA: []WGPeerConfig{
				{
					PublicKey: createPubKey(1),
					Endpoint:  nil,
				},
			},
			peersB: []WGPeerConfig{
				{
					PublicKey: createPubKey(1),
					Endpoint:  createUDPAddr("192.168.1.1", 51820),
				},
			},
			expected: false,
		},
		{
			name: "Both nil endpoints",
			peersA: []WGPeerConfig{
				{
					PublicKey: createPubKey(1),
					Endpoint:  nil,
				},
			},
			peersB: []WGPeerConfig{
				{
					PublicKey: createPubKey(1),
					Endpoint:  nil,
				},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SamePeers(tt.peersA, tt.peersB)
			if result != tt.expected {
				t.Errorf("SamePeers() = %v, want %v", result, tt.expected)
			}
		})
	}
}
