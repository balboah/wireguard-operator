package proto

import (
	"encoding/json"
	"net"
)

type PeerRequest struct {
	PublicKey []byte `json:"public_key"`
}

type PeerResponse struct {
	VIP4 net.IP `json:"vip_ipv4"`
	VIP6 net.IP `json:"vip_ipv6"`
}

func (r *PeerResponse) String() string {
	b, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(b)
}
