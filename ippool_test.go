package operator

import (
	"net"
	"testing"
)

func TestRemovePanic(t *testing.T) {
	// Test that Remove with multiple IPs doesn't panic
	pool, err := NewPool("192.168.0.252/30")
	if err != nil {
		t.Fatal(err)
	}

	ip1, _ := pool.Allocate()
	ip2, _ := pool.Allocate()

	pool.Free(ip2)
	pool.Free(ip1)

	err = pool.Remove(ip1, ip2)
	if err != nil {
		t.Fatalf("Remove should not return error, got: %v", err)
	}

	// Pool should be empty
	if len(pool.available) != 0 {
		t.Errorf("Pool should be empty after removing both IPs, but has %d IPs", len(pool.available))
	}
}

func TestDoubleFree(t *testing.T) {
	// Test that double-free causes IP to be allocated twice
	pool, err := NewPool("10.99.0.0/30")
	if err != nil {
		t.Fatal(err)
	}

	ip1, _ := pool.Allocate()
	pool.Allocate() // empty the pool

	err = pool.Free(ip1)
	if err != nil {
		t.Fatal(err)
	}

	// Free the same IP again
	err = pool.Free(ip1)
	if err != nil {
		t.Fatal(err)
	}

	alloc1, _ := pool.Allocate()
	alloc2, _ := pool.Allocate()

	if alloc1.Equal(alloc2) {
		t.Errorf("Same IP allocated twice: %s", alloc1.String())
	}
}

func TestBasicPool(t *testing.T) {
	pool, err := NewPool("10.0.0.0/30")
	if err != nil {
		t.Fatal(err)
	}

	// Should have 2 IPs
	if l := len(pool.available); l != 2 {
		t.Errorf("Expected 2 available addresses, got %d", l)
	}

	// Allocate and free
	ip, _ := pool.Allocate()
	if err := pool.Free(ip); err != nil {
		t.Fatal(err)
	}

	// Remove an IP
	if err := pool.Remove(ip); err != nil {
		t.Fatal(err)
	}
	if l := len(pool.available); l != 1 {
		t.Errorf("Expected 1 available address after remove, got %d", l)
	}
}

func TestIP4To6(t *testing.T) {
	_, prefix, err := net.ParseCIDR("fdad:b10c:a::/48")
	if err != nil {
		t.Fatal(err)
	}
	ip6 := ip4To6(net.IP{10, 1, 2, 3}, prefix)
	if s := net.IP(ip6).String(); s != "fdad:b10c:a::a01:203" {
		t.Log(s)
		t.Error("invalid IP")
	}
}
