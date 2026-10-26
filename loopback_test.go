package operator

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLoopbackWritesOnly(t *testing.T) {
	called := false
	h := LoopbackWritesOnly(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		called = true
	}))

	for _, tc := range []struct {
		method, remote string
		want           int
		reached        bool
	}{
		{"PUT", "127.0.0.1:4321", http.StatusOK, true},
		{"POST", "[::1]:4321", http.StatusOK, true},
		{"DELETE", "127.0.0.5:4321", http.StatusOK, true},
		{"PUT", "[::ffff:127.0.0.1]:4321", http.StatusOK, true},
		{"PUT", "10.143.0.1:4321", http.StatusForbidden, false},
		{"POST", "[2001:db8::1]:4321", http.StatusForbidden, false},
		{"DELETE", "192.0.2.1:4321", http.StatusForbidden, false},
		{"PUT", "garbage", http.StatusForbidden, false},
		{"PATCH", "192.0.2.1:4321", http.StatusForbidden, false},
		{"GET", "192.0.2.1:4321", http.StatusOK, true},
		{"HEAD", "192.0.2.1:4321", http.StatusOK, true},
		{"OPTIONS", "192.0.2.1:4321", http.StatusOK, true},
	} {
		called = false
		req := httptest.NewRequest(tc.method, "http://operator/v1/peer", nil)
		req.RemoteAddr = tc.remote
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)
		if w.Code != tc.want {
			t.Errorf("%s from %s: got %d, want %d", tc.method, tc.remote, w.Code, tc.want)
		}
		if called != tc.reached {
			t.Errorf("%s from %s: handler reached %v, want %v", tc.method, tc.remote, called, tc.reached)
		}
	}
}
