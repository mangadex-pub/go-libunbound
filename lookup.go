package unbound

import (
	"github.com/miekg/dns"
	"net"
)

// These are function are a re-implementation of the net.Lookup* ones
// They are adapted to the package unbound and the package dns.

// LookupAddr performs a reverse lookup for the given address, returning a
// list of names mapping to that address. It is up to the caller to prime
// Unbound with trust anchor(s).
func (u *Unbound) LookupAddr(addr string) (name []string, err error) {
	reverse, err := dns.ReverseAddr(addr)
	if err != nil {
		return nil, err
	}
	r, err := u.Resolve(reverse, dns.TypePTR, dns.ClassINET)
	if err != nil {
		return nil, err
	}
	for _, rr := range r.Rr {
		name = append(name, rr.(*dns.RR_PTR).Ptr)
	}
	return
}

// LookupCNAME returns the canonical DNS host for the given name. Callers
// that do not care about the canonical name can call LookupHost or
// LookupIP directly; both take care of resolving the canonical name as
// part of the lookup. It is up to the caller to prime
// Unbound with trust anchor(s).
func (u *Unbound) LookupCNAME(name string) (cname string, err error) {
	return "", nil
}

// LookupHost looks up the given host using the local resolver. It returns
// an array of that host's addresses. It is up to the caller to prime
// Unbound with trust anchor(s).
func (u *Unbound) LookupHost(host string) (addrs []string, err error) {
	ipaddrs, err := u.LookupIP(host)
	if err != nil {
		return nil, err
	}
	for _, ip := range ipaddrs {
		addrs = append(addrs, ip.String())
	}
	return addrs, nil
}

// LookupIP looks up host using the local resolver. It returns an array of
// that host's IPv4 and IPv6 addresses. It is up to the caller to prime
// Unbound with trust anchor(s).
func (u *Unbound) LookupIP(host string) (addrs []net.IP, err error) {
	ca := make(chan net.IP)
	caaaa := make(chan net.IP)

	u.ResolveAsync(host, dns.TypeA, dns.ClassINET, ca, lookupA)
	u.ResolveAsync(host, dns.TypeAAAA, dns.ClassINET, caaaa, lookupAAAA)
	for ip := range ca {
		addrs = append(addrs, ip)
	}
	for ip := range caaaa {
		addrs = append(addrs, ip)
	}
	return
}

func lookupA(i interface{}, e error, r *Result) {
	c := i.(chan net.IP)
	defer close(c)
	if e != nil {
		return
	}
	for _, rr := range r.Rr {
		c <- rr.(*dns.RR_A).A
	}
}

func lookupAAAA(i interface{}, e error, r *Result) {
	c := i.(chan net.IP)
	defer close(c)
	if e != nil {
		return
	}
	for _, rr := range r.Rr {
		c <- rr.(*dns.RR_AAAA).AAAA
	}
}

// LookupMX returns the DNS MX records for the given domain name sorted by
// preference. It is up to the caller to prime Unbound with trust anchor(s).
func (u *Unbound) LookupMX(name string) (mx []*dns.RR_MX, err error) {
	r, err := u.Resolve(name, dns.TypeMX, dns.ClassINET)
	if err != nil {
		return nil, err
	}
	for _, rr := range r.Rr {
		mx = append(mx, rr.(*dns.RR_MX))
	}
	return
}

// LookupSRV tries to resolve an SRV query of the given service, protocol,
// and domain name. The proto is "tcp" or "udp". The returned records are
// sorted by priority and randomized by weight within a priority.
// 
// LookupSRV constructs the DNS name to look up following RFC 2782. That
// is, it looks up _service._proto.name. To accommodate services publishing
// SRV records under non-standard names, if both service and proto are
// empty strings, LookupSRV looks up name directly. It is up to the caller to prime
// Unbound with trust anchor(s).
func (u *Unbound) LookupSRV(service, proto, name string) (cname string, addrs []*dns.RR_SRV, err error) {
	return "", nil, nil
}

// LookupTXT returns the DNS TXT records for the given domain name. It is up to the caller to prime
// Unbound with trust anchor(s).
func (u *Unbound) LookupTXT(name string) (txt []string, err error) {
	r, err := u.Resolve(name, dns.TypeTXT, dns.ClassINET)
	if err != nil {
		return nil, err
	}
	for _, rr := range r.Rr {
		txt = append(txt, rr.(*dns.RR_TXT).Txt...)
	}
	return
}
