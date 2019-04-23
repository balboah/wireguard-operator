package operator

import (
	"encoding/binary"
	"errors"
	"net"
	"sync"
)

// IPPool knows how to allocate IPs and free previously allocated ones.
type IPPool interface {
	Allocate() (net.IP, error)
	Free(net.IP) error
}

// Pool is a pool of available IP numbers for allocation.
type Pool struct {
	network   *net.IPNet
	available []net.IP
	allocMu   sync.Mutex
}

// NewPool creates a new pool with a CIDR.
func NewPool(cidr string) (*Pool, error) {
	networkip, network, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, err
	}

	mask, size := network.Mask.Size()
	maskbits := binary.BigEndian.Uint32(network.Mask)
	max := uint32(1 << uint(size-mask))
	// Remove unusable host ranges
	max -= 2

	available := make([]net.IP, max)
	networkipbits := maskbits & binary.BigEndian.Uint32(networkip.To4())
	for ; max > 0; max-- {
		ip := make([]byte, 4)
		binary.BigEndian.PutUint32(ip, uint32(networkipbits|max))
		available[max-1] = net.IP(ip)
	}

	return &Pool{
		network:   network,
		available: available,
	}, nil
}

// Allocate assigns a new IP to the pool for use.
func (p *Pool) Allocate() (ip net.IP, err error) {
	p.allocMu.Lock()
	defer p.allocMu.Unlock()

	if len(p.available) > 0 {
		ip = p.available[0]
		p.available = p.available[1:]
	} else {
		err = errors.New("No more IPs currently available")
	}
	return
}

// Free returns the IP to the pool to be used by other allocations.
func (p *Pool) Free(ip net.IP) error {
	if !p.network.Contains(ip) {
		return errors.New("IP is not part of this pool")
	}
	p.allocMu.Lock()
	defer p.allocMu.Unlock()

	p.available = append([]net.IP{ip}, p.available...)
	return nil
}

// ip4To6 will prefix IPv4 with the IPv6 network to create an IPv6 address.
func ip4To6(ip4 net.IP, ip6prefix *net.IPNet) (ip6 net.IP) {
	b6 := ip6prefix.IP.To16()
	return append(b6[:12], ip4.To4()...)
}
