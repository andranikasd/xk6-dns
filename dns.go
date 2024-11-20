package xk6_dns

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/miekg/dns"
)

type K6DNS struct {
	client  *dns.Client
	Version string
}

func NewK6DNS(version string) *K6DNS {
	return &K6DNS{
		client:  &dns.Client{},
		Version: version,
	}
}

func (k *K6DNS) SetDialTimeout(s string) error {
	d, err := time.ParseDuration(s)
	if err != nil {
		return err
	}
	k.client.DialTimeout = d
	return nil
}

func (k *K6DNS) SetReadTimeout(s string) error {
	d, err := time.ParseDuration(s)
	if err != nil {
		return err
	}
	k.client.ReadTimeout = d
	return nil
}

func (k *K6DNS) SetWriteTimeout(s string) error {
	d, err := time.ParseDuration(s)
	if err != nil {
		return err
	}
	k.client.WriteTimeout = d
	return nil
}

func (k *K6DNS) Resolve(ctx context.Context, addr, query, qtypeStr, protocol string) (string, error) {
	// Validate query type
	qtype, ok := dns.StringToType[qtypeStr]
	if !ok {
		return "", fmt.Errorf("unknown query type: %s", qtypeStr)
	}

	// Construct the DNS request message
	msg := &dns.Msg{}
	msg.Id = dns.Id()                  // Unique query ID for every request
	msg.RecursionDesired = true        // Allow recursive resolution
	msg.Question = []dns.Question{{
		Name:   dns.Fqdn(query),       // Ensure query is a fully qualified domain name
		Qtype:  qtype,
		Qclass: dns.ClassINET,
	}}

	reportDial(ctx)

	// Establish a connection based on the specified protocol
	var conn net.Conn
	var err error
	switch protocol {
	case "udp":
		conn, err = NewK6UDPConn(addr)
		reportConnection(ctx, "udp")
	case "tcp":
		conn, err = net.Dial("tcp", addr)
		reportConnection(ctx, "tcp")
	default:
		return "", fmt.Errorf("unsupported protocol: %s", protocol)
	}

	if err != nil {
		reportDialError(ctx)
		reportConnectionError(ctx, protocol)
		return err.Error(), nil
	}
	defer conn.Close()

	reportRequest(ctx)

	// Send the DNS query
	resp, rtt, err := k.client.ExchangeWithConn(msg, &dns.Conn{Conn: conn})
	if err != nil {
		reportRequestError(ctx)
		return err.Error(), nil
	}
	reportResponseTime(ctx, rtt)

	// Capture data sent and received
	switch c := conn.(type) {
	case *k6UDPConn:
		reportDataSent(ctx, float64(c.GetTXBytes()), "udp")
		reportDataReceived(ctx, float64(c.GetRXBytes()), "udp")
	case net.Conn:
		// Estimate for TCP data
		reportDataSent(ctx, float64(len(msg.Question)*100), "tcp")
		reportDataReceived(ctx, float64(len(resp.Question)*100), "tcp")
	}

	return resp.String(), nil
}
