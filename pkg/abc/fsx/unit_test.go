// Copyright (c) 2024 ExonLabs, All rights reserved.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

package fsx_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/exonlabs/go-utils/pkg/abc/fsx"
)

func TestParsePath(t *testing.T) {
	path, err := fsx.ParsePath(".")
	assert.NoError(t, err, "should not return an error for a valid path")
	assert.NotEmpty(t, path, "should return a valid absolute path")

	_, err = fsx.ParsePath("")
	assert.Error(t, err, "should return an error for an empty path")
}

func TestIsExist(t *testing.T) {
	dir := t.TempDir()

	assert.True(t, fsx.IsExist(dir), "directory should exist")

	nonExistentPath := filepath.Join(dir, "nonexistent")
	assert.False(t, fsx.IsExist(nonExistentPath),
		"non-existent file should not exist")
}

func TestCopyFile(t *testing.T) {
	srcFile := filepath.Join(t.TempDir(), "srcfile.txt")
	dstFile := filepath.Join(t.TempDir(), "dstfile.txt")

	err := os.WriteFile(srcFile, []byte("test content"), 0o664)
	assert.NoError(t, err,
		"should not return error on writing to source file")

	err = fsx.Copy(srcFile, dstFile)
	assert.NoError(t, err, "should not return error during file copy")

	content, err := os.ReadFile(dstFile)
	assert.NoError(t, err,
		"should not return error reading the destination file")
	assert.Equal(t, "test content", string(content),
		"destination file content should match the source")
}

func TestCopySymlink(t *testing.T) {
	srcFile := filepath.Join(t.TempDir(), "srcfile.txt")
	err := os.WriteFile(srcFile, []byte("test content"), 0o664)
	assert.NoError(t, err)

	srcSymlink := srcFile + "_symlink"
	err = os.Symlink(srcFile, srcSymlink)
	assert.NoError(t, err, "should not return error creating symlink")

	dstSymlink := srcFile + "_dst_symlink"
	err = fsx.Copy(srcSymlink, dstSymlink)
	assert.NoError(t, err, "should not return error during symlink copy")

	linkDest, err := os.Readlink(dstSymlink)
	assert.NoError(t, err,
		"should not return error reading destination symlink")
	assert.Equal(t, srcFile, linkDest,
		"symlink destination should match the original")
}

func TestCopyDir(t *testing.T) {
	srcDir := t.TempDir()
	dstDir := t.TempDir()

	testFilePath := filepath.Join(srcDir, "testfile.txt")
	err := os.WriteFile(testFilePath, []byte("test content"), 0o664)
	assert.NoError(t, err, "should not return error writing to test file")

	err = fsx.CopyDir(srcDir, dstDir)
	assert.NoError(t, err, "should not return error during directory copy")

	content, err := os.ReadFile(filepath.Join(dstDir, "testfile.txt"))
	assert.NoError(t, err, "should not return error reading copied file")
	assert.Equal(t, "test content", string(content),
		"copied file content should match the original")
}

func TestCopyDirToExistingDirectory(t *testing.T) {
	srcDir := t.TempDir()
	dstDir := t.TempDir()

	testFilePath := filepath.Join(srcDir, "testfile.txt")
	err := os.WriteFile(testFilePath, []byte("test content"), 0o664)
	assert.NoError(t, err)

	err = fsx.CopyDir(srcDir, dstDir)
	assert.NoError(t, err,
		"should not return error during directory copy to an existing directory")

	content, err := os.ReadFile(filepath.Join(dstDir, "testfile.txt"))
	assert.NoError(t, err, "should not return error reading copied file")
	assert.Equal(t, "test content", string(content),
		"copied file content should match the original")
}

func TestRemoveFile(t *testing.T) {
	srcFile := filepath.Join(t.TempDir(), "srcfile.txt")
	err := os.WriteFile(srcFile, []byte("test content"), 0o664)
	assert.NoError(t, err,
		"should not return error on writing to source file")
	assert.True(t, fsx.IsExist(srcFile),
		"source file should exist before remove")

	err = fsx.Remove(srcFile)
	assert.NoError(t, err,
		"should not return error on remove source file")
	assert.False(t, fsx.IsExist(srcFile),
		"source file should not exist after remove")
}

func TestRemoveDir(t *testing.T) {
	srcDir := filepath.Join(t.TempDir(), "srcdir")
	os.MkdirAll(srcDir, 0o775)

	srcFile := filepath.Join(srcDir, "srcfile.txt")
	err := os.WriteFile(srcFile, []byte("test content"), 0o664)
	assert.NoError(t, err,
		"should not return error on writing to source file")
	assert.True(t, fsx.IsExist(srcDir),
		"source file should exist before remove")

	err = fsx.Remove(srcDir)
	assert.NoError(t, err,
		"should not return error on remove source dir")
	assert.False(t, fsx.IsExist(srcDir),
		"source file should not exist after remove")
}

func TestTouch(t *testing.T) {
	srcFile := filepath.Join(t.TempDir(), "srcdir", "srcfile.txt")
	assert.False(t, fsx.IsExist(srcFile),
		"source file should not exist before touch")
	err := fsx.Touch(srcFile)
	assert.NoError(t, err,
		"should not return error on touching source file")
	assert.True(t, fsx.IsExist(srcFile),
		"source file should exist after touch")
}

func TestFilesEqual(t *testing.T) {
	tmpDir := t.TempDir()

	_, err := fsx.FilesEqual()
	assert.Error(t, err, "should return error with no files to compare")
	_, err = fsx.FilesEqual("", "")
	assert.Error(t, err, "should return error with empty files paths")

	testContent := []byte("test content")
	files := []string{}
	for i := 0; i < 5; i++ {
		fpath := filepath.Join(tmpDir, fmt.Sprintf("file%d.txt", i))
		err := os.WriteFile(fpath, testContent, 0o664)
		assert.NoError(t, err,
			"should not return error on writing to file")
		files = append(files, fpath)
	}

	result, err := fsx.FilesEqual(files...)
	assert.NoError(t, err,
		"should not return error during files compare")
	assert.Equal(t, result, true, "files compare should match")

	// testing non-matching files
	diffContent := []byte("different test content")
	fpath := filepath.Join(tmpDir, "filex.txt")
	err = os.WriteFile(fpath, diffContent, 0o664)
	assert.NoError(t, err,
		"should not return error on writing to file")
	files = append(files, fpath)

	result, err = fsx.FilesEqual(files...)
	assert.NoError(t, err,
		"should not return error during files compare")
	assert.Equal(t, result, false, "files compare should not match")
}
