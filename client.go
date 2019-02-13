package operator

import (
	"net"

	"github.com/mdlayher/wireguardctrl"
	"github.com/mdlayher/wireguardctrl/wgtypes"
)

// WgDeviceConfigurator configures one WireGuard interface.
type WgDeviceConfigurator interface {
	ConfigureDevice(wgtypes.Config) error
}

// WgIdentity is the information required for a remote peer
// to connect via WireGuard.
type WgIdentity interface {
	PublicKey() []byte
	// Endpoint4 for IPv4
	Endpoint4() net.IP
	// Endpoint6 for IPv6
	Endpoint6() net.IP
	Port() int
}

// WgClient is a thin wrapper around wireguardctrl for binding config
// to a specific interface name.
type WgClient struct {
	ifName     string
	port       int
	privateKey wgtypes.Key
	*wireguardctrl.Client
}

func NewWgClient(ifName string, port int) (*WgClient, error) {
	wg, err := wireguardctrl.New()
	if err != nil {
		return nil, err
	}
	key, err := wgtypes.GeneratePrivateKey()
	if err != nil {
		return nil, err
	}
	c := WgClient{ifName: ifName, privateKey: key, port: port, Client: wg}

	if err = c.ConfigureDevice(wgtypes.Config{
		PrivateKey:   &key,
		ListenPort:   &port,
		ReplacePeers: true,
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
	return c.Client.ConfigureDevice(c.ifName, cfg)
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
