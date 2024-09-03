module github.com/exonlabs/go-utils

go 1.20

// ignore old dev versions
retract [v0.0.0, v0.2.9]

require (
	go.bug.st/serial v1.6.2
	golang.org/x/net v0.28.0
	golang.org/x/sys v0.24.0
	golang.org/x/term v0.23.0
)

require github.com/creack/goselect v0.1.2 // indirect
