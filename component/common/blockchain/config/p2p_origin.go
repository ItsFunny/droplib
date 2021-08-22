// /*
// # -*- coding: utf-8 -*-
// # @Author : joker
// # @Time : 2021/3/15 12:43 下午
// # @File : p2p.go
// # @Description :	原先的p2p 配置
// # @Attention :
// */
package config
//
// import "time"
//
// type P2PConfiguration struct {
// 	RootDir string `mapstructure:"home"`
//
// 	// Address to listen for incoming connections
// 	ListenAddress string `mapstructure:"laddr"`
//
// 	// Address to advertise to peers for them to dial
// 	ExternalAddress string `mapstructure:"external-address"`
//
// 	// Comma separated list of seed nodes to connect to
// 	// We only use these if we can’t connect to peers in the addrbook
// 	Seeds string `mapstructure:"seeds"`
//
// 	// Comma separated list of nodes to keep persistent connections to
// 	PersistentPeers string `mapstructure:"persistent-peers"`
//
// 	// UPNP port forwarding
// 	UPNP bool `mapstructure:"upnp"`
//
// 	// Path to address book
// 	AddrBook string `mapstructure:"addr-book-file"`
//
// 	// Set true for strict address routability rules
// 	// Set false for private or local networks
// 	AddrBookStrict bool `mapstructure:"addr-book-strict"`
//
// 	// Maximum number of inbound peers
// 	MaxNumInboundPeers int `mapstructure:"max-num-inbound-peers"`
//
// 	// Maximum number of outbound peers to connect to, excluding persistent peers
// 	MaxNumOutboundPeers int `mapstructure:"max-num-outbound-peers"`
//
// 	// List of node IDs, to which a connection will be (re)established ignoring any existing limits
// 	UnconditionalPeerIDs string `mapstructure:"unconditional-peer-ids"`
//
// 	// Maximum pause when redialing a persistent peer (if zero, exponential backoff is used)
// 	PersistentPeersMaxDialPeriod time.Duration `mapstructure:"persistent-peers-max-dial-period"`
//
// 	// Time to wait before flushing messages out on the connection
// 	FlushThrottleTimeout time.Duration `mapstructure:"flush-throttle-timeout"`
//
// 	// Maximum size of a message packet payload, in bytes
// 	MaxPacketMsgPayloadSize int `mapstructure:"max-packet-msg-payload-size"`
//
// 	// Rate at which packets can be sent, in bytes/second
// 	SendRate int64 `mapstructure:"send-rate"`
//
// 	// Rate at which packets can be received, in bytes/second
// 	RecvRate int64 `mapstructure:"recv-rate"`
//
// 	// Set true to enable the peer-exchange reactor
// 	PexReactor bool `mapstructure:"pex"`
//
// 	// Seed mode, in which node constantly crawls the network and looks for
// 	// peers. If another node asks it for addresses, it responds and disconnects.
// 	//
// 	// Does not work if the peer-exchange reactor is disabled.
// 	SeedMode bool `mapstructure:"seed-mode"`
//
// 	// Comma separated list of peer IDs to keep private (will not be gossiped to
// 	// other peers)
// 	PrivatePeerIDs string `mapstructure:"private-peer-ids"`
//
// 	// Toggle to disable guard against peers connecting from the same ip.
// 	AllowDuplicateIP bool `mapstructure:"allow-duplicate-ip"`
//
// 	// Peer connection configuration.
// 	HandshakeTimeout time.Duration `mapstructure:"handshake-timeout"`
// 	DialTimeout      time.Duration `mapstructure:"dial-timeout"`
//
// 	// Testing params.
// 	// Force dial to fail
// 	TestDialFail bool `mapstructure:"test-dial-fail"`
// }
