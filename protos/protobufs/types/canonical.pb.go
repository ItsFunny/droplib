// Code generated by protoc-gen-go. DO NOT EDIT.
// source: canonical.proto

package types

import (
	fmt "fmt"
	_ "github.com/gogo/protobuf/gogoproto"
	proto "github.com/golang/protobuf/proto"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
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

type CanonicalBlockIDProto struct {
	Hash                 []byte                       `protobuf:"bytes,1,opt,name=hash,proto3" json:"hash,omitempty"`
	PartSetHeader        *CanonicalPartSetHeaderProto `protobuf:"bytes,2,opt,name=part_set_header,json=partSetHeader,proto3" json:"part_set_header,omitempty"`
	XXX_NoUnkeyedLiteral struct{}                     `json:"-"`
	XXX_unrecognized     []byte                       `json:"-"`
	XXX_sizecache        int32                        `json:"-"`
}

func (m *CanonicalBlockIDProto) Reset()         { *m = CanonicalBlockIDProto{} }
func (m *CanonicalBlockIDProto) String() string { return proto.CompactTextString(m) }
func (*CanonicalBlockIDProto) ProtoMessage()    {}
func (*CanonicalBlockIDProto) Descriptor() ([]byte, []int) {
	return fileDescriptor_17c7d1ea96171b5b, []int{0}
}

func (m *CanonicalBlockIDProto) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_CanonicalBlockIDProto.Unmarshal(m, b)
}
func (m *CanonicalBlockIDProto) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_CanonicalBlockIDProto.Marshal(b, m, deterministic)
}
func (m *CanonicalBlockIDProto) XXX_Merge(src proto.Message) {
	xxx_messageInfo_CanonicalBlockIDProto.Merge(m, src)
}
func (m *CanonicalBlockIDProto) XXX_Size() int {
	return xxx_messageInfo_CanonicalBlockIDProto.Size(m)
}
func (m *CanonicalBlockIDProto) XXX_DiscardUnknown() {
	xxx_messageInfo_CanonicalBlockIDProto.DiscardUnknown(m)
}

var xxx_messageInfo_CanonicalBlockIDProto proto.InternalMessageInfo

func (m *CanonicalBlockIDProto) GetHash() []byte {
	if m != nil {
		return m.Hash
	}
	return nil
}

func (m *CanonicalBlockIDProto) GetPartSetHeader() *CanonicalPartSetHeaderProto {
	if m != nil {
		return m.PartSetHeader
	}
	return nil
}

