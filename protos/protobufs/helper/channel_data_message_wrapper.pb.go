// Code generated by protoc-gen-go. DO NOT EDIT.
// source: channel_data_message_wrapper.proto

package helper

import (
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	consensus "github.com/hyperledger/fabric-droplib/protos/protobufs/consensus"
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

type ConsensusVoteChannelDataMessageWrapper struct {
	// Types that are valid to be assigned to Sum:
	//	*ConsensusVoteChannelDataMessageWrapper_Vote
	Sum                  isConsensusVoteChannelDataMessageWrapper_Sum `protobuf_oneof:"sum"`
	XXX_NoUnkeyedLiteral struct{}                                     `json:"-"`
	XXX_unrecognized     []byte                                       `json:"-"`
	XXX_sizecache        int32                                        `json:"-"`
}

func (m *ConsensusVoteChannelDataMessageWrapper) Reset() {
	*m = ConsensusVoteChannelDataMessageWrapper{}
}
func (m *ConsensusVoteChannelDataMessageWrapper) String() string { return proto.CompactTextString(m) }
func (*ConsensusVoteChannelDataMessageWrapper) ProtoMessage()    {}
func (*ConsensusVoteChannelDataMessageWrapper) Descriptor() ([]byte, []int) {
	return fileDescriptor_1c67989915eedd3e, []int{0}
}

func (m *ConsensusVoteChannelDataMessageWrapper) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ConsensusVoteChannelDataMessageWrapper.Unmarshal(m, b)
}
func (m *ConsensusVoteChannelDataMessageWrapper) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ConsensusVoteChannelDataMessageWrapper.Marshal(b, m, deterministic)
}
func (m *ConsensusVoteChannelDataMessageWrapper) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ConsensusVoteChannelDataMessageWrapper.Merge(m, src)
}
func (m *ConsensusVoteChannelDataMessageWrapper) XXX_Size() int {
	return xxx_messageInfo_ConsensusVoteChannelDataMessageWrapper.Size(m)
}
func (m *ConsensusVoteChannelDataMessageWrapper) XXX_DiscardUnknown() {
	xxx_messageInfo_ConsensusVoteChannelDataMessageWrapper.DiscardUnknown(m)
}

var xxx_messageInfo_ConsensusVoteChannelDataMessageWrapper proto.InternalMessageInfo

type isConsensusVoteChannelDataMessageWrapper_Sum interface {
	isConsensusVoteChannelDataMessageWrapper_Sum()
}

type ConsensusVoteChannelDataMessageWrapper_Vote struct {
	Vote *consensus.VoteProtoWrapper `protobuf:"bytes,1,opt,name=vote,proto3,oneof"`
}

func (*ConsensusVoteChannelDataMessageWrapper_Vote) isConsensusVoteChannelDataMessageWrapper_Sum() {}

func (m *ConsensusVoteChannelDataMessageWrapper) GetSum() isConsensusVoteChannelDataMessageWrapper_Sum {
	if m != nil {
		return m.Sum
	}
	return nil
}

func (m *ConsensusVoteChannelDataMessageWrapper) GetVote() *consensus.VoteProtoWrapper {
	if x, ok := m.GetSum().(*ConsensusVoteChannelDataMessageWrapper_Vote); ok {
		return x.Vote
	}
	return nil
}

// XXX_OneofWrappers is for the internal use of the proto package.
func (*ConsensusVoteChannelDataMessageWrapper) XXX_OneofWrappers() []interface{} {
	return []interface{}{
		(*ConsensusVoteChannelDataMessageWrapper_Vote)(nil),
	}
}

//
//NewRoundStepProtoWrapper
//NewValidBlockProtoWrapper
//HasVoteProtoWrapper
//VoteSetMaj23ProtoWrapper
type ConsensusStateChannelDataMessageWrapper struct {
	// Types that are valid to be assigned to Sum:
	//	*ConsensusStateChannelDataMessageWrapper_NewRoundStep
	//	*ConsensusStateChannelDataMessageWrapper_NewValidBlock
	//	*ConsensusStateChannelDataMessageWrapper_HasVote
	//	*ConsensusStateChannelDataMessageWrapper_VoteSetMaj23
	Sum                  isConsensusStateChannelDataMessageWrapper_Sum `protobuf_oneof:"sum"`
	XXX_NoUnkeyedLiteral struct{}                                      `json:"-"`
	XXX_unrecognized     []byte                                        `json:"-"`
	XXX_sizecache        int32                                         `json:"-"`
}

