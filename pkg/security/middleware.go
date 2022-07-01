package security

import (
	"fmt"
	"github.com/itd27m01/go-metrics-service/pkg/logging/log"
	"net"
	"net/http"
)

// RealIPRoundTripper sets X-Real-IP header for client
type RealIPRoundTripper struct {
	proxied http.RoundTripper
}

func NewRealIPRoundTripper(proxied http.RoundTripper) *RealIPRoundTripper {
	return &RealIPRoundTripper{
		proxied: proxied,
	}
}

func (realIPtr *RealIPRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	conn, err := net.Dial("tcp", net.JoinHostPort(req.URL.Hostname(), req.URL.Port()))
	if err != nil {
		return nil, fmt.Errorf("couldn't make tcp connection in RealIP middleware: %w", err)
	}
	defer func() {
		if err := conn.Close(); err != nil {
			log.Error().Msgf("Couldn't close tcp connection in RealIP middleware: %s", err)
		}
	}()
	localAddr := conn.LocalAddr().(*net.TCPAddr)

	req.Header.Set("X-Real-IP", localAddr.IP.String())

	return realIPtr.proxied.RoundTrip(req)
}

// CheckRealIP checks real client IP against trustedCIDR
func CheckRealIP(trustedCIDR string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if trustedCIDR == "" {
				next.ServeHTTP(w, r)

				return
			}

			_, trustedCIDRIPNet, err := net.ParseCIDR(trustedCIDR)
			if err != nil {
				http.Error(w, fmt.Sprintf("check trusted networks: %q", err), http.StatusInternalServerError)

				return
			}

			if !trustedCIDRIPNet.Contains(net.ParseIP(r.RemoteAddr)) {
				http.Error(w, fmt.Sprintf("access for IP forbidden: %s", r.RemoteAddr), http.StatusForbidden)

				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
