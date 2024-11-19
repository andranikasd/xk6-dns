package xk6_dns

import (
	"context"
	"fmt"
	"time"

	"go.k6.io/k6/lib"
	"go.k6.io/k6/lib/metrics"
	"go.k6.io/k6/stats"
)

var (
	DialCount          = stats.New("dns.dial.count", stats.Counter)
	DialError          = stats.New("dns.dial.error", stats.Counter)
	ConnectionCount    = stats.New("dns.connection.count", stats.Counter)
	ConnectionTCPCount = stats.New("dns.connection.tcp.count", stats.Counter)
	ConnectionUDPCount = stats.New("dns.connection.udp.count", stats.Counter)
	ConnectionTCPError = stats.New("dns.connection.tcp.error", stats.Counter)
	ConnectionUDPError = stats.New("dns.connection.udp.error", stats.Counter)

	RequestCount       = stats.New("dns.request.count", stats.Counter)
	RequestError       = stats.New("dns.request.error", stats.Counter)
	ResponseTime       = stats.New("dns.response.time", stats.Trend, stats.Time)
	DataTCPSent        = stats.New("dns.data.tcp.sent", stats.Counter)
	DataTCPReceived    = stats.New("dns.data.tcp.received", stats.Counter)
	DataUDPSent        = stats.New("dns.data.udp.sent", stats.Counter)
	DataUDPReceived    = stats.New("dns.data.udp.received", stats.Counter)
)

func reportDial(ctx context.Context) error {
	return pushMetric(ctx, DialCount, 1)
}

func reportDialError(ctx context.Context) error {
	return pushMetric(ctx, DialError, 1)
}

func reportConnection(ctx context.Context, protocol string) error {
	switch protocol {
	case "tcp":
		pushMetric(ctx, ConnectionTCPCount, 1)
	case "udp":
		pushMetric(ctx, ConnectionUDPCount, 1)
	}
	return pushMetric(ctx, ConnectionCount, 1)
}

func reportConnectionError(ctx context.Context, protocol string) error {
	switch protocol {
	case "tcp":
		return pushMetric(ctx, ConnectionTCPError, 1)
	case "udp":
		return pushMetric(ctx, ConnectionUDPError, 1)
	}
	return nil
}

func reportRequest(ctx context.Context) error {
	return pushMetric(ctx, RequestCount, 1)
}

func reportRequestError(ctx context.Context) error {
	return pushMetric(ctx, RequestError, 1)
}

func reportResponseTime(ctx context.Context, rtt time.Duration) error {
	return pushMetric(ctx, ResponseTime, float64(rtt.Milliseconds()))
}

func reportDataSent(ctx context.Context, value float64, protocol string) error {
	switch protocol {
	case "tcp":
		return pushMetric(ctx, DataTCPSent, value)
	case "udp":
		return pushMetric(ctx, DataUDPSent, value)
	}
	return nil
}

func reportDataReceived(ctx context.Context, value float64, protocol string) error {
	switch protocol {
	case "tcp":
		return pushMetric(ctx, DataTCPReceived, value)
	case "udp":
		return pushMetric(ctx, DataUDPReceived, value)
	}
	return nil
}

func pushMetric(ctx context.Context, metric *stats.Metric, value float64) error {
	state := lib.GetState(ctx)
	if state == nil {
		return fmt.Errorf("state is nil")
	}

	now := time.Now()
	stats.PushIfNotDone(ctx, state.Samples, stats.Sample{
		Metric: metric,
		Time:   now,
		Value:  value,
	})

	return nil
}