func (m *ConsensusStateChannelDataMessageWrapper) Reset() {
	*m = ConsensusStateChannelDataMessageWrapper{}
}
func (m *ConsensusStateChannelDataMessageWrapper) String() string { return proto.CompactTextString(m) }
func (*ConsensusStateChannelDataMessageWrapper) ProtoMessage()    {}
func (*ConsensusStateChannelDataMessageWrapper) Descriptor() ([]byte, []int) {
	return fileDescriptor_1c67989915eedd3e, []int{1}
}

func (m *ConsensusStateChannelDataMessageWrapper) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ConsensusStateChannelDataMessageWrapper.Unmarshal(m, b)
}
func (m *ConsensusStateChannelDataMessageWrapper) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ConsensusStateChannelDataMessageWrapper.Marshal(b, m, deterministic)
}
func (m *ConsensusStateChannelDataMessageWrapper) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ConsensusStateChannelDataMessageWrapper.Merge(m, src)
}
func (m *ConsensusStateChannelDataMessageWrapper) XXX_Size() int {
	return xxx_messageInfo_ConsensusStateChannelDataMessageWrapper.Size(m)
}
func (m *ConsensusStateChannelDataMessageWrapper) XXX_DiscardUnknown() {
	xxx_messageInfo_ConsensusStateChannelDataMessageWrapper.DiscardUnknown(m)
}

var xxx_messageInfo_ConsensusStateChannelDataMessageWrapper proto.InternalMessageInfo

type isConsensusStateChannelDataMessageWrapper_Sum interface {
	isConsensusStateChannelDataMessageWrapper_Sum()
}

type ConsensusStateChannelDataMessageWrapper_NewRoundStep struct {
	NewRoundStep *consensus.NewRoundStepProtoWrapper `protobuf:"bytes,1,opt,name=new_round_step,json=newRoundStep,proto3,oneof"`
}

type ConsensusStateChannelDataMessageWrapper_NewValidBlock struct {
	NewValidBlock *consensus.NewValidBlockProtoWrapper `protobuf:"bytes,2,opt,name=new_valid_block,json=newValidBlock,proto3,oneof"`
}

type ConsensusStateChannelDataMessageWrapper_HasVote struct {
	HasVote *consensus.HasVoteProtoWrapper `protobuf:"bytes,3,opt,name=has_vote,json=hasVote,proto3,oneof"`
}

type ConsensusStateChannelDataMessageWrapper_VoteSetMaj23 struct {
	VoteSetMaj23 *consensus.VoteSetMaj23ProtoWrapper `protobuf:"bytes,4,opt,name=vote_set_maj23,json=voteSetMaj23,proto3,oneof"`
}

func (*ConsensusStateChannelDataMessageWrapper_NewRoundStep) isConsensusStateChannelDataMessageWrapper_Sum() {
}

func (*ConsensusStateChannelDataMessageWrapper_NewValidBlock) isConsensusStateChannelDataMessageWrapper_Sum() {
}

func (*ConsensusStateChannelDataMessageWrapper_HasVote) isConsensusStateChannelDataMessageWrapper_Sum() {
}

func (*ConsensusStateChannelDataMessageWrapper_VoteSetMaj23) isConsensusStateChannelDataMessageWrapper_Sum() {
}

func (m *ConsensusStateChannelDataMessageWrapper) GetSum() isConsensusStateChannelDataMessageWrapper_Sum {
	if m != nil {
		return m.Sum
	}
	return nil
}

func (m *ConsensusStateChannelDataMessageWrapper) GetNewRoundStep() *consensus.NewRoundStepProtoWrapper {
	if x, ok := m.GetSum().(*ConsensusStateChannelDataMessageWrapper_NewRoundStep); ok {
		return x.NewRoundStep
	}
	return nil
}

func (m *ConsensusStateChannelDataMessageWrapper) GetNewValidBlock() *consensus.NewValidBlockProtoWrapper {
	if x, ok := m.GetSum().(*ConsensusStateChannelDataMessageWrapper_NewValidBlock); ok {
		return x.NewValidBlock
	}
	return nil
}

