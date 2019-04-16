package operator

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/balboah/wireguard-operator/proto"
	log "github.com/sirupsen/logrus"
)

func IDHandler(wgID WgIdentity) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		defer req.Body.Close()

		switch req.Method {
		case "GET":
			if err := json.NewEncoder(rw).Encode(&proto.IDResponse{
				PublicKey: wgID.PublicKey(),
				Port:      wgID.Port(),
			}); err != nil {
				log.Error("IDHandler.GET:", err)
				rw.WriteHeader(http.StatusInternalServerError)
				return
			}
		default:
			log.Debugf("IDHandler.%s: StatusMethodNotAllowed", strings.ToUpper(req.Method))
			rw.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
	}
}
