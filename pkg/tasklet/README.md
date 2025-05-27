<br>

The tasklet package offers a robust framework for managing long-running jobs with complete lifecycle management through initialization, execution, and termination phases.

It provides:

- **Controlled Lifecycle**: Initialize job resources, execute job in a loop, and properly release job resources on termination.
- **State Management**: Enable/disable functionality for controlled starts and stops.
- **Error Handling**: Built-in panic recovery and error logging.
- **Synchronization Primitives**: Simple API for cooperative sleep and termination.

## Installation

```bash
go get github.com/exonlabs/go-utils/pkg/tasklet
```
