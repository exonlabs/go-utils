package xcopy

import (
	"errors"
	"io"
	"os"
	"path/filepath"
)

func CopyFile(src, dst string) error {
	// check src path and file
	src = filepath.Clean(src)
	if src == string(filepath.Separator) || src == filepath.Dir(src) {
		return errors.New("invalid src path")
	}
	srcInfo, err := os.Stat(src)
	if err != nil || srcInfo.IsDir() {
		return errors.New("invalid src path")
	}

	// check dst path
	dst = filepath.Clean(dst)
	if dst == string(filepath.Separator) || dst == filepath.Dir(dst) {
		return errors.New("invalid dst path")
	} else if dst == src {
		return errors.New("same src and dst paths")
	}
	dstInfo, err := os.Stat(src)
	if err != nil || dstInfo.IsDir() {
		return errors.New("invalid dst path")
	}
	if err := os.Remove(dst); err != nil {
		return err
	}

	// open src file
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	// open dst file
	dstFile, err := os.OpenFile(
		dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, srcInfo.Mode())
	if err != nil {
		return err
	}
	defer dstFile.Close()

	// do copy
	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return err
	}
	return nil
}

func CopyDir(src, dst string) error {
	// check src path and file
	src = filepath.Clean(src)
	if src == string(filepath.Separator) || src == filepath.Dir(src) {
		return errors.New("invalid src path")
	}
	srcInfo, err := os.Stat(src)
	if err != nil {
		return errors.New("invalid src path")
	}

	// check dst path
	dst = filepath.Clean(dst)
	if dst == string(filepath.Separator) || dst == filepath.Dir(dst) {
		return errors.New("invalid dst path")
	} else if dst == src {
		return errors.New("same src and dst paths")
	}

	// Create the destination directory if it doesn't exist
	if err := os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return err
	}
	if err := os.Chmod(dst, srcInfo.Mode()); err != nil {
		return err
	}

	// Read the contents of the source directory
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	// Iterate through the contents
	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())
		if entry.IsDir() {
			if err := CopyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := CopyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}
	return nil
}