type CanonicalPartSetHeaderProto struct {
	Total                uint32   `protobuf:"varint,1,opt,name=total,proto3" json:"total,omitempty"`
	Hash                 []byte   `protobuf:"bytes,2,opt,name=hash,proto3" json:"hash,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *CanonicalPartSetHeaderProto) Reset()         { *m = CanonicalPartSetHeaderProto{} }
func (m *CanonicalPartSetHeaderProto) String() string { return proto.CompactTextString(m) }
func (*CanonicalPartSetHeaderProto) ProtoMessage()    {}
func (*CanonicalPartSetHeaderProto) Descriptor() ([]byte, []int) {
	return fileDescriptor_17c7d1ea96171b5b, []int{1}
}

func (m *CanonicalPartSetHeaderProto) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_CanonicalPartSetHeaderProto.Unmarshal(m, b)
}
func (m *CanonicalPartSetHeaderProto) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_CanonicalPartSetHeaderProto.Marshal(b, m, deterministic)
}
func (m *CanonicalPartSetHeaderProto) XXX_Merge(src proto.Message) {
	xxx_messageInfo_CanonicalPartSetHeaderProto.Merge(m, src)
}
func (m *CanonicalPartSetHeaderProto) XXX_Size() int {
	return xxx_messageInfo_CanonicalPartSetHeaderProto.Size(m)
}
func (m *CanonicalPartSetHeaderProto) XXX_DiscardUnknown() {
	xxx_messageInfo_CanonicalPartSetHeaderProto.DiscardUnknown(m)
}

var xxx_messageInfo_CanonicalPartSetHeaderProto proto.InternalMessageInfo

func (m *CanonicalPartSetHeaderProto) GetTotal() uint32 {
	if m != nil {
		return m.Total
	}
	return 0
}

func (m *CanonicalPartSetHeaderProto) GetHash() []byte {
	if m != nil {
		return m.Hash
	}
	return nil
}

type CanonicalProposalProto struct {
	Type                 SignedMsgType          `protobuf:"varint,1,opt,name=type,proto3,enum=tendermint.types.SignedMsgType" json:"type,omitempty"`
	Height               int64                  `protobuf:"fixed64,2,opt,name=height,proto3" json:"height,omitempty"`
	Round                int64                  `protobuf:"fixed64,3,opt,name=round,proto3" json:"round,omitempty"`
	PolRound             int64                  `protobuf:"varint,4,opt,name=pol_round,json=polRound,proto3" json:"pol_round,omitempty"`
	BlockId              *CanonicalBlockIDProto `protobuf:"bytes,5,opt,name=block_id,json=blockId,proto3" json:"block_id,omitempty"`
	Timestamp            *timestamppb.Timestamp `protobuf:"bytes,6,opt,name=timestamp,proto3" json:"timestamp,omitempty"`
	ChainId              string                 `protobuf:"bytes,7,opt,name=chain_id,json=chainId,proto3" json:"chain_id,omitempty"`
	XXX_NoUnkeyedLiteral struct{}               `json:"-"`
	XXX_unrecognized     []byte                 `json:"-"`
	XXX_sizecache        int32                  `json:"-"`
}

func (m *CanonicalProposalProto) Reset()         { *m = CanonicalProposalProto{} }
func (m *CanonicalProposalProto) String() string { return proto.CompactTextString(m) }
func (*CanonicalProposalProto) ProtoMessage()    {}
func (*CanonicalProposalProto) Descriptor() ([]byte, []int) {
	return fileDescriptor_17c7d1ea96171b5b, []int{2}
}

func (m *CanonicalProposalProto) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_CanonicalProposalProto.Unmarshal(m, b)
}
func (m *CanonicalProposalProto) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_CanonicalProposalProto.Marshal(b, m, deterministic)
}
func (m *CanonicalProposalProto) XXX_Merge(src proto.Message) {
	xxx_messageInfo_CanonicalProposalProto.Merge(m, src)
}
func (m *CanonicalProposalProto) XXX_Size() int {
	return xxx_messageInfo_CanonicalProposalProto.Size(m)
}
func (m *CanonicalProposalProto) XXX_DiscardUnknown() {
	xxx_messageInfo_CanonicalProposalProto.DiscardUnknown(m)
}

var xxx_messageInfo_CanonicalProposalProto proto.InternalMessageInfo

func (m *CanonicalProposalProto) GetType() SignedMsgType {
	if m != nil {
		return m.Type
	}
	return SignedMsgType_SIGNED_MSG_TYPE_UNKNOWN
}

func (m *CanonicalProposalProto) GetHeight() int64 {
	if m != nil {
		return m.Height
	}
	return 0
}

func (m *CanonicalProposalProto) GetRound() int64 {
	if m != nil {
		return m.Round
	}
	return 0
}

func (m *CanonicalProposalProto) GetPolRound() int64 {
	if m != nil {
		return m.PolRound
	}
	return 0
}

func (m *CanonicalProposalProto) GetBlockId() *CanonicalBlockIDProto {
	if m != nil {
		return m.BlockId
	}
	return nil
}

func (m *CanonicalProposalProto) GetTimestamp() *timestamppb.Timestamp {
	if m != nil {
		return m.Timestamp
	}
	return nil
}

func (m *CanonicalProposalProto) GetChainId() string {
	if m != nil {
		return m.ChainId
	}
	return ""
}

type CanonicalVoteProto struct {
	Type                 SignedMsgType          `protobuf:"varint,1,opt,name=type,proto3,enum=tendermint.types.SignedMsgType" json:"type,omitempty"`
	Height               int64                  `protobuf:"fixed64,2,opt,name=height,proto3" json:"height,omitempty"`
	Round                int64                  `protobuf:"fixed64,3,opt,name=round,proto3" json:"round,omitempty"`
	BlockId              *CanonicalBlockIDProto `protobuf:"bytes,4,opt,name=block_id,json=blockId,proto3" json:"block_id,omitempty"`
	Timestamp            *timestamppb.Timestamp `protobuf:"bytes,5,opt,name=timestamp,proto3" json:"timestamp,omitempty"`
	ChainId              string                 `protobuf:"bytes,6,opt,name=chain_id,json=chainId,proto3" json:"chain_id,omitempty"`
	XXX_NoUnkeyedLiteral struct{}               `json:"-"`
	XXX_unrecognized     []byte                 `json:"-"`
	XXX_sizecache        int32                  `json:"-"`
}

func (m *CanonicalVoteProto) Reset()         { *m = CanonicalVoteProto{} }
func (m *CanonicalVoteProto) String() string { return proto.CompactTextString(m) }
func (*CanonicalVoteProto) ProtoMessage()    {}
func (*CanonicalVoteProto) Descriptor() ([]byte, []int) {
	return fileDescriptor_17c7d1ea96171b5b, []int{3}
}

func (m *CanonicalVoteProto) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_CanonicalVoteProto.Unmarshal(m, b)
}
func (m *CanonicalVoteProto) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_CanonicalVoteProto.Marshal(b, m, deterministic)
}
func (m *CanonicalVoteProto) XXX_Merge(src proto.Message) {
	xxx_messageInfo_CanonicalVoteProto.Merge(m, src)
}
func (m *CanonicalVoteProto) XXX_Size() int {
	return xxx_messageInfo_CanonicalVoteProto.Size(m)
}
func (m *CanonicalVoteProto) XXX_DiscardUnknown() {
	xxx_messageInfo_CanonicalVoteProto.DiscardUnknown(m)
}

var xxx_messageInfo_CanonicalVoteProto proto.InternalMessageInfo

func (m *CanonicalVoteProto) GetType() SignedMsgType {
	if m != nil {
		return m.Type
	}
	return SignedMsgType_SIGNED_MSG_TYPE_UNKNOWN
}

func (m *CanonicalVoteProto) GetHeight() int64 {
	if m != nil {
		return m.Height
	}
	return 0
}

func (m *CanonicalVoteProto) GetRound() int64 {
	if m != nil {
		return m.Round
	}
	return 0
}

func (m *CanonicalVoteProto) GetBlockId() *CanonicalBlockIDProto {
	if m != nil {
		return m.BlockId
	}
	return nil
}

func (m *CanonicalVoteProto) GetTimestamp() *timestamppb.Timestamp {
	if m != nil {
		return m.Timestamp
	}
	return nil
}

func (m *CanonicalVoteProto) GetChainId() string {
	if m != nil {
		return m.ChainId
	}
	return ""
}

func init() {
	proto.RegisterType((*CanonicalBlockIDProto)(nil), "tendermint.types.CanonicalBlockIDProto")
	proto.RegisterType((*CanonicalPartSetHeaderProto)(nil), "tendermint.types.CanonicalPartSetHeaderProto")
	proto.RegisterType((*CanonicalProposalProto)(nil), "tendermint.types.CanonicalProposalProto")
	proto.RegisterType((*CanonicalVoteProto)(nil), "tendermint.types.CanonicalVoteProto")
}

func init() { proto.RegisterFile("canonical.proto", fileDescriptor_17c7d1ea96171b5b) }

var fileDescriptor_17c7d1ea96171b5b = []byte{
	// 486 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xbc, 0x53, 0x4d, 0x6e, 0xd3, 0x40,
	0x14, 0xc6, 0xa9, 0x93, 0x38, 0xd3, 0x96, 0x56, 0xa3, 0x52, 0x59, 0x61, 0xe1, 0x28, 0x0b, 0x08,
	0x8b, 0xda, 0x52, 0xbb, 0x45, 0x2c, 0x5c, 0x24, 0x1a, 0x09, 0xd4, 0xc8, 0xad, 0x58, 0xc0, 0x22,
	0x1a, 0xdb, 0x53, 0x7b, 0x84, 0xe3, 0x37, 0x1a, 0x4f, 0x16, 0xb9, 0x01, 0x4b, 0x0e, 0xc3, 0x21,
	0x38, 0x45, 0x90, 0xca, 0x45, 0x90, 0xdf, 0x24, 0x4e, 0x04, 0xa2, 0x9b, 0xa0, 0x6e, 0x32, 0xef,
	0xe7, 0x9b, 0xef, 0x7d, 0xf9, 0xde, 0x98, 0x1c, 0x25, 0xac, 0x84, 0x52, 0x24, 0xac, 0xf0, 0xa5,
	0x02, 0x0d, 0xf4, 0x58, 0xf3, 0x32, 0xe5, 0x6a, 0x26, 0x4a, 0xed, 0xeb, 0x85, 0xe4, 0x55, 0xff,
	0x24, 0x83, 0x0c, 0xb0, 0x19, 0xd4, 0x91, 0xc1, 0xf5, 0xaf, 0x32, 0xa1, 0xf3, 0x79, 0xec, 0x27,
	0x30, 0x0b, 0xf2, 0x85, 0xe4, 0xaa, 0xe0, 0x69, 0xc6, 0x55, 0x70, 0xc7, 0x62, 0x25, 0x92, 0xb3,
	0x54, 0x81, 0x2c, 0x44, 0x1c, 0x20, 0xb8, 0x32, 0x47, 0x3c, 0xbf, 0xab, 0x02, 0xe4, 0x34, 0xbf,
	0x2b, 0x26, 0x2f, 0x03, 0xc8, 0x0a, 0xde, 0x60, 0x02, 0x2d, 0x66, 0xbc, 0xd2, 0x6c, 0x26, 0x0d,
	0x60, 0xf8, 0xd5, 0x22, 0xcf, 0x2e, 0xd7, 0x32, 0xc3, 0x02, 0x92, 0x2f, 0xe3, 0xb7, 0x13, 0x14,
	0x4b, 0x89, 0x9d, 0xb3, 0x2a, 0x77, 0xad, 0x81, 0x35, 0x3a, 0x88, 0x30, 0xa6, 0x9f, 0xc9, 0x91,
	0x64, 0x4a, 0x4f, 0x2b, 0xae, 0xa7, 0x39, 0x67, 0x29, 0x57, 0x6e, 0x6b, 0x60, 0x8d, 0xf6, 0xcf,
	0xcf, 0xfc, 0x3f, 0xff, 0x9a, 0xdf, 0xb0, 0x4e, 0x98, 0xd2, 0x37, 0x5c, 0x5f, 0x21, 0x1e, 0xb9,
	0x43, 0xfb, 0xc7, 0xd2, 0x7b, 0x12, 0x1d, 0xca, 0xed, 0xce, 0xf0, 0x1d, 0x79, 0xfe, 0xc0, 0x1d,
	0x7a, 0x42, 0xda, 0x1a, 0x34, 0x2b, 0x50, 0xd0, 0x61, 0x64, 0x92, 0x46, 0x65, 0x6b, 0xa3, 0x72,
	0xf8, 0xab, 0x45, 0x4e, 0x37, 0x4c, 0x0a, 0x24, 0x54, 0x78, 0x6a, 0xa0, 0x17, 0xc4, 0xae, 0xd5,
	0x21, 0xc7, 0xd3, 0x73, 0xef, 0x6f, 0xd5, 0x37, 0x22, 0x2b, 0x79, 0xfa, 0xa1, 0xca, 0x6e, 0x17,
	0x92, 0x47, 0x08, 0xa6, 0xa7, 0xa4, 0x93, 0x73, 0x91, 0xe5, 0x1a, 0xa7, 0x1c, 0x47, 0xab, 0xac,
	0x56, 0xa4, 0x60, 0x5e, 0xa6, 0xee, 0x1e, 0x96, 0x4d, 0x42, 0x5f, 0x91, 0x9e, 0x84, 0x62, 0x6a,
	0x3a, 0xf6, 0xc0, 0x1a, 0xed, 0x85, 0x07, 0xf7, 0x4b, 0xcf, 0x99, 0x5c, 0xbf, 0x8f, 0xea, 0x5a,
	0xe4, 0x48, 0x28, 0x30, 0xa2, 0xd7, 0xc4, 0x89, 0x6b, 0xcb, 0xa7, 0x22, 0x75, 0xdb, 0xe8, 0xe3,
	0xcb, 0x07, 0x7c, 0xdc, 0xde, 0x4e, 0xb8, 0x7f, 0xbf, 0xf4, 0xba, 0xab, 0x4a, 0xd4, 0x45, 0x96,
	0x71, 0x4a, 0x43, 0xd2, 0x6b, 0x16, 0xec, 0x76, 0x90, 0xb1, 0xef, 0x9b, 0x27, 0xe0, 0xaf, 0x9f,
	0x80, 0x7f, 0xbb, 0x46, 0x84, 0x4e, 0xbd, 0x86, 0x6f, 0x3f, 0x3d, 0x2b, 0xda, 0x5c, 0xa3, 0x2f,
	0x88, 0x93, 0xe4, 0x4c, 0x94, 0xb5, 0xa8, 0xee, 0xc0, 0x1a, 0xf5, 0xcc, 0xac, 0xcb, 0xba, 0x56,
	0xcf, 0xc2, 0xe6, 0x38, 0x1d, 0x7e, 0x6f, 0x11, 0xda, 0x68, 0xfb, 0x08, 0x9a, 0x3f, 0x9a, 0xc3,
	0xdb, 0xb6, 0xd9, 0xff, 0xdd, 0xb6, 0xf6, 0xee, 0xb6, 0x75, 0xfe, 0x6d, 0x5b, 0xf8, 0xe6, 0xd3,
	0xeb, 0x5d, 0xbe, 0xee, 0xb8, 0x83, 0x85, 0x8b, 0xdf, 0x01, 0x00, 0x00, 0xff, 0xff, 0x59, 0x28,
	0x49, 0xb1, 0x5d, 0x04, 0x00, 0x00,
}
