package middleware

import (
	"net/http"
	"net/netip"

	"github.com/The-Gleb/go_metrics_and_alerting/internal/logger"
)

type checkSubnetMiddleware struct {
	subNet *netip.Prefix
}

func NewCheckSubnetMiddleware(stringCIDR string) (*checkSubnetMiddleware, error) {
	if stringCIDR == "" {
		return &checkSubnetMiddleware{subNet: nil}, nil
	}
	IPNet, err := netip.ParsePrefix(stringCIDR)
	if err != nil {
		return nil, err
	}
	return &checkSubnetMiddleware{subNet: &IPNet}, nil
}

func (m *checkSubnetMiddleware) Do(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if m.subNet == nil {
			next.ServeHTTP(w, r)
			return
		}

		IP, err := netip.ParseAddr(r.Header.Get("X-Real-IP"))
		if err != nil {
			http.Error(w, "error parsing IP", http.StatusForbidden)
			return
		}

		if !m.subNet.Contains(IP) {
			http.Error(w, "IP is not in trusted subnet", http.StatusForbidden)
			return
		}

		logger.Log.Debugw("subnet contains IP", "subnet", m.subNet, "IP", IP)

		next.ServeHTTP(w, r)

	})
}