func (m *ConsensusStateChannelDataMessageWrapper) GetHasVote() *consensus.HasVoteProtoWrapper {
	if x, ok := m.GetSum().(*ConsensusStateChannelDataMessageWrapper_HasVote); ok {
		return x.HasVote
	}
	return nil
}

func (m *ConsensusStateChannelDataMessageWrapper) GetVoteSetMaj23() *consensus.VoteSetMaj23ProtoWrapper {
	if x, ok := m.GetSum().(*ConsensusStateChannelDataMessageWrapper_VoteSetMaj23); ok {
		return x.VoteSetMaj23
	}
	return nil
}

// XXX_OneofWrappers is for the internal use of the proto package.
func (*ConsensusStateChannelDataMessageWrapper) XXX_OneofWrappers() []interface{} {
	return []interface{}{
		(*ConsensusStateChannelDataMessageWrapper_NewRoundStep)(nil),
		(*ConsensusStateChannelDataMessageWrapper_NewValidBlock)(nil),
		(*ConsensusStateChannelDataMessageWrapper_HasVote)(nil),
		(*ConsensusStateChannelDataMessageWrapper_VoteSetMaj23)(nil),
	}
}

//
//ProposalProtoWrapper
//ProposalPOLProtoWrapper
//BlockPartProtoWrapper
type ConsensusDataChannelDataMessageWrapper struct {
	// Types that are valid to be assigned to Sum:
	//	*ConsensusDataChannelDataMessageWrapper_Proposal
	//	*ConsensusDataChannelDataMessageWrapper_ProposalPol
	//	*ConsensusDataChannelDataMessageWrapper_BlockPart
	Sum                  isConsensusDataChannelDataMessageWrapper_Sum `protobuf_oneof:"sum"`
	XXX_NoUnkeyedLiteral struct{}                                     `json:"-"`
	XXX_unrecognized     []byte                                       `json:"-"`
	XXX_sizecache        int32                                        `json:"-"`
}

func (m *ConsensusDataChannelDataMessageWrapper) Reset() {
	*m = ConsensusDataChannelDataMessageWrapper{}
}
func (m *ConsensusDataChannelDataMessageWrapper) String() string { return proto.CompactTextString(m) }
func (*ConsensusDataChannelDataMessageWrapper) ProtoMessage()    {}
func (*ConsensusDataChannelDataMessageWrapper) Descriptor() ([]byte, []int) {
	return fileDescriptor_1c67989915eedd3e, []int{2}
}

func (m *ConsensusDataChannelDataMessageWrapper) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ConsensusDataChannelDataMessageWrapper.Unmarshal(m, b)
}
func (m *ConsensusDataChannelDataMessageWrapper) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ConsensusDataChannelDataMessageWrapper.Marshal(b, m, deterministic)
}
func (m *ConsensusDataChannelDataMessageWrapper) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ConsensusDataChannelDataMessageWrapper.Merge(m, src)
}
func (m *ConsensusDataChannelDataMessageWrapper) XXX_Size() int {
	return xxx_messageInfo_ConsensusDataChannelDataMessageWrapper.Size(m)
}
func (m *ConsensusDataChannelDataMessageWrapper) XXX_DiscardUnknown() {
	xxx_messageInfo_ConsensusDataChannelDataMessageWrapper.DiscardUnknown(m)
}

var xxx_messageInfo_ConsensusDataChannelDataMessageWrapper proto.InternalMessageInfo

type isConsensusDataChannelDataMessageWrapper_Sum interface {
	isConsensusDataChannelDataMessageWrapper_Sum()
}

type ConsensusDataChannelDataMessageWrapper_Proposal struct {
	Proposal *consensus.ProposalProtoWrapper `protobuf:"bytes,1,opt,name=proposal,proto3,oneof"`
}

type ConsensusDataChannelDataMessageWrapper_ProposalPol struct {
	ProposalPol *consensus.ProposalPOLProtoWrapper `protobuf:"bytes,2,opt,name=proposalPol,proto3,oneof"`
}

type ConsensusDataChannelDataMessageWrapper_BlockPart struct {
	BlockPart *consensus.BlockPartProtoWrapper `protobuf:"bytes,3,opt,name=blockPart,proto3,oneof"`
}

func (*ConsensusDataChannelDataMessageWrapper_Proposal) isConsensusDataChannelDataMessageWrapper_Sum() {
}

