// +build integration

package operator

import (
	"testing"

	"github.com/vishvananda/netlink"
)

func TestConfigureInterface(t *testing.T) {
	wg0 := WgLink{netlink.LinkAttrs{Name: "wg0"}}
	defer func() { netlink.LinkDel(&wg0) }()

	if err := netlink.LinkAdd(&wg0); err != nil {
		t.Fatal(err)
	}
	addr4, err := netlink.ParseAddr("192.168.2.1/24")
	if err != nil {
		t.Fatal(err)
	}
	if err := netlink.AddrAdd(&wg0, addr4); err != nil {
		t.Fatal(err)
	}
	addr6, err := netlink.ParseAddr("fd01::/64")
	if err != nil {
		t.Fatal(err)
	}
	if err := netlink.AddrAdd(&wg0, addr6); err != nil {
		t.Fatal(err)
	}

	c, err := NewWgClient("wg0", 1234)
	if err != nil {
		t.Fatal(err)
	}
	d, err := c.Device("wg0")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(d.PublicKey.String())

	// c.ConfigureDevice("wg0", wgtypes.Config{
	// 	Peers: []wgtypes.PeerConfig{
	// 		{
	// 			PublicKey:         peerKey,
	// 			ReplaceAllowedIPs: false, // aka append peer
	// 			AllowedIPs:        ips,
	// 		},
	// 	},
	// })
}
