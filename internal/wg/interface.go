package wg

import (
	"encoding/base64"
	"fmt"
	"log"
	"os/exec"
	"strings"
	"sync"

	"golang.zx2c4.com/wireguard/conn"
	"golang.zx2c4.com/wireguard/device"
	"golang.zx2c4.com/wireguard/tun"
)

type WireGuardInterface struct {
	ifName string
	dev    *device.Device
	lock   sync.Mutex
}

// NewWireGuardInterface creates and initializes a new WireGuard TUN interface (Linux only)
func NewWireGuardInterface(name string) (*WireGuardInterface, error) {
	tunDev, err := tun.CreateTUN(name, device.DefaultMTU)
	if err != nil {
		log.Fatalf("failed to create TUN device: %v", err)
	}
	log.Printf("Created TUN device: %s\n", name)

	bind := conn.NewDefaultBind() // エラー返さない
	if err != nil {
		log.Fatalf("failed to create bind: %v", err)
	}

	logger := device.NewLogger(device.LogLevelVerbose, fmt.Sprintf("[WG-%s] ", name))

	wgDev := device.NewDevice(tunDev, bind, logger)

	return &WireGuardInterface{
		ifName: name,
		dev:    wgDev,
	}, nil
}

// AddAddress assigns an IP address to the interface (Linux only)
func (w *WireGuardInterface) AddAddress(ipWithCIDR string) error {
	cmd := exec.Command("ip", "addr", "add", ipWithCIDR, "dev", w.ifName)
	return cmd.Run()
}

// SetUpInterface brings the interface up (Linux only)
func (w *WireGuardInterface) SetUpInterface() error {
	cmd := exec.Command("ip", "link", "set", "up", "dev", w.ifName)
	return cmd.Run()
}

func (w *WireGuardInterface) Up(cfg *WGConfig) error {
	w.lock.Lock()
	defer w.lock.Unlock()

	// Encode private key to base64
	base64PrivateKey := base64.StdEncoding.EncodeToString((*cfg.PrivateKey)[:])
	if err := w.dev.IpcSet(fmt.Sprintf("private_key=%s\n", base64PrivateKey)); err != nil {
		return err
	}

	// Apply peer settings
	for _, peer := range cfg.Peers {
		var sb strings.Builder

		// Encode public key to base64
		base64PublicKey := base64.StdEncoding.EncodeToString(peer.PublicKey[:])
		sb.WriteString("public_key=" + base64PublicKey + "\n")

		if peer.Endpoint != nil {
			sb.WriteString("endpoint=" + peer.Endpoint.String() + "\n")
		}
		if peer.PersistentKeepaliveInterval != nil {
			sb.WriteString(fmt.Sprintf("persistent_keepalive_interval=%d\n", *peer.PersistentKeepaliveInterval))
		}
		for _, ipnet := range peer.AllowedIPs {
			sb.WriteString("allowed_ip=" + ipnet.String() + "\n")
		}
		sb.WriteString("replace_allowed_ips=true\n")

		if err := w.dev.IpcSet(sb.String()); err != nil {
			return err
		}
	}

	return nil
}
