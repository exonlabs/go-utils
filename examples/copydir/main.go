package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/exonlabs/go-utils/pkg/os/xcopy"
)

func main() {
	srcPath := filepath.Join(os.TempDir(), "foobar")
	fmt.Printf("\nUsing src path: %s\n\n", srcPath)

	dirtree := []string{
		"a/a1/a11/a111",
		"a/a1/a12/a112",
		"a/a2/a21",
		"b/b2/b21/b211",
	}
	for _, d := range dirtree {
		os.MkdirAll(filepath.Join(srcPath, d), os.ModePerm)
	}
	fmt.Println("---- created src dir tree:")
	filepath.Walk(srcPath,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			fmt.Println(path)
			return nil
		},
	)

	dstPath := filepath.Join(os.TempDir(), "foobar_copy")
	fmt.Printf("\nUsing dst path: %s\n\n", dstPath)

	fmt.Println("---- copy dir tree to:")
	if err := xcopy.CopyDir(srcPath, dstPath); err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Println("---- verify dst dir tree:")
	filepath.Walk(dstPath,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			fmt.Println(path)
			return nil
		},
	)

}
