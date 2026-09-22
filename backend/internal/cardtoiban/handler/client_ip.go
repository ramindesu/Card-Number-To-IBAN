package handler

import (
	"net"
	"net/netip"
)

func parseClientIP(remoteAddress string) *netip.Addr {
	host, _, err := net.SplitHostPort(remoteAddress)
	if err != nil {
		host = remoteAddress
	}

	address, err := netip.ParseAddr(host)
	if err != nil {
		return nil
	}

	address = address.Unmap()
	return &address
}
