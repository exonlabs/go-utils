package xcomm

import (
	"strings"

	"github.com/exonlabs/go-utils/pkg/xlog"
)

// create new connection handler
func NewConnection(uri string, log *xlog.Logger) (Connection, error) {
	t := strings.ToLower(strings.SplitN(uri, "@", 2)[0])
	switch t {
	case "serial":
		return NewSerialConnection(uri, log)
	default:
		return NewNetConnection(uri, log)
	}
}

// create new listener handler
func NewListener(uri string, log *xlog.Logger) (Listener, error) {
	t := strings.ToLower(strings.SplitN(uri, "@", 2)[0])
	switch t {
	case "serial":
		return NewSerialListener(uri, log)
	default:
		return NewNetListener(uri, log)
	}
}
