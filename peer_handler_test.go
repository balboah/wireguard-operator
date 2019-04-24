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
	"github.com/mdlayher/wireguardctrl/wgtypes"
)

func TestPeerHandler(t *testing.T) {
	key, err := wgtypes.GeneratePrivateKey()
	if err != nil {
		t.Fatal(err)
	}
	pool, err := NewPool("10.2.0.0/16")
	if err != nil {
		t.Fatal(err)
	}
	_, prefix, err := net.ParseCIDR("fd:b10c:ad:add1:de1e:7ed::/112")
	if err != nil {
		t.Fatal(err)
	}
	wg := dummy{}
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
		if len(res.VIP4) < 4 {
			t.Error("missing VIP4")
		}
		if len(res.VIP6) < 16 {
			t.Error("missing VIP6")
		}
		if len(wg) != 1 {
			t.Error("peer not configured")
		}
		if len(wg[key.PublicKey()]) != 2 {
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
		if len(wg) != 0 {
			t.Error("peer not deleted")
		}
	})

	t.Run("replace peers", func(t *testing.T) {
		peers := []proto.PeerReplacement{
			{
				PublicKey: pub("1y7aZNACS4ZDyNgQJN7/vtEUrj0lHWmIwJQO5VgrigM="),
				VIPs: []net.IP{
					net.ParseIP("10.2.0.1"),
					net.ParseIP("fd:b10c:ad:add1:de1e:7ed::1"),
				},
			},
			{
				PublicKey: pub("2y7aZNACS4ZDyNgQJN7/vtEUrj0lHWmIwJQO5VgrigM="),
				VIPs: []net.IP{
					net.ParseIP("10.2.0.2"),
					net.ParseIP("fd:b10c:ad:add1:de1e:7ed::2"),
				},
			},
		}
		js, err := json.Marshal(proto.PeerReplaceRequest{
			Peers: peers,
		})
		if err != nil {
			t.Fatal(err)
		}
		req := httptest.NewRequest("POST", "http://operator/v1/peer", bytes.NewBuffer(js))
		w := httptest.NewRecorder()
		h(w, req)
		if w.Code != http.StatusOK {
			t.Error(http.StatusText(w.Code))
		}
		if len(wg) != len(peers) {
			t.Error("peers not replaced")
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

type dummy map[wgtypes.Key][]net.IPNet

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
		d[p.PublicKey] = p.AllowedIPs
	}
	return nil
}

func (d dummy) ResolvePeerNets(key wgtypes.Key) ([]net.IPNet, error) {
	if d[key] == nil {
		return nil, ErrPeerNotFound
	}
	return d[key], nil
}

func (dummy) PublicKey() []byte {
	b, _ := base64.StdEncoding.DecodeString("IiGCPXY61aghq0n+9m2YOFCLcyqERD9qS9k6bxiks3g=")
	return b
}

func (dummy) Endpoint4() net.IP { return net.IP([]byte{1, 2, 3, 4}) }

func (dummy) Endpoint6() net.IP { return net.IP([]byte{1, 2, 3, 4, 1, 2, 3, 4, 1, 2, 3, 4, 1, 2, 3, 4}) }

func (dummy) Port() int { return 1234 }
