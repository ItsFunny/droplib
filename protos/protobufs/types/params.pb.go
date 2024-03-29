// Code generated by protoc-gen-go. DO NOT EDIT.
// source: params.proto

package types

import (
	fmt "fmt"
	_ "github.com/gogo/protobuf/gogoproto"
	proto "github.com/golang/protobuf/proto"
	durationpb "google.golang.org/protobuf/types/known/durationpb"
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

// ConsensusParams contains consensus critical parameters that determine the
// validity of blocks.
type ConsensusParamsProto struct {
	Block                *BlockParamsProto     `protobuf:"bytes,1,opt,name=block,proto3" json:"block,omitempty"`
	Evidence             *EvidenceParamsProto  `protobuf:"bytes,2,opt,name=evidence,proto3" json:"evidence,omitempty"`
	Validator            *ValidatorParamsProto `protobuf:"bytes,3,opt,name=validator,proto3" json:"validator,omitempty"`
	Version              *VersionParamsProto   `protobuf:"bytes,4,opt,name=version,proto3" json:"version,omitempty"`
	XXX_NoUnkeyedLiteral struct{}              `json:"-"`
	XXX_unrecognized     []byte                `json:"-"`
	XXX_sizecache        int32                 `json:"-"`
}

func (m *ConsensusParamsProto) Reset()         { *m = ConsensusParamsProto{} }
func (m *ConsensusParamsProto) String() string { return proto.CompactTextString(m) }
func (*ConsensusParamsProto) ProtoMessage()    {}
func (*ConsensusParamsProto) Descriptor() ([]byte, []int) {
	return fileDescriptor_8679b07c520418a1, []int{0}
}

func (m *ConsensusParamsProto) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ConsensusParamsProto.Unmarshal(m, b)
}
func (m *ConsensusParamsProto) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ConsensusParamsProto.Marshal(b, m, deterministic)
}
func (m *ConsensusParamsProto) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ConsensusParamsProto.Merge(m, src)
}
func (m *ConsensusParamsProto) XXX_Size() int {
	return xxx_messageInfo_ConsensusParamsProto.Size(m)
}
func (m *ConsensusParamsProto) XXX_DiscardUnknown() {
	xxx_messageInfo_ConsensusParamsProto.DiscardUnknown(m)
}

var xxx_messageInfo_ConsensusParamsProto proto.InternalMessageInfo

func (m *ConsensusParamsProto) GetBlock() *BlockParamsProto {
	if m != nil {
		return m.Block
	}
	return nil
}

func (m *ConsensusParamsProto) GetEvidence() *EvidenceParamsProto {
	if m != nil {
		return m.Evidence
	}
	return nil
}

func (m *ConsensusParamsProto) GetValidator() *ValidatorParamsProto {
	if m != nil {
		return m.Validator
	}
	return nil
}

func (m *ConsensusParamsProto) GetVersion() *VersionParamsProto {
	if m != nil {
		return m.Version
	}
	return nil
}