func (*ConsensusDataChannelDataMessageWrapper_ProposalPol) isConsensusDataChannelDataMessageWrapper_Sum() {
}

func (*ConsensusDataChannelDataMessageWrapper_BlockPart) isConsensusDataChannelDataMessageWrapper_Sum() {
}

func (m *ConsensusDataChannelDataMessageWrapper) GetSum() isConsensusDataChannelDataMessageWrapper_Sum {
	if m != nil {
		return m.Sum
	}
	return nil
}

func (m *ConsensusDataChannelDataMessageWrapper) GetProposal() *consensus.ProposalProtoWrapper {
	if x, ok := m.GetSum().(*ConsensusDataChannelDataMessageWrapper_Proposal); ok {
		return x.Proposal
	}
	return nil
}

func (m *ConsensusDataChannelDataMessageWrapper) GetProposalPol() *consensus.ProposalPOLProtoWrapper {
	if x, ok := m.GetSum().(*ConsensusDataChannelDataMessageWrapper_ProposalPol); ok {
		return x.ProposalPol
	}
	return nil
}

func (m *ConsensusDataChannelDataMessageWrapper) GetBlockPart() *consensus.BlockPartProtoWrapper {
	if x, ok := m.GetSum().(*ConsensusDataChannelDataMessageWrapper_BlockPart); ok {
		return x.BlockPart
	}
	return nil
}

// XXX_OneofWrappers is for the internal use of the proto package.
func (*ConsensusDataChannelDataMessageWrapper) XXX_OneofWrappers() []interface{} {
	return []interface{}{
		(*ConsensusDataChannelDataMessageWrapper_Proposal)(nil),
		(*ConsensusDataChannelDataMessageWrapper_ProposalPol)(nil),
		(*ConsensusDataChannelDataMessageWrapper_BlockPart)(nil),
	}
}

//
//VoteSetBitsProtoWrapper
//
type ConsensusVoteSetBitsChannelDataMessageWrapper struct {
	// Types that are valid to be assigned to Sum:
	//	*ConsensusVoteSetBitsChannelDataMessageWrapper_VoteSetBits
	Sum                  isConsensusVoteSetBitsChannelDataMessageWrapper_Sum `protobuf_oneof:"sum"`
	XXX_NoUnkeyedLiteral struct{}                                            `json:"-"`
	XXX_unrecognized     []byte                                              `json:"-"`
	XXX_sizecache        int32                                               `json:"-"`
}

func (m *ConsensusVoteSetBitsChannelDataMessageWrapper) Reset() {
	*m = ConsensusVoteSetBitsChannelDataMessageWrapper{}
}
func (m *ConsensusVoteSetBitsChannelDataMessageWrapper) String() string {
	return proto.CompactTextString(m)
}
func (*ConsensusVoteSetBitsChannelDataMessageWrapper) ProtoMessage() {}
func (*ConsensusVoteSetBitsChannelDataMessageWrapper) Descriptor() ([]byte, []int) {
	return fileDescriptor_1c67989915eedd3e, []int{3}
}

func (m *ConsensusVoteSetBitsChannelDataMessageWrapper) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ConsensusVoteSetBitsChannelDataMessageWrapper.Unmarshal(m, b)
}
func (m *ConsensusVoteSetBitsChannelDataMessageWrapper) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ConsensusVoteSetBitsChannelDataMessageWrapper.Marshal(b, m, deterministic)
}
func (m *ConsensusVoteSetBitsChannelDataMessageWrapper) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ConsensusVoteSetBitsChannelDataMessageWrapper.Merge(m, src)
}
func (m *ConsensusVoteSetBitsChannelDataMessageWrapper) XXX_Size() int {
	return xxx_messageInfo_ConsensusVoteSetBitsChannelDataMessageWrapper.Size(m)
}
func (m *ConsensusVoteSetBitsChannelDataMessageWrapper) XXX_DiscardUnknown() {
	xxx_messageInfo_ConsensusVoteSetBitsChannelDataMessageWrapper.DiscardUnknown(m)
}

var xxx_messageInfo_ConsensusVoteSetBitsChannelDataMessageWrapper proto.InternalMessageInfo

type isConsensusVoteSetBitsChannelDataMessageWrapper_Sum interface {
	isConsensusVoteSetBitsChannelDataMessageWrapper_Sum()
}

