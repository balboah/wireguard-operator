package operator

import (
	"net"
	"net/http"

	log "github.com/sirupsen/logrus"
)

// LoopbackWritesOnly answers 403 to any non-read request whose peer address
// is not loopback, so writes are limited to callers on the same host.
func LoopbackWritesOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		switch req.Method {
		case http.MethodGet, http.MethodHead, http.MethodOptions:
		default:
			if !isLoopback(req.RemoteAddr) {
				log.Debugf("%s %s from %s: StatusForbidden", req.Method, req.URL.Path, req.RemoteAddr)
				rw.WriteHeader(http.StatusForbidden)
				return
			}
		}
		next.ServeHTTP(rw, req)
	})
}

func isLoopback(remoteAddr string) bool {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		return false
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}
