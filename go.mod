module github.com/exonlabs/go-utils

// version = "0.5.0"

go 1.20

require (
	github.com/cespare/xxhash/v2 v2.3.0
	github.com/fatih/color v1.18.0
	github.com/stretchr/testify v1.10.0
	go.bug.st/serial v1.6.3
	golang.org/x/net v0.36.0
	golang.org/x/sys v0.30.0
	golang.org/x/term v0.29.0
)

require (
	github.com/creack/goselect v0.1.2 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/stretchr/objx v0.5.2 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

retract [v0.0.0, v0.4.99] // drop old dev versions before v0.5.0
