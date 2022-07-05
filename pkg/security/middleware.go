package security

import (
	"fmt"
	"net"
	"net/http"
)

// RealIPRoundTripper sets X-Real-IP header for client
type RealIPRoundTripper struct {
	proxied http.RoundTripper
	localIP string
}

func NewRealIPRoundTripper(proxied http.RoundTripper, localIP string) *RealIPRoundTripper {
	return &RealIPRoundTripper{
		proxied: proxied,
		localIP: localIP,
	}
}

func (realIPtr *RealIPRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("X-Real-IP", realIPtr.localIP)

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
