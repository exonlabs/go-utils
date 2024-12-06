<br>

This package provides a simple model for working with named pipes on Unix-like
systems. It allows for creating, reading, writing and managing named pipes.

**Named Pipes** are also called FIFO (First In, First Out) are a special type
of files which are used to facilitate two-way communication between processes
without storing anything on the disk.

## Features

- **Non-blocking I/O**: Supports non-blocking read and write operations.
- **Customizable Polling**: Set custom poll intervals and chunk sizes for
efficient data transfer.
- **Timeouts and Break Events**: Easily manage read and write timeouts and
handle cancelable operations with break events.

## Installation

```bash
go get github.com/exonlabs/go-utils/pkg/unix/namedpipes
```

## Usage Examples

https://github.com/exonlabs/go-utils/tree/master/examples/namedpipe
