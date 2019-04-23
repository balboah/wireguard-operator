package operator

import (
	"encoding/json"
	"net"
	"net/http"
	"strings"

	"github.com/balboah/wireguard-operator/proto"
	"github.com/mdlayher/wireguardctrl/wgtypes"
	log "github.com/sirupsen/logrus"
)

func PeerHandler(c WgDeviceConfigurator, wgID WgIdentity, p *Pool, ip6prefix *net.IPNet) http.HandlerFunc {
	// FIXME: do we need mutex for client?
	return func(rw http.ResponseWriter, req *http.Request) {
		defer req.Body.Close()

		switch req.Method {
		case "PUT":
			js := proto.PeerRequest{}
			if err := json.NewDecoder(req.Body).Decode(&js); err != nil {
				log.Debug("PeerHandler.PUT:", err)
				rw.WriteHeader(http.StatusBadRequest)
				return
			}

			ip4, ip6, err := putPeer(c, js.PublicKey, p, ip6prefix)
			if err != nil {
				log.Error("PeerHandler.PUT:", err)
				rw.WriteHeader(http.StatusBadRequest)
				return
			}

			if err := json.NewEncoder(rw).Encode(&proto.PeerResponse{
				VIP4: ip4,
				VIP6: ip6,
			}); err != nil {
				log.Error("PeerHandler.PUT:", err)
				rw.WriteHeader(http.StatusInternalServerError)
				return
			}
		case "DELETE":
			js := proto.PeerRequest{}
			if err := json.NewDecoder(req.Body).Decode(&js); err != nil {
				log.Debug("PeerHandler.DELETE:", err)
				rw.WriteHeader(http.StatusBadRequest)
				return
			}
			// if err := deletePeer(c, )
		default:
			log.Debugf("PeerHandler.%s: StatusMethodNotAllowed", strings.ToUpper(req.Method))
			rw.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
	}
}

// putPeer appends the public key with a list of allowed IP adresses.
// FIXME: need to make sure we don't append same pk multiple times?
func putPeer(c WgDeviceConfigurator, publicKey []byte, p *Pool, ip6prefix *net.IPNet) (ip4, ip6 net.IP, err error) {
	pk, err := wgtypes.NewKey(publicKey)
	if err != nil {
		return nil, nil, err
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
				ReplaceAllowedIPs: false, // a.k.a. append peer
			},
		},
	})

	return ip4, ip6, err
}
