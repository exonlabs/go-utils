// Copyright (c) 2024 ExonLabs, All rights reserved.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

package xos

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// ParsePath cleans and verifies path, empty paths are rejected.
// returns cleaned path if no errors.
func ParsePath(path string) (string, error) {
	p := strings.TrimSpace(path)
	if len(path) == 0 || len(p) == 0 {
		return "", errors.New("invalid path")
	}
	return filepath.Abs(p)
}

// IsExist checks if path exists on system or not. path can point to
// a directory, regular file or symbolic link.
func IsExist(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}

// copy regular file from src to dst path and set dst mode. dst file
// will be overwritten if exists. dst parent dir must exist.
func copy_regular(src, dst string, mode os.FileMode) error {
	fin, err := os.Open(src)
	if err != nil {
		return err
	}
	defer fin.Close()

	fout, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	defer fout.Close()

	if _, err := io.Copy(fout, fin); err != nil {
		return err
	}
	return fout.Sync()
}

// copy symbolic link from src to dst path and set dst link mode.
// dst parent dir must exist.
func copy_symlink(src, dst string) error {
	link, err := os.Readlink(src)
	if err != nil {
		return err
	}
	return os.Symlink(link, dst)
}

// Copy copies the contents of src file to dst file. dst file parent dir
// must exist. If dst file exists, its content will be replaced by contents
// of src file. The file mode is copied from src file.
func Copy(src, dst string) error {
	var err error

	if src, err = ParsePath(src); err != nil {
		return err
	}
	src_info, err := os.Stat(src)
	if err != nil {
		return err
	}
	if src_info.IsDir() {
		return errors.New("src path is a directory")
	}

	if dst, err = ParsePath(dst); err != nil {
		return err
	}
	if dst == src {
		return errors.New("same src and dst paths")
	}
	if _, err = os.Stat(filepath.Dir(dst)); os.IsNotExist(err) {
		return errors.New("dst parent directory does not exist")
	}
	if dst_info, _ := os.Stat(dst); dst_info != nil && dst_info.IsDir() {
		return errors.New("dst path is a directory")
	}

	if (src_info.Mode() & os.ModeType) == os.ModeSymlink {
		return copy_symlink(src, dst)
	}
	return copy_regular(src, dst, src_info.Mode())
}

// copy dir recursively from src to dst path.
func copy_dir(src, dst string) error {
	var err error

	src_info, err := os.Stat(src)
	if err != nil {
		return err
	}

	if err = os.MkdirAll(dst, src_info.Mode()); err != nil {
		return err
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		s := filepath.Join(src, entry.Name())
		d := filepath.Join(dst, entry.Name())
		switch entry.Type() {
		case os.ModeDir:
			err = copy_dir(s, d)
		case os.ModeSymlink:
			err = copy_symlink(s, d)
		default:
			s_info, _ := os.Stat(s)
			err = copy_regular(s, d, s_info.Mode())
		}
		if err != nil {
			return err
		}
	}
	return nil
}

// CopyDir copies dir recursively from src to dst path.
// dst parent dir must exist.
func CopyDir(src, dst string) error {
	var err error

	if src, err = ParsePath(src); err != nil {
		return err
	}
	src_info, err := os.Stat(src)
	if err != nil {
		return err
	}
	if !src_info.IsDir() {
		return errors.New("src path is not a directory")
	}

	if dst, err = ParsePath(dst); err != nil {
		return err
	}
	if dst == src {
		return errors.New("same src and dst paths")
	}
	if _, err = os.Stat(filepath.Dir(dst)); os.IsNotExist(err) {
		return errors.New("dst parent directory does not exist")
	}

	return copy_dir(src, dst)
}
