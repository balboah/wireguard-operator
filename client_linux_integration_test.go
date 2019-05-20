// +build integration

package operator

import (
	"encoding/base64"
	"testing"

	"github.com/vishvananda/netlink"
)

func TestWgClient(t *testing.T) {
	wg0, err := NewWgLink("wgo-test")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { wg0.Close() }()

	addr4, err := netlink.ParseAddr("192.168.2.1/24")
	if err != nil {
		t.Fatal(err)
	}
	addr6, err := netlink.ParseAddr("fd01::/64")
	if err != nil {
		t.Fatal(err)
	}
	if err := wg0.AddrAdd(addr4, addr6); err != nil {
		t.Fatal(err)
	}

	c, err := NewWgClient(wg0, 1234, "")
	if err != nil {
		t.Fatal(err)
	}
	if p := c.PublicKey(); len(p) != 32 {
		t.Log(p)
		t.Error("did not generate key")
	}
	if c, err = NewWgClient(wg0, 1234, "QPaixVanELV/q/fTgSP3xtP2iJX2Alr8uXyWlU5NzEw="); err != nil {
		t.Fatal(err)
	}
	pub := base64.StdEncoding.EncodeToString(c.PublicKey())
	if pub != "VTqwfWmU9DKd48UrZ93BM7/9Rpn86Ang89WbUvSKwx8=" {
		t.Log(pub)
		t.Error("public key is invalid")
	}
}
