package proto

// IDResponse of the WireGuard external endpoint.
type IDResponse struct {
	PublicKey []byte `json:"public_key"`
	Port      int    `json:"port"`
}
