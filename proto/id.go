package proto

import "time"

// IDResponse of the WireGuard external endpoint.
type IDResponse struct {
	Started   time.Time `json:"started"`
	PublicKey []byte    `json:"public_key"`
	Port      int       `json:"port"`
}
