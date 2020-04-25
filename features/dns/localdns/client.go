package localdns

import (
	"v2ray.com/core/common/net"
	"v2ray.com/core/features/dns"
  mdns "github.com/miekg/dns"
)

// Client is an implementation of dns.Client, which queries localhost for DNS.
type Client struct{}

// Type implements common.HasType.
func (*Client) Type() interface{} {
	return dns.ClientType()
}

// Start implements common.Runnable.
func (*Client) Start() error {
	//newDebugMsg("localdns server started")
	return nil
}

// Close implements common.Closable.
func (*Client) Close() error { return nil }

// TODO: we can use different DNS servers according to user's contury info, see https://public-dns.info/
var globalDNSServers = []string{
  "8.8.8.8",
  "8.8.4.4",
  "23.226.80.100",
}

// Lookup IP using global servers
func (*Client) GlobalLookupIP (host string) ([]net.IP) {
  ret := []net.IP{}
  for _, server := range globalDNSServers {
    //newDebugMsg("feature: resolving IP for " + host + ", using " + server)
    c := mdns.Client{}
    m := mdns.Msg{}
    m.SetQuestion(host + ".", mdns.TypeA)
    r, _, _ := c.Exchange(&m, server+":53")
    for _, ans := range r.Answer {
        Arecord := ans.(*mdns.A)
        ret = append(ret, Arecord.A)
    }
  }
  return ret
}

// LookupIP implements Client.
func (*Client) LookupIP(host string) ([]net.IP, error) {
	//newDebugMsg("feature: resolving IP for: " + host)
	ips, err := net.LookupIP(host)
	if err != nil {
		return nil, err
	}
	parsedIPs := make([]net.IP, 0, len(ips))
	for _, ip := range ips {
		parsed := net.IPAddress(ip)
		if parsed != nil {
			parsedIPs = append(parsedIPs, parsed.IP())
		}
	}
	if len(parsedIPs) == 0 {
		return nil, dns.ErrEmptyResponse
	}
	return parsedIPs, nil
}

// LookupIPv4 implements IPv4Lookup.
func (c *Client) LookupIPv4(host string) ([]net.IP, error) {
	newDebugMsg("feature: resolving IPv4 for: " + host)
	ips, err := c.LookupIP(host)
	if err != nil {
		return nil, err
	}
	ipv4 := make([]net.IP, 0, len(ips))
	for _, ip := range ips {
		if len(ip) == net.IPv4len {
			ipv4 = append(ipv4, ip)
		}
	}
	if len(ipv4) == 0 {
		return nil, dns.ErrEmptyResponse
	}
	return ipv4, nil
}

// LookupIPv6 implements IPv6Lookup.
func (c *Client) LookupIPv6(host string) ([]net.IP, error) {
	newDebugMsg("feature: resolving IPv6 for: " + host)
	ips, err := c.LookupIP(host)
	if err != nil {
		return nil, err
	}
	ipv6 := make([]net.IP, 0, len(ips))
	for _, ip := range ips {
		if len(ip) == net.IPv6len {
			ipv6 = append(ipv6, ip)
		}
	}
	if len(ipv6) == 0 {
		return nil, dns.ErrEmptyResponse
	}
	return ipv6, nil
}

// New create a new dns.Client that queries localhost for DNS.
func New() *Client {
	return &Client{}
}
