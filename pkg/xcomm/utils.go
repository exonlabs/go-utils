package xcomm

import (
	"strings"
)

// create new connection handler
func NewConnection(uri string, opts Options, log *Logger) (Connection, error) {
	t := strings.ToLower(strings.SplitN(uri, "@", 2)[0])
	switch t {
	case "serial":
		return NewSerialConnection(uri, opts, log)
	default:
		return NewNetConnection(uri, opts, log)
	}
}

// create new listener handler
func NewListener(uri string, opts Options, log *Logger) (Listener, error) {
	t := strings.ToLower(strings.SplitN(uri, "@", 2)[0])
	switch t {
	case "serial":
		return NewSerialListener(uri, opts, log)
	default:
		return NewNetListener(uri, opts, log)
	}
}
