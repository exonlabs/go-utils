// Copyright (c) 2024 ExonLabs, All rights reserved.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

package fsx

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// ParsePath validates and returns the absolute path.
func ParsePath(path string) (string, error) {
	p := strings.TrimSpace(path)
	if len(p) == 0 {
		return "", errors.New("invalid path")
	}
	return filepath.Abs(p)
}

// IsExist checks if a file or directory exists.
func IsExist(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// copyFile copies regular files from src to dst, preserving file mode.
func copyFile(src, dst string, perm os.FileMode) error {
	fin, err := os.Open(src)
	if err != nil {
		return err
	}
	defer fin.Close()

	fout, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm)
	if err != nil {
		return err
	}
	defer fout.Close()

	if _, err := io.Copy(fout, fin); err != nil {
		return err
	}
	return fout.Sync()
}

// copySymlink copies symbolic links from src to dst.
func copySymlink(src, dst string) error {
	link, err := os.Readlink(src)
	if err != nil {
		return err
	}
	return os.Symlink(link, dst)
}

// Copy copies a file from src to dst. It handles files and symbolic links.
func Copy(src, dst string) error {
	srcInfo, err := os.Lstat(src)
	if err != nil {
		return err
	}
	if srcInfo.IsDir() {
		return errors.New("source is a directory")
	}

	if dst == src {
		return errors.New("source and destination are the same")
	}

	if !IsExist(filepath.Dir(dst)) {
		return errors.New("destination parent directory does not exist")
	}

	if srcInfo.Mode()&os.ModeSymlink != 0 {
		return copySymlink(src, dst)
	}
	return copyFile(src, dst, srcInfo.Mode().Perm())
}

// copyDir recursively copies a directory from src to dst.
func copyDir(src, dst string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dst, srcInfo.Mode().Perm()); err != nil {
		return err
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			err = copyDir(srcPath, dstPath)
		} else if entry.Type()&os.ModeSymlink != 0 {
			err = copySymlink(srcPath, dstPath)
		} else {
			entryInfo, _ := entry.Info()
			err = copyFile(srcPath, dstPath, entryInfo.Mode().Perm())
		}
		if err != nil {
			return err
		}
	}
	return nil
}

// CopyDir copies a directory and its contents from src to dst.
func CopyDir(src, dst string) error {
	srcInfo, err := os.Lstat(src)
	if err != nil {
		return err
	}
	if !srcInfo.IsDir() {
		return errors.New("source is not a directory")
	}

	if dst == src {
		return errors.New("source and destination are the same")
	}

	if !IsExist(filepath.Dir(dst)) {
		return errors.New("destination parent directory does not exist")
	}

	return copyDir(src, dst)
}

// Remove removes regular file or directory if exists.
func Remove(path string) error {
	finfo, err := os.Stat(path)
	if os.IsNotExist(err) {
		return nil
	}
	if finfo.IsDir() {
		return os.RemoveAll(path)
	}
	return os.Remove(path)
}

// Touch creates empty file at path if not exists.
// it returns error if path exist and is a directory.
func Touch(path string) error {
	finfo, err := os.Stat(path)
	if !os.IsNotExist(err) {
		if finfo.IsDir() {
			return errors.New("path is a directory")
		}
		return nil
	}
	// create dir tree
	if err := os.MkdirAll(filepath.Dir(path), 0o775); err != nil {
		return err
	}
	if f, err := os.OpenFile(path, os.O_CREATE, 0o664); err != nil {
		return err
	} else {
		f.Close()
	}
	return nil
}
