// Code generated by protoc-gen-go. DO NOT EDIT.
// source: go/pkg/balancer/proto/grpcbalancer.proto

package grpcbalancer

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

type AffinityConfig_Command int32

const (
	AffinityConfig_BOUND  AffinityConfig_Command = 0
	AffinityConfig_BIND   AffinityConfig_Command = 1
	AffinityConfig_UNBIND AffinityConfig_Command = 2
)

var AffinityConfig_Command_name = map[int32]string{
	0: "BOUND",
	1: "BIND",
	2: "UNBIND",
}

var AffinityConfig_Command_value = map[string]int32{
	"BOUND":  0,
	"BIND":   1,
	"UNBIND": 2,
}

func (x AffinityConfig_Command) String() string {
	return proto.EnumName(AffinityConfig_Command_name, int32(x))
}

func (AffinityConfig_Command) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_aebddaecf1fa8cdb, []int{3, 0}
}

type ApiConfig struct {
	ChannelPool          *ChannelPoolConfig `protobuf:"bytes,2,opt,name=channel_pool,json=channelPool,proto3" json:"channel_pool,omitempty"`
	Method               []*MethodConfig    `protobuf:"bytes,1001,rep,name=method,proto3" json:"method,omitempty"`
	XXX_NoUnkeyedLiteral struct{}           `json:"-"`
	XXX_unrecognized     []byte             `json:"-"`
	XXX_sizecache        int32              `json:"-"`
}

func (m *ApiConfig) Reset()         { *m = ApiConfig{} }
func (m *ApiConfig) String() string { return proto.CompactTextString(m) }
func (*ApiConfig) ProtoMessage()    {}
func (*ApiConfig) Descriptor() ([]byte, []int) {
	return fileDescriptor_aebddaecf1fa8cdb, []int{0}
}

func (m *ApiConfig) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ApiConfig.Unmarshal(m, b)
}
func (m *ApiConfig) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ApiConfig.Marshal(b, m, deterministic)
}
func (m *ApiConfig) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ApiConfig.Merge(m, src)
}
func (m *ApiConfig) XXX_Size() int {
	return xxx_messageInfo_ApiConfig.Size(m)
}
func (m *ApiConfig) XXX_DiscardUnknown() {
	xxx_messageInfo_ApiConfig.DiscardUnknown(m)
}

var xxx_messageInfo_ApiConfig proto.InternalMessageInfo

func (m *ApiConfig) GetChannelPool() *ChannelPoolConfig {
	if m != nil {
		return m.ChannelPool
	}
	return nil
}

func (m *ApiConfig) GetMethod() []*MethodConfig {
	if m != nil {
		return m.Method
	}
	return nil
}

type ChannelPoolConfig struct {
	MaxSize                          uint32   `protobuf:"varint,1,opt,name=max_size,json=maxSize,proto3" json:"max_size,omitempty"`
	IdleTimeout                      uint64   `protobuf:"varint,2,opt,name=idle_timeout,json=idleTimeout,proto3" json:"idle_timeout,omitempty"`
	MaxConcurrentStreamsLowWatermark uint32   `protobuf:"varint,3,opt,name=max_concurrent_streams_low_watermark,json=maxConcurrentStreamsLowWatermark,proto3" json:"max_concurrent_streams_low_watermark,omitempty"`
	XXX_NoUnkeyedLiteral             struct{} `json:"-"`
	XXX_unrecognized                 []byte   `json:"-"`
	XXX_sizecache                    int32    `json:"-"`
}

func (m *ChannelPoolConfig) Reset()         { *m = ChannelPoolConfig{} }
func (m *ChannelPoolConfig) String() string { return proto.CompactTextString(m) }
func (*ChannelPoolConfig) ProtoMessage()    {}
func (*ChannelPoolConfig) Descriptor() ([]byte, []int) {
	return fileDescriptor_aebddaecf1fa8cdb, []int{1}
}

func (m *ChannelPoolConfig) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ChannelPoolConfig.Unmarshal(m, b)
}
func (m *ChannelPoolConfig) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ChannelPoolConfig.Marshal(b, m, deterministic)
}
func (m *ChannelPoolConfig) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ChannelPoolConfig.Merge(m, src)
}
func (m *ChannelPoolConfig) XXX_Size() int {
	return xxx_messageInfo_ChannelPoolConfig.Size(m)
}
func (m *ChannelPoolConfig) XXX_DiscardUnknown() {
	xxx_messageInfo_ChannelPoolConfig.DiscardUnknown(m)
}

var xxx_messageInfo_ChannelPoolConfig proto.InternalMessageInfo

func (m *ChannelPoolConfig) GetMaxSize() uint32 {
	if m != nil {
		return m.MaxSize
	}
	return 0
}

func (m *ChannelPoolConfig) GetIdleTimeout() uint64 {
	if m != nil {
		return m.IdleTimeout
	}
	return 0
}

