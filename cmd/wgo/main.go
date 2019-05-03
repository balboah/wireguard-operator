package main

import (
	"encoding/base64"
	"flag"
	"io/ioutil"
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
		"interface-delete", false, "delete specificed interface on exit")
	wgPort := flag.Int("wireguard-port", 51820, "port for incoming WireGuard traffic")
	wgKey := flag.String(
		"wireguard-private-key-file", "", "file with base64 encoded private key")
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
	wg, err := operator.NewWgLink(*ifName)
	if err != nil {
		log.Fatal("main.NewWgLink: ", err)
	}
	// Delete the link on exit.
	if *ifDelete {
		defer func() { wg.Close() }()
	}

	addr4, err := netlink.ParseAddr(*ip4Addr)
	if err != nil {
		log.Fatal("main.ParseAddr4: ", err)
	}
	addr6, err := netlink.ParseAddr(*ip6Addr)
	if err != nil {
		log.Fatal("main.ParseAddr6: ", err)
	}
	if err := wg.AddrAdd(addr4, addr6); err != nil {
		log.Fatal("main.AddrAdd: ", err)
	}

	key, err := ioutil.ReadFile(*wgKey)
	if err != nil {
		log.Fatal("main.ReadFile: ", err)
	}
	c, err := operator.NewWgClient(wg, *wgPort, string(key))
	if err != nil {
		log.Fatal("main.NewWgClient: ", err)
	}

	p, err := operator.NewPool(*ip4Addr)
	if err != nil {
		log.Fatal("main.NewPool: ", err)
	}
	_, net6, err := net.ParseCIDR(*ip6Addr)
	if err != nil {
		log.Fatal("main.ParseCIDR6: ", err)
	}

	id := myID{
		port:      *wgPort,
		publicKey: c.PublicKey(),
	}
	log.Infof("WireGuard Operator version %s", version)
	log.Infof("Public key: %s", base64.StdEncoding.EncodeToString(c.PublicKey()))

	http.HandleFunc("/v1/peer", operator.PeerHandler(c, id, p, net6))
	http.HandleFunc("/v1/id", operator.IDHandler(id))
	if err := http.ListenAndServe(*listenAddr, nil); err != nil {
		log.Fatal(err)
	}
}

type myID struct {
	publicKey []byte
	port      int
}

func (id myID) PublicKey() []byte { return id.publicKey }

func (id myID) Port() int { return id.port }
