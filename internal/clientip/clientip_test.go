package clientip

import (
	"context"
	"net/http/httptest"
	"net/netip"
	"testing"
)

func TestResolve(t *testing.T) {
	trusted := []netip.Prefix{netip.MustParsePrefix("10.0.0.0/8")}

	tests := []struct {
		name   string
		remote string
		xff    []string
		want   string
	}{
		{"프록시 없음", "203.0.113.7:5000", nil, "203.0.113.7"},
		{"모르는 곳의 XFF는 무시", "203.0.113.7:5000", []string{"1.2.3.4"}, "203.0.113.7"},
		{"믿는 프록시면 오른쪽 끝 값", "10.0.0.2:443", []string{"6.6.6.6, 198.51.100.9"}, "198.51.100.9"},
		{"헤더 여러 줄이면 마지막 줄의 끝", "10.0.0.2:443", []string{"6.6.6.6", "198.51.100.9"}, "198.51.100.9"},
		{"믿는 프록시인데 XFF 없음", "10.0.0.2:443", nil, "10.0.0.2"},
		{"XFF가 깨졌으면 프록시 IP", "10.0.0.2:443", []string{"쓰레기"}, "10.0.0.2"},
		{"IPv4-mapped IPv6", "[::ffff:203.0.113.7]:5000", nil, "203.0.113.7"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequestWithContext(context.Background(), "GET", "/", nil)
			r.RemoteAddr = tt.remote
			for _, v := range tt.xff {
				r.Header.Add("X-Forwarded-For", v)
			}
			if got := Resolve(r, trusted); got.String() != tt.want {
				t.Errorf("got %s, 기대 %s", got, tt.want)
			}
		})
	}
}
