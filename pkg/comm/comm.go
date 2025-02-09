// Copyright (c) 2024 ExonLabs, All rights reserved.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

package comm

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/exonlabs/go-utils/pkg/abc/dictx"
)

// Connection represents a generic interface for handling client side connections.
type Connection interface {
	// Uri returns the URI of the connection.
	Uri() string

	// Type returns the type of the connection.
	Type() string

	// Parent returns the listener that created this connection.
	Parent() Listener

	// IsOpened checks if the connection is currently open.
	IsOpened() bool

	// Open establishes the connection, with a specified timeout.
	Open(timeout float64) error

	// Close terminates the connection.
	Close()

	// Cancel interrupts any ongoing operation with this connection.
	Cancel()

	// CancelSend interrupts any ongoing sending operation with this connection.
	CancelSend()

	// CancelRecv interrupts any ongoing receiving operation with this connection.
	CancelRecv()

	// Send transmits data over the connection, with a specified timeout.
	Send(data []byte, timeout float64) error

	// SendTo transmits data to addr over the connection, with a specified timeout.
	SendTo(data []byte, addr any, timeout float64) error

	// Recv receives data over the connection, with a specified timeout.
	Recv(timeout float64) (data []byte, err error)

	// RecvFrom receives data from addr over the connection, with a specified timeout.
	RecvFrom(timeout float64) (data []byte, addr any, err error)
}

// Listener represents a generic interface for handling server side connections.
type Listener interface {
	// Uri returns the URI of the listener.
	Uri() string

	// Type returns the type of listener.
	Type() string

	// IsActive checks if the listener is currently active.
	IsActive() bool

	// Start initializes the listener and begins accepting connections.
	Start() error

	// Stop terminates the listener and closes all connections.
	Stop()

	// SetConnHandler sets a callback function to handle incoming connections.
	SetConnHandler(handler func(conn Connection))
}

// PollingConfig is used to configure read polling.
type PollingConfig struct {
	// Timeout defines the timeout in seconds for read data polling.
	// polling timeout value must be > 0.
	Timeout float64
	// ChunkSize defines the size of chunks to read during polling.
	// polling chunk size value must be > 0.
	ChunkSize int
	// MaxSize defines the maximum size for read data.
	// use 0 or negative value to disable max limit.
	MaxSize int
}

// ParsePollingConfig returns polling configuration from parsed options.
//
// The parsed options are:
//   - poll_timeout: (float64) the timeout in seconds for read data polling.
//     polling timeout value must be > 0.
//   - poll_chunksize: (int) the size of chunks to read during polling.
//     polling chunk size value must be > 0.
//   - poll_maxsize: (int) the maximum size for read data.
//     use 0 or negative value to disable max limit for read polling.
func ParsePollingConfig(opts dictx.Dict) (*PollingConfig, error) {
	// initial default values
	cfg := &PollingConfig{
		Timeout:   0.01,
		ChunkSize: 102400,
		MaxSize:   -1,
	}

	if v := dictx.GetFloat(opts, "poll_timeout", 0); v > 0 {
		cfg.Timeout = v
	}
	if v := dictx.GetInt(opts, "poll_chunksize", 0); v > 0 {
		cfg.ChunkSize = v
	}
	if v := dictx.GetInt(opts, "poll_maxsize", 0); v > 0 {
		cfg.MaxSize = v
	}

	return cfg, nil
}

// KeepaliveConfig is used to configure keep-alive probes for TCP connections.
type KeepaliveConfig struct {
	// Interval defines in seconds the time between keep-alive probes.
	// use 0 to enable keep-alive probes with OS defined values. (default is 0)
	// use -1 to disable keep-alive probes.
	Interval int
}

// ParseKeepaliveConfig returns keep-alive configuration from parsed options.
//
// The parsed options are:
//   - keepalive_interval: (int) the keep-alive interval in seconds.
//     use 0 to enable keep-alive probes with OS defined values. (default is 0)
//     use -1 to disable keep-alive probes.
func ParseKeepaliveConfig(opts dictx.Dict) (*KeepaliveConfig, error) {
	// initial default values
	cfg := &KeepaliveConfig{
		Interval: -1,
	}

	if v := dictx.GetInt(opts, "keepalive_interval", 0); v >= 0 {
		cfg.Interval = v
	}

	return cfg, nil
}

// LimiterConfig is used to configure limits for listeners.
type LimiterConfig struct {
	// SimultaneousConn defines the at most number of connections that
	// the listener accepts simultaneously.
	// use 0 or negative value to disable connections limit.
	SimultaneousConn int
}

// ParseLimiterConfig returns keep-alive configuration from parsed options.
//
// The parsed options are:
//   - simultaneous_connections: (int) the limit on number of concurrent connections.
//     use 0 or negative value to disable connections limit.
func ParseLimiterConfig(opts dictx.Dict) (*LimiterConfig, error) {
	// initial default values
	cfg := &LimiterConfig{
		SimultaneousConn: -1,
	}

	if v := dictx.GetInt(opts, "simultaneous_connections", 0); v > 0 {
		cfg.SimultaneousConn = v
	}

	return cfg, nil
}

