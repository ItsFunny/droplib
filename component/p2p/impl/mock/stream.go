/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/3/26 9:45 上午
# @File : stream.go
# @Description :
# @Attention :
*/
package mock
//
// import (
// 	"github.com/libp2p/go-libp2p-core/network"
// 	"github.com/libp2p/go-libp2p-core/protocol"
// 	"sync"
// 	"time"
// )
//
// type MockStream struct {
// 	sync.RWMutex
// 	buf []byte
// }
//
// func (m *MockStream) Read(p []byte) (n int, err error) {
// 	m.Lock()
// 	defer m.Unlock()
// 	l := copy(p, m.buf)
// 	m.buf = m.buf[l:]
// 	return l, nil
// }
//
// func (m *MockStream) Write(p []byte) (n int, err error) {
// 	m.Lock()
// 	defer m.Unlock()
// 	m.buf = append(m.buf, p...)
// 	return len(p), nil
// }
//
// func (m *MockStream) Close() error {
// 	panic("implement me")
// }
//
// func (m *MockStream) CloseWrite() error {
// 	panic("implement me")
// }
//
// func (m *MockStream) CloseRead() error {
// 	panic("implement me")
// }
//
// func (m *MockStream) Reset() error {
// 	panic("implement me")
// }
//
// func (m *MockStream) SetDeadline(time time.Time) error {
// 	panic("implement me")
// }
//
// func (m *MockStream) SetReadDeadline(time time.Time) error {
// 	panic("implement me")
// }
//
// func (m *MockStream) SetWriteDeadline(time time.Time) error {
// 	panic("implement me")
// }
//
// func (m *MockStream) ID() string {
// 	panic("implement me")
// }
//
// func (m *MockStream) Protocol() protocol.ID {
// 	panic("implement me")
// }
//
// func (m *MockStream) SetProtocol(id protocol.ID) {
// 	panic("implement me")
// }
//
// func (m *MockStream) Stat() network.Stat {
// 	panic("implement me")
// }
//
// func (m *MockStream) Conn() network.Conn {
// 	panic("implement me")
// }
