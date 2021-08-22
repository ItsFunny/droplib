/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/3/28 3:29 下午
# @File : p2p_test.go.go
# @Description :
# @Attention :
*/
package p2p

import (
	"fmt"
	"github.com/hyperledger/fabric-droplib/base/debug"
	logplugin "github.com/hyperledger/fabric-droplib/base/log"
	"github.com/hyperledger/fabric-droplib/base/services"
	"github.com/hyperledger/fabric-droplib/component/common/blockchain/config"
	"github.com/hyperledger/fabric-droplib/component/p2p/models"
	p2ptypes "github.com/hyperledger/fabric-droplib/component/p2p/types"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/peerstore"
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"
)

func waitFor() {
	fmt.Println("exit")
	sigCh := make(chan os.Signal)
	signal.Notify(sigCh, syscall.SIGKILL, syscall.SIGINT)
	<-sigCh
}

var (
	tlsPrv1 = `
-----BEGIN PRIVATE KEY-----
MIGHAgEAMBMGByqGSM49AgEGCCqGSM49AwEHBG0wawIBAQQgV3eorGZxYm2EW9/l
iS7qKbhv9EmsxeR3mo3iHESQa2ihRANCAAS0FCJsx+enLGLXtp5x5nr7sj/yR/ev
2dn5zW66JaOQ/NSlk1nZREVpsBMJGWa20icV2yVbwQqLHxOCTtlPy0wV
-----END PRIVATE KEY-----
`
	tlsPrv2 = `
-----BEGIN PRIVATE KEY-----
MIGHAgEAMBMGByqGSM49AgEGCCqGSM49AwEHBG0wawIBAQQgynqH3Nhs1cOtb3iB
/5O2aGyYsZBuIdQrIa++S08UJ76hRANCAASN0WcRCpVZVtucKo6KGT76yc/yQu8X
8FG/1JXj7OgKlJJp8viVDHRfvANWgOmZ9U4so8Sicihwe2CplkiYZFGD
-----END PRIVATE KEY-----
`
)

func TestMultiP2PTls(t *testing.T) {
	// config.Tls([]byte(tlsPrv1))
	cfg := config.NewDefaultP2PConfiguration()
	// libp2p.Security(libp2ptls.ID, libp2ptls.New)
	p1 := makeRandP2P(10011, cfg)
	// config.Tls([]byte(tlsPrv2)
	cfg = config.NewDefaultP2PConfiguration()
	p2 := makeRandP2P(10012, cfg)

	p2Addr := p2.GetExternalAddress()
	id := p2.router.GetNode().ID()
	logplugin.Info("手动添加的节点id为:" + id.Pretty())

	m := models.MemberChanged{
		PeerId:             id,
		ExternMultiAddress: p2Addr,
		Type:               p2ptypes.MEMBER_CHANGED_ADD,
		Ttl:                peerstore.PermanentAddrTTL,
	}
	p1.router.OutMemberChanged(m)

	fmt.Println(p1)

	waitFor()
}
func TestWithDisablePingpong(t *testing.T) {
	cfg := config.NewDefaultP2PConfiguration(config.DisablePingPong())
	p1 := makeRandP2P(10001, cfg)
	debug.RED = false
	p2 := makeRandP2P(10002, cfg)

	p1Addr := p1.GetExternalAddress()
	logplugin.Info("手动添加的节点id为:" + p1Addr)
	m := models.MemberChanged{
		PeerId:             p1.router.GetNode().ID(),
		ExternMultiAddress: p1Addr,
		Type:               p2ptypes.MEMBER_CHANGED_ADD,
		Ttl:                peerstore.PermanentAddrTTL,
	}
	p2.router.OutMemberChanged(m)

	time.Sleep(time.Second * 180)
}
func TestMultiP2PEnablepingpong(t *testing.T) {
	cfg := config.NewDefaultP2PConfiguration()
	p1 := makeRandP2P(10001, cfg)
	p2 := makeRandP2P(10002, cfg)

	p1Addr := p1.GetExternalAddress()
	logplugin.Info("手动添加的节点id为:" + p1Addr)
	m := models.MemberChanged{
		PeerId:             p1.router.GetNode().ID(),
		ExternMultiAddress: p1Addr,
		Type:               p2ptypes.MEMBER_CHANGED_ADD,
		Ttl:                peerstore.PermanentAddrTTL,
	}
	p2.router.OutMemberChanged(m)

	time.Sleep(time.Second * 180)
}

func Test_P2p(t *testing.T) {
	cfg := config.NewDefaultP2PConfiguration()
	p := NewP2P(cfg)
	p.Start(services.SYNC_START_WAIT_READY)
	time.Sleep(time.Second * 20)
}

func makeRandP2P(port int, cfg *config.P2PConfiguration, opts ...libp2p.Option) *P2P {
	cfg.IdentityProperty.Port = port
	p := NewP2P(cfg, opts...)
	p.Start(services.ASYNC_START)
	// if err := p.Start(services.SYNC_START); nil != err {
	// 	panic(err)
	// }
	// p.ReadyIfPanic(services.READY_UNTIL_START)
	return p
}

func TestMultiProtocol(t *testing.T) {

}
