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
	addr, err := netlink.ParseAddr("192.168.2.1/24")
	if err != nil {
		t.Fatal(err)
	}
	if err := netlink.AddrAdd(&wg0, addr); err != nil {
		t.Fatal(err)
	}

	if _, err := NewClient("wg0", 1234); err != nil {
		t.Fatal(err)
	}
}
