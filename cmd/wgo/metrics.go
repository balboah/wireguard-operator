package main

import (
	"expvar"
	"net"

	operator "github.com/balboah/wireguard-operator"
)

type metricPool struct {
	operator.IPPool
	metrics *expvar.Map
}

func poolWithMetrics(p operator.IPPool, err error) (*metricPool, error) {
	return &metricPool{
		p,
		expvar.NewMap("pool"),
	}, err
}

func (p *metricPool) Allocate() (net.IP, error) {
	ip, err := p.IPPool.Allocate()
	if err != nil {
		return nil, err
	}
	p.metrics.Add("allocated", 1)
	return ip, nil
}

func (p *metricPool) Free(ip net.IP) error {
	err := p.IPPool.Free(ip)
	if err != nil {
		return err
	}
	p.metrics.Add("allocated", -1)
	return nil
}

// Remove from the pool a.k.a allocate these IPs.
func (p *metricPool) Remove(ips ...net.IP) error {
	err := p.IPPool.Remove(ips...)
	if err != nil {
		return err
	}
	total := expvar.Int{}
	total.Set(int64(len(ips)))
	// Assume this is the exact number, use Set instead of Add.
	p.metrics.Set("allocated", &total)
	return nil
}

func (p *metricPool) String() string {
	return p.metrics.String()
}
