package operator

import (
	"net"
	"testing"
)

func TestPoolIP4(t *testing.T) {
	pool, err := NewPool("10.2.0.0/16")
	if err != nil {
		t.Fatal(err)
	}
	if l := len(pool.available); l != 65534 {
		t.Error("Expected 65534 available addresses, got", l)
	}

	ip, err := pool.Allocate()
	if err != nil {
		t.Fatal(err)
	}

	if l := len(pool.available); l != 65533 {
		t.Error("Expected 65533 available addresses, got", l)
	}

	if err := pool.Free(ip); err != nil {
		t.Fatal(err)
	}
	if l := len(pool.available); l != 65534 {
		t.Error("Expected 65534 available addresses, got", l)
	}

	if ip.String() != "10.2.0.1" {
		t.Error("Wrong ip", ip)
	}

	if err = pool.Free(net.ParseIP("1.2.3.4")); err == nil {
		t.Error("expected error trying to free invalid IP")
	}
}

func TestIP4To6(t *testing.T) {
	_, prefix, err := net.ParseCIDR("fd:b10c:ad:add1:de1e:7ed::/112")
	if err != nil {
		t.Fatal(err)
	}
	ip6 := ip4To6(net.IP{10, 1, 2, 3}, prefix)
	if s := net.IP(ip6).String(); s != "fd:b10c:ad:add1:de1e:7ed:a01:203" {
		t.Log(s)
		t.Error("invalid IP")
	}
}
