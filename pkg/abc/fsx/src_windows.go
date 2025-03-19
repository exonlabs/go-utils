// Copyright (c) 2024 ExonLabs, All rights reserved.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

//go:build windows

package fsx

import (
	"errors"
	"os"
	"time"

	"golang.org/x/sys/windows"
)

// Use of 0x00000000 for the shared lock is a guess based on the MS Windows
// `LockFileEX` docs, which document the `LOCKFILE_EXCLUSIVE_LOCK` flag as:
//
// > The function requests an exclusive lock. Otherwise, it requests a shared lock.
//
// https://msdn.microsoft.com/en-us/library/windows/desktop/aa365203(v=vs.85).aspx
const lockfile_shared_lock = 0x00000000

// The error code returned from the Windows syscall when a lock would block,
// and you ask to fail immediately.
const err_lock_violation windows.Errno = 0x21 // 33

// get a lock on file or fail.
// add the "windows.LOCKFILE_FAIL_IMMEDIATELY" flag to fail immediately
func lock(f *os.File, flag uint32) (bool, error) {
	err := windows.LockFileEx(
		windows.Handle(f.Fd()), flag, 0, 1, 0, &windows.Overlapped{})
	if err != nil && !errors.Is(err, windows.Errno(0)) {
		if errors.Is(err, err_lock_violation) ||
			errors.Is(err, windows.ERROR_IO_PENDING) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// Try getting a lock on file for certain timeout before fail.
func lockWait(f *os.File, flag uint32, timeout float64) error {
	tDeadline := time.Now().Add(
		time.Duration(timeout * float64(time.Second)))
	for {
		locked, err := lock(f, flag|windows.LOCKFILE_FAIL_IMMEDIATELY)
		if err != nil {
			return err
		} else if locked {
			break
		}
		if time.Now().After(tDeadline) {
			return errors.New("timeout")
		}
		time.Sleep(50 * time.Millisecond)
	}
	return nil
}

// Lock tries getting a Read/Write file lock. it will block all operations
// on file till lock is released.
func Lock(f *os.File) (bool, error) {
	return lock(f,
		windows.LOCKFILE_EXCLUSIVE_LOCK|windows.LOCKFILE_FAIL_IMMEDIATELY)
}

// LockWait tries getting a Read/Write file lock for a certain timeout.
// Setting timeout 0 or negative value will wait indefinitely.
func LockWait(f *os.File, timeout float64) error {
	if timeout > 0 {
		return lockWait(f, windows.LOCKFILE_EXCLUSIVE_LOCK, timeout)
	}
	// blocking call
	_, err := lock(f, windows.LOCKFILE_EXCLUSIVE_LOCK)
	return err
}

// RLock tries getting a Read file lock. it will block all write operations
// on file till lock is released.
func RLock(f *os.File) (bool, error) {
	return lock(f,
		lockfile_shared_lock|windows.LOCKFILE_FAIL_IMMEDIATELY)
}

// RLockWait tries getting a Read file lock for a certain timeout.
// Setting timeout 0 or negative value will wait indefinitely.
func RLockWait(f *os.File, timeout float64) error {
	if timeout > 0 {
		return lockWait(f, lockfile_shared_lock, timeout)
	}
	// blocking call
	_, err := lock(f, lockfile_shared_lock)
	return err
}

// UnLock releases the lock on a file
func UnLock(f *os.File) error {
	err := windows.UnlockFileEx(
		windows.Handle(f.Fd()), 0, 1, 0, &windows.Overlapped{})
	if err != nil && !errors.Is(err, windows.Errno(0)) {
		return err
	}
	return nil
}