// BlockParams contains limits on the block size.
type BlockParamsProto struct {
	// Max block size, in bytes.
	// Note: must be greater than 0
	MaxBytes int64 `protobuf:"varint,1,opt,name=max_bytes,json=maxBytes,proto3" json:"max_bytes,omitempty"`
	// Max gas per block.
	// Note: must be greater or equal to -1
	MaxGas               int64    `protobuf:"varint,2,opt,name=max_gas,json=maxGas,proto3" json:"max_gas,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *BlockParamsProto) Reset()         { *m = BlockParamsProto{} }
func (m *BlockParamsProto) String() string { return proto.CompactTextString(m) }
func (*BlockParamsProto) ProtoMessage()    {}
func (*BlockParamsProto) Descriptor() ([]byte, []int) {
	return fileDescriptor_8679b07c520418a1, []int{1}
}

func (m *BlockParamsProto) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_BlockParamsProto.Unmarshal(m, b)
}
func (m *BlockParamsProto) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_BlockParamsProto.Marshal(b, m, deterministic)
}
func (m *BlockParamsProto) XXX_Merge(src proto.Message) {
	xxx_messageInfo_BlockParamsProto.Merge(m, src)
}
func (m *BlockParamsProto) XXX_Size() int {
	return xxx_messageInfo_BlockParamsProto.Size(m)
}
func (m *BlockParamsProto) XXX_DiscardUnknown() {
	xxx_messageInfo_BlockParamsProto.DiscardUnknown(m)
}

var xxx_messageInfo_BlockParamsProto proto.InternalMessageInfo

func (m *BlockParamsProto) GetMaxBytes() int64 {
	if m != nil {
		return m.MaxBytes
	}
	return 0
}

func (m *BlockParamsProto) GetMaxGas() int64 {
	if m != nil {
		return m.MaxGas
	}
	return 0
}

// EvidenceParams determine how we handle evidence of malfeasance.
type EvidenceParamsProto struct {
	// Max age of evidence, in blocks.
	//
	// The basic formula for calculating this is: MaxAgeDuration / {average block
	// time}.
	MaxAgeNumBlocks int64 `protobuf:"varint,1,opt,name=max_age_num_blocks,json=maxAgeNumBlocks,proto3" json:"max_age_num_blocks,omitempty"`
	// Max age of evidence, in time.
	//
	// It should correspond with an app's "unbonding period" or other similar
	// mechanism for handling [Nothing-At-Stake
	// attacks](https://github.com/ethereum/wiki/wiki/Proof-of-Stake-FAQ#what-is-the-nothing-at-stake-problem-and-how-can-it-be-fixed).
	MaxAgeDuration *durationpb.Duration `protobuf:"bytes,2,opt,name=max_age_duration,json=maxAgeDuration,proto3" json:"max_age_duration,omitempty"`
	// This sets the maximum size of total evidence in bytes that can be committed in a single block.
	// and should fall comfortably under the max block bytes.
	// Default is 1048576 or 1MB
	MaxBytes             int64    `protobuf:"varint,3,opt,name=max_bytes,json=maxBytes,proto3" json:"max_bytes,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *EvidenceParamsProto) Reset()         { *m = EvidenceParamsProto{} }
func (m *EvidenceParamsProto) String() string { return proto.CompactTextString(m) }
func (*EvidenceParamsProto) ProtoMessage()    {}
func (*EvidenceParamsProto) Descriptor() ([]byte, []int) {
	return fileDescriptor_8679b07c520418a1, []int{2}
}

func (m *EvidenceParamsProto) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_EvidenceParamsProto.Unmarshal(m, b)
}
func (m *EvidenceParamsProto) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_EvidenceParamsProto.Marshal(b, m, deterministic)
}
func (m *EvidenceParamsProto) XXX_Merge(src proto.Message) {
	xxx_messageInfo_EvidenceParamsProto.Merge(m, src)
}
func (m *EvidenceParamsProto) XXX_Size() int {
	return xxx_messageInfo_EvidenceParamsProto.Size(m)
}
func (m *EvidenceParamsProto) XXX_DiscardUnknown() {
	xxx_messageInfo_EvidenceParamsProto.DiscardUnknown(m)
}

var xxx_messageInfo_EvidenceParamsProto proto.InternalMessageInfo

func (m *EvidenceParamsProto) GetMaxAgeNumBlocks() int64 {
	if m != nil {
		return m.MaxAgeNumBlocks
	}
	return 0
}

func (m *EvidenceParamsProto) GetMaxAgeDuration() *durationpb.Duration {
	if m != nil {
		return m.MaxAgeDuration
	}
	return nil
}

func (m *EvidenceParamsProto) GetMaxBytes() int64 {
	if m != nil {
		return m.MaxBytes
	}
	return 0
}