func (m *ChannelPoolConfig) GetMaxConcurrentStreamsLowWatermark() uint32 {
	if m != nil {
		return m.MaxConcurrentStreamsLowWatermark
	}
	return 0
}

type MethodConfig struct {
	Name                 []string        `protobuf:"bytes,1,rep,name=name,proto3" json:"name,omitempty"`
	Affinity             *AffinityConfig `protobuf:"bytes,1001,opt,name=affinity,proto3" json:"affinity,omitempty"`
	XXX_NoUnkeyedLiteral struct{}        `json:"-"`
	XXX_unrecognized     []byte          `json:"-"`
	XXX_sizecache        int32           `json:"-"`
}

func (m *MethodConfig) Reset()         { *m = MethodConfig{} }
func (m *MethodConfig) String() string { return proto.CompactTextString(m) }
func (*MethodConfig) ProtoMessage()    {}
func (*MethodConfig) Descriptor() ([]byte, []int) {
	return fileDescriptor_aebddaecf1fa8cdb, []int{2}
}

func (m *MethodConfig) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_MethodConfig.Unmarshal(m, b)
}
func (m *MethodConfig) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_MethodConfig.Marshal(b, m, deterministic)
}
func (m *MethodConfig) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MethodConfig.Merge(m, src)
}
func (m *MethodConfig) XXX_Size() int {
	return xxx_messageInfo_MethodConfig.Size(m)
}
func (m *MethodConfig) XXX_DiscardUnknown() {
	xxx_messageInfo_MethodConfig.DiscardUnknown(m)
}

var xxx_messageInfo_MethodConfig proto.InternalMessageInfo

func (m *MethodConfig) GetName() []string {
	if m != nil {
		return m.Name
	}
	return nil
}

func (m *MethodConfig) GetAffinity() *AffinityConfig {
	if m != nil {
		return m.Affinity
	}
	return nil
}

type AffinityConfig struct {
	Command              AffinityConfig_Command `protobuf:"varint,2,opt,name=command,proto3,enum=grpcbalancer.AffinityConfig_Command" json:"command,omitempty"`
	AffinityKey          string                 `protobuf:"bytes,3,opt,name=affinity_key,json=affinityKey,proto3" json:"affinity_key,omitempty"`
	XXX_NoUnkeyedLiteral struct{}               `json:"-"`
	XXX_unrecognized     []byte                 `json:"-"`
	XXX_sizecache        int32                  `json:"-"`
}

func (m *AffinityConfig) Reset()         { *m = AffinityConfig{} }
func (m *AffinityConfig) String() string { return proto.CompactTextString(m) }
func (*AffinityConfig) ProtoMessage()    {}
func (*AffinityConfig) Descriptor() ([]byte, []int) {
	return fileDescriptor_aebddaecf1fa8cdb, []int{3}
}

func (m *AffinityConfig) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_AffinityConfig.Unmarshal(m, b)
}
func (m *AffinityConfig) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_AffinityConfig.Marshal(b, m, deterministic)
}
func (m *AffinityConfig) XXX_Merge(src proto.Message) {
	xxx_messageInfo_AffinityConfig.Merge(m, src)
}
func (m *AffinityConfig) XXX_Size() int {
	return xxx_messageInfo_AffinityConfig.Size(m)
}
func (m *AffinityConfig) XXX_DiscardUnknown() {
	xxx_messageInfo_AffinityConfig.DiscardUnknown(m)
}

var xxx_messageInfo_AffinityConfig proto.InternalMessageInfo

func (m *AffinityConfig) GetCommand() AffinityConfig_Command {
	if m != nil {
		return m.Command
	}
	return AffinityConfig_BOUND
}

func (m *AffinityConfig) GetAffinityKey() string {
	if m != nil {
		return m.AffinityKey
	}
	return ""
}

func init() {
	proto.RegisterEnum("grpcbalancer.AffinityConfig_Command", AffinityConfig_Command_name, AffinityConfig_Command_value)
	proto.RegisterType((*ApiConfig)(nil), "grpcbalancer.ApiConfig")
	proto.RegisterType((*ChannelPoolConfig)(nil), "grpcbalancer.ChannelPoolConfig")
	proto.RegisterType((*MethodConfig)(nil), "grpcbalancer.MethodConfig")
	proto.RegisterType((*AffinityConfig)(nil), "grpcbalancer.AffinityConfig")
}

func init() {
	proto.RegisterFile("go/pkg/balancer/proto/grpcbalancer.proto", fileDescriptor_aebddaecf1fa8cdb)
}

