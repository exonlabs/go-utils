// Copyright (c) 2024 ExonLabs, All rights reserved.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

package xos

import (
	"os"
	"path/filepath"
	"testing"
)

// testing dir tree:
//
//	tmp_fs
//	  |_ dir_1
//	  |  |_ dir_11
//	  |  |  |_ file_111
//	  |  |  |_ file_112
//	  |  |_ dir_12
//	  |     |_ file_121
//	  |     |_ file_122
//	  |_ dir_2
//	     |_ dir_21
//	        |_ file_211
//	        |_ file_212
func tmpFS(t testing.TB) (string, error) {
	t.Helper()

	t.Logf("creating tmp files structure")
	tmp_fs := filepath.Join(os.TempDir(), "tmp_fs")
	t.Cleanup(func() {
		t.Logf("cleanup tmp files structure")
		os.RemoveAll(tmp_fs)
	})

	paths := []string{
		filepath.Join(tmp_fs, "dir_1", "dir_11", "file_111"),
		filepath.Join(tmp_fs, "dir_1", "dir_11", "file_112"),
		filepath.Join(tmp_fs, "dir_1", "dir_12", "file_121"),
		filepath.Join(tmp_fs, "dir_1", "dir_12", "file_122"),
		filepath.Join(tmp_fs, "dir_2", "dir_21", "file_211"),
		filepath.Join(tmp_fs, "dir_2", "dir_21", "file_212"),
	}
	for _, p := range paths {
		if err := os.MkdirAll(filepath.Dir(p), 0o775); err != nil {
			return "", err
		}
		f, err := os.OpenFile(p, os.O_WRONLY|os.O_CREATE, 0o664)
		if err != nil {
			return "", err
		}
		_, err = f.Write([]byte(p + "\n"))
		if err != nil {
			return "", err
		}
		f.Close()
	}

	return tmp_fs, nil
}

func Test_ParsePath(t *testing.T) {
	paths := [][]string{
		{"", ""},
		{" ", ""},
		{"/", "/"},
		{"/tmp/..", "/"},
		{"/tmp/.", "/tmp"},
		{"/tmp/dir1/../..", "/"},
		{"/tmp/dir1/../", "/tmp"},
		{"/tmp/dir1/.././", "/tmp"},
	}
	for _, v := range paths {
		r, _ := ParsePath(v[0])
		if r != v[1] {
			t.Errorf(
				"Error!! in: \"%s\", out: \"%s\", want: \"%s\"",
				v[0], r, v[1])
		}
	}
}

func Test_IsExist(t *testing.T) {
	tmp_fs, err := tmpFS(t)
	if err != nil {
		t.Errorf("Error!! %s", err.Error())
		return
	}

	existing_paths := []string{
		filepath.Join(tmp_fs, "dir_1", "dir_11", "file_111"),
		filepath.Join(tmp_fs, "dir_1", "dir_11", "file_112"),
		filepath.Join(tmp_fs, "dir_1", "dir_12", "file_121"),
		filepath.Join(tmp_fs, "dir_1", "dir_12", "file_122"),
		filepath.Join(tmp_fs, "dir_2", "dir_21", "file_211"),
		filepath.Join(tmp_fs, "dir_2", "dir_21", "file_212"),
	}
	for _, p := range existing_paths {
		if !IsExist(p) {
			t.Errorf("Error!! invalid existing path status \"%s\"", p)
		}
	}

	non_existing_paths := []string{
		filepath.Join(tmp_fs, "no_dir_1", "dir_11", "file_111"),
		filepath.Join(tmp_fs, "dir_1", "no_dir_11", "file_112"),
		filepath.Join(tmp_fs, "dir_1", "dir_12", "no_file_121"),
		filepath.Join(tmp_fs, "dir_1", "dir_12", "no_file_122"),
		filepath.Join(tmp_fs, "no_dir_2", "dir_21", "file_211"),
		filepath.Join(tmp_fs, "dir_2", "dir_21", "no_file_212"),
	}
	for _, p := range non_existing_paths {
		if IsExist(p) {
			t.Errorf("Error!! invalid non-existing path status \"%s\"", p)
		}
	}
}

