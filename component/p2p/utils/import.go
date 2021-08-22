/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/3/25 4:00 下午
# @File : import.go
# @Description :
# @Attention :
*/
package utils

import (
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"strings"
)

// ecdsa私钥读取
func ECDSA_KEY_IMPORTER(raw, pwd []byte) (*ecdsa.PrivateKey, error) {
	var res *ecdsa.PrivateKey
	if len(raw) == 0 {
		return nil, errors.New("Invalid PEM. It must be different from nil.")
	}

	if strings.Contains(string(raw), "BEGIN") {
		block, _ := pem.Decode(raw)
		if block == nil {
			return res, fmt.Errorf("Failed decoding PEM. Block must be different from nil. [% x]", raw)
		}

		// TODO: derive from header the type of the key

		if x509.IsEncryptedPEMBlock(block) {
			if len(pwd) == 0 {
				return res, errors.New("Encrypted Key. Need a password")
			}

			decrypted, err := x509.DecryptPEMBlock(block, pwd)
			if err != nil {
				return res, fmt.Errorf("Failed PEM decryption [%s]", err)
			}

			res, err = DERToPrivateKey(decrypted)
			if err != nil {
				return res, err
			}
			return res, err
		}
		raw = block.Bytes
	}
	resWrapper, err := DERToPrivateKey(raw)
	if err != nil {
		return resWrapper, err
	}
	return resWrapper, err
}
func DERToPrivateKey(der []byte) (*ecdsa.PrivateKey, error) {
	if key, err := x509.ParsePKCS8PrivateKey(der); err == nil {
		switch key.(type) {
		case *rsa.PrivateKey:
			return nil, errors.New("实际为rsa密钥")
		case *ecdsa.PrivateKey:
			return key.(*ecdsa.PrivateKey), nil
		default:
			return nil, errors.New("Found unknown private key type in PKCS#8 wrapping")
		}
	}

	if key, err := x509.ParseECPrivateKey(der); err == nil {
		return key, nil
	}

	return nil, errors.New("不是ecdsa 私钥")
}
