/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/3/31 4:43 下午
# @File : opts.go
# @Description :
# @Attention :
*/
package config

func Port(port int) Option {
	return func(cfg *P2PConfiguration) {
		cfg.IdentityProperty.Port = port
	}
}
func DisablePingPong() Option {
	return func(cfg *P2PConfiguration) {
		cfg.LifeCycleProperty.EnablePingPong = false
	}
}
func Tls(bs []byte) Option {
	return func(cfg *P2PConfiguration) {
		cfg.IdentityProperty.EnableTls = true
		cfg.IdentityProperty.TlsBytes = bs
	}
}
func TlsPath(path string) Option {
	return func(cfg *P2PConfiguration) {
		cfg.IdentityProperty.EnableTls = true
		cfg.IdentityProperty.TlsPath = path
	}
}
