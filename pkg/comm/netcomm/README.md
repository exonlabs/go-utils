<br>

This package provides functionalities for establishing network-based
communications in Go applications.

## Installation

```bash
go get github.com/exonlabs/go-utils/pkg/comm/netcomm
```

## Usage

#### Connection URI

```<network>@<host>:<port>```

- **network**: TCP and UDP networks {tcp|tcp4|tcp6|udp|udp4|udp6}
- **host**:    The host FQDN or IP address.
- **port**:    The port number. can be number or protocol name.

#### Usage Example

https://github.com/exonlabs/go-utils/tree/master/examples/comm
