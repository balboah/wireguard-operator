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

	js, err := json.Marshal(proto.PeerRequest{
		PublicKey: keyToSlice(key.PublicKey()),
	})
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

	req := httptest.NewRequest("PUT", "http://operator/v1/peer", bytes.NewBuffer(js))
	w := httptest.NewRecorder()
	h := PeerHandler(dummy{}, dummy{}, pool, prefix)
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
	if len(res.Endpoint4) < 4 {
		t.Error("missing Endpoint4")
	}
	if len(res.Endpoint6) < 16 {
		t.Error("missing Endpoint6")
	}
	if res.Port == 0 {
		t.Error("missing port")
	}
	if res.PublicKey == nil {
		t.Error("missing PublicKey")
	}

	req = httptest.NewRequest("GET", "http://operator/v1/peer", nil)
	w = httptest.NewRecorder()
	h(w, req)
	if w.Code != http.StatusMethodNotAllowed {
		t.Log(w.Code)
		t.Error("unexpected response")
	}
}

type dummy struct{}

func (dummy) ConfigureDevice(cfg wgtypes.Config) error {
	return nil
}

func (dummy) PublicKey() []byte {
	b, _ := base64.StdEncoding.DecodeString("IiGCPXY61aghq0n+9m2YOFCLcyqERD9qS9k6bxiks3g=")
	return b
}

func (dummy) Endpoint4() net.IP { return net.IP([]byte{1, 2, 3, 4}) }

func (dummy) Endpoint6() net.IP { return net.IP([]byte{1, 2, 3, 4, 1, 2, 3, 4, 1, 2, 3, 4, 1, 2, 3, 4}) }

func (dummy) Port() int { return 1234 }
