// Copyright (c) 2024 ExonLabs, All rights reserved.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

//go:build !windows

package fsx

import (
	"errors"
	"os"
	"time"

	"golang.org/x/sys/unix"
)

// lock attempts to acquire a file lock using the given flag.
// If the lock is already held, it returns false and no error.
// If another error occurs, it returns false and the error.
func lock(f *os.File, flag int) (bool, error) {
	err := unix.Flock(int(f.Fd()), flag)
	if err != nil {
		if errors.Is(err, unix.EWOULDBLOCK) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// lockWait attempts to acquire a file lock within the given timeout.
// It retries until the timeout elapses. Returns an error if it fails.
func lockWait(f *os.File, flag int, timeout float64) error {
	deadline := time.Now().Add(
		time.Duration(timeout * float64(time.Second)))
	for time.Now().Before(deadline) {
		if locked, err := lock(f, flag|unix.LOCK_NB); err != nil {
			return err
		} else if locked {
			return nil
		}
		time.Sleep(50 * time.Millisecond)
	}
	return errors.New("timeout: failed to acquire file lock")
}

// Lock acquires an exclusive (write) lock on the file without blocking.
// Returns false if the lock is already held.
func Lock(f *os.File) (bool, error) {
	return lock(f, unix.LOCK_EX|unix.LOCK_NB)
}

// LockWait attempts to acquire an exclusive (write) lock within the given timeout.
// If timeout is 0 or negative, it waits indefinitely.
func LockWait(f *os.File, timeout float64) error {
	if timeout > 0 {
		return lockWait(f, unix.LOCK_EX, timeout)
	}
	// Blocking call
	_, err := lock(f, unix.LOCK_EX)
	return err
}

// RLock acquires a shared (read) lock on the file without blocking.
// Returns false if the lock is already held.
func RLock(f *os.File) (bool, error) {
	return lock(f, unix.LOCK_SH|unix.LOCK_NB)
}

// RLockWait attempts to acquire a shared (read) lock within the given timeout.
// If timeout is 0 or negative, it waits indefinitely.
func RLockWait(f *os.File, timeout float64) error {
	if timeout > 0 {
		return lockWait(f, unix.LOCK_SH, timeout)
	}
	// Blocking call
	_, err := lock(f, unix.LOCK_SH)
	return err
}

// UnLock releases any lock held on the file.
func UnLock(f *os.File) error {
	return unix.Flock(int(f.Fd()), unix.LOCK_UN|unix.LOCK_NB)
}
