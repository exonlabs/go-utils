// Copyright (c) 2024 ExonLabs, All rights reserved.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

package comm

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
