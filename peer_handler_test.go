package operator

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/balboah/wireguard-operator/proto"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

func TestPeerHandler(t *testing.T) {
	wg := dummy{}
	testPeerHandler(t, wg)
}

func testPeerHandler(t *testing.T, wg WgDeviceConfigurator) {
	key, err := wgtypes.GeneratePrivateKey()
	if err != nil {
		t.Fatal(err)
	}
	pool, err := NewPool("10.2.0.0/16")
	if err != nil {
		t.Fatal(err)
	}
	_, prefix, err := net.ParseCIDR("fdad:b10c:a::/48")
	if err != nil {
		t.Fatal(err)
	}
	h := PeerHandler(wg, dummy{}, pool, prefix)

	t.Run("add peer", func(t *testing.T) {
		pr := proto.PeerRequest{
			PublicKey: keyToSlice(key.PublicKey()),
		}
		js, err := json.Marshal(&pr)
		if err != nil {
			t.Fatal(err)
		}
		req := httptest.NewRequest("PUT", "http://operator/v1/peer", bytes.NewBuffer(js))
		w := httptest.NewRecorder()
		h(w, req)
		if w.Code != http.StatusOK {
			t.Error(http.StatusText(w.Code))
		}

		res := proto.PeerResponse{}
		t.Log(string(w.Body.Bytes()))
		if err := json.NewDecoder(w.Body).Decode(&res); err != nil {
			t.Fatal(err)
		}
		if !res.VIP4.Equal(net.ParseIP("10.2.0.1")) {
			t.Error("invalid VIP4")
		}
		if !res.VIP6.Equal(net.ParseIP("fdad:b10c:a::a02:1")) {
			t.Error("invalid VIP6")
		}
		nets, err := wg.ResolvePeerNets(key.PublicKey())
		if err != nil {
			t.Log(wg)
			t.Fatal(err)
		}
		if len(nets) != 2 {
			t.Error("expected 1 vip4 & 1 vip6")
		}
		if l := len(pool.available); l != 65534-1 {
			t.Error("invalid available in pool, ", l)
		}
	})

	t.Run("replace peers", func(t *testing.T) {
		replaceReq := proto.PeerReplaceRequest{
			Peers: []proto.PeerReplacement{
				{
					PublicKey: pub("1y7aZNACS4ZDyNgQJN7/vtEUrj0lHWmIwJQO5VgrigM="),
					VIPs: []net.IP{
						net.ParseIP("10.2.0.2"),
						net.ParseIP("fdad:b10c:a::1"),
					},
				},
				{
					PublicKey: pub("2y7aZNACS4ZDyNgQJN7/vtEUrj0lHWmIwJQO5VgrigM="),
					VIPs: []net.IP{
						net.ParseIP("10.2.0.3"),
						net.ParseIP("fdad:b10c:a::2"),
					},
				},
			},
		}
		js, err := json.Marshal(replaceReq)
		if err != nil {
			t.Fatal(err)
		}
		req := httptest.NewRequest("POST", "http://operator/v1/peer", bytes.NewBuffer(js))
		w := httptest.NewRecorder()
		h(w, req)
		if w.Code != http.StatusOK {
			t.Error(http.StatusText(w.Code))
		}
		peers, err := wg.Peers()
		if err != nil {
			t.Fatal(err)
		}
		if len(peers) != 2 {
			t.Error("peers not replaced")
		}
		if l := len(pool.available); l != 65534-2 {
			t.Error("invalid available in pool, ", l)
		}
	})

	t.Run("add replaced peer again", func(t *testing.T) {
		pr := proto.PeerRequest{
			PublicKey: keyToSlice(key.PublicKey()),
		}
		js, err := json.Marshal(&pr)
		if err != nil {
			t.Fatal(err)
		}
		req := httptest.NewRequest("PUT", "http://operator/v1/peer", bytes.NewBuffer(js))
		w := httptest.NewRecorder()
		h(w, req)
		if w.Code != http.StatusOK {
			t.Error(http.StatusText(w.Code))
		}

		res := proto.PeerResponse{}
		t.Log(string(w.Body.Bytes()))
		if err := json.NewDecoder(w.Body).Decode(&res); err != nil {
			t.Fatal(err)
		}
		if !res.VIP4.Equal(net.ParseIP("10.2.0.1")) {
			t.Log(res.VIP4)
			t.Error("expected re-used VIP4")
		}
		if !res.VIP6.Equal(net.ParseIP("fdad:b10c:a::a02:1")) {
			t.Log(res.VIP6)
			t.Error("expected re-used VIP6")
		}
		nets, err := wg.ResolvePeerNets(key.PublicKey())
		if err != nil {
			t.Log(wg)
			t.Fatal(err)
		}
		if len(nets) != 2 {
			t.Error("expected 1 vip4 & 1 vip6")
		}
	})

	t.Run("delete peer", func(t *testing.T) {
		js, err := json.Marshal(proto.PeerRequest{
			PublicKey: keyToSlice(key.PublicKey()),
		})
		if err != nil {
			t.Fatal(err)
		}
		req := httptest.NewRequest("DELETE", "http://operator/v1/peer", bytes.NewBuffer(js))
		w := httptest.NewRecorder()
		h(w, req)
		if w.Code != http.StatusOK {
			t.Error(http.StatusText(w.Code))
		}

		peers, err := wg.Peers()
		if err != nil {
			t.Fatal(err)
		}
		if n := len(peers); n != 2 {
			t.Log(n)
			t.Error("peer not deleted")
		}
	})

	t.Run("error on unsupported method", func(t *testing.T) {
		req := httptest.NewRequest("GET", "http://operator/v1/peer", nil)
		w := httptest.NewRecorder()
		h(w, req)
		if w.Code != http.StatusMethodNotAllowed {
			t.Log(w.Code)
			t.Error("unexpected response")
		}
	})
}

func pub(b64 string) []byte {
	b, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		panic(err)
	}
	return b
}

type dummy map[wgtypes.Key]wgtypes.Peer

func (d dummy) ConfigureDevice(cfg wgtypes.Config) error {
	if cfg.ReplacePeers {
		for k := range d {
			delete(d, k)
		}
	}
	for _, p := range cfg.Peers {
		if p.Remove {
			delete(d, p.PublicKey)
			continue
		}
		d[p.PublicKey] = wgtypes.Peer{
			PublicKey:  p.PublicKey,
			AllowedIPs: p.AllowedIPs,
		}
	}
	return nil
}

func (d dummy) Peers() ([]wgtypes.Peer, error) {
	peers := []wgtypes.Peer{}
	for _, p := range d {
		peers = append(peers, p)
	}
	return peers, nil
}

func (d dummy) ResolvePeerNets(key wgtypes.Key) ([]net.IPNet, error) {
	p, ok := d[key]
	if !ok {
		return nil, ErrPeerNotFound
	}

	return p.AllowedIPs, nil
}

func (dummy) PublicKey() []byte {
	b, _ := base64.StdEncoding.DecodeString("IiGCPXY61aghq0n+9m2YOFCLcyqERD9qS9k6bxiks3g=")
	return b
}

func (dummy) Endpoint4() net.IP { return net.IP([]byte{1, 2, 3, 4}) }

func (dummy) Endpoint6() net.IP { return net.IP([]byte{1, 2, 3, 4, 1, 2, 3, 4, 1, 2, 3, 4, 1, 2, 3, 4}) }

func (dummy) Port() int { return 1234 }
