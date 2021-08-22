/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/3/2 3:44 下午
# @File : byte_slice.go
# @Description :
# @Attention :
*/
package libs


// Fingerprint returns the first 6 bytes of a byte slice.
// If the slice is less than 6 bytes, the fingerprint
// contains trailing zeroes.
func Fingerprint(slice []byte) []byte {
	fingerprint := make([]byte, 6)
	copy(fingerprint, slice)
	return fingerprint
}