var fileDescriptor_aebddaecf1fa8cdb = []byte{
	// 386 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x7c, 0x92, 0x4f, 0xab, 0xd3, 0x40,
	0x14, 0xc5, 0xcd, 0x6b, 0x6d, 0x9a, 0x9b, 0xf8, 0x78, 0xce, 0x2a, 0x8a, 0x60, 0x0c, 0x6f, 0x11,
	0x5c, 0x34, 0xd0, 0xb7, 0x72, 0x23, 0xb4, 0xe9, 0x46, 0xd4, 0x28, 0xa9, 0xc5, 0x95, 0x84, 0x69,
	0x3a, 0x4d, 0x87, 0x66, 0x66, 0xc2, 0x74, 0x4a, 0xff, 0xac, 0xfd, 0x1e, 0xe2, 0xc7, 0xf3, 0x5b,
	0x48, 0x26, 0x49, 0x6d, 0x10, 0xde, 0x6e, 0x38, 0xf7, 0x77, 0x0e, 0xf7, 0x5c, 0x06, 0x82, 0x5c,
	0x84, 0xe5, 0x36, 0x0f, 0x97, 0xb8, 0xc0, 0x3c, 0x23, 0x32, 0x2c, 0xa5, 0x50, 0x22, 0xcc, 0x65,
	0x99, 0xb5, 0xd2, 0x48, 0x4b, 0xc8, 0xb9, 0xd6, 0xfc, 0x9f, 0x06, 0x58, 0x93, 0x92, 0x46, 0x82,
	0xaf, 0x69, 0x8e, 0xa6, 0xe0, 0x64, 0x1b, 0xcc, 0x39, 0x29, 0xd2, 0x52, 0x88, 0xc2, 0xbd, 0xf1,
	0x8c, 0xc0, 0x1e, 0xbf, 0x1e, 0x75, 0x62, 0xa2, 0x9a, 0xf8, 0x2a, 0x44, 0x51, 0xdb, 0x12, 0x3b,
	0xfb, 0x27, 0xa1, 0x07, 0x18, 0x30, 0xa2, 0x36, 0x62, 0xe5, 0xfe, 0x31, 0xbd, 0x5e, 0x60, 0x8f,
	0x5f, 0x76, 0xed, 0x9f, 0xf5, 0xb0, 0x71, 0x36, 0xa8, 0xff, 0xdb, 0x80, 0xe7, 0xff, 0xe5, 0xa2,
	0x17, 0x30, 0x64, 0xf8, 0x98, 0xee, 0xe8, 0x99, 0xb8, 0x86, 0x67, 0x04, 0xcf, 0x12, 0x93, 0xe1,
	0xe3, 0x9c, 0x9e, 0x09, 0x7a, 0x03, 0x0e, 0x5d, 0x15, 0x24, 0x55, 0x94, 0x11, 0xb1, 0x57, 0x7a,
	0xd3, 0x7e, 0x62, 0x57, 0xda, 0xb7, 0x5a, 0x42, 0x31, 0xdc, 0x57, 0xee, 0x4c, 0xf0, 0x6c, 0x2f,
	0x25, 0xe1, 0x2a, 0xdd, 0x29, 0x49, 0x30, 0xdb, 0xa5, 0x85, 0x38, 0xa4, 0x07, 0xac, 0x88, 0x64,
	0x58, 0x6e, 0xdd, 0x9e, 0x4e, 0xf6, 0x18, 0x3e, 0x46, 0x17, 0x74, 0x5e, 0x93, 0x9f, 0xc4, 0xe1,
	0x7b, 0xcb, 0xf9, 0x3f, 0xc0, 0xb9, 0xde, 0x1d, 0x21, 0xe8, 0x73, 0xcc, 0xaa, 0xcd, 0x7a, 0x81,
	0x95, 0xe8, 0x37, 0x7a, 0x07, 0x43, 0xbc, 0x5e, 0x53, 0x4e, 0xd5, 0xa9, 0xaa, 0x5f, 0x5d, 0xef,
	0x55, 0xb7, 0xfe, 0xa4, 0x19, 0x37, 0x07, 0xb8, 0xe0, 0xfe, 0x2f, 0x03, 0x6e, 0xbb, 0x43, 0xf4,
	0x1e, 0xcc, 0x4c, 0x30, 0x86, 0xf9, 0x4a, 0xf7, 0xbb, 0x1d, 0xdf, 0x3f, 0x96, 0x35, 0x8a, 0x6a,
	0x36, 0x69, 0x4d, 0xd5, 0x91, 0xda, 0xf8, 0x74, 0x4b, 0x4e, 0xba, 0xa9, 0x95, 0xd8, 0xad, 0xf6,
	0x91, 0x9c, 0xfc, 0xb7, 0x60, 0x36, 0x36, 0x64, 0xc1, 0xd3, 0xe9, 0x97, 0x45, 0x3c, 0xbb, 0x7b,
	0x82, 0x86, 0xd0, 0x9f, 0x7e, 0x88, 0x67, 0x77, 0x06, 0x02, 0x18, 0x2c, 0x62, 0xfd, 0xbe, 0x59,
	0x0e, 0xf4, 0x07, 0x7a, 0xf8, 0x1b, 0x00, 0x00, 0xff, 0xff, 0x38, 0x6e, 0x53, 0xd4, 0x6c, 0x02,
	0x00, 0x00,
}
