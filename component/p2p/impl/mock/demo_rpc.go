/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/3/31 4:56 下午
# @File : demo_rpc.go
# @Description :
# @Attention :
*/
package mock
//
// import (
// 	"github.com/golang/protobuf/proto"
// 	"github.com/hyperledger/fabric-droplib/base/debug"
// 	logplugin "github.com/hyperledger/fabric-droplib/base/log"
// 	logrusplugin "github.com/hyperledger/fabric-droplib/base/log/v2/logrus"
// 	"github.com/hyperledger/fabric-droplib/base/log/modules"
// 	services2 "github.com/hyperledger/fabric-droplib/base/services"
// 	"github.com/hyperledger/fabric-droplib/base/types"
// 	models2 "github.com/hyperledger/fabric-droplib/common/models"
// 	"github.com/hyperledger/fabric-droplib/component/base"
// 	"github.com/hyperledger/fabric-droplib/component/p2p/bo"
// 	p2p "github.com/hyperledger/fabric-droplib/component/p2p/cmd"
// 	streamimpl "github.com/hyperledger/fabric-droplib/component/p2p/impl"
// 	"github.com/hyperledger/fabric-droplib/component/p2p/impl/protocol"
// 	"github.com/hyperledger/fabric-droplib/component/p2p/models"
// 	"github.com/hyperledger/fabric-droplib/component/p2p/services"
// 	"github.com/hyperledger/fabric-droplib/protos"
// 	"github.com/hyperledger/fabric-droplib/protos/p2p/protos/mock"
// 	protocol2 "github.com/libp2p/go-libp2p-core/protocol"
// )
//
// const (
// 	PROTOCOL_DEMO = "/demo/1.0.0"
// )
//
// var (
// 	_ services.ProtocolFactory = (*factory)(nil)
// )
//
// var (
// 	_ IDemoLogicService            = (*DemoLogicServiceImpl)(nil)
// 	_ services.ILocalNetworkBundle = (*DemoLocalBundle)(nil)
// 	_ IDemoLocalLogicService       = (*DemoProtocolLogicServiceImpl)(nil)
// )
// var (
// 	_ services.IRemoteNetworkBundle = (*demoRemoteNetoworkBundle)(nil)
// 	_ IDemoRemoteProtocolService    = (*DemoRemoteProtocolService)(nil)
// )
//
// var (
// 	bindChannels []models.ChannelInfo
// 	channelId1   = types.ChannelID(100)
// 	channelId2   = types.ChannelID(200)
// )
//
// var DemoProtocolInfo = p2p.ProtocolProviderRegister{
// 	ProtocolId:      PROTOCOL_DEMO,
// 	MessageType:     demo_message_type,
// 	ProtocolFactory: &factory{},
// 	Channels:        bindChannels,
// }
//
// var demo_message_type = &protos.StreamCommonProtocolWrapper{}
// var messageTypes = map[types.ChannelID]services2.IChannelMessageWrapper{
// 	channelId1: &mock.DemoC1ChannelMessageWrapper{},
// 	channelId2: &mock.DemoC2ChannelMessageWrapper{},
// }
//
// func init() {
// 	c1Topics := []models.TopicInfo{{
// 		"c1_topic1",
// 	}, {"c1_topic2"}, {"c1_topic3"}}
// 	c1 := models.ChannelInfo{
// 		ChannelId:   channelId1,
// 		ChannelSize: 20,
// 		MessageType: &mock.DemoC1ChannelMessageWrapper{},
// 		Topics:      c1Topics,
// 	}
//
// 	c2Topics := []models.TopicInfo{{
// 		"c2_topic1",
// 	}, {"c2_topic2"}, {"c2_topic3"}}
// 	c2 := models.ChannelInfo{
// 		ChannelId:   channelId2,
// 		ChannelSize: 20,
// 		MessageType: &mock.DemoC2ChannelMessageWrapper{},
// 		Topics:      c2Topics,
// 	}
//
// 	bindChannels = append(bindChannels, c1, c2)
// 	DemoProtocolInfo.Channels = bindChannels
// }
//
// type IDemoLogicService interface {
// 	base.ILogicService
// }
//
// type DemoLocalBundle struct {
// 	*protocol.LocalCommonNetworkBundle
// }
//
// type IDemoLocalLogicService interface {
// 	services.IProtocolService
// }
// type DemoProtocolLogicServiceImpl struct {
// 	*streamimpl.DefaultLocalProtocolServiceImpl
// }
//
// func NewDemoProtocolLocalBundle(peerId types.NodeID,
// 	protocolId protocol2.ID,
// 	dataSupport services2.StreamDataSupport,
// 	dataWp services2.IProtocolWrapper,
// 	ins map[types.ChannelID]chan<- models2.Envelope, bundle services.ILocalNetworkBundle) *DemoProtocolLogicServiceImpl {
// 	r := &DemoProtocolLogicServiceImpl{
// 		DefaultLocalProtocolServiceImpl: nil,
// 	}
// 	r.DefaultLocalProtocolServiceImpl = streamimpl.NewDefaultLocalProtocolServiceImpl(peerId, protocolId, dataSupport, dataWp,
// 		ins, bundle)
// 	return r
// }
//
// func NewDemoLocalBundle() *DemoLocalBundle {
// 	r := &DemoLocalBundle{
// 	}
// 	resp := make(map[types.ChannelID]protocol.Response)
// 	resp[channelId1] = func() ([]byte, error) {
// 		logplugin.Info(debug.Yellow("resp回复数据 demo_protocol_channel1_message2"))
// 		protoResp := &mock.DemoC1Message2Proto{
// 			C12: "demo_protocol_channel1_message2",
// 		}
// 		return proto.Marshal(protoResp)
// 	}
// 	resp[channelId2] = func() ([]byte, error) {
// 		logplugin.Info(debug.Yellow("resp回复数据 demo_protocol_channel2_message2"))
// 		protoResp := &mock.DemoC2Message2Proto{
// 			C22: "demo_protocol_channel2_message2",
// 		}
// 		return proto.Marshal(protoResp)
// 	}
// 	r.LocalCommonNetworkBundle = protocol.NewLocalCommonNetworkBundle(resp, r)
//
// 	return r
// }
//
// func (l *DemoLocalBundle) AfterReceive() {
// 	logplugin.Info("after receive ")
// }
//
// type demoRemoteNetoworkBundle struct {
// 	*protocol.RemoteCommonNetworkBundle
// }
//
// type IDemoRemoteProtocolService interface {
// 	services.IRemoteProtocolService
// }
//
// func NewRemoteDemoProtocolBundle(peerId types.NodeID) *demoRemoteNetoworkBundle {
// 	r := &demoRemoteNetoworkBundle{}
// 	m := make(map[types.ChannelID]protocol.VerifyResponse)
// 	m[channelId1] = func(bytes []byte) (interface{}, error) {
// 		logplugin.Info(debug.Magenta("channel1 收到数据:" + string(bytes)))
// 		resp := &mock.DemoC1Message2Proto{}
// 		err := proto.Unmarshal(bytes, resp)
// 		if nil != err {
// 			return nil, err
// 		}
// 		return resp, nil
// 	}
// 	m[channelId2] = func(bytes []byte) (interface{}, error) {
// 		logplugin.Info(debug.Magenta("channel2 收到数据:" + string(bytes)))
// 		protoResp := &mock.DemoC2Message2Proto{}
// 		if err := proto.Unmarshal(bytes, protoResp); nil != err {
// 			return nil, err
// 		}
// 		return protoResp, nil
// 	}
// 	rl := logrusplugin.NewLogrusLogger(modules.NewModule("REMOTE_DEMO_BUNDLE", 1))
// 	r.RemoteCommonNetworkBundle = protocol.NewRemoteCommonNetworkBundle(m, rl, peerId, PROTOCOL_DEMO, r)
//
// 	return r
// }
//
// type DemoRemoteProtocolService struct {
// 	*streamimpl.DefaultRemoteProtocolServiceImpl
// 	// Channel1 *models2.Channel
// 	// Channel2 *models2.Channel
// }
//
// func NewDemoRemoteProtocolService(peerId types.NodeID,
// 	protocolId protocol2.ID,
// 	peerStatus func() bool,
// 	streamInfo services2.IStream,
// 	dataSupport services2.StreamDataSupport,
// 	dataWp services2.IProtocolWrapper,
// 	helper services.IRemoteNetworkBundle, ) *DemoRemoteProtocolService {
// 	r := &DemoRemoteProtocolService{
// 		DefaultRemoteProtocolServiceImpl: nil,
// 	}
// 	r.DefaultRemoteProtocolServiceImpl = streamimpl.NewDefaultRemoteProtocolServiceImpl(
// 		peerId, protocolId, peerStatus, streamInfo, dataSupport, dataWp, helper,
// 	)
// 	return r
// }
//
// type factory struct {
// }
//
// func (f *factory) Local(req bo.LocalProtocolCreateBO) (services.IProtocolService, error) {
// 	return NewDemoProtocolLocalBundle(req.NodeID,
// 		req.ProtocolId, req.DataSupport, req.ProtocolWrapper, req.Channels, NewDemoLocalBundle(),
// 	), nil
// }
//
// func (f *factory) BundleLocal(req bo.LocalBundleCreateBO) (services.ILocalNetworkBundle, error) {
// 	return NewDemoLocalBundle(), nil
// }
//
// func (f *factory) Remote(req bo.RemoteProtocolCreateBO) (services.IRemoteProtocolService, error) {
// 	return NewDemoRemoteProtocolService(req.NodeID, req.ProtocolId, req.Status, req.Stream, req.DataSupport, req.MessageType, NewRemoteDemoProtocolBundle(req.NodeID)), nil
// }
//
// func (f *factory) BundleRemote(req bo.RemoteBundleCreateBO) (services.IRemoteNetworkBundle, error) {
// 	return NewRemoteDemoProtocolBundle(req.NodeID), nil
// }
//
// const (
// 	DEMO_SERVICE_ID types.ServiceID = "DEMO_SERVICE_ID"
// )
//
// type Option func(e *models2.ChannelEnvelope)
//
// func UseCache(e *models2.ChannelEnvelope) {
// 	e.AcquireStreamFromCache = true
// }
// func LastSendWithOutResponse(e *models2.ChannelEnvelope) {
// 	e.StreamFlag = protos.StreamFlag_CLOSE_IMMEDIATELY
// }
//
// type DemoLogicServiceImpl struct {
// 	*streamimpl.BaseLogicServiceImpl
// 	Channel1 *models2.Channel
// 	Channel2 *models2.Channel
// }
//
// func NewDemoLogicServiceImpl() *DemoLogicServiceImpl {
// 	r := &DemoLogicServiceImpl{
// 		BaseLogicServiceImpl: nil,
// 	}
// 	r.BaseLogicServiceImpl = streamimpl.NewBaseLogicServiceImpl(modules.NewModule("DEMO_LOGIC", 1), r)
// 	return r
// }
//
// func (d *DemoLogicServiceImpl) ChoseInterestEvents(bus base.IP2PEventBusComponent) {
// 	return
// }
//
// func (d *DemoLogicServiceImpl) SendTo(cid types.ChannelID, remoteId types.NodeID, option ...Option) {
// 	if cid == channelId1 {
// 		e := &models2.ChannelEnvelope{
// 			To: remoteId,
// 			Message: &mock.DemoC1Message1Proto{
// 				C11: "这是c11",
// 			},
// 			StreamFlag:             protos.StreamFlag_RESPONSE_AND_CLOSE,
// 			AcquireStreamFromCache: false,
// 			ChannelID:              channelId1,
// 		}
// 		for _, opt := range option {
// 			opt(e)
// 		}
// 		d.Channel1.Out <- *e
// 	} else {
// 		d.Channel2.Out <- models2.ChannelEnvelope{
// 			To: types.NodeID(remoteId),
// 			Message: &mock.DemoC2Message2Proto{
// 				C22: "这是c22",
// 			},
// 			StreamFlag:             protos.StreamFlag_RESPONSE_AND_CLOSE,
// 			AcquireStreamFromCache: false,
// 			ChannelID:              channelId2,
// 		}
// 	}
// }
//
// func (d *DemoLogicServiceImpl) OnStart() error {
// 	go d.receive()
// 	return nil
// }
// func (d *DemoLogicServiceImpl) receive() {
// 	for {
// 		select {
// 		case msg := <-d.Channel1.In:
// 			m, _, err := d.UnwrapFromBytes(channelId1, msg.Message)
// 			if nil != err {
// 				d.Logger.Error("解码失败:" + err.Error())
// 				continue
// 			}
// 			d.Logger.Info("channel1 收到msg:" + m.String())
// 		case msg := <-d.Channel2.In:
// 			m, _, err := d.UnwrapFromBytes(channelId1, msg.Message)
// 			if nil != err {
// 				d.Logger.Error("解码失败:" + err.Error())
// 				continue
// 			}
// 			d.Logger.Info("channel1 收到msg:" + m.String())
// 		}
// 	}
// }
// func (d *DemoLogicServiceImpl) GetServiceId() types.ServiceID {
// 	return DEMO_SERVICE_ID
// }
//
// func (d *DemoLogicServiceImpl) ChoseInterestProtocolsOrTopics(pProtocol *models2.P2PProtocol) error {
// 	info := pProtocol.GetProtocolByProtocolID(PROTOCOL_DEMO)
// 	d.Channel1 = info.GetChannelById(channelId1)
// 	d.Channel2 = info.GetChannelById(channelId2)
// 	return nil
// }
