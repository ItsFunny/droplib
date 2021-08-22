/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/3/22 6:55 下午
# @File : factory_impl.go
# @Description :
# @Attention :
*/
package streamimpl

import (
	"context"
	"crypto/rand"
	"crypto/x509"
	"fmt"
	"github.com/hyperledger/fabric-droplib/base/services/impl"
	libutils "github.com/hyperledger/fabric-droplib/base/utils"
	"github.com/hyperledger/fabric-droplib/component/common/blockchain/config"
	"github.com/hyperledger/fabric-droplib/component/p2p/services"
	"github.com/hyperledger/fabric-droplib/component/p2p/utils"
	"github.com/hyperledger/fabric-droplib/component/p2p/wrapper"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	libp2ptls "github.com/libp2p/go-libp2p-tls"
	"github.com/multiformats/go-multiaddr"
	"io/ioutil"
	"os"
)

const (
	CLOSE_INDEX_READ  = 1
	CLOSE_INDEX_WRITE = 2
	CLOSE_INDEX_ALL   = 1 | 2
)

var (
	_ services.INodeFactory = (*DefaultNodeFactory)(nil)
)

type DefaultNodeFactory struct {
}

func NewDefaultNodeFactory() *DefaultNodeFactory {
	f := &DefaultNodeFactory{
	}

	return f
}

func (d *DefaultNodeFactory) CreateNode(ctx context.Context, cfg *config.P2PConfiguration, opts ...libp2p.Option) services.INode {

	n := &LibP2PNode{ctx: ctx}
	n.Host = d.newHost(ctx, cfg, opts...)
	addr := cfg.IdentityProperty.Address
	port := cfg.IdentityProperty.Port
	n.Addr = fmt.Sprintf("/ip4/%s/tcp/%d/p2p/%s", addr, port, n.ID().Pretty())
	return n
}

func (d *DefaultNodeFactory) newHost(ctx context.Context, cfg *config.P2PConfiguration, opts ...libp2p.Option) host.Host {
	// 默认只支持tcp
	ip := "0.0.0.0"
	if len(cfg.IdentityProperty.Address) != 0 {
		ip = cfg.IdentityProperty.Address
	}

	addr, err := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/%s/tcp/%d", ip, cfg.IdentityProperty.Port))
	if nil != err {
		panic("setup failed:" + err.Error())
	}
	opts = append(opts, libp2p.ListenAddrs(addr))
	prv, err := d.loadPrvKey(cfg.IdentityProperty.PrvBytes, cfg.IdentityProperty.PrvPath)
	if nil != err {
		panic("加载证书失败了啊:" + err.Error())
	}
	opts = append(opts, libp2p.Identity(prv))
	if cfg.IdentityProperty.EnableTls {
		key, err := d.loadPrvKey(cfg.IdentityProperty.TlsBytes, cfg.IdentityProperty.TlsPath)
		libutils.PanicIfError("加载tls失败", err)
		transport, err := libp2ptls.New(key)
		libutils.PanicIfError("tlsIdentity", err)
		opts = append(opts, libp2p.Security(libp2ptls.ID, transport))
	}

	// opts=append(opts,libp2p.DefaultTransports)
	h, err := libp2p.New(
		ctx, opts...,
	)
	libutils.PanicIfError("创建节点", err)

	return h
}

func (d *DefaultNodeFactory) registerProtocol(ctx context.Context, boots string) {

}

func (d *DefaultNodeFactory) loadPrvKey(prvBytes []byte, path string) (crypto.PrivKey, error) {
	if len(prvBytes) == 0 {
		if !Exists(path) {
			c1K, _, _ := crypto.GenerateKeyPairWithReader(crypto.ECDSA, 256, rand.Reader)
			return c1K, nil
		}
		fbs, err := ioutil.ReadFile(path)
		if nil != err {
			return nil, err
		}
		prvBytes = fbs
	}
	// return crypto.UnmarshalECDSAPrivateKey(prvBytes)
	//
	k, err := utils.ECDSA_KEY_IMPORTER(prvBytes, nil)
	if nil != err {
		return nil, err
	}
	key, err := x509.MarshalECPrivateKey(k)
	if nil != err {
		return nil, err
	}
	return crypto.UnmarshalECDSAPrivateKey(key)
	// return &models.EcdsaPrvK{k}, nil
}

func Exists(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}

type DefaultStreamHandlerFactory struct {
	parentCtx context.Context
}

func NewDefaultStreamHandlerFactory(ctx context.Context) *DefaultStreamHandlerFactory {
	r := &DefaultStreamHandlerFactory{
		parentCtx: ctx,
	}
	return r
}

func (d *DefaultStreamHandlerFactory) createP2PStreamHandler(h host.Host, wrapper wrapper.ProtocolWrapper) (*services.P2PStreamHandlerWrapper, error) {
	ctx, cancelFunc := context.WithCancel(d.parentCtx)
	stream, err := h.NewStream(ctx, wrapper.PeerId, wrapper.ProtoCol)
	if nil != err {
		return nil, err
	}
	res := &services.P2PStreamHandlerWrapper{
		Ctx:       ctx,
		Cancel:    cancelFunc,
		Stream:    stream,
		Handler:   d.WrapHandler(wrapper),
		ProtoCol:  wrapper.ProtoCol,
		FifoQueue: impl.NewFIFOQueue(wrapper.QueueSize),
	}

	return res, nil
}

type StreamDecorator func(stream network.Stream)

var (
	CLOSE_READ_DECORATOR = func(stream network.Stream) {
		stream.CloseRead()
	}
	CLOSE_WRITE_DECORATOR = func(stream network.Stream) {
		stream.CloseWrite()
	}
	CLOSE_ALL_DECORATOR = func(s network.Stream) {
		s.Close()
	}
)

func (this *DefaultStreamHandlerFactory) WrapHandler(wp wrapper.ProtocolWrapper) network.StreamHandler {
	if wp.CloseQuickly {
		var deferF StreamDecorator
		if wp.CloseIndex&CLOSE_INDEX_READ > 0 {
			deferF = CLOSE_READ_DECORATOR
		} else if wp.CloseIndex&CLOSE_INDEX_WRITE > 0 {
			deferF = CLOSE_WRITE_DECORATOR
		} else if wp.CloseIndex&CLOSE_INDEX_ALL >= CLOSE_INDEX_ALL {
			deferF = CLOSE_ALL_DECORATOR
		}
		return func(stream network.Stream) {
			defer deferF(stream)
			wp.Handler(stream)
		}
	}
	// TODO

	return wp.Handler
}
