<br>

This package provides utilities for process and concurrency management in Go applications:

- `Process`: Manages OS signal handling and delegates lifecycle control to an embedded Tasklet
- `TaskletManager`: Runs and manages multiple tasklets inside a process, providing monitoring and lifecycle management
- Handles graceful shutdown with configurable timeouts
- Provides robust error handling and recovery from panics
- Supports custom signal handlers for different OS signals (SIGINT, SIGTERM, etc.)

## Features

- Process management with signal handling
- Tasklet management for concurrent operations
- Graceful shutdown with configurable timeouts
- Panic recovery and error handling
- Monitoring capabilities for tasklet health

## Installation

```bash
go get github.com/exonlabs/go-utils/pkg/proc
```

## Usage Examples

https://github.com/exonlabs/go-utils/tree/master/examples/proc
