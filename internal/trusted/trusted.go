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

			err := CheckAccessToTrustedNetwork(realIP, trustedNetwork)
			if err != nil {
				http.Error(w, err.Error(), http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func CheckAccessToTrustedNetwork(IPStr string, trustedNetwork *net.IPNet) error {
	IP := net.ParseIP(IPStr)
	if IP == nil {
		return fmt.Errorf("invalid IP address %s", IP)
	}

	if !trustedNetwork.Contains(IP) {
		return fmt.Errorf("IP address %s is not in trusted subnet %s", IP, trustedNetwork.String())
	}
	return nil
}

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

func GetTrustedSubnet(trustedSubnetStr string) (*net.IPNet, error) {
	if trustedSubnetStr == "" {
		return nil, nil
	}

	_, ipNet, err := net.ParseCIDR(trustedSubnetStr)
	if err != nil {
		return nil, fmt.Errorf("invalid trusted subnet format: %w", err)
	}

	return ipNet, nil
}
