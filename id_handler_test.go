package operator

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/balboah/wireguard-operator/proto"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

func TestIDHandler(t *testing.T) {
	key, err := wgtypes.GeneratePrivateKey()
	if err != nil {
		t.Fatal(err)
	}
	pubkey := keyToSlice(key.PublicKey())
	id := myID{
		publicKey: pubkey,
		port:      1337,
	}

	req := httptest.NewRequest("GET", "http://operator/v1/id", nil)
	w := httptest.NewRecorder()
	started := time.Now()
	h := IDHandler(id)
	h(w, req)
	if w.Code != http.StatusOK {
		t.Error(http.StatusText(w.Code))
	}

	res := proto.IDResponse{}
	t.Log(string(w.Body.Bytes()))
	if err := json.NewDecoder(w.Body).Decode(&res); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(res.PublicKey, pubkey) {
		t.Error("invalid public key")
	}
	if res.Port != 1337 {
		t.Error("invalid port")
	}
	if res.Started.Before(started) {
		t.Error("invalid started timestamp")
	}
}

type myID struct {
	publicKey []byte
	port      int
}

func (id myID) PublicKey() []byte { return id.publicKey }

func (id myID) Port() int { return id.port }