func Test_Copy(t *testing.T) {
	tmp_fs, err := tmpFS(t)
	if err != nil {
		t.Errorf("Error!! %s", err.Error())
		return
	}

	err = os.MkdirAll(filepath.Join(tmp_fs, "dir_4"), 0o775)
	if err != nil {
		t.Errorf("Error!! %s", err.Error())
		return
	}

	paths := [][]string{
		{
			filepath.Join(tmp_fs, "dir_1", "dir_11", "file_111"),
			filepath.Join(tmp_fs, "dir_4", "file_111"),
		},
		{
			filepath.Join(tmp_fs, "dir_1", "dir_12", "file_122"),
			filepath.Join(tmp_fs, "dir_4", "file_122"),
		},
		{
			filepath.Join(tmp_fs, "dir_2", "dir_21", "file_212"),
			filepath.Join(tmp_fs, "dir_4", "file_212"),
		},
	}
	for _, v := range paths {
		src, dst := v[0], v[1]
		if err := Copy(src, dst); err != nil {
			t.Errorf("Error!! copy \"%s\" failed, %s", src, err.Error())
		} else if !IsExist(dst) {
			t.Errorf("Error!! copy \"%s\" failed, dst does not exist", src)
		}
	}
}

func Test_Copy_Reject(t *testing.T) {
	tmp_fs, err := tmpFS(t)
	if err != nil {
		t.Errorf("Error!! %s", err.Error())
		return
	}

	paths := [][]string{
		{
			filepath.Join(tmp_fs, "dir_1", "no_dir_11", "file_111"),
			filepath.Join(tmp_fs, "dir_4", "file_111"),
		},
		{
			filepath.Join(tmp_fs, "dir_1", "dir_11", "file_111"),
			filepath.Join(tmp_fs, "dir_4", "file_111"),
		},
		{
			"",
			filepath.Join(tmp_fs, "dir_4", "file_122"),
		},
		{
			filepath.Join(tmp_fs, "dir_2", "dir_21", "file_212"),
			"",
		},
	}
	for _, v := range paths {
		src, dst := v[0], v[1]
		if err := Copy(src, dst); err == nil || IsExist(dst) {
			t.Errorf("Error!! copy completed for \"%s\"", src)
		}
	}
}

func Test_CopyDir(t *testing.T) {
	tmp_fs, err := tmpFS(t)
	if err != nil {
		t.Errorf("Error!! %s", err.Error())
		return
	}

	err = os.MkdirAll(filepath.Join(tmp_fs, "dir_4"), 0o775)
	if err != nil {
		t.Errorf("Error!! %s", err.Error())
		return
	}

	paths := [][]string{
		{
			filepath.Join(tmp_fs, "dir_1", "dir_11"),
			filepath.Join(tmp_fs, "dir_4", "dir_11_copy"),
		},
		{
			filepath.Join(tmp_fs, "dir_1", "dir_12"),
			filepath.Join(tmp_fs, "dir_4", "dir_12_copy"),
		},
		{
			filepath.Join(tmp_fs, "dir_2", "dir_21"),
			filepath.Join(tmp_fs, "dir_4", "dir_21_copy"),
		},
	}
	for _, v := range paths {
		src, dst := v[0], v[1]
		if err := CopyDir(src, dst); err != nil {
			t.Errorf("Error!! copy \"%s\" failed, %s", src, err.Error())
		} else if !IsExist(dst) {
			t.Errorf("Error!! copy \"%s\" failed, dst does not exist", src)
		}
	}
}

func Test_CopyDir_Reject(t *testing.T) {
	tmp_fs, err := tmpFS(t)
	if err != nil {
		t.Errorf("Error!! %s", err.Error())
		return
	}

	paths := [][]string{
		{
			filepath.Join(tmp_fs, "dir_1", "no_dir_11"),
			filepath.Join(tmp_fs, "dir_4"),
		},
		{
			filepath.Join(tmp_fs, "dir_1", "dir_11"),
			filepath.Join(tmp_fs, "no_dir", "dir_4"),
		},
		{
			"",
			filepath.Join(tmp_fs, "dir_4"),
		},
		{
			filepath.Join(tmp_fs, "dir_2", "dir_21"),
			"",
		},
	}
	for _, v := range paths {
		src, dst := v[0], v[1]
		if err := CopyDir(src, dst); err == nil || IsExist(dst) {
			t.Errorf("Error!! copy completed for \"%s\"", src)
		}
	}
}
