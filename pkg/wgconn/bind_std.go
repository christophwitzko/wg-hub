/* SPDX-License-Identifier: MIT
 *
 * Copyright (C) 2017-2022 WireGuard LLC. All Rights Reserved.
 * Forked from: https://github.com/WireGuard/wireguard-go/blob/bb719d3a6e2cd20ec00f26d65c0073c1dde6b529/conn/bind_std.go
 */

package wgconn

import (
	"errors"
	"fmt"
	"net"
	"net/netip"
	"sync"
	"syscall"

	"golang.zx2c4.com/wireguard/conn"
)

// StdNetBind is meant to be a temporary solution on platforms for which
// the sticky socket / source caching behavior has not yet been implemented.
// It uses the Go's net package to implement networking.
// See LinuxSocketBind for a proper implementation on the Linux platform.
type StdNetBind struct {
	mu          sync.Mutex // protects following fields
	ipv4        *net.UDPConn
	ipv6        *net.UDPConn
	blackhole4  bool
	blackhole6  bool
	bindAddress string
}

func NewStdNetBind(bindAddress string) conn.Bind {
	return &StdNetBind{
		bindAddress: bindAddress,
	}
}

type StdNetEndpoint netip.AddrPort

var (
	_ conn.Bind     = (*StdNetBind)(nil)
	_ conn.Endpoint = StdNetEndpoint{}
)

func (*StdNetBind) ParseEndpoint(s string) (conn.Endpoint, error) {
	e, err := netip.ParseAddrPort(s)
	return asEndpoint(e), err
}

func (StdNetEndpoint) ClearSrc() {}

func (e StdNetEndpoint) DstIP() netip.Addr {
	return (netip.AddrPort)(e).Addr()
}

func (e StdNetEndpoint) SrcIP() netip.Addr {
	return netip.Addr{} // not supported
}

func (e StdNetEndpoint) DstToBytes() []byte {
	b, _ := (netip.AddrPort)(e).MarshalBinary()
	return b
}

func (e StdNetEndpoint) DstToString() string {
	return (netip.AddrPort)(e).String()
}

func (e StdNetEndpoint) SrcToString() string {
	return ""
}

func (bind *StdNetBind) listenNet(network string, port int) (*net.UDPConn, int, error) {
	addr := &net.UDPAddr{Port: port}

	// try to resolve network address
	resAddr, err := net.ResolveUDPAddr(network, fmt.Sprintf("%s:%d", bind.bindAddress, port))
	if err == nil {
		addr = resAddr
	}

	c, err := net.ListenUDP(network, addr)
	if err != nil {
		return nil, 0, err
	}
	// Retrieve port.
	laddr := c.LocalAddr()
	uaddr, err := net.ResolveUDPAddr(
		laddr.Network(),
		laddr.String(),
	)
	if err != nil {
		return nil, 0, err
	}
	return c, uaddr.Port, nil
}

//gocyclo:ignore
func (bind *StdNetBind) Open(uport uint16) ([]conn.ReceiveFunc, uint16, error) {
	bind.mu.Lock()
	defer bind.mu.Unlock()

	var err error
	var tries int

	if bind.ipv4 != nil || bind.ipv6 != nil {
		return nil, 0, conn.ErrBindAlreadyOpen
	}

	// Attempt to open ipv4 and ipv6 listeners on the same port.
	// If uport is 0, we can retry on failure.
again:
	port := int(uport)
	var ipv4, ipv6 *net.UDPConn

	ipv4, port, err = bind.listenNet("udp4", port)
	if err != nil && !errors.Is(err, syscall.EAFNOSUPPORT) {
		return nil, 0, err
	}

	// Listen on the same port as we're using for ipv4.
	ipv6, port, err = bind.listenNet("udp6", port)
	if uport == 0 && errors.Is(err, syscall.EADDRINUSE) && tries < 100 {
		ipv4.Close()
		tries++
		goto again
	}
	if err != nil && !errors.Is(err, syscall.EAFNOSUPPORT) {
		ipv4.Close()
		return nil, 0, err
	}
	var fns []conn.ReceiveFunc
	if ipv4 != nil {
		fns = append(fns, bind.makeReceiveIPv4(ipv4))
		bind.ipv4 = ipv4
	}
	if ipv6 != nil {
		fns = append(fns, bind.makeReceiveIPv6(ipv6))
		bind.ipv6 = ipv6
	}
	if len(fns) == 0 {
		return nil, 0, syscall.EAFNOSUPPORT
	}
	return fns, uint16(port), nil
}

func (bind *StdNetBind) BatchSize() int {
	return 1
}

func (bind *StdNetBind) Close() error {
	bind.mu.Lock()
	defer bind.mu.Unlock()

	var err1, err2 error
	if bind.ipv4 != nil {
		err1 = bind.ipv4.Close()
		bind.ipv4 = nil
	}
	if bind.ipv6 != nil {
		err2 = bind.ipv6.Close()
		bind.ipv6 = nil
	}
	bind.blackhole4 = false
	bind.blackhole6 = false
	if err1 != nil {
		return err1
	}
	return err2
}

func (*StdNetBind) makeReceiveIPv4(c *net.UDPConn) conn.ReceiveFunc {
	return func(buffs [][]byte, sizes []int, eps []conn.Endpoint) (n int, err error) {
		size, endpoint, err := c.ReadFromUDPAddrPort(buffs[0])
		if err == nil {
			sizes[0] = size
			eps[0] = asEndpoint(endpoint)
			return 1, nil
		}
		return 0, err
	}
}

func (*StdNetBind) makeReceiveIPv6(c *net.UDPConn) conn.ReceiveFunc {
	return func(buffs [][]byte, sizes []int, eps []conn.Endpoint) (n int, err error) {
		size, endpoint, err := c.ReadFromUDPAddrPort(buffs[0])
		if err == nil {
			sizes[0] = size
			eps[0] = asEndpoint(endpoint)
			return 1, nil
		}
		return 0, err
	}
}

func (bind *StdNetBind) Send(buffs [][]byte, endpoint conn.Endpoint) error {
	var err error
	nend, ok := endpoint.(StdNetEndpoint)
	if !ok {
		return conn.ErrWrongEndpointType
	}
	addrPort := netip.AddrPort(nend)

	bind.mu.Lock()
	blackhole := bind.blackhole4
	c := bind.ipv4
	if addrPort.Addr().Is6() {
		blackhole = bind.blackhole6
		c = bind.ipv6
	}
	bind.mu.Unlock()

	if blackhole {
		return nil
	}
	if c == nil {
		return syscall.EAFNOSUPPORT
	}
	for _, buff := range buffs {
		_, err = c.WriteToUDPAddrPort(buff, addrPort)
		if err != nil {
			return err
		}
	}
	return nil
}

// endpointPool contains a re-usable set of mapping from netip.AddrPort to Endpoint.
// This exists to reduce allocations: Putting a netip.AddrPort in an Endpoint allocates,
// but Endpoints are immutable, so we can re-use them.
var endpointPool = sync.Pool{
	New: func() any {
		return make(map[netip.AddrPort]conn.Endpoint)
	},
}

// asEndpoint returns an Endpoint containing ap.
func asEndpoint(ap netip.AddrPort) conn.Endpoint {
	m := endpointPool.Get().(map[netip.AddrPort]conn.Endpoint)
	defer endpointPool.Put(m)
	e, ok := m[ap]
	if !ok {
		e = conn.Endpoint(StdNetEndpoint(ap))
		m[ap] = e
	}
	return e
}

func (bind *StdNetBind) SetMark(_ uint32) error {
	return nil
}