type ConsensusVoteSetBitsChannelDataMessageWrapper_VoteSetBits struct {
	VoteSetBits *consensus.VoteSetBitsProtoWrapper `protobuf:"bytes,1,opt,name=voteSetBits,proto3,oneof"`
}

func (*ConsensusVoteSetBitsChannelDataMessageWrapper_VoteSetBits) isConsensusVoteSetBitsChannelDataMessageWrapper_Sum() {
}

func (m *ConsensusVoteSetBitsChannelDataMessageWrapper) GetSum() isConsensusVoteSetBitsChannelDataMessageWrapper_Sum {
	if m != nil {
		return m.Sum
	}
	return nil
}

func (m *ConsensusVoteSetBitsChannelDataMessageWrapper) GetVoteSetBits() *consensus.VoteSetBitsProtoWrapper {
	if x, ok := m.GetSum().(*ConsensusVoteSetBitsChannelDataMessageWrapper_VoteSetBits); ok {
		return x.VoteSetBits
	}
	return nil
}

// XXX_OneofWrappers is for the internal use of the proto package.
func (*ConsensusVoteSetBitsChannelDataMessageWrapper) XXX_OneofWrappers() []interface{} {
	return []interface{}{
		(*ConsensusVoteSetBitsChannelDataMessageWrapper_VoteSetBits)(nil),
	}
}

func init() {
	proto.RegisterType((*ConsensusVoteChannelDataMessageWrapper)(nil), "tendermint.helper.ConsensusVoteChannelDataMessageWrapper")
	proto.RegisterType((*ConsensusStateChannelDataMessageWrapper)(nil), "tendermint.helper.ConsensusStateChannelDataMessageWrapper")
	proto.RegisterType((*ConsensusDataChannelDataMessageWrapper)(nil), "tendermint.helper.ConsensusDataChannelDataMessageWrapper")
	proto.RegisterType((*ConsensusVoteSetBitsChannelDataMessageWrapper)(nil), "tendermint.helper.ConsensusVoteSetBitsChannelDataMessageWrapper")
}

func init() {
	proto.RegisterFile("channel_data_message_wrapper.proto", fileDescriptor_1c67989915eedd3e)
}

