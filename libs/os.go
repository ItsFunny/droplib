/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/3/2 12:49 下午
# @File : os.go
# @Description :
# @Attention :
*/
package libs

import (
	"fmt"
	"io"
	"os"
)

func EnsureDir(dir string, mode os.FileMode) error {
	err := os.MkdirAll(dir, mode)
	if err != nil {
		return fmt.Errorf("could not create directory %q: %w", dir, err)
	}
	return nil
}


// CopyFile copies a file. It truncates the destination file if it exists.
func CopyFile(src, dst string) error {
	info, err := os.Stat(src)
	if err != nil {
		return err
	}

	srcfile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcfile.Close()

	// create new file, truncate if exists and apply same permissions as the original one
	dstfile, err := os.OpenFile(dst, os.O_RDWR|os.O_CREATE|os.O_TRUNC, info.Mode().Perm())
	if err != nil {
		return err
	}
	defer dstfile.Close()

	_, err = io.Copy(dstfile, srcfile)
	return err
}
func Exit(s string) {
	fmt.Printf(s + "\n")
	os.Exit(1)
}


func FileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return !os.IsNotExist(err)
}