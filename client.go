package operator

import (
	"github.com/mdlayher/wireguardctrl"
	"github.com/mdlayher/wireguardctrl/wgtypes"
)

func NewClient(device string, port int) (*wireguardctrl.Client, error) {
	c, err := wireguardctrl.New()
	if err != nil {
		return nil, err
	}

	key, err := wgtypes.GeneratePrivateKey()
	if err != nil {
		return nil, err
	}

	if err = c.ConfigureDevice(device, wgtypes.Config{
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
	return c, nil
}
