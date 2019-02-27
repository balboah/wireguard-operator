// +build integration

package operator

import (
	"testing"

	"github.com/vishvananda/netlink"
)

func TestWgLinkReuseInterface(t *testing.T) {
	wg0 := WgLink{&netlink.LinkAttrs{Name: "wg0"}}
	defer func() { netlink.LinkDel(&wg0) }()

	if err := netlink.LinkAdd(&wg0); err != nil {
		t.Fatal(err)
	}

	if _, err := NewWgLink("wg0"); err != nil {
		t.Fatal(err)
	}
}

func TestWgLinkDeleteInterface(t *testing.T) {
	wg, err := NewWgLink("wg0")
	if err != nil {
		t.Fatal(err)
	}

	if _, err := netlink.LinkByName("wg0"); err != nil {
		t.Fatal(err)
	}
	if err := wg.Close(); err != nil {
		t.Fatal(err)
	}
	_, err = netlink.LinkByName("wg0")
	switch err.(type) {
	case netlink.LinkNotFoundError:
		// expected
	default:
		t.Error(err)
	}
}
