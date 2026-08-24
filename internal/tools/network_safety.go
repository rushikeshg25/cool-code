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

func blockedNetworkIP(ip net.IP) bool {
	return ip == nil || ip.IsUnspecified() || ip.IsLoopback() || ip.IsPrivate() ||
		ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsMulticast()
}
