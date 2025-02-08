// Copyright (c) 2024 ExonLabs, All rights reserved.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

package commutils

import (
	"errors"
	"strings"

	"github.com/exonlabs/go-utils/pkg/abc/dictx"
	"github.com/exonlabs/go-utils/pkg/comm"
	"github.com/exonlabs/go-utils/pkg/comm/netcomm"
	"github.com/exonlabs/go-utils/pkg/comm/serialcomm"
	"github.com/exonlabs/go-utils/pkg/comm/sockcomm"
	"github.com/exonlabs/go-utils/pkg/logging"
)

// NewConnection creates a new Connection based on the provided URI prefix.
// It supports different connection types (e.g., tcp, udp, sock, serial)
func NewConnection(uri string, log *logging.Logger, opts dictx.Dict) (comm.Connection, error) {
	if uri == "" {
		return nil, errors.New("uri should not be empty")
	}

	// Determine the connection type from the URI prefix
	t := strings.ToLower(strings.SplitN(uri, "@", 2)[0])
	switch t {
	case "tcp", "tcp4", "tcp6", "udp", "udp4", "udp6":
		return netcomm.NewConnection(uri, log, opts)
	case "sock":
		return sockcomm.NewConnection(uri, log, opts)
	case "serial":
		return serialcomm.NewConnection(uri, log, opts)
	}

	return nil, comm.ErrUri
}

// NewListener creates a new Listener based on the provided URI prefix.
// It supports different listener types (e.g., tcp, udp, sock, serial)
func NewListener(uri string, log *logging.Logger, opts dictx.Dict) (comm.Listener, error) {
	if uri == "" {
		return nil, errors.New("uri should not be empty")
	}

	// Determine the listener type from the URI prefix
	t := strings.ToLower(strings.SplitN(uri, "@", 2)[0])
	switch t {
	case "tcp", "tcp4", "tcp6", "udp", "udp4", "udp6":
		return netcomm.NewListener(uri, log, opts)
	case "sock":
		return sockcomm.NewListener(uri, log, opts)
	case "serial":
		return serialcomm.NewListener(uri, log, opts)
	}

	return nil, comm.ErrUri
}
