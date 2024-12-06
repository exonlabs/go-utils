<br>

This package provides a simple interface for serial communication in Go.
It supports creating connections to serial ports and handling incoming data.

## Installation

```bash
go get github.com/exonlabs/go-utils/pkg/comm/serialcomm
```

## Usage

#### Connection URI

```serial@<port>:<baud>:<mode>```

- **port**: Serial port name (e.g., /dev/ttyS0 or COM1)
- **baud**: Baud rate (e.g., 4800,9600,19200,115200...)
- **mode**: Data bits, parity, and stop bits in the format {8|7}{N|E|O|M|S}{1|2}

#### Usage Example

https://github.com/exonlabs/go-utils/tree/master/examples/comm
