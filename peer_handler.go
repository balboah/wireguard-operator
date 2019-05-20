package operator

import (
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"strings"

	"github.com/balboah/wireguard-operator/proto"
	log "github.com/sirupsen/logrus"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

func PeerHandler(c WgDeviceConfigurator, wgID WgIdentity, p IPPool, ip6prefix *net.IPNet) http.HandlerFunc {
	// FIXME: do we need mutex for client?
	return func(rw http.ResponseWriter, req *http.Request) {
		defer req.Body.Close()

		switch req.Method {
		case "PUT":
			js := proto.PeerRequest{}
			if err := json.NewDecoder(req.Body).Decode(&js); err != nil {
				log.Debug("PeerHandler.PUT: ", err)
				rw.WriteHeader(http.StatusBadRequest)
				return
			}

			ip4, ip6, err := putPeer(c, js.PublicKey, p, ip6prefix)
			if err != nil {
				log.Error("PeerHandler.PUT: ", err)
				rw.WriteHeader(http.StatusBadRequest)
				return
			}

			if err := json.NewEncoder(rw).Encode(&proto.PeerResponse{
				VIP4: ip4,
				VIP6: ip6,
			}); err != nil {
				log.Error("PeerHandler.PUT: ", err)
				rw.WriteHeader(http.StatusInternalServerError)
				return
			}
		case "POST":
			js := proto.PeerReplaceRequest{}
			if err := json.NewDecoder(req.Body).Decode(&js); err != nil {
				log.Debug("PeerHandler.POST: ", err)
				rw.WriteHeader(http.StatusBadRequest)
				return
			}

			if err := replacePeers(c, js.Peers, p, ip6prefix); err != nil {
				log.Error("PeerHandler.POST: ", err)
				rw.WriteHeader(http.StatusBadRequest)
				return
			}
		case "DELETE":
			js := proto.PeerRequest{}
			if err := json.NewDecoder(req.Body).Decode(&js); err != nil {
				log.Debug("PeerHandler.DELETE: ", err)
				rw.WriteHeader(http.StatusBadRequest)
				return
			}
			if err := deletePeer(c, js.PublicKey, p); err != nil {
				switch err {
				case ErrPeerNotFound:
					log.Debug("PeerHandler.DELETE: ", err)
				default:
					log.Error("PeerHandler.DELETE: ", err)
					rw.WriteHeader(http.StatusInternalServerError)
				}
				return
			}
		default:
			log.Debugf("PeerHandler.%s: StatusMethodNotAllowed", strings.ToUpper(req.Method))
			rw.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
	}
}

// putPeer appends the public key with a list of allowed IP adresses.
func putPeer(c WgDeviceConfigurator, publicKey []byte, p IPPool, ip6prefix *net.IPNet) (ip4, ip6 net.IP, err error) {
	pk, err := wgtypes.NewKey(publicKey)
	if err != nil {
		return nil, nil, err
	}

	// Free any IPs that this peer already has.
	if err := freeAll(c, pk, p); err != nil {
		// Not found error is expected.
		if err != ErrPeerNotFound {
			return nil, nil, err
		}
	}

	if ip4, err = p.Allocate(); err != nil {
		return nil, nil, err
	}
	// re-use allocated IPv4 with the IPv6 network prefix.
	ip6 = ip4To6(ip4, ip6prefix)

	err = c.ConfigureDevice(wgtypes.Config{
		Peers: []wgtypes.PeerConfig{
			{
				PublicKey: pk,
				AllowedIPs: []net.IPNet{
					net.IPNet{IP: ip4, Mask: net.CIDRMask(32, 32)},
					net.IPNet{IP: ip6, Mask: net.CIDRMask(128, 128)},
				},
				ReplaceAllowedIPs: true,
			},
		},
	})

	return ip4, ip6, err
}

func replacePeers(c WgDeviceConfigurator, replacements []proto.PeerReplacement, p IPPool, ip6prefix *net.IPNet) error {
	wgPeers := make([]wgtypes.PeerConfig, len(replacements))
	allocatedIPs := make([]net.IP, len(replacements))
	for n, r := range replacements {
		ip4 := net.IP{}
		ip6 := net.IP{}
		if len(r.VIPs) != 2 {
			return errors.New("invalid replacement: IPs length")
		}
		for _, ip := range r.VIPs {
			if ip.To4() != nil {
				ip4 = ip
				continue
			}
			ip6 = ip
		}
		if ip4.To4() == nil {
			return errors.New("invalid replacement: missing IPv4")
		}
		if ip6.To16() == nil {
			return errors.New("invalid replacement: missing IPv6")
		}
		allocatedIPs[n] = ip4
		pk, err := wgtypes.NewKey(r.PublicKey)
		if err != nil {
			return err
		}
		wgPeers[n] = wgtypes.PeerConfig{
			PublicKey: pk,
			AllowedIPs: []net.IPNet{
				{IP: ip4, Mask: net.CIDRMask(32, 32)},
				{IP: ip6, Mask: net.CIDRMask(128, 128)},
			},
			ReplaceAllowedIPs: true,
		}
	}
	if err := p.Remove(allocatedIPs...); err != nil {
		return err
	}

	return c.ConfigureDevice(wgtypes.Config{
		Peers:        wgPeers,
		ReplacePeers: true,
	})
}

// deletePeer frees all IPs in the pool and then removes the peer from wg.
func deletePeer(c WgDeviceConfigurator, publicKey []byte, p IPPool) error {
	pk, err := wgtypes.NewKey(publicKey)
	if err != nil {
		return err
	}
	if err := freeAll(c, pk, p); err != nil {
		// Log but don't return, we can still delete the peer but we got a
		// memory leak in our IP pool if this fails.
		log.Error("deletePeer: ", err)
	}
	err = c.ConfigureDevice(wgtypes.Config{
		Peers: []wgtypes.PeerConfig{
			{
				PublicKey: pk,
				Remove:    true,
			},
		},
	})

	return freeAll(c, pk, p)
}

// freeAll loops all IP networks for a peer and frees IPv4 addresses from the pool.
func freeAll(c WgDeviceConfigurator, publicKey wgtypes.Key, p IPPool, n ...net.IPNet) error {
	nets, err := c.ResolvePeerNets(publicKey)
	if err != nil {
		return err
	}
	for _, n := range nets {
		if n.IP.To4() != nil {
			// ip6 is derived from ip4 thus don't need to be free'd from pool.
			if err := p.Free(n.IP.To4()); err != nil {
				return err
			}
		}
	}
	return nil
}
