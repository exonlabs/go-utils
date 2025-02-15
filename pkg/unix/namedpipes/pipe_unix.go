// Copyright (c) 2024 ExonLabs, All rights reserved.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

//go:build !windows

package namedpipes

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"syscall"
	"time"

	"github.com/exonlabs/go-utils/pkg/abc/gx"
	"golang.org/x/sys/unix"
)

// open opens the pipe file with the given mode if it's not already open.
func (p *NamedPipe) open(mode int) error {
	if p.fd == nil {
		var err error
		p.fd, err = os.OpenFile(p.path, mode, os.ModeNamedPipe)
		if err != nil {
			return fmt.Errorf("%w, %v", ErrOpen, err)
		}
	}
	return nil
}

// open_read opens the pipe for reading.
func (p *NamedPipe) open_read() error {
	return p.open(os.O_RDONLY | unix.O_NONBLOCK)
}

// open_write opens the pipe for writing.
func (p *NamedPipe) open_write() error {
	return p.open(os.O_WRONLY | unix.O_NONBLOCK)
}

// close closes the pipe if it's open.
func (p *NamedPipe) close() {
	if p.fd != nil {
		p.fd.Close()
	}
	p.fd = nil
}

// Read waits to receive data from the named pipe until a timeout occurs,
// cancel/close events or an error occurs.
// timeout=0 waits forever until data is received.
func (p *NamedPipe) Read(timeout float64) ([]byte, error) {
	// Acquire lock
	p.mu.Lock()
	defer p.mu.Unlock()

	p.breakEvent.Clear()

	// determine read buffer size
	nRead := p.PollChunkSize
	if p.PollMaxSize > 0 {
		nRead = p.PollMaxSize
	}

	// set read polling duration and deadline
	var tPolling float64
	var tDeadline time.Time

	tPolling = gx.Max(0.01, p.PollTimeout)
	if timeout > 0 {
		tDeadline = time.Now().Add(
			time.Duration(timeout * float64(time.Second)))
	}

	var data []byte
	for {
		// open pipe for read if not already openned
		if p.fd == nil {
			if err := p.open_read(); err == nil {
				defer p.close()
			}
		}

		if p.fd != nil {
			b := make([]byte, nRead)
			n, err := p.fd.Read(b)
			if err != nil && err != io.EOF {
				return nil, fmt.Errorf("%w, %v", ErrRead, err)
			}
			if n > 0 {
				data = append(data, b[:n]...)
				if p.PollMaxSize > 0 {
					nRead -= n
					if nRead <= 0 {
						break
					}
				}
			} else if len(data) > 0 {
				break
			}
		}

		if !p.breakEvent.Wait(tPolling) {
			return nil, ErrBreak
		}
		if timeout > 0 && time.Now().After(tDeadline) {
			return nil, ErrTimeout
		}
	}

	return data, nil
}

// Write wait to write data to the named pipe until a timeout occurs,
// cancel/close events or an error occurs.
// timeout=0 waits forever until data is written.
func (p *NamedPipe) Write(data []byte, timeout float64) error {
	// Acquire lock
	p.mu.Lock()
	defer p.mu.Unlock()

	p.breakEvent.Clear()

	// set write polling duration and deadline
	var tPolling float64
	var tDeadline time.Time

	tPolling = gx.Max(0.01, p.PollTimeout)
	if timeout > 0 {
		tDeadline = time.Now().Add(
			time.Duration(timeout * float64(time.Second)))
	}

	for {
		// open pipe for write if not already openned
		if p.fd == nil {
			if err := p.open_write(); err == nil {
				defer p.close()
			}
		}

		if p.fd != nil {
			if _, err := p.fd.Write(data); err != nil {
				return fmt.Errorf("%w, %v", ErrWrite, err)
			}
			return nil
		}

		if !p.breakEvent.Wait(tPolling) {
			return ErrBreak
		}
		if timeout > 0 && time.Now().After(tDeadline) {
			return ErrTimeout
		}
	}
}

/////////////////////////////////////////////////////

// Create creates a named pipe at the specified path with the given permissions.
func Create(path string, perm os.FileMode) error {
	path = filepath.Clean(path)
	_, err := os.Lstat(path)
	if os.IsNotExist(err) {
		if err := syscall.Mkfifo(path, uint32(perm)); err != nil {
			return err
		}
	} else if err != nil {
		return err
	}
	return nil
}

// Delete removes the named pipe at the specified path if it exists.
func Delete(path string) error {
	path = filepath.Clean(path)
	err := os.Remove(path)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}
