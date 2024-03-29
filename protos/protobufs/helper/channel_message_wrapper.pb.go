// Code generated by protoc-gen-go. DO NOT EDIT.
// source: channel_message_wrapper.proto

package helper

import (
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

// vote
type ConsensusVoteChannelMessageWrapper struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ConsensusVoteChannelMessageWrapper) Reset()         { *m = ConsensusVoteChannelMessageWrapper{} }
func (m *ConsensusVoteChannelMessageWrapper) String() string { return proto.CompactTextString(m) }
func (*ConsensusVoteChannelMessageWrapper) ProtoMessage()    {}
func (*ConsensusVoteChannelMessageWrapper) Descriptor() ([]byte, []int) {
	return fileDescriptor_f0505cdede181be7, []int{0}
}

func (m *ConsensusVoteChannelMessageWrapper) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ConsensusVoteChannelMessageWrapper.Unmarshal(m, b)
}
func (m *ConsensusVoteChannelMessageWrapper) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ConsensusVoteChannelMessageWrapper.Marshal(b, m, deterministic)
}
func (m *ConsensusVoteChannelMessageWrapper) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ConsensusVoteChannelMessageWrapper.Merge(m, src)
}
func (m *ConsensusVoteChannelMessageWrapper) XXX_Size() int {
	return xxx_messageInfo_ConsensusVoteChannelMessageWrapper.Size(m)
}
func (m *ConsensusVoteChannelMessageWrapper) XXX_DiscardUnknown() {
	xxx_messageInfo_ConsensusVoteChannelMessageWrapper.DiscardUnknown(m)
}

var xxx_messageInfo_ConsensusVoteChannelMessageWrapper proto.InternalMessageInfo

// state
type ConsensusStateChannelMessageWrapper struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ConsensusStateChannelMessageWrapper) Reset()         { *m = ConsensusStateChannelMessageWrapper{} }
func (m *ConsensusStateChannelMessageWrapper) String() string { return proto.CompactTextString(m) }
func (*ConsensusStateChannelMessageWrapper) ProtoMessage()    {}
func (*ConsensusStateChannelMessageWrapper) Descriptor() ([]byte, []int) {
	return fileDescriptor_f0505cdede181be7, []int{1}
}

func (m *ConsensusStateChannelMessageWrapper) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ConsensusStateChannelMessageWrapper.Unmarshal(m, b)
}
func (m *ConsensusStateChannelMessageWrapper) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ConsensusStateChannelMessageWrapper.Marshal(b, m, deterministic)
}
func (m *ConsensusStateChannelMessageWrapper) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ConsensusStateChannelMessageWrapper.Merge(m, src)
}
func (m *ConsensusStateChannelMessageWrapper) XXX_Size() int {
	return xxx_messageInfo_ConsensusStateChannelMessageWrapper.Size(m)
}
func (m *ConsensusStateChannelMessageWrapper) XXX_DiscardUnknown() {
	xxx_messageInfo_ConsensusStateChannelMessageWrapper.DiscardUnknown(m)
}

var xxx_messageInfo_ConsensusStateChannelMessageWrapper proto.InternalMessageInfo

// data
type ConsensusDataChannelMessageWrapper struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ConsensusDataChannelMessageWrapper) Reset()         { *m = ConsensusDataChannelMessageWrapper{} }
func (m *ConsensusDataChannelMessageWrapper) String() string { return proto.CompactTextString(m) }
func (*ConsensusDataChannelMessageWrapper) ProtoMessage()    {}
func (*ConsensusDataChannelMessageWrapper) Descriptor() ([]byte, []int) {
	return fileDescriptor_f0505cdede181be7, []int{2}
}

func (m *ConsensusDataChannelMessageWrapper) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ConsensusDataChannelMessageWrapper.Unmarshal(m, b)
}
func (m *ConsensusDataChannelMessageWrapper) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ConsensusDataChannelMessageWrapper.Marshal(b, m, deterministic)
}
func (m *ConsensusDataChannelMessageWrapper) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ConsensusDataChannelMessageWrapper.Merge(m, src)
}
func (m *ConsensusDataChannelMessageWrapper) XXX_Size() int {
	return xxx_messageInfo_ConsensusDataChannelMessageWrapper.Size(m)
}
func (m *ConsensusDataChannelMessageWrapper) XXX_DiscardUnknown() {
	xxx_messageInfo_ConsensusDataChannelMessageWrapper.DiscardUnknown(m)
}

var xxx_messageInfo_ConsensusDataChannelMessageWrapper proto.InternalMessageInfo

