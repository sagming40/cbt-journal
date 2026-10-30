// Package clientip는 요청을 보낸 사람의 실제 IP를 정한다 (ARCHITECTURE 3.3절).
// 리버스 프록시 뒤에서는 모든 요청이 프록시 IP 하나로 보여서, 그대로 요청 제한에 쓰면 모든 사용자가 한 줄로 묶인다.
// 비유: 택배 기사(프록시)가 가져온 상자는 송장(X-Forwarded-For)을 보되, 우리가 아는 기사가 붙인 송장만 믿는다.
package clientip

import (
	"context"
	"net"
	"net/http"
	"net/netip"
	"strings"
)

type ctxKey struct{}

// Middleware는 실제 접속 IP를 구해 context에 넣는다.
func Middleware(trusted []netip.Prefix) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := Resolve(r, trusted)
			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), ctxKey{}, ip)))
		})
	}
}

// FromContext는 Middleware가 넣은 IP를 꺼낸다. 없으면 빈 주소(IsValid() == false).
func FromContext(ctx context.Context) netip.Addr {
	ip, _ := ctx.Value(ctxKey{}).(netip.Addr)
	return ip
}

// Resolve는 RemoteAddr가 믿는 프록시일 때만 X-Forwarded-For의 **오른쪽 끝 값**을 쓴다.
// 오른쪽 끝은 그 프록시가 직접 붙인 값이고, 그보다 왼쪽은 클라이언트가 마음대로 넣을 수 있어 믿지 않는다.
func Resolve(r *http.Request, trusted []netip.Prefix) netip.Addr {
	remote := parseRemote(r.RemoteAddr)
	if !remote.IsValid() || !contains(trusted, remote) {
		return remote
	}
	xff := r.Header.Values("X-Forwarded-For")
	if len(xff) == 0 {
		return remote
	}
	// 헤더가 여러 줄이면 마지막 줄의 마지막 값이 가장 가까운 프록시가 붙인 것이다.
	last := xff[len(xff)-1]
	if i := strings.LastIndexByte(last, ','); i >= 0 {
		last = last[i+1:]
	}
	ip, err := netip.ParseAddr(strings.TrimSpace(last))
	if err != nil {
		return remote // 프록시 설정이 이상하면 프록시 IP로라도 센다
	}
	return ip.Unmap()
}

func parseRemote(remoteAddr string) netip.Addr {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		host = remoteAddr
	}
	ip, err := netip.ParseAddr(host)
	if err != nil {
		return netip.Addr{}
	}
	// IPv4가 IPv6 형식(::ffff:1.2.3.4)으로 들어와도 같은 주소로 센다.
	return ip.Unmap()
}

func contains(prefixes []netip.Prefix, ip netip.Addr) bool {
	for _, p := range prefixes {
		if p.Contains(ip) {
			return true
		}
	}
	return false
}
