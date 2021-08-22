// Code generated by protoc-gen-go. DO NOT EDIT.
// source: types.proto

package states

import (
	fmt "fmt"
	_ "github.com/gogo/protobuf/gogoproto"
	proto "github.com/golang/protobuf/proto"
	types "github.com/hyperledger/fabric-droplib/protos/protobufs/types"
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

// ValidatorsInfo represents the latest validator set, or the last height it changed
type ValidatorsInfoProto struct {
	ValidatorSet         *types.ValidatorSetProto `protobuf:"bytes,1,opt,name=validator_set,json=validatorSet,proto3" json:"validator_set,omitempty"`
	LastHeightChanged    int64                    `protobuf:"varint,2,opt,name=last_height_changed,json=lastHeightChanged,proto3" json:"last_height_changed,omitempty"`
	XXX_NoUnkeyedLiteral struct{}                 `json:"-"`
	XXX_unrecognized     []byte                   `json:"-"`
	XXX_sizecache        int32                    `json:"-"`
}

func (m *ValidatorsInfoProto) Reset()         { *m = ValidatorsInfoProto{} }
func (m *ValidatorsInfoProto) String() string { return proto.CompactTextString(m) }
func (*ValidatorsInfoProto) ProtoMessage()    {}
func (*ValidatorsInfoProto) Descriptor() ([]byte, []int) {
	return fileDescriptor_d938547f84707355, []int{0}
}

func (m *ValidatorsInfoProto) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ValidatorsInfoProto.Unmarshal(m, b)
}
func (m *ValidatorsInfoProto) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ValidatorsInfoProto.Marshal(b, m, deterministic)
}
func (m *ValidatorsInfoProto) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ValidatorsInfoProto.Merge(m, src)
}
func (m *ValidatorsInfoProto) XXX_Size() int {
	return xxx_messageInfo_ValidatorsInfoProto.Size(m)
}
func (m *ValidatorsInfoProto) XXX_DiscardUnknown() {
	xxx_messageInfo_ValidatorsInfoProto.DiscardUnknown(m)
}

var xxx_messageInfo_ValidatorsInfoProto proto.InternalMessageInfo

func (m *ValidatorsInfoProto) GetValidatorSet() *types.ValidatorSetProto {
	if m != nil {
		return m.ValidatorSet
	}
	return nil
}

func (m *ValidatorsInfoProto) GetLastHeightChanged() int64 {
	if m != nil {
		return m.LastHeightChanged
	}
	return 0
}

// ConsensusParamsInfo represents the latest consensus params, or the last height it changed
type ConsensusParamsInfoProto struct {
	ConsensusParams      *types.ConsensusParamsProto `protobuf:"bytes,1,opt,name=consensus_params,json=consensusParams,proto3" json:"consensus_params,omitempty"`
	LastHeightChanged    int64                       `protobuf:"varint,2,opt,name=last_height_changed,json=lastHeightChanged,proto3" json:"last_height_changed,omitempty"`
	XXX_NoUnkeyedLiteral struct{}                    `json:"-"`
	XXX_unrecognized     []byte                      `json:"-"`
	XXX_sizecache        int32                       `json:"-"`
}

func (m *ConsensusParamsInfoProto) Reset()         { *m = ConsensusParamsInfoProto{} }
func (m *ConsensusParamsInfoProto) String() string { return proto.CompactTextString(m) }
func (*ConsensusParamsInfoProto) ProtoMessage()    {}
func (*ConsensusParamsInfoProto) Descriptor() ([]byte, []int) {
	return fileDescriptor_d938547f84707355, []int{1}
}

