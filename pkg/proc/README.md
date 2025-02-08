<br>

This package provides utilities for managing tasklets with lifecycle control.
tasklets are utilized in routines and processes. It provides features for safely starting, stopping, and terminating tasklets, handling system signals, and managing
tasklets state through various handlers.

## Features

- **TaskletHandler**: Handles tasklet lifecycle, including initialization, execution, and graceful termination.
- **ProcessHandler**: Extends TaskletHandler to manage system signals like `SIGINT`, `SIGTERM`, and others.

## Installation

```bash
go get github.com/exonlabs/go-utils/pkg/proc
```

## Usage Examples

https://github.com/exonlabs/go-utils/tree/master/examples/proc
