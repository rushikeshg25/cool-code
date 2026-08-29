package tools

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/url"
	"time"
)

var safeWebClient = &http.Client{
	Timeout: webTimeout,
	Transport: &http.Transport{
		Proxy:       nil,
		DialContext: safeDialContext,
	},
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		if len(via) >= 5 {
			return errors.New("too many redirects")
		}
		return validateWebURL(req.Context(), req.URL)
	},
}

func validateWebURL(ctx context.Context, u *url.URL) error {
	if u == nil || u.Scheme != "https" || u.Host == "" {
		return errors.New("only absolute HTTPS URLs are allowed")
	}
	if u.User != nil {
		return errors.New("URLs containing user credentials are blocked")
	}
	ips, err := net.DefaultResolver.LookupIPAddr(ctx, u.Hostname())
	if err != nil || len(ips) == 0 {
		return errors.New("could not resolve URL host")
	}
	for _, ip := range ips {
		if blockedNetworkIP(ip.IP) {
			return errors.New("private, loopback, link-local, and metadata addresses are blocked")
		}
	}
	return nil
}

func safeDialContext(ctx context.Context, network, address string) (net.Conn, error) {
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return nil, errors.New("invalid network address")
	}
	ips, err := net.DefaultResolver.LookupIPAddr(ctx, host)
	if err != nil || len(ips) == 0 {
		return nil, errors.New("could not resolve network host")
	}
	dialer := net.Dialer{Timeout: 10 * time.Second}
	for _, resolved := range ips {
		if blockedNetworkIP(resolved.IP) {
			continue
		}
		conn, dialErr := dialer.DialContext(ctx, network, net.JoinHostPort(resolved.IP.String(), port))
		if dialErr == nil {
			return conn, nil
		}
		err = dialErr
	}
	if err != nil {
		return nil, err
	}
	return nil, errors.New("network target is blocked")
}

// reservedRanges covers allocations that net.IP.IsPrivate does not. IsPrivate
// is RFC1918 and RFC4193 only, which leaves the carrier-grade NAT range used by
// Tailscale, including its 100.100.100.100 resolver, reachable.
var reservedRanges = func() []*net.IPNet {
	cidrs := []string{
		"100.64.0.0/10", // RFC 6598 carrier-grade NAT
		"192.0.0.0/24",  // IETF protocol assignments
		"198.18.0.0/15", // benchmarking
		"240.0.0.0/4",   // reserved for future use
		"255.255.255.255/32",
		"64:ff9b::/96",   // NAT64, which embeds an IPv4 address
		"64:ff9b:1::/48", // local-use NAT64
	}
	out := make([]*net.IPNet, 0, len(cidrs))
	for _, c := range cidrs {
		if _, n, err := net.ParseCIDR(c); err == nil {
			out = append(out, n)
		}
	}
	return out
}()

func blockedNetworkIP(ip net.IP) bool {
	if ip == nil || ip.IsUnspecified() || ip.IsLoopback() || ip.IsPrivate() ||
		ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsMulticast() {
		return true
	}
	for _, n := range reservedRanges {
		if n.Contains(ip) {
			return true
		}
	}
	// A NAT64 address embeds an IPv4 address in its low 32 bits, so check the
	// address it would actually reach.
	if len(ip) == net.IPv6len {
		if embedded := net.IPv4(ip[12], ip[13], ip[14], ip[15]); embedded != nil {
			for _, n := range reservedRanges[:5] {
				if n.Contains(embedded) {
					return true
				}
			}
			if embedded.IsPrivate() || embedded.IsLoopback() || embedded.IsLinkLocalUnicast() {
				return isNAT64(ip)
			}
		}
	}
	return false
}

func isNAT64(ip net.IP) bool {
	for _, c := range []string{"64:ff9b::/96", "64:ff9b:1::/48"} {
		if _, n, err := net.ParseCIDR(c); err == nil && n.Contains(ip) {
			return true
		}
	}
	return false
}
