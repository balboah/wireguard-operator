package main

import (
	"flag"
	"net"
	"net/http"

	operator "github.com/balboah/wireguard-operator"
	log "github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
)

const version = "dev"

func main() {
	logLevel := flag.String("log-level", "info", "logging level")
	ifName := flag.String(
		"interface", "wg0", "which WireGuard interface to use")
	ifDelete := flag.Bool(
		"interface_delete", false, "delete specificed interface on exit if we created it")
	externalIP4 := flag.String(
		"external-ip4", "127.0.0.1", "endpoint reported to clients for incoming IPv4 WireGuard traffic")
	externalIP6 := flag.String(
		"external-ip6", "", "endpoint reported to clients for incoming IPv6 WireGuard traffic")
	wgPort := flag.Int("wireguard-port", 51820, "port for incoming WireGuard traffic")
	listenAddr := flag.String(
		"listen-addr", "0.0.0.0:8080", "listen address for API traffic")
	ip4Addr := flag.String(
		"ip4-addr", "10.143.0.0/16", "tunnel IPv4 network")
	ip6Addr := flag.String(
		"ip6-addr", "fd:b10c:ad:add1:de1e:7ed::/112", "tunnel IPv6 network")
	flag.Parse()

	if lvl, err := log.ParseLevel(*logLevel); err != nil {
		log.Fatal(err)
	} else {
		log.SetLevel(lvl)
	}

	// Use existing link or create a new one.
	wg, err := netlink.LinkByName(*ifName)
	if err != nil {
		switch err.(type) {
		case netlink.LinkNotFoundError:
			wg = &operator.WgLink{LinkAttrs: netlink.LinkAttrs{Name: *ifName}}
			if err := netlink.LinkAdd(wg); err != nil {
				log.Fatal(err)
			}
			// Delete the link on exit if we created it.
			if *ifDelete {
				defer func() { netlink.LinkDel(wg) }()
			}
		default:
			log.Fatal(err)
		}
	}

	addr4, err := netlink.ParseAddr(*ip4Addr)
	if err != nil {
		log.Fatal(err)
	}
	if err := netlink.AddrAdd(wg, addr4); err != nil {
		log.Fatal(err)
	}
	addr6, err := netlink.ParseAddr(*ip6Addr)
	if err != nil {
		log.Fatal(err)
	}
	if err := netlink.AddrAdd(wg, addr6); err != nil {
		log.Fatal(err)
	}

	c, err := operator.NewWgClient(*ifName, *wgPort)
	if err != nil {
		log.Fatal(err)
	}
	d, err := c.Device(*ifName)
	if err != nil {
		log.Fatal(err)
	}
	p, err := operator.NewPool(*ip4Addr)
	if err != nil {
		log.Fatal(err)
	}
	_, net6, err := net.ParseCIDR(*ip6Addr)
	if err != nil {
		log.Fatal(err)
	}
	ip4 := net.ParseIP(*externalIP4)
	ip6 := net.ParseIP(*externalIP6)
	if ip4 == nil && ip6 == nil {
		log.Fatal("need at least one external IP")
	}
	id := myID{
		externalIP4: ip4,
		externalIP6: ip6,
		port:        *wgPort,
		publicKey:   c.PublicKey(),
	}
	log.Infof("WireGuard Operator version %s", version)
	log.Infof("Public key: %s", d.PublicKey.String())
	http.HandleFunc("/v1/peer", operator.PeerHandler(c, id, p, net6))
	if err := http.ListenAndServe(*listenAddr, nil); err != nil {
		log.Fatal(err)
	}
}

type myID struct {
	externalIP4 net.IP
	externalIP6 net.IP
	publicKey   []byte
	port        int
}

func (id myID) Endpoint4() net.IP { return id.externalIP4 }

func (id myID) Endpoint6() net.IP { return nil }

func (id myID) PublicKey() []byte { return id.publicKey }

func (id myID) Port() int { return id.port }
