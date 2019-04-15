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

			ip4, err := p.Allocate()
			if err != nil {
				log.Error(err)
				rw.WriteHeader(http.StatusInternalServerError)
				return
			}
			// re-use allocated IPv4 with the IPv6 network prefix.
			ip6 := ip4To6(ip4, ip6prefix)

			pubKey, err := wgtypes.NewKey(js.PublicKey)
			if err != nil {
				log.Debug("PeerHandler.PUT:", err)
				rw.WriteHeader(http.StatusBadRequest)
				return
			}
			if err := putPeer(
				c, pubKey,
				net.IPNet{IP: ip4, Mask: net.CIDRMask(32, 32)},
				net.IPNet{IP: ip6, Mask: net.CIDRMask(128, 128)},
			); err != nil {
				log.Error("PeerHandler.PUT:", err)
				rw.WriteHeader(http.StatusBadRequest)
				return
			}

			if err := json.NewEncoder(rw).Encode(&proto.PeerResponse{
				PublicKey: wgID.PublicKey(),
				Endpoint4: wgID.Endpoint4(),
				Endpoint6: wgID.Endpoint6(),
				VIP4:      ip4,
				VIP6:      ip6,
				Port:      wgID.Port(),
			}); err != nil {
				log.Error("PeerHandler.PUT:", err)
				rw.WriteHeader(http.StatusInternalServerError)
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
// FIXME: need to make sure we don't append same pk multiple times?
func putPeer(c WgDeviceConfigurator, publicKey wgtypes.Key, ips ...net.IPNet) error {
	return c.ConfigureDevice(wgtypes.Config{
		Peers: []wgtypes.PeerConfig{
			{
				PublicKey:         publicKey,
				AllowedIPs:        ips,
				ReplaceAllowedIPs: false, // a.k.a. append peer
			},
		},
	})
}
