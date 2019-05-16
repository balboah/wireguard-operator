package operator

import (
	"encoding/base64"
	"errors"
	"net"

	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

// ErrPeerNotFound is returned when unable to find a matching peer while looking up
// information about it.
var ErrPeerNotFound = errors.New("peer not found")

// WgDeviceConfigurator configures one WireGuard interface and tracks IPs
// of registered peers.
type WgDeviceConfigurator interface {
	ConfigureDevice(wgtypes.Config) error
	ResolvePeerNets(wgtypes.Key) ([]net.IPNet, error)
}

// WgIdentity is the information required for a remote peer
// to connect via WireGuard.
type WgIdentity interface {
	PublicKey() []byte
	Port() int
}

// WgClient is a thin wrapper around wgctrl for binding config
// to a specific interface link.
type WgClient struct {
	link       *WgLink
	port       int
	privateKey wgtypes.Key
	*wgctrl.Client
}

func NewWgClient(link *WgLink, port int, pk string) (*WgClient, error) {
	wg, err := wgctrl.New()
	if err != nil {
		return nil, err
	}
	c := WgClient{link: link, port: port, Client: wg}
	if pk == "" {
		key, err := wgtypes.GeneratePrivateKey()
		if err != nil {
			return nil, err
		}
		c.privateKey = key
	} else {
		b, err := base64.StdEncoding.DecodeString(pk)
		if err != nil {
			return nil, err
		}
		key, err := wgtypes.NewKey(b)
		if err != nil {
			return nil, err
		}
		c.privateKey = key
	}

	if err = c.ConfigureDevice(wgtypes.Config{
		PrivateKey:   &c.privateKey,
		ListenPort:   &port,
		ReplacePeers: false,
		// Peers: []wgtypes.PeerConfig{
		// 	{
		// 		PublicKey:         peerKey,
		// 		ReplaceAllowedIPs: true,
		// 		AllowedIPs:        ips,
		// 	},
		// },
	}); err != nil {
		return nil, err
	}
	return &c, nil
}

func (c *WgClient) ConfigureDevice(cfg wgtypes.Config) error {
	return c.Client.ConfigureDevice(c.link.Name, cfg)
}

func (c *WgClient) ResolvePeerNets(key wgtypes.Key) ([]net.IPNet, error) {
	d, err := c.Device(c.link.Name)
	if err != nil {
		return nil, err
	}
	for _, p := range d.Peers {
		if p.PublicKey == key {
			return p.AllowedIPs, nil
		}
	}
	return nil, ErrPeerNotFound
}

func (c *WgClient) PublicKey() []byte {
	// convert to slice which should have been the default.
	return keyToSlice(c.privateKey.PublicKey())
}

// keyToSlice converts to slice which should have been the default.
func keyToSlice(key wgtypes.Key) []byte {
	b := make([]byte, wgtypes.KeyLen)
	for n, k := range key {
		b[n] = k
	}
	return b
}
