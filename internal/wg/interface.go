package wg

import (
	"encoding/hex"
	"fmt"
	"os/exec"
	"strings"
	"sync"

	"github.com/pabotesu/kurohabaki-client/internal/logger"
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
func NewWireGuardInterface(ifname string) (*WireGuardInterface, error) {
	// Create the TUN device
	tunDev, err := tun.CreateTUN(ifname, device.DefaultMTU)
	if err != nil {
		return nil, fmt.Errorf("failed to create TUN device: %w", err)
	}

	// TUNデバイス作成のログはデバッグモードでのみ表示されるようにloggerを使用
	logger.Println("Created TUN device:", ifname)

	// Set logging for WireGuard device based on debug mode
	// Set log level to error by default; change to LogLevelVerbose for more detailed logs if needed.
	logLevel := device.LogLevelError // only log errors by default

	// デバッグモードが有効な場合はより詳細なログを表示
	if logger.IsDebugMode() {
		logLevel = device.LogLevelVerbose
	}

	// Create WireGuard device with appropriate log level
	dev := device.NewDevice(tunDev, conn.NewDefaultBind(), device.NewLogger(logLevel, fmt.Sprintf("[WG-%s] ", ifname)))

	return &WireGuardInterface{
		ifName: ifname,
		dev:    dev,
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
	privateKeyHex := hex.EncodeToString((*cfg.PrivateKey)[:])
	if err := w.dev.IpcSet(fmt.Sprintf("private_key=%s\n", privateKeyHex)); err != nil {
		return fmt.Errorf("failed to set private_key: %w", err)
	}
	// Apply peer settings
	for _, peer := range cfg.Peers {
		var sb strings.Builder

		// Encode public key to base64
		publicKeyHex := hex.EncodeToString(peer.PublicKey[:])
		sb.WriteString("public_key=" + publicKeyHex + "\n")

		if peer.Endpoint != nil {
			sb.WriteString("endpoint=" + peer.Endpoint.String() + "\n")
		}
		if peer.PersistentKeepaliveInterval != nil {
			sb.WriteString(fmt.Sprintf("persistent_keepalive_interval=%d\n", *peer.PersistentKeepaliveInterval))
		}

		sb.WriteString("replace_allowed_ips=true\n")

		for _, ipnet := range peer.AllowedIPs {
			sb.WriteString("allowed_ip=" + ipnet.String() + "\n")
		}

		if err := w.dev.IpcSet(sb.String()); err != nil {
			return err
		}
	}

	// Add route to the peer subnet (Linux only)
	if len(cfg.Routes) > 0 {
		for _, route := range cfg.Routes {
			// ルート追加のログ
			logger.Printf("Adding route to %s via %s", route, w.ifName)
			cmd := exec.Command("ip", "route", "add", route, "dev", w.ifName)
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("failed to add route: %w", err)
			}
		}
	}

	// インターフェース起動のログ
	logger.Println("WireGuard interface is up")

	return nil
}

func (w *WireGuardInterface) UpdatePeers(peers []WGPeerConfig) error {
	w.lock.Lock()
	defer w.lock.Unlock()

	var sb strings.Builder

	for _, peer := range peers {
		publicKeyHex := hex.EncodeToString(peer.PublicKey[:])
		sb.WriteString("public_key=" + publicKeyHex + "\n")

		if peer.Endpoint != nil {
			sb.WriteString("endpoint=" + peer.Endpoint.String() + "\n")
		}
		if peer.PersistentKeepaliveInterval != nil {
			sb.WriteString(fmt.Sprintf("persistent_keepalive_interval=%d\n", *peer.PersistentKeepaliveInterval))
		}

		sb.WriteString("replace_allowed_ips=true\n")
		for _, ipnet := range peer.AllowedIPs {
			sb.WriteString("allowed_ip=" + ipnet.String() + "\n")
			logger.Printf("📌 AllowedIP: %s", ipnet.String())
		}
	}

	return w.dev.IpcSet(sb.String())
}

// Close shuts down the WireGuard device.
func (w *WireGuardInterface) Close() {
	w.lock.Lock()
	defer w.lock.Unlock()

	if w.dev != nil {
		w.dev.Close()
	}
}
