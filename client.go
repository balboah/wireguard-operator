package operator

import (
	"encoding/base64"

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
	Port() int
}

// WgClient is a thin wrapper around wireguardctrl for binding config
// to a specific interface link.
type WgClient struct {
	link       *WgLink
	port       int
	privateKey wgtypes.Key
	*wireguardctrl.Client
}

func NewWgClient(link *WgLink, port int, pk string) (*WgClient, error) {
	wg, err := wireguardctrl.New()
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
