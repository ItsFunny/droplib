/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/3/11 2:26 下午
# @File : common.go
# @Description :
# @Attention :
*/
package config

import "path/filepath"

// helper function to make config creation independent of root dir
func rootify(path, root string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(root, path)
}
