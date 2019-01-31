package main

import (
	"github.com/mdlayher/wireguardctrl"
	"github.com/mdlayher/wireguardctrl/wgtypes"
	log "github.com/sirupsen/logrus"
)

func main() {
	c, err := wireguardctrl.New()
	if err != nil {
		log.Fatalf("failed to open client: %v", err)
	}
	defer c.Close()

	devices, err := c.Devices()
	if err != nil {
		log.Fatalf("failed to get devices: %v", err)
	}
	for _, d := range devices {
		log.Debug(d.Name)
	}

	key, err := wgtypes.GeneratePrivateKey()
	if err != nil {
		log.Fatalf("can't generate private key: %v", err)
	}
	port := 1234
	if err = c.ConfigureDevice("wg0", wgtypes.Config{
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
		log.Fatalf("can't configure device: %v", err)
	}
}
