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
	p := strings.TrimSpace(filepath.Clean(path))
	if len(path) == 0 || len(p) == 0 {
		return "", errors.New("invalid path")
	}
	p, err := filepath.Abs(p)
	if err != nil {
		return "", err
	}
	return p, nil
}

// IsExist checks if path exists on system or not. path can point to
// a directory, regular file or symbolic link.
func IsExist(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}

// CreateDir creates directory at path with any necessary parents and sets
// the permissions perm.
func CreateDir(path string, perm os.FileMode) error {
	path, err := ParsePath(path)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(path, perm); err != nil {
		return err
	}
	return nil
}

// copy regular file from src to dst path and set dst mode. dst file
// will be overwritten if exists. dst parent dir must exist.
func copy_regular(src, dst string, perm os.FileMode) error {
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
	if err := fout.Sync(); err != nil {
		return err
	}
	return nil
}

// copy symbolic link from src to dst path. dst parent dir must exist.
func copy_symlink(src, dst string) error {
	link, err := os.Readlink(src)
	if err != nil {
		return err
	}
	return os.Symlink(link, dst)
}

// Copy copies the contents of src file to dst file. The file will be
// created if not exist. If dst file exists, it will be replaced by the
// contents of the src file. The file mode will be copied from src file.
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
		return errors.New("src path is directory")
	}

	if dst, err = ParsePath(dst); err != nil {
		return err
	}
	if dst == src {
		return errors.New("same src and dst paths")
	}

	if err = os.MkdirAll(filepath.Dir(dst), 0o775); err != nil {
		return err
	}
	if src_info.Mode()&os.ModeType == os.ModeSymlink {
		return copy_symlink(src, dst)
	}
	return copy_regular(src, dst, src_info.Mode())
}

// CopyDir copies recursively a dir contents from src to dst path after doing
// path checking. root path as src or dst and same value for
// src and dst are not allowed.
func CopyDir(src, dst string) error {
	// // clean and check src path
	// src = filepath.Clean(src)
	// if src == string(filepath.Separator) || src == filepath.Dir(src) {
	// 	return errors.New("invalid src path")
	// }
	// sinfo, err := os.Stat(src)
	// if err != nil || !sinfo.IsDir() {
	// 	return errors.New("invalid src path")
	// }

	// // clean and check dst path
	// dst = filepath.Clean(dst)
	// if dst == src {
	// 	return errors.New("same src and dst paths")
	// }
	// if dst == string(filepath.Separator) || dst == filepath.Dir(dst) {
	// 	return errors.New("invalid dst path")
	// }

	// // create dst dir if not exist
	// if err := os.MkdirAll(dst, sinfo.Mode()); err != nil {
	// 	return err
	// }
	// if err := os.Chmod(dst, sinfo.Mode()); err != nil {
	// 	return err
	// }

	// // Read contents of src dir and iterate through the contents
	// entries, err := os.ReadDir(src)
	// if err != nil {
	// 	return err
	// }
	// for _, entry := range entries {
	// 	s := filepath.Join(src, entry.Name())
	// 	d := filepath.Join(dst, entry.Name())
	// 	if entry.IsDir() {
	// 		err = CopyDir(s, d)
	// 	} else {
	// 		err = CopyFile(s, d)
	// 	}
	// 	if err != nil {
	// 		return err
	// 	}
	// }
	return nil
}
