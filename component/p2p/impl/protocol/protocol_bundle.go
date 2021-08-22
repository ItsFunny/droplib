/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/3/31 8:43 下午
# @File : local_protocol.go
# @Description :
# @Attention :
*/
package protocol

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"github.com/hyperledger/fabric-droplib/base/debug"
	logplugin "github.com/hyperledger/fabric-droplib/base/log"
	v2 "github.com/hyperledger/fabric-droplib/base/log/v2"
	services2 "github.com/hyperledger/fabric-droplib/base/services"
	"github.com/hyperledger/fabric-droplib/base/types"
	"github.com/hyperledger/fabric-droplib/component/p2p/services"
	"github.com/libp2p/go-libp2p-core/protocol"
	"strconv"
	"time"
)

var (
	_ services.ILocalNetworkBundle  = (*LocalCommonNetworkBundle)(nil)
	_ services.IRemoteNetworkBundle = (*RemoteCommonNetworkBundle)(nil)
)

type Response func() ([]byte, error)

type LocalCommonNetworkBundle struct {
	responses map[types.ChannelID]Response
	impl      services.ILocalNetworkBundle
}

func NewLocalCommonNetworkBundle(responses map[types.ChannelID]Response, impl services.ILocalNetworkBundle) *LocalCommonNetworkBundle {
	r := &LocalCommonNetworkBundle{
		responses: responses,
		impl:      impl,
	}
	return r
}

func (l *LocalCommonNetworkBundle) Response(cid types.ChannelID, stream services2.IStream, support services2.StreamDataSupport) error {
	f, exist := l.responses[cid]
	if !exist {
		return errors.New("channelid为:" + strconv.Itoa(int(cid)) + ",对应的response不存在")
	}
	data, err := f()
	if nil != err {
		return errors.New("生成响应数据失败:" + err.Error())
	}
	data, err = support.Pack(data)
	if nil != err {
		return err
	}
	bufWriter := bufio.NewWriter(stream)
	// FIXME 不需要
	write, err := bufWriter.Write(data)
	if nil != err {
		return err
	} else if write != len(data) {
		return errors.New(fmt.Sprintf("写入数据不全,写入了%d,实际需要为:%d", write, len(data)))
	}
	if err := bufWriter.Flush(); nil != err {
		return err
	}
	return nil
}

// noop
func (l *LocalCommonNetworkBundle) AfterReceive() {
}

type VerifyResponse func([]byte) (interface{}, error)

type RemoteCommonNetworkBundle struct {
	Logger     v2.Logger
	peerId     types.NodeID
	protocolId protocol.ID

	verifiers map[types.ChannelID]VerifyResponse
	impl      services.IRemoteNetworkBundle
}

func NewRemoteCommonNetworkBundle(ver map[types.ChannelID]VerifyResponse, parentLog v2.Logger, peerId types.NodeID, protocolId protocol.ID, rr services.IRemoteNetworkBundle) *RemoteCommonNetworkBundle {
	// FIXME
	r := &RemoteCommonNetworkBundle{
		Logger:     parentLog,
		peerId:     peerId,
		protocolId: protocolId,
		verifiers:  ver,
		impl:       rr,
	}
	return r
}

func (r *RemoteCommonNetworkBundle) WaitForResponse(cid types.ChannelID, stream services2.IStream, support services2.StreamDataSupport) error {
	logplugin.Info(debug.Magenta("等待接收响应数据,channelId为:" + strconv.Itoa(int(cid))))
	ver, exist := r.verifiers[cid]
	if !exist {
		return errors.New("cid为:" + strconv.Itoa(int(cid)) + ",对应的verifier不存在")
	}
	ctx, _ := context.WithTimeout(context.TODO(), time.Second*30)

	pack, err := support.UnPack(ctx, stream, nil)
	if nil != err {
		return err
	}
	data, err := ver(pack)
	if nil != err {
		return err
	}

	return r.OnAfterWaitResponse(cid, data)
}

func (r *RemoteCommonNetworkBundle) OnAfterWaitResponse(cid types.ChannelID, responseData interface{}) error {
	return nil
}

func (p *RemoteCommonNetworkBundle) AfterSendStream() {

}

func (p *RemoteCommonNetworkBundle) WaitBy() {

}

func (p *RemoteCommonNetworkBundle) StreamPolicy(registry services.IStreamRegistry, stream services2.IStream, cache bool) (services2.IStream, error) {
	if nil != stream {
		if p.StreamHealth(stream) {
			return stream, nil
		}
		stream.Reset()
	}
	v, err := registry.AcquireNewStream(p.peerId, p.protocolId, cache)
	if nil != err {
		return nil, err
	}
	return v, nil
}

func (p *RemoteCommonNetworkBundle) StreamHealth(s services2.IStream) bool {
	return true
}
func (p *RemoteCommonNetworkBundle) SendWithStream(stream services2.IStream, bytes []byte) error {
	rw := bufio.NewWriter(stream)
	if _, err := rw.Write(bytes); nil != err {
		return err
	}
	if err := rw.Flush(); nil != err {
		p.Logger.Error("flush失败,peerId:%s,err:%s", p.peerId.Pretty(), err.Error())
		return err
	}
	return nil
}
