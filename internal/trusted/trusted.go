package trusted

import (
	"fmt"
	"net"
	"net/http"
)

func TrustedSubnetMiddleware(trustedNetwork *net.IPNet) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if trustedNetwork == nil {
				next.ServeHTTP(w, r)
				return
			}

			realIP := r.Header.Get("X-Real-IP")
			if realIP == "" {
				http.Error(w, "X-Real-IP header is required", http.StatusForbidden)
				return
			}

			ip := net.ParseIP(realIP)
			if ip == nil {
				http.Error(w, "Invalid IP address in X-Real-IP header", http.StatusForbidden)
				return
			}

			if !trustedNetwork.Contains(ip) {
				http.Error(w, "IP address is not in trusted subnet", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// getLocalIP получает локальный IP-адрес хоста
func GetLocalIP() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String(), nil
			}
		}
	}
	return "", fmt.Errorf("no local IP address found")
}