// ValidatorParams restrict the public key types validators can use.
// NOTE: uses ABCI pubkey naming, not Amino names.
type ValidatorParamsProto struct {
	PubKeyTypes          []string `protobuf:"bytes,1,rep,name=pub_key_types,json=pubKeyTypes,proto3" json:"pub_key_types,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ValidatorParamsProto) Reset()         { *m = ValidatorParamsProto{} }
func (m *ValidatorParamsProto) String() string { return proto.CompactTextString(m) }
func (*ValidatorParamsProto) ProtoMessage()    {}
func (*ValidatorParamsProto) Descriptor() ([]byte, []int) {
	return fileDescriptor_8679b07c520418a1, []int{3}
}

func (m *ValidatorParamsProto) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ValidatorParamsProto.Unmarshal(m, b)
}
func (m *ValidatorParamsProto) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ValidatorParamsProto.Marshal(b, m, deterministic)
}
func (m *ValidatorParamsProto) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ValidatorParamsProto.Merge(m, src)
}
func (m *ValidatorParamsProto) XXX_Size() int {
	return xxx_messageInfo_ValidatorParamsProto.Size(m)
}
func (m *ValidatorParamsProto) XXX_DiscardUnknown() {
	xxx_messageInfo_ValidatorParamsProto.DiscardUnknown(m)
}

var xxx_messageInfo_ValidatorParamsProto proto.InternalMessageInfo

func (m *ValidatorParamsProto) GetPubKeyTypes() []string {
	if m != nil {
		return m.PubKeyTypes
	}
	return nil
}

// VersionParams contains the ABCI application version.
type VersionParamsProto struct {
	AppVersion           uint64   `protobuf:"varint,1,opt,name=app_version,json=appVersion,proto3" json:"app_version,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *VersionParamsProto) Reset()         { *m = VersionParamsProto{} }
func (m *VersionParamsProto) String() string { return proto.CompactTextString(m) }
func (*VersionParamsProto) ProtoMessage()    {}
func (*VersionParamsProto) Descriptor() ([]byte, []int) {
	return fileDescriptor_8679b07c520418a1, []int{4}
}

func (m *VersionParamsProto) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_VersionParamsProto.Unmarshal(m, b)
}
func (m *VersionParamsProto) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_VersionParamsProto.Marshal(b, m, deterministic)
}
func (m *VersionParamsProto) XXX_Merge(src proto.Message) {
	xxx_messageInfo_VersionParamsProto.Merge(m, src)
}
func (m *VersionParamsProto) XXX_Size() int {
	return xxx_messageInfo_VersionParamsProto.Size(m)
}
func (m *VersionParamsProto) XXX_DiscardUnknown() {
	xxx_messageInfo_VersionParamsProto.DiscardUnknown(m)
}

var xxx_messageInfo_VersionParamsProto proto.InternalMessageInfo

func (m *VersionParamsProto) GetAppVersion() uint64 {
	if m != nil {
		return m.AppVersion
	}
	return 0
}