// TlsConfig is used to configure the TLS attributes for TCP connections.
type TlsConfig = tls.Config

// ParseTlsConfig returns tls configuration from parsed options.
//
// The parsed options are:
//   - tls_enable: (bool) enable/disable TLS, default disabled.
//   - tls_mutual_auth: (bool) enable/disable mutual TLS auth, default disabled.
//   - tls_server_name: (string) server name to use by client TLS session.
//   - tls_min_version: (float64) min TLS version to use. default TLS v1.2
//   - tls_max_version: (float64) max TLS version to use. default TLS v1.3
//   - tls_ca_certs: (string) comma separated list of CA certs to use.
//     cert values could be file paths to load or cert content in PEM format.
//   - tls_local_cert: (string) cert to use for TLS session.
//     cert could be file path to load or cert content in PEM format.
//   - tls_local_key: (string) private key to use for TLS session.
//     key could be file path to load or key content in PEM format.
func ParseTlsConfig(opts dictx.Dict) (*TlsConfig, error) {
	if !dictx.Fetch(opts, "tls_enable", false) {
		return nil, nil
	}

	cfg := &TlsConfig{
		ClientAuth:         tls.RequireAnyClientCert,
		InsecureSkipVerify: true,
		MinVersion:         tls.VersionTLS12,
		MaxVersion:         tls.VersionTLS13,
	}

	if dictx.Fetch(opts, "tls_mutual_auth", false) {
		cfg.ClientAuth = tls.RequireAndVerifyClientCert
		cfg.InsecureSkipVerify = false
	}
	if v := strings.TrimSpace(dictx.Fetch(opts, "tls_server_name", "")); v != "" {
		cfg.ServerName = v
	}

	if v := dictx.GetFloat(opts, "tls_min_version", 0); v > 0 {
		switch v {
		case 1.0:
			cfg.MinVersion = tls.VersionTLS10
		case 1.1:
			cfg.MinVersion = tls.VersionTLS11
		case 1.2:
			cfg.MinVersion = tls.VersionTLS12
		case 1.3:
			cfg.MinVersion = tls.VersionTLS13
		default:
			return nil, errors.New("invalid tls_min_version value")
		}
	}
	if v := dictx.GetFloat(opts, "tls_max_version", 0); v > 0 {
		switch v {
		case 1.0:
			cfg.MaxVersion = tls.VersionTLS10
		case 1.1:
			cfg.MaxVersion = tls.VersionTLS11
		case 1.2:
			cfg.MaxVersion = tls.VersionTLS12
		case 1.3:
			cfg.MaxVersion = tls.VersionTLS13
		default:
			return nil, errors.New("invalid tls_max_version value")
		}
	}

	if v := strings.TrimSpace(dictx.Fetch(opts, "tls_ca_certs", "")); v != "" {
		certPool := x509.NewCertPool()

		for _, crtStr := range strings.Split(v, ",") {
			crtStr = strings.TrimSpace(crtStr)

			var crtByte []byte
			var err error
			if strings.HasPrefix(crtStr, "-----BEGIN") {
				crtByte = []byte(crtStr)
			} else {
				crtByte, err = os.ReadFile(crtStr)
				if err != nil {
					return nil, fmt.Errorf(
						"error loading tls_ca_certs - %v", err)
				}
			}
			if !certPool.AppendCertsFromPEM(crtByte) {
				return nil, errors.New("invalid tls_ca_certs value")
			}
		}

		cfg.RootCAs = certPool
		cfg.ClientCAs = certPool
	}

	if crtStr := strings.TrimSpace(
		dictx.Fetch(opts, "tls_local_cert", "")); crtStr != "" {
		keyStr := strings.TrimSpace(dictx.Fetch(opts, "tls_local_key", ""))
		if keyStr == "" {
			return nil, errors.New("empty tls_local_key value")
		}

		var cert tls.Certificate
		var err error
		if strings.HasPrefix(crtStr, "-----BEGIN") &&
			strings.HasPrefix(keyStr, "-----BEGIN") {
			cert, err = tls.X509KeyPair([]byte(crtStr), []byte(keyStr))
		} else if strings.HasPrefix(crtStr, "-----BEGIN") ||
			strings.HasPrefix(keyStr, "-----BEGIN") {
			return nil, errors.New(
				"both options tls_local_cert and tls_local_key should be " +
					"file paths or PEM formatted contents for cert and key")
		} else {
			cert, err = tls.LoadX509KeyPair(crtStr, keyStr)
		}
		if err != nil {
			return nil, fmt.Errorf(
				"error loading tls_local_cert, tls_local_key - %v", err)
		}

		cfg.Certificates = []tls.Certificate{cert}
	}

	return cfg, nil
}