// VoteSetBits
type ConsensusVoteSetBitsChannelMessageWrapper struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ConsensusVoteSetBitsChannelMessageWrapper) Reset() {
	*m = ConsensusVoteSetBitsChannelMessageWrapper{}
}
func (m *ConsensusVoteSetBitsChannelMessageWrapper) String() string {
	return proto.CompactTextString(m)
}
func (*ConsensusVoteSetBitsChannelMessageWrapper) ProtoMessage() {}
func (*ConsensusVoteSetBitsChannelMessageWrapper) Descriptor() ([]byte, []int) {
	return fileDescriptor_f0505cdede181be7, []int{3}
}

func (m *ConsensusVoteSetBitsChannelMessageWrapper) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ConsensusVoteSetBitsChannelMessageWrapper.Unmarshal(m, b)
}
func (m *ConsensusVoteSetBitsChannelMessageWrapper) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ConsensusVoteSetBitsChannelMessageWrapper.Marshal(b, m, deterministic)
}
func (m *ConsensusVoteSetBitsChannelMessageWrapper) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ConsensusVoteSetBitsChannelMessageWrapper.Merge(m, src)
}
func (m *ConsensusVoteSetBitsChannelMessageWrapper) XXX_Size() int {
	return xxx_messageInfo_ConsensusVoteSetBitsChannelMessageWrapper.Size(m)
}
func (m *ConsensusVoteSetBitsChannelMessageWrapper) XXX_DiscardUnknown() {
	xxx_messageInfo_ConsensusVoteSetBitsChannelMessageWrapper.DiscardUnknown(m)
}

var xxx_messageInfo_ConsensusVoteSetBitsChannelMessageWrapper proto.InternalMessageInfo

func init() {
	proto.RegisterType((*ConsensusVoteChannelMessageWrapper)(nil), "tendermint.helper.ConsensusVoteChannelMessageWrapper")
	proto.RegisterType((*ConsensusStateChannelMessageWrapper)(nil), "tendermint.helper.ConsensusStateChannelMessageWrapper")
	proto.RegisterType((*ConsensusDataChannelMessageWrapper)(nil), "tendermint.helper.ConsensusDataChannelMessageWrapper")
	proto.RegisterType((*ConsensusVoteSetBitsChannelMessageWrapper)(nil), "tendermint.helper.ConsensusVoteSetBitsChannelMessageWrapper")
}

func init() { proto.RegisterFile("channel_message_wrapper.proto", fileDescriptor_f0505cdede181be7) }

var fileDescriptor_f0505cdede181be7 = []byte{
	// 187 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x74, 0x8f, 0xb1, 0xcb, 0xc2, 0x30,
	0x10, 0xc5, 0xb7, 0x6f, 0xc8, 0xf6, 0xb9, 0xbb, 0x54, 0x1d, 0x44, 0x6c, 0x06, 0x67, 0x11, 0x5a,
	0x57, 0xa7, 0x82, 0x82, 0x4b, 0x49, 0xda, 0x6b, 0x13, 0x48, 0x93, 0x70, 0x77, 0x45, 0xfc, 0xef,
	0x85, 0x06, 0x44, 0xa1, 0x4e, 0x07, 0xc7, 0xef, 0xbd, 0xc7, 0x4f, 0x2c, 0x1b, 0xa3, 0xbc, 0x07,
	0x57, 0x0f, 0x40, 0xa4, 0x7a, 0xa8, 0x1f, 0xa8, 0x62, 0x04, 0xcc, 0x23, 0x06, 0x0e, 0x8b, 0x7f,
	0x06, 0xdf, 0x02, 0x0e, 0xd6, 0x73, 0x6e, 0xc0, 0x45, 0xc0, 0x6c, 0x2d, 0xb2, 0x32, 0x78, 0x02,
	0x4f, 0x23, 0x5d, 0x03, 0x43, 0x99, 0x0a, 0x2e, 0x29, 0x7f, 0x4b, 0xf1, 0x6c, 0x23, 0x56, 0x6f,
	0xaa, 0x62, 0xf5, 0x0b, 0xfb, 0x2c, 0x3b, 0x2b, 0x56, 0xf3, 0xd4, 0x4e, 0x6c, 0xbf, 0x26, 0x2b,
	0xe0, 0xc2, 0x32, 0xcd, 0xc2, 0xc5, 0xe9, 0x7e, 0xec, 0x2d, 0x9b, 0x51, 0xe7, 0x4d, 0x18, 0xa4,
	0x79, 0x46, 0x40, 0x07, 0x6d, 0x0f, 0x28, 0x3b, 0xa5, 0xd1, 0x36, 0xfb, 0x16, 0x43, 0x74, 0x56,
	0xcb, 0xc9, 0x90, 0xd2, 0xd1, 0x63, 0x47, 0x32, 0x09, 0xea, 0xbf, 0xe9, 0x73, 0x78, 0x05, 0x00,
	0x00, 0xff, 0xff, 0xd6, 0x70, 0xc8, 0xb6, 0x1b, 0x01, 0x00, 0x00,
}
