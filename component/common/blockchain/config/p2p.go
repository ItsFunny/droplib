/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/3/22 1:27 下午
# @File : config.go
# @Description :
# @Attention :
*/
package config

import (
	"bufio"
	"fmt"
	"github.com/libp2p/go-libp2p-connmgr"
	"github.com/libp2p/go-libp2p-core/network"
	"os"
	"time"
)

var (
	HELLO = "/hello/1.0.0"
)
var HELLO_HANDLER = func(s network.Stream) {
	rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))
	go readData(rw)
	go writeData(rw)
}

func readData(rw *bufio.ReadWriter) {
	for {
		str, _ := rw.ReadString('\n')

		if str == "" {
			return
		}
		if str != "\n" {
			// Green console colour: 	\x1b[32m
			// Reset console colour: 	\x1b[0m
			fmt.Printf("\x1b[32m%s\x1b[0m> ", str)
		}

	}
}
func writeData(rw *bufio.ReadWriter) {
	stdReader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("> ")
		sendData, err := stdReader.ReadString('\n')

		if err != nil {
			panic(err)
		}

		rw.WriteString(fmt.Sprintf("%s\n", sendData))
		rw.Flush()
	}

}

type BootstrapProperty struct {
}

type ConnectionProperty struct {
	HighWater     int
	LowWater      int
	GraceDuration time.Duration
	Opts          connmgr.Option

}
type IdentityProperty struct {
	Port int
	Address string

	PrvBytes []byte
	PrvPath  string

	EnableTls bool
	TlsBytes  []byte
	TlsPath   string
}

type Option func(cfg *P2PConfiguration)

type P2PConfiguration struct {
	IdentityProperty *IdentityProperty
	BootstrapNodes []*BootstrapProperty

	LifeCycleProperty  *LifeCycleProperty
	ConnectionProperty *ConnectionProperty
}

type LifeCycleProperty struct {
	EnablePingPong bool
}

func (p *P2PConfiguration) ValidateBasic() error {
	return nil
}

func NewDefaultP2PConfiguration(opts ...Option) *P2PConfiguration {
	c := &P2PConfiguration{
		ConnectionProperty: &ConnectionProperty{
			HighWater:     200,
			LowWater:      100,
			GraceDuration: 5,
		},
	}
	c.IdentityProperty = defaultIdentityProperty()
	c.LifeCycleProperty = defaultLifeCycleProperty()

	for _, o := range opts {
		o(c)
	}

	return c
}
func defaultLifeCycleProperty() *LifeCycleProperty {
	r := &LifeCycleProperty{
		EnablePingPong: true,
	}

	return r
}
func defaultIdentityProperty() *IdentityProperty {
	bytes := []byte(`-----BEGIN PRIVATE KEY-----
MIGHAgEAMBMGByqGSM49AgEGCCqGSM49AwEHBG0wawIBAQQgB7OLxrOPUrFVjxqu
qcmFMnJvurQxbt78JCbotqyBhc+hRANCAAQoJKlBdc1PIsmJsm895tOk2En2xGNT
NAmry2mTW7mqhRybA7Iy+wYKPaP9VlNBTgTFECObfttm2wScJnkwZaMp
-----END PRIVATE KEY-----
`)
	bytes = nil
	r := &IdentityProperty{
		Port:      10000,
		Address:   "127.0.0.1",
		PrvBytes:  bytes,
		PrvPath:   "",
		EnableTls: false,
		TlsBytes:  nil,
		TlsPath:   "",
	}

	return r
}