var fileDescriptor_1c67989915eedd3e = []byte{
	// 467 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x9c, 0x94, 0xdf, 0x6e, 0xd3, 0x30,
	0x14, 0xc6, 0xe9, 0x3a, 0x60, 0xb8, 0xfc, 0x11, 0xb9, 0xaa, 0xb8, 0x42, 0xbd, 0x00, 0x84, 0xd4,
	0x44, 0x6b, 0xaf, 0x11, 0x28, 0xe3, 0x62, 0x42, 0x1d, 0xab, 0x56, 0x6d, 0x48, 0x93, 0x90, 0xe5,
	0x24, 0x67, 0x4d, 0x20, 0xb1, 0x8d, 0x7d, 0xd2, 0x6a, 0x0f, 0xc1, 0x15, 0xcf, 0xc0, 0x7b, 0x22,
	0x3b, 0x4e, 0x16, 0x22, 0xb2, 0xa2, 0x5d, 0xd5, 0xf2, 0xf9, 0xce, 0xef, 0x9c, 0x7e, 0x5f, 0x6b,
	0x32, 0x89, 0x53, 0xc6, 0x39, 0xe4, 0x34, 0x61, 0xc8, 0x68, 0x01, 0x5a, 0xb3, 0x35, 0xd0, 0xad,
	0x62, 0x52, 0x82, 0xf2, 0xa5, 0x12, 0x28, 0xbc, 0xe7, 0x08, 0x3c, 0x01, 0x55, 0x64, 0x1c, 0xfd,
	0x14, 0x72, 0x09, 0xea, 0xc5, 0x62, 0x9d, 0x61, 0x5a, 0x46, 0x7e, 0x2c, 0x8a, 0x20, 0xbd, 0x96,
	0xa0, 0x72, 0x48, 0xd6, 0xa0, 0x82, 0x2b, 0x16, 0xa9, 0x2c, 0x9e, 0x26, 0x4a, 0xc8, 0x3c, 0x8b,
	0x02, 0xdb, 0xaf, 0xab, 0x8f, 0xa8, 0xbc, 0xd2, 0x41, 0x2c, 0xb8, 0x06, 0xae, 0x4b, 0x1d, 0xe0,
	0xb5, 0x04, 0x5d, 0x0d, 0x98, 0xfc, 0x20, 0xaf, 0x8e, 0xea, 0xc2, 0x85, 0x40, 0x38, 0xaa, 0x76,
	0xfa, 0xc8, 0x90, 0x9d, 0x54, 0x1b, 0x7d, 0xa9, 0x16, 0xf2, 0x3e, 0x90, 0xfd, 0x8d, 0x40, 0x18,
	0x0f, 0x5e, 0x0e, 0xde, 0x8c, 0x66, 0x6f, 0xfd, 0x86, 0xec, 0x37, 0x64, 0x7b, 0xe5, 0x1b, 0xcc,
	0xd2, 0x9c, 0x5c, 0xe7, 0xf1, 0xbd, 0x33, 0xdb, 0x19, 0xde, 0x27, 0x43, 0x5d, 0x16, 0x93, 0x5f,
	0x43, 0xf2, 0xba, 0x99, 0xb9, 0x42, 0x76, 0xdb, 0xd0, 0x4b, 0xf2, 0x94, 0xc3, 0x96, 0x2a, 0x51,
	0xf2, 0x84, 0x6a, 0x04, 0xe9, 0xc6, 0xcf, 0xfa, 0xc7, 0x7f, 0x86, 0xed, 0x99, 0x91, 0xaf, 0x10,
	0x64, 0x67, 0x8d, 0xc7, 0xbc, 0x55, 0xf3, 0xbe, 0x92, 0x67, 0x86, 0xbd, 0x61, 0x79, 0x96, 0xd0,
	0x28, 0x17, 0xf1, 0xf7, 0xf1, 0x9e, 0x85, 0xcf, 0x6f, 0x85, 0x5f, 0x18, 0x7d, 0x68, 0xe4, 0x1d,
	0xfa, 0x13, 0xde, 0x2e, 0x7a, 0x9f, 0xc8, 0x41, 0xca, 0x34, 0xb5, 0x9e, 0x0d, 0x2d, 0x77, 0xda,
	0xcf, 0x3d, 0x66, 0xfa, 0x1f, 0xb6, 0x3d, 0x4c, 0xab, 0x6b, 0x63, 0x83, 0xe1, 0x50, 0x0d, 0x48,
	0x0b, 0xf6, 0x6d, 0x36, 0x1f, 0xef, 0xef, 0xb2, 0xc1, 0xf4, 0xad, 0x00, 0x4f, 0x8c, 0xba, 0x6b,
	0xc3, 0xa6, 0x55, 0xab, 0x53, 0xf9, 0xbd, 0xd7, 0xfa, 0x25, 0x98, 0x24, 0xfa, 0x43, 0x59, 0x90,
	0x03, 0xa9, 0x84, 0x14, 0x9a, 0xe5, 0x2e, 0x0e, 0xbf, 0x7f, 0x8f, 0xa5, 0x53, 0x76, 0x76, 0x68,
	0x08, 0xde, 0x39, 0x19, 0xd5, 0xe7, 0xa5, 0xc8, 0x5d, 0x04, 0x87, 0xff, 0x01, 0x3c, 0x5d, 0x74,
	0x98, 0x6d, 0x8e, 0x77, 0x4a, 0x1e, 0xd9, 0x4c, 0x97, 0x4c, 0xa1, 0xf3, 0x3f, 0xe8, 0x87, 0x86,
	0xb5, 0xb4, 0x83, 0xbc, 0x61, 0xd4, 0x3e, 0xfd, 0x1c, 0x90, 0xe9, 0x5f, 0xff, 0x98, 0x15, 0x60,
	0x98, 0xa1, 0xee, 0xb7, 0xeb, 0x9c, 0x8c, 0x36, 0x37, 0x3a, 0xe7, 0xd8, 0xe1, 0xce, 0xe4, 0x8c,
	0xb8, 0xfb, 0x05, 0x5b, 0x1c, 0xb7, 0x4f, 0xf8, 0xfe, 0xf2, 0xdd, 0x1d, 0x1f, 0x84, 0xea, 0x3d,
	0x89, 0x1e, 0xd8, 0x9b, 0xf9, 0x9f, 0x00, 0x00, 0x00, 0xff, 0xff, 0x04, 0x93, 0xd7, 0xc5, 0x8f,
	0x04, 0x00, 0x00,
}