func (m *ConsensusParamsInfoProto) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ConsensusParamsInfoProto.Unmarshal(m, b)
}
func (m *ConsensusParamsInfoProto) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ConsensusParamsInfoProto.Marshal(b, m, deterministic)
}
func (m *ConsensusParamsInfoProto) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ConsensusParamsInfoProto.Merge(m, src)
}
func (m *ConsensusParamsInfoProto) XXX_Size() int {
	return xxx_messageInfo_ConsensusParamsInfoProto.Size(m)
}
func (m *ConsensusParamsInfoProto) XXX_DiscardUnknown() {
	xxx_messageInfo_ConsensusParamsInfoProto.DiscardUnknown(m)
}

var xxx_messageInfo_ConsensusParamsInfoProto proto.InternalMessageInfo

func (m *ConsensusParamsInfoProto) GetConsensusParams() *types.ConsensusParamsProto {
	if m != nil {
		return m.ConsensusParams
	}
	return nil
}

func (m *ConsensusParamsInfoProto) GetLastHeightChanged() int64 {
	if m != nil {
		return m.LastHeightChanged
	}
	return 0
}

type LatestStateProto struct {
	//  Version version = 1 [(gogoproto.nullable) = false];
	// immutable
	ChainId       string `protobuf:"bytes,2,opt,name=chain_id,json=chainId,proto3" json:"chain_id,omitempty"`
	InitialHeight int64  `protobuf:"varint,14,opt,name=initial_height,json=initialHeight,proto3" json:"initial_height,omitempty"`
	// LastBlockHeight=0 at genesis (ie. block(H=0) does not exist)
	LastBlockHeight int64                  `protobuf:"varint,3,opt,name=last_block_height,json=lastBlockHeight,proto3" json:"last_block_height,omitempty"`
	LastBlockId     *types.BlockIDProto    `protobuf:"bytes,4,opt,name=last_block_id,json=lastBlockId,proto3" json:"last_block_id,omitempty"`
	LastBlockTime   *timestamppb.Timestamp `protobuf:"bytes,5,opt,name=last_block_time,json=lastBlockTime,proto3" json:"last_block_time,omitempty"`
	// LastValidators is used to validate block.LastCommit.
	// Validators are persisted to the database separately every time they change,
	// so we can query for historical validator sets.
	// Note that if s.LastBlockHeight causes a valset change,
	// we set s.LastHeightValidatorsChanged = s.LastBlockHeight + 1 + 1
	// Extra +1 due to nextValSet delay.
	NextValidators              *types.ValidatorSetProto `protobuf:"bytes,6,opt,name=next_validators,json=nextValidators,proto3" json:"next_validators,omitempty"`
	Validators                  *types.ValidatorSetProto `protobuf:"bytes,7,opt,name=validators,proto3" json:"validators,omitempty"`
	LastValidators              *types.ValidatorSetProto `protobuf:"bytes,8,opt,name=last_validators,json=lastValidators,proto3" json:"last_validators,omitempty"`
	LastHeightValidatorsChanged int64                    `protobuf:"varint,9,opt,name=last_height_validators_changed,json=lastHeightValidatorsChanged,proto3" json:"last_height_validators_changed,omitempty"`
	// Consensus parameters used for validating blocks.
	// Changes returned by EndBlock and updated after Commit.
	ConsensusParams                  *types.ConsensusParamsProto `protobuf:"bytes,10,opt,name=consensus_params,json=consensusParams,proto3" json:"consensus_params,omitempty"`
	LastHeightConsensusParamsChanged int64                       `protobuf:"varint,11,opt,name=last_height_consensus_params_changed,json=lastHeightConsensusParamsChanged,proto3" json:"last_height_consensus_params_changed,omitempty"`
	// Merkle root of the results from executing prev block
	LastResultsHash []byte `protobuf:"bytes,12,opt,name=last_results_hash,json=lastResultsHash,proto3" json:"last_results_hash,omitempty"`
	// the latest AppHash we've received from calling abci.Commit()
	AppHash              []byte   `protobuf:"bytes,13,opt,name=app_hash,json=appHash,proto3" json:"app_hash,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *LatestStateProto) Reset()         { *m = LatestStateProto{} }
func (m *LatestStateProto) String() string { return proto.CompactTextString(m) }
func (*LatestStateProto) ProtoMessage()    {}
func (*LatestStateProto) Descriptor() ([]byte, []int) {
	return fileDescriptor_d938547f84707355, []int{2}
}

func (m *LatestStateProto) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_LatestStateProto.Unmarshal(m, b)
}
func (m *LatestStateProto) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_LatestStateProto.Marshal(b, m, deterministic)
}
func (m *LatestStateProto) XXX_Merge(src proto.Message) {
	xxx_messageInfo_LatestStateProto.Merge(m, src)
}
func (m *LatestStateProto) XXX_Size() int {
	return xxx_messageInfo_LatestStateProto.Size(m)
}
func (m *LatestStateProto) XXX_DiscardUnknown() {
	xxx_messageInfo_LatestStateProto.DiscardUnknown(m)
}

var xxx_messageInfo_LatestStateProto proto.InternalMessageInfo

func (m *LatestStateProto) GetChainId() string {
	if m != nil {
		return m.ChainId
	}
	return ""
}

func (m *LatestStateProto) GetInitialHeight() int64 {
	if m != nil {
		return m.InitialHeight
	}
	return 0
}

func (m *LatestStateProto) GetLastBlockHeight() int64 {
	if m != nil {
		return m.LastBlockHeight
	}
	return 0
}

func (m *LatestStateProto) GetLastBlockId() *types.BlockIDProto {
	if m != nil {
		return m.LastBlockId
	}
	return nil
}

func (m *LatestStateProto) GetLastBlockTime() *timestamppb.Timestamp {
	if m != nil {
		return m.LastBlockTime
	}
	return nil
}

func (m *LatestStateProto) GetNextValidators() *types.ValidatorSetProto {
	if m != nil {
		return m.NextValidators
	}
	return nil
}

func (m *LatestStateProto) GetValidators() *types.ValidatorSetProto {
	if m != nil {
		return m.Validators
	}
	return nil
}

func (m *LatestStateProto) GetLastValidators() *types.ValidatorSetProto {
	if m != nil {
		return m.LastValidators
	}
	return nil
}

func (m *LatestStateProto) GetLastHeightValidatorsChanged() int64 {
	if m != nil {
		return m.LastHeightValidatorsChanged
	}
	return 0
}

func (m *LatestStateProto) GetConsensusParams() *types.ConsensusParamsProto {
	if m != nil {
		return m.ConsensusParams
	}
	return nil
}

func (m *LatestStateProto) GetLastHeightConsensusParamsChanged() int64 {
	if m != nil {
		return m.LastHeightConsensusParamsChanged
	}
	return 0
}

func (m *LatestStateProto) GetLastResultsHash() []byte {
	if m != nil {
		return m.LastResultsHash
	}
	return nil
}

func (m *LatestStateProto) GetAppHash() []byte {
	if m != nil {
		return m.AppHash
	}
	return nil
}

func init() {
	proto.RegisterType((*ValidatorsInfoProto)(nil), "tendermint.state.ValidatorsInfoProto")
	proto.RegisterType((*ConsensusParamsInfoProto)(nil), "tendermint.state.ConsensusParamsInfoProto")
	proto.RegisterType((*LatestStateProto)(nil), "tendermint.state.LatestStateProto")
}

func init() { proto.RegisterFile("types.proto", fileDescriptor_d938547f84707355) }

var fileDescriptor_d938547f84707355 = []byte{
	// 596 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xac, 0x94, 0xdf, 0x6e, 0xd3, 0x30,
	0x14, 0xc6, 0x09, 0x1b, 0x6b, 0xe7, 0xf4, 0x1f, 0x29, 0x17, 0xa1, 0x48, 0x6b, 0x35, 0x60, 0xaa,
	0x90, 0x48, 0x24, 0xb8, 0x46, 0x48, 0x6d, 0x2f, 0x1a, 0xa9, 0x42, 0x53, 0x86, 0x98, 0xc4, 0x4d,
	0xe4, 0x24, 0x6e, 0x62, 0x91, 0xc6, 0x51, 0xec, 0x4e, 0xec, 0x29, 0xe0, 0x1d, 0x78, 0x19, 0xc4,
	0x43, 0x14, 0x69, 0x4f, 0x82, 0x7c, 0x9c, 0xa4, 0xe9, 0xc6, 0xc5, 0xc6, 0xb8, 0xab, 0xce, 0x77,
	0xce, 0xcf, 0x9f, 0x4f, 0x3f, 0x07, 0xe9, 0xe2, 0x32, 0x23, 0xdc, 0xca, 0x72, 0x26, 0x98, 0xd1,
	0x13, 0x24, 0x0d, 0x49, 0xbe, 0xa2, 0xa9, 0xb0, 0xb8, 0xc0, 0x82, 0x0c, 0x86, 0x11, 0x63, 0x51,
	0x42, 0x6c, 0xd0, 0xfd, 0xf5, 0xd2, 0x16, 0x74, 0x45, 0xb8, 0xc0, 0xab, 0x4c, 0x8d, 0x0c, 0x9e,
	0x44, 0x2c, 0x62, 0xf0, 0xd3, 0x96, 0xbf, 0x8a, 0xea, 0x3c, 0xa2, 0x22, 0x5e, 0xfb, 0x56, 0xc0,
	0x56, 0x76, 0x7c, 0x99, 0x91, 0x3c, 0x21, 0x61, 0x44, 0x72, 0x7b, 0x89, 0xfd, 0x9c, 0x06, 0xaf,
	0xc3, 0x9c, 0x65, 0x09, 0xf5, 0x15, 0x95, 0x57, 0x70, 0x6e, 0x83, 0x19, 0xbb, 0x66, 0x69, 0xe0,
	0xdc, 0x8b, 0x94, 0xe1, 0x1c, 0xaf, 0x4a, 0xd4, 0xe2, 0x5e, 0xa8, 0x0b, 0x9c, 0xd0, 0x10, 0x0b,
	0x96, 0x2b, 0xda, 0xf1, 0x37, 0x0d, 0xf5, 0x3f, 0x95, 0x35, 0xee, 0xa4, 0x4b, 0x76, 0x0a, 0x3b,
	0x9c, 0xa3, 0x76, 0xd5, 0xea, 0x71, 0x22, 0x4c, 0x6d, 0xa4, 0x8d, 0xf5, 0x37, 0xcf, 0xad, 0xda,
	0x6e, 0xd5, 0x05, 0xab, 0xe9, 0x33, 0x22, 0x60, 0xd6, 0x6d, 0x5d, 0xd4, 0x4a, 0x86, 0x85, 0xfa,
	0x09, 0xe6, 0xc2, 0x8b, 0x09, 0x8d, 0x62, 0xe1, 0x05, 0x31, 0x4e, 0x23, 0x12, 0x9a, 0x0f, 0x47,
	0xda, 0x78, 0xcf, 0x7d, 0x2c, 0xa5, 0x39, 0x28, 0x53, 0x25, 0x1c, 0xff, 0xd0, 0x90, 0x39, 0x65,
	0x29, 0x27, 0x29, 0x5f, 0xf3, 0x53, 0xb8, 0xf9, 0xd6, 0xd6, 0x39, 0xea, 0x05, 0xa5, 0xe6, 0xa9,
	0xb5, 0x14, 0xce, 0x4e, 0x6e, 0x3a, 0xbb, 0x46, 0x01, 0xc2, 0x64, 0xff, 0xe7, 0x66, 0xf8, 0xc0,
	0xed, 0x06, 0xbb, 0xda, 0x9d, 0x5d, 0xfe, 0x3a, 0x40, 0xbd, 0x05, 0x16, 0x84, 0x8b, 0x33, 0x99,
	0x30, 0xe5, 0xee, 0x04, 0x35, 0x83, 0x18, 0xd3, 0xd4, 0xa3, 0x6a, 0xf2, 0x70, 0xa2, 0x5f, 0x6d,
	0x86, 0x8d, 0xa9, 0xac, 0x39, 0x33, 0xb7, 0x01, 0xa2, 0x13, 0x1a, 0x2f, 0x51, 0x87, 0xa6, 0x54,
	0x50, 0x9c, 0x14, 0xe7, 0x99, 0x1d, 0x38, 0xa7, 0x5d, 0x54, 0xd5, 0x51, 0xc6, 0x2b, 0x04, 0x07,
	0x7b, 0x7e, 0xc2, 0x82, 0x2f, 0x65, 0xe7, 0x1e, 0x74, 0x76, 0xa5, 0x30, 0x91, 0xf5, 0xa2, 0xf7,
	0x1c, 0xb5, 0x6b, 0xbd, 0x34, 0x34, 0xf7, 0x61, 0x2b, 0x47, 0x37, 0xb7, 0x02, 0x53, 0xce, 0x4c,
	0x6d, 0xa3, 0x2f, 0xb7, 0x71, 0xb5, 0x19, 0xea, 0x8b, 0x92, 0xe7, 0xcc, 0x5c, 0xbd, 0x82, 0x3b,
	0xa1, 0xb1, 0x40, 0xdd, 0x1a, 0x58, 0xbe, 0x1b, 0xf3, 0x11, 0xa0, 0x07, 0x96, 0x7a, 0x54, 0x56,
	0x19, 0x31, 0xeb, 0x63, 0xf9, 0xa8, 0x26, 0x4d, 0x89, 0xfd, 0xfe, 0x7b, 0xa8, 0xb9, 0xed, 0x8a,
	0x25, 0x55, 0x49, 0x4b, 0xc9, 0x57, 0xe1, 0x55, 0x09, 0xe1, 0xe6, 0xc1, 0xed, 0x83, 0xd5, 0x91,
	0xb3, 0xdb, 0xb4, 0x1a, 0x53, 0x84, 0x6a, 0xa0, 0xc6, 0xed, 0x41, 0xb5, 0xb1, 0xea, 0x82, 0x35,
	0x52, 0xf3, 0x0e, 0x96, 0xe4, 0xec, 0x8e, 0xa5, 0xa3, 0x7a, 0x8e, 0xb6, 0xd0, 0x2a, 0x52, 0x87,
	0xf0, 0x07, 0x3e, 0xdb, 0x46, 0x6a, 0x3b, 0x5d, 0x84, 0xeb, 0xaf, 0x29, 0x47, 0xff, 0x23, 0xe5,
	0x1f, 0xd0, 0x8b, 0x9d, 0x94, 0x5f, 0x3b, 0xa4, 0xf2, 0xa8, 0x83, 0xc7, 0x51, 0x2d, 0xf6, 0xbb,
	0xa0, 0xd2, 0x68, 0x99, 0xd0, 0x9c, 0xf0, 0x75, 0x22, 0xb8, 0x17, 0x63, 0x1e, 0x9b, 0xad, 0x91,
	0x36, 0x6e, 0xa9, 0x84, 0xba, 0xaa, 0x3e, 0xc7, 0x3c, 0x36, 0x9e, 0xa2, 0x26, 0xce, 0x32, 0xd5,
	0xd2, 0x86, 0x96, 0x06, 0xce, 0x32, 0x29, 0x4d, 0xde, 0x7f, 0x7e, 0xf7, 0x8f, 0x1f, 0x35, 0xf8,
	0xbc, 0x73, 0xff, 0x00, 0x2a, 0x6f, 0xff, 0x04, 0x00, 0x00, 0xff, 0xff, 0x21, 0x85, 0xf0, 0x72,
	0x07, 0x06, 0x00, 0x00,
}