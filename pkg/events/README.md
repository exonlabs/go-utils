<br>

This package provides a simple and efficient implementation of an event signaling mechanism in Go. It allows you to manage an internal flag that can be set to true and reset to false, enabling synchronization between goroutines.

Features:

- **Set**: Set the internal flag to true, waking up any goroutines waiting for it to be true.
- **Clear**: Reset the internal flag to false, causing subsequent wait calls to block until the flag is set again.
- **IsSet**: Check if the event is currently set.
- **Wait**: Block until the internal flag is set or a specified timeout elapses.
