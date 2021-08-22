package listener

import (
	"github.com/hyperledger/fabric-droplib/base/log/modules"
	models2 "github.com/hyperledger/fabric-droplib/base/services/models"
	"github.com/hyperledger/fabric-droplib/component/base"
)

var (
	_ IListenerComponent = (*listenerComponent)(nil)
)

type listenerComponent struct {
	*base.BaseComponent
	clientId string
	pubsub   *PubSub
	// cpubsub  services.ICommonEventBusComponent
}

func DefaultNewListenerComponent(opts ...Opt) *listenerComponent {
	return NewListenerComponent(256)
}
func NewListenerComponent(cap int, opts ...Opt) *listenerComponent {
	r := &listenerComponent{}
	// r.cpubsub = impl.NewCommonEventBusComponentImpl(impl.BufferCapacity(cap))
	r.pubsub = New(cap)
	r.BaseComponent = base.NewBaseComponent(modules.NewModule("LISTENER", 1), r)
	for _, opt := range opts {
		opt(r)
	}
	if r.clientId == "" {
		r.clientId = "default"
	}
	return r
}
func (l *listenerComponent) OnStart(ctx *models2.StartCTX) error {
	return nil
}
func (l *listenerComponent) RegisterListener(topic ...string) <-chan interface{} {
	// l.Logger.Info("注册", "ids", strings.Join(topic, ","))
	return l.pubsub.SubOnce(topic...)
}

func (l *listenerComponent) NotifyListener(data interface{}, listenerIds ...string) {
	// l.Logger.Info("通知", "ids", strings.Join(listenerIds, ","))
	l.pubsub.Pub(data, listenerIds...)
}
func (l *listenerComponent) OnStop(c *models2.StopCTX) {
	l.pubsub.Stop()
}