// HashedParams is a subset of ConsensusParams.
//
// It is hashed into the Header.ConsensusHash.
type HashedParamsProto struct {
	BlockMaxBytes        int64    `protobuf:"varint,1,opt,name=block_max_bytes,json=blockMaxBytes,proto3" json:"block_max_bytes,omitempty"`
	BlockMaxGas          int64    `protobuf:"varint,2,opt,name=block_max_gas,json=blockMaxGas,proto3" json:"block_max_gas,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *HashedParamsProto) Reset()         { *m = HashedParamsProto{} }
func (m *HashedParamsProto) String() string { return proto.CompactTextString(m) }
func (*HashedParamsProto) ProtoMessage()    {}
func (*HashedParamsProto) Descriptor() ([]byte, []int) {
	return fileDescriptor_8679b07c520418a1, []int{5}
}

func (m *HashedParamsProto) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_HashedParamsProto.Unmarshal(m, b)
}
func (m *HashedParamsProto) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_HashedParamsProto.Marshal(b, m, deterministic)
}
func (m *HashedParamsProto) XXX_Merge(src proto.Message) {
	xxx_messageInfo_HashedParamsProto.Merge(m, src)
}
func (m *HashedParamsProto) XXX_Size() int {
	return xxx_messageInfo_HashedParamsProto.Size(m)
}
func (m *HashedParamsProto) XXX_DiscardUnknown() {
	xxx_messageInfo_HashedParamsProto.DiscardUnknown(m)
}

var xxx_messageInfo_HashedParamsProto proto.InternalMessageInfo

func (m *HashedParamsProto) GetBlockMaxBytes() int64 {
	if m != nil {
		return m.BlockMaxBytes
	}
	return 0
}

func (m *HashedParamsProto) GetBlockMaxGas() int64 {
	if m != nil {
		return m.BlockMaxGas
	}
	return 0
}

func init() {
	proto.RegisterType((*ConsensusParamsProto)(nil), "tendermint.types.ConsensusParamsProto")
	proto.RegisterType((*BlockParamsProto)(nil), "tendermint.types.BlockParamsProto")
	proto.RegisterType((*EvidenceParamsProto)(nil), "tendermint.types.EvidenceParamsProto")
	proto.RegisterType((*ValidatorParamsProto)(nil), "tendermint.types.ValidatorParamsProto")
	proto.RegisterType((*VersionParamsProto)(nil), "tendermint.types.VersionParamsProto")
	proto.RegisterType((*HashedParamsProto)(nil), "tendermint.types.HashedParamsProto")
}

func init() { proto.RegisterFile("params.proto", fileDescriptor_8679b07c520418a1) }

var fileDescriptor_8679b07c520418a1 = []byte{
	// 490 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x6c, 0x53, 0x6f, 0x6b, 0xd3, 0x40,
	0x18, 0x37, 0xeb, 0xdc, 0xda, 0xa7, 0xd6, 0xd5, 0xb3, 0x60, 0x9c, 0xb0, 0x8e, 0xa0, 0x63, 0x20,
	0x26, 0xa0, 0x08, 0x22, 0x22, 0x2c, 0x4e, 0x36, 0x90, 0xc9, 0x08, 0xe2, 0x0b, 0xdf, 0x84, 0x4b,
	0xf3, 0x2c, 0x0d, 0xcb, 0xe5, 0x8e, 0xbb, 0xa4, 0x34, 0xdf, 0x44, 0xbf, 0x81, 0xaf, 0xfc, 0x1c,
	0x7e, 0x0a, 0x05, 0x3f, 0x89, 0xe4, 0xd2, 0x5b, 0xb3, 0xb6, 0xef, 0x72, 0xf7, 0xfb, 0xf3, 0xe4,
	0xf7, 0x7b, 0x38, 0xb8, 0x27, 0xa8, 0xa4, 0x4c, 0xb9, 0x42, 0xf2, 0x82, 0x93, 0x61, 0x81, 0x79,
	0x8c, 0x92, 0xa5, 0x79, 0xe1, 0x16, 0x95, 0x40, 0xb5, 0x3f, 0x4a, 0x78, 0xc2, 0x35, 0xe8, 0xd5,
	0x5f, 0x0d, 0x6f, 0xff, 0x20, 0xe1, 0x3c, 0xc9, 0xd0, 0xd3, 0xa7, 0xa8, 0xbc, 0xf2, 0xe2, 0x52,
	0xd2, 0x22, 0xe5, 0x79, 0x83, 0x3b, 0x3f, 0xb6, 0x60, 0xf4, 0x81, 0xe7, 0x0a, 0x73, 0x55, 0xaa,
	0x4b, 0x3d, 0xe1, 0x52, 0x0f, 0x78, 0x03, 0x77, 0xa3, 0x8c, 0x4f, 0xae, 0x6d, 0xeb, 0xd0, 0x3a,
	0xee, 0xbf, 0x74, 0xdc, 0xd5, 0x81, 0xae, 0x5f, 0xc3, 0x2d, 0x49, 0xd0, 0x08, 0xc8, 0x09, 0x74,
	0x71, 0x96, 0xc6, 0x98, 0x4f, 0xd0, 0xde, 0xd2, 0xe2, 0x67, 0xeb, 0xe2, 0x8f, 0x0b, 0x46, 0x5b,
	0x7f, 0x23, 0x23, 0xa7, 0xd0, 0x9b, 0xd1, 0x2c, 0x8d, 0x69, 0xc1, 0xa5, 0xdd, 0xd1, 0x1e, 0x47,
	0xeb, 0x1e, 0x5f, 0x0d, 0xa5, 0x6d, 0xb2, 0x14, 0x92, 0xf7, 0xb0, 0x3b, 0x43, 0xa9, 0x52, 0x9e,
	0xdb, 0xdb, 0xda, 0xe3, 0xe9, 0x06, 0x8f, 0x86, 0xd0, 0x76, 0x30, 0x22, 0xe7, 0x1c, 0x86, 0xab,
	0x19, 0xc9, 0x13, 0xe8, 0x31, 0x3a, 0x0f, 0xa3, 0xaa, 0x40, 0xa5, 0xab, 0xe9, 0x04, 0x5d, 0x46,
	0xe7, 0x7e, 0x7d, 0x26, 0x8f, 0x60, 0xb7, 0x06, 0x13, 0xaa, 0x74, 0xf0, 0x4e, 0xb0, 0xc3, 0xe8,
	0xfc, 0x8c, 0x2a, 0xe7, 0x97, 0x05, 0x0f, 0x37, 0x24, 0x26, 0xcf, 0x81, 0xd4, 0x02, 0x9a, 0x60,
	0x98, 0x97, 0x2c, 0xd4, 0xfd, 0x19, 0xdb, 0x3d, 0x46, 0xe7, 0x27, 0x09, 0x7e, 0x2e, 0x99, 0xfe,
	0x09, 0x45, 0x2e, 0x60, 0x68, 0xc8, 0x66, 0x89, 0x8b, 0x7e, 0x1f, 0xbb, 0xcd, 0x96, 0x5d, 0xb3,
	0x65, 0xf7, 0x74, 0x41, 0xf0, 0xbb, 0xbf, 0xff, 0x8c, 0xef, 0x7c, 0xff, 0x3b, 0xb6, 0x82, 0xfb,
	0x8d, 0x9f, 0x41, 0x6e, 0x27, 0xe9, 0xdc, 0x4e, 0xe2, 0xbc, 0x85, 0xd1, 0xa6, 0x76, 0x89, 0x03,
	0x03, 0x51, 0x46, 0xe1, 0x35, 0x56, 0xa1, 0xee, 0xcf, 0xb6, 0x0e, 0x3b, 0xc7, 0xbd, 0xa0, 0x2f,
	0xca, 0xe8, 0x13, 0x56, 0x5f, 0xea, 0x2b, 0xe7, 0x35, 0x90, 0xf5, 0x56, 0xc9, 0x18, 0xfa, 0x54,
	0x88, 0xd0, 0x2c, 0xa4, 0xce, 0xb8, 0x1d, 0x00, 0x15, 0x62, 0xc1, 0x75, 0x42, 0x78, 0x70, 0x4e,
	0xd5, 0x14, 0xe3, 0xb6, 0xea, 0x08, 0xf6, 0x74, 0x29, 0xe1, 0x6a, 0xe9, 0x03, 0x7d, 0x7d, 0x61,
	0x9a, 0x77, 0x60, 0xb0, 0xe4, 0x2d, 0xfb, 0xef, 0x1b, 0xd6, 0x19, 0x55, 0xbe, 0xff, 0xf3, 0xdf,
	0x81, 0xf5, 0xed, 0x5d, 0x92, 0x16, 0xd3, 0x32, 0x72, 0x27, 0x9c, 0x79, 0xd3, 0x4a, 0xa0, 0xcc,
	0x30, 0x4e, 0x50, 0x7a, 0x57, 0x34, 0x92, 0xe9, 0xe4, 0x45, 0x2c, 0xb9, 0xc8, 0xd2, 0xa8, 0x79,
	0x2f, 0xea, 0xe6, 0xd9, 0x28, 0x4f, 0xc7, 0x8d, 0x76, 0xf4, 0xc5, 0xab, 0xff, 0x01, 0x00, 0x00,
	0xff, 0xff, 0x27, 0xd7, 0x5c, 0x23, 0x8d, 0x03, 0x00, 0x00,
}
