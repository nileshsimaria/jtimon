// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: sr_te_bsid_stats.proto

package schema

import (
	fmt "fmt"
	github_com_gogo_protobuf_proto "github.com/gogo/protobuf/proto"
	proto "github.com/gogo/protobuf/proto"
	io "io"
	math "math"
	math_bits "math/bits"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion3 // please upgrade the proto package

//
// Top-level message
//
type SrTeBsidStats struct {
	// List of SR TE stats per BSID, IP and Color records
	TeBsidStats          []*SegmentRoutingTeBsidRecord `protobuf:"bytes,1,rep,name=te_bsid_stats,json=teBsidStats" json:"te_bsid_stats,omitempty"`
	XXX_NoUnkeyedLiteral struct{}                      `json:"-"`
	XXX_unrecognized     []byte                        `json:"-"`
	XXX_sizecache        int32                         `json:"-"`
}

func (m *SrTeBsidStats) Reset()         { *m = SrTeBsidStats{} }
func (m *SrTeBsidStats) String() string { return proto.CompactTextString(m) }
func (*SrTeBsidStats) ProtoMessage()    {}
func (*SrTeBsidStats) Descriptor() ([]byte, []int) {
	return fileDescriptor_c99674bf69207578, []int{0}
}
func (m *SrTeBsidStats) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *SrTeBsidStats) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_SrTeBsidStats.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *SrTeBsidStats) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SrTeBsidStats.Merge(m, src)
}
func (m *SrTeBsidStats) XXX_Size() int {
	return m.Size()
}
func (m *SrTeBsidStats) XXX_DiscardUnknown() {
	xxx_messageInfo_SrTeBsidStats.DiscardUnknown(m)
}

var xxx_messageInfo_SrTeBsidStats proto.InternalMessageInfo

func (m *SrTeBsidStats) GetTeBsidStats() []*SegmentRoutingTeBsidRecord {
	if m != nil {
		return m.TeBsidStats
	}
	return nil
}

//
// SR TE BSID statistics record
//
type SegmentRoutingTeBsidRecord struct {
	// Name of the BSID
	BsidIdentifier *uint64 `protobuf:"varint,1,req,name=bsid_identifier,json=bsidIdentifier" json:"bsid_identifier,omitempty"`
	// Ip prefix of endpoint
	ToIpPrefix *string `protobuf:"bytes,2,req,name=to_ip_prefix,json=toIpPrefix" json:"to_ip_prefix,omitempty"`
	// Policy color value
	ColorIdentifier *uint32 `protobuf:"varint,3,opt,name=color_identifier,json=colorIdentifier" json:"color_identifier,omitempty"`
	// Instance Identifier for cases when RPD creates multiple instances
	InstanceIdentifier *uint32 `protobuf:"varint,4,opt,name=instance_identifier,json=instanceIdentifier" json:"instance_identifier,omitempty"`
	// Name of the counter. This is useful when an SR label has multiple counters.
	// For some scenarios like routing restart, it is possible that a new counter is
	// created in the hardware.
	CounterName *string `protobuf:"bytes,5,opt,name=counter_name,json=counterName" json:"counter_name,omitempty"`
	// Statistics
	Stats                *SegmentRoutingTeBsidStats `protobuf:"bytes,6,opt,name=stats" json:"stats,omitempty"`
	XXX_NoUnkeyedLiteral struct{}                   `json:"-"`
	XXX_unrecognized     []byte                     `json:"-"`
	XXX_sizecache        int32                      `json:"-"`
}

func (m *SegmentRoutingTeBsidRecord) Reset()         { *m = SegmentRoutingTeBsidRecord{} }
func (m *SegmentRoutingTeBsidRecord) String() string { return proto.CompactTextString(m) }
func (*SegmentRoutingTeBsidRecord) ProtoMessage()    {}
func (*SegmentRoutingTeBsidRecord) Descriptor() ([]byte, []int) {
	return fileDescriptor_c99674bf69207578, []int{1}
}
func (m *SegmentRoutingTeBsidRecord) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *SegmentRoutingTeBsidRecord) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_SegmentRoutingTeBsidRecord.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *SegmentRoutingTeBsidRecord) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SegmentRoutingTeBsidRecord.Merge(m, src)
}
func (m *SegmentRoutingTeBsidRecord) XXX_Size() int {
	return m.Size()
}
func (m *SegmentRoutingTeBsidRecord) XXX_DiscardUnknown() {
	xxx_messageInfo_SegmentRoutingTeBsidRecord.DiscardUnknown(m)
}

var xxx_messageInfo_SegmentRoutingTeBsidRecord proto.InternalMessageInfo

func (m *SegmentRoutingTeBsidRecord) GetBsidIdentifier() uint64 {
	if m != nil && m.BsidIdentifier != nil {
		return *m.BsidIdentifier
	}
	return 0
}

func (m *SegmentRoutingTeBsidRecord) GetToIpPrefix() string {
	if m != nil && m.ToIpPrefix != nil {
		return *m.ToIpPrefix
	}
	return ""
}

func (m *SegmentRoutingTeBsidRecord) GetColorIdentifier() uint32 {
	if m != nil && m.ColorIdentifier != nil {
		return *m.ColorIdentifier
	}
	return 0
}

func (m *SegmentRoutingTeBsidRecord) GetInstanceIdentifier() uint32 {
	if m != nil && m.InstanceIdentifier != nil {
		return *m.InstanceIdentifier
	}
	return 0
}

func (m *SegmentRoutingTeBsidRecord) GetCounterName() string {
	if m != nil && m.CounterName != nil {
		return *m.CounterName
	}
	return ""
}

func (m *SegmentRoutingTeBsidRecord) GetStats() *SegmentRoutingTeBsidStats {
	if m != nil {
		return m.Stats
	}
	return nil
}

type SegmentRoutingTeBsidStats struct {
	// Packet and Byte statistics
	Packets *uint64 `protobuf:"varint,1,opt,name=packets" json:"packets,omitempty"`
	Bytes   *uint64 `protobuf:"varint,2,opt,name=bytes" json:"bytes,omitempty"`
	// Rates of the above counters
	PacketRate           *uint64  `protobuf:"varint,3,opt,name=packet_rate,json=packetRate" json:"packet_rate,omitempty"`
	ByteRate             *uint64  `protobuf:"varint,4,opt,name=byte_rate,json=byteRate" json:"byte_rate,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *SegmentRoutingTeBsidStats) Reset()         { *m = SegmentRoutingTeBsidStats{} }
func (m *SegmentRoutingTeBsidStats) String() string { return proto.CompactTextString(m) }
func (*SegmentRoutingTeBsidStats) ProtoMessage()    {}
func (*SegmentRoutingTeBsidStats) Descriptor() ([]byte, []int) {
	return fileDescriptor_c99674bf69207578, []int{2}
}
func (m *SegmentRoutingTeBsidStats) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *SegmentRoutingTeBsidStats) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_SegmentRoutingTeBsidStats.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *SegmentRoutingTeBsidStats) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SegmentRoutingTeBsidStats.Merge(m, src)
}
func (m *SegmentRoutingTeBsidStats) XXX_Size() int {
	return m.Size()
}
func (m *SegmentRoutingTeBsidStats) XXX_DiscardUnknown() {
	xxx_messageInfo_SegmentRoutingTeBsidStats.DiscardUnknown(m)
}

var xxx_messageInfo_SegmentRoutingTeBsidStats proto.InternalMessageInfo

func (m *SegmentRoutingTeBsidStats) GetPackets() uint64 {
	if m != nil && m.Packets != nil {
		return *m.Packets
	}
	return 0
}

func (m *SegmentRoutingTeBsidStats) GetBytes() uint64 {
	if m != nil && m.Bytes != nil {
		return *m.Bytes
	}
	return 0
}

func (m *SegmentRoutingTeBsidStats) GetPacketRate() uint64 {
	if m != nil && m.PacketRate != nil {
		return *m.PacketRate
	}
	return 0
}

func (m *SegmentRoutingTeBsidStats) GetByteRate() uint64 {
	if m != nil && m.ByteRate != nil {
		return *m.ByteRate
	}
	return 0
}

var E_JnprSrTeBsidStatsExt = &proto.ExtensionDesc{
	ExtendedType:  (*JuniperNetworksSensors)(nil),
	ExtensionType: (*SrTeBsidStats)(nil),
	Field:         24,
	Name:          "jnpr_sr_te_bsid_stats_ext",
	Tag:           "bytes,24,opt,name=jnpr_sr_te_bsid_stats_ext",
	Filename:      "sr_te_bsid_stats.proto",
}

func init() {
	proto.RegisterType((*SrTeBsidStats)(nil), "SrTeBsidStats")
	proto.RegisterType((*SegmentRoutingTeBsidRecord)(nil), "SegmentRoutingTeBsidRecord")
	proto.RegisterType((*SegmentRoutingTeBsidStats)(nil), "SegmentRoutingTeBsidStats")
	proto.RegisterExtension(E_JnprSrTeBsidStatsExt)
}

func init() { proto.RegisterFile("sr_te_bsid_stats.proto", fileDescriptor_c99674bf69207578) }

var fileDescriptor_c99674bf69207578 = []byte{
	// 477 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x74, 0x92, 0xc1, 0x6e, 0x13, 0x3d,
	0x14, 0x85, 0xe5, 0x69, 0xf2, 0xff, 0xad, 0xd3, 0xb4, 0xc8, 0x45, 0x30, 0x4d, 0xa5, 0x30, 0xca,
	0x02, 0x66, 0x93, 0x99, 0xaa, 0x0b, 0x16, 0x6c, 0x40, 0x91, 0x58, 0x94, 0x45, 0x55, 0x4d, 0x58,
	0xb1, 0x71, 0x9d, 0xc9, 0xed, 0xc4, 0x49, 0xc6, 0xb6, 0xec, 0x1b, 0x48, 0xb7, 0xbc, 0x0b, 0x4f,
	0xc1, 0x0b, 0xb0, 0xe4, 0x11, 0x50, 0x9e, 0x04, 0xcd, 0xb8, 0xa1, 0x13, 0x44, 0xb7, 0xe7, 0x7c,
	0xe7, 0xc8, 0x3a, 0xbe, 0xf4, 0x99, 0xb3, 0x1c, 0x81, 0x4f, 0x9c, 0x9c, 0x72, 0x87, 0x02, 0x5d,
	0x62, 0xac, 0x46, 0xdd, 0x3b, 0x41, 0x58, 0x42, 0x09, 0x68, 0xef, 0x38, 0x6a, 0xe3, 0xc5, 0xc1,
	0x35, 0xed, 0x8e, 0xed, 0x47, 0x18, 0x39, 0x39, 0x1d, 0x57, 0x2c, 0x7b, 0x4b, 0xbb, 0x3b, 0xe1,
	0x90, 0x44, 0x7b, 0x71, 0xe7, 0xe2, 0x2c, 0x19, 0x43, 0x51, 0x82, 0xc2, 0x4c, 0xaf, 0x50, 0xaa,
	0xc2, 0x47, 0x32, 0xc8, 0xb5, 0x9d, 0x66, 0x1d, 0x7c, 0x28, 0x18, 0x7c, 0x0f, 0x68, 0xef, 0x71,
	0x96, 0x25, 0xf4, 0xb8, 0x2e, 0x97, 0x53, 0x50, 0x28, 0x6f, 0x25, 0xd8, 0x90, 0x44, 0x41, 0xdc,
	0x1a, 0xb5, 0xbf, 0xbe, 0x0b, 0xf6, 0x49, 0x76, 0x54, 0xb9, 0x97, 0x7f, 0x4c, 0xf6, 0x8a, 0x1e,
	0xa2, 0xe6, 0xd2, 0x70, 0x63, 0xe1, 0x56, 0xae, 0xc3, 0x20, 0x0a, 0xe2, 0x83, 0x2d, 0x4c, 0x51,
	0x5f, 0x9a, 0xeb, 0xda, 0x60, 0xe7, 0xf4, 0x49, 0xae, 0x97, 0xda, 0x36, 0x9b, 0xf7, 0x22, 0x12,
	0x77, 0xb7, 0xf0, 0x71, 0x6d, 0x37, 0xaa, 0x5f, 0xd3, 0x13, 0xa9, 0x1c, 0x0a, 0x95, 0x43, 0x33,
	0xd4, 0x6a, 0x86, 0xd8, 0x96, 0x68, 0xe4, 0x62, 0x7a, 0x98, 0xeb, 0x95, 0x42, 0xb0, 0x5c, 0x89,
	0x12, 0xc2, 0x76, 0x44, 0x1e, 0x9e, 0xd4, 0xb9, 0xb7, 0xae, 0x44, 0x09, 0xec, 0x9c, 0xb6, 0xfd,
	0x88, 0xff, 0x45, 0x24, 0xee, 0x5c, 0xf4, 0xfe, 0x39, 0x62, 0x3d, 0x5b, 0xe6, 0xc1, 0xc1, 0x37,
	0x42, 0x4f, 0x1f, 0x85, 0xd8, 0x0b, 0xfa, 0xbf, 0x11, 0xf9, 0x02, 0xea, 0x6f, 0x21, 0xf7, 0xa3,
	0x85, 0x24, 0xdb, 0xaa, 0xec, 0x8c, 0xb6, 0x27, 0x77, 0x08, 0x2e, 0x0c, 0x9a, 0xb6, 0xd7, 0xd8,
	0x4b, 0xda, 0xf1, 0x1c, 0xb7, 0x02, 0xa1, 0x1e, 0xc7, 0x23, 0x11, 0xc9, 0xa8, 0x77, 0x32, 0x81,
	0xc0, 0x06, 0xf4, 0xa0, 0x0a, 0x78, 0xaa, 0xd5, 0xa4, 0xf6, 0x2b, 0xbd, 0x62, 0xde, 0xdc, 0xd0,
	0xd3, 0xb9, 0x32, 0x96, 0xff, 0x7d, 0x6b, 0x1c, 0xd6, 0xc8, 0x9e, 0x27, 0x1f, 0x56, 0x4a, 0x1a,
	0xb0, 0x57, 0x80, 0x5f, 0xb4, 0x5d, 0xb8, 0x31, 0x28, 0xa7, 0xad, 0x0b, 0xc3, 0x7a, 0x86, 0xa3,
	0x64, 0xe7, 0xe4, 0xb2, 0xa7, 0x55, 0xd3, 0x8e, 0xf4, 0x7e, 0x8d, 0xa3, 0x9b, 0x1f, 0x9b, 0x3e,
	0xf9, 0xb9, 0xe9, 0x93, 0x5f, 0x9b, 0x3e, 0xf9, 0x94, 0x15, 0x12, 0x93, 0xb9, 0x6f, 0x4d, 0x14,
	0x60, 0x2a, 0x73, 0x98, 0x80, 0x2d, 0xd2, 0x19, 0x88, 0x25, 0xce, 0x26, 0x1a, 0x87, 0x52, 0x15,
	0xe0, 0x70, 0x38, 0x47, 0x39, 0x54, 0x02, 0xe5, 0x67, 0x48, 0xcd, 0xa2, 0x48, 0xe7, 0x28, 0x53,
	0x31, 0x15, 0x06, 0xc1, 0xba, 0x94, 0xcf, 0x57, 0x4a, 0xbb, 0xd4, 0xe5, 0x33, 0x28, 0xc5, 0xef,
	0x00, 0x00, 0x00, 0xff, 0xff, 0x27, 0x5a, 0xb9, 0x67, 0x29, 0x03, 0x00, 0x00,
}

func (m *SrTeBsidStats) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *SrTeBsidStats) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *SrTeBsidStats) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.XXX_unrecognized != nil {
		i -= len(m.XXX_unrecognized)
		copy(dAtA[i:], m.XXX_unrecognized)
	}
	if len(m.TeBsidStats) > 0 {
		for iNdEx := len(m.TeBsidStats) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.TeBsidStats[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintSrTeBsidStats(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0xa
		}
	}
	return len(dAtA) - i, nil
}

func (m *SegmentRoutingTeBsidRecord) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *SegmentRoutingTeBsidRecord) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *SegmentRoutingTeBsidRecord) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.XXX_unrecognized != nil {
		i -= len(m.XXX_unrecognized)
		copy(dAtA[i:], m.XXX_unrecognized)
	}
	if m.Stats != nil {
		{
			size, err := m.Stats.MarshalToSizedBuffer(dAtA[:i])
			if err != nil {
				return 0, err
			}
			i -= size
			i = encodeVarintSrTeBsidStats(dAtA, i, uint64(size))
		}
		i--
		dAtA[i] = 0x32
	}
	if m.CounterName != nil {
		i -= len(*m.CounterName)
		copy(dAtA[i:], *m.CounterName)
		i = encodeVarintSrTeBsidStats(dAtA, i, uint64(len(*m.CounterName)))
		i--
		dAtA[i] = 0x2a
	}
	if m.InstanceIdentifier != nil {
		i = encodeVarintSrTeBsidStats(dAtA, i, uint64(*m.InstanceIdentifier))
		i--
		dAtA[i] = 0x20
	}
	if m.ColorIdentifier != nil {
		i = encodeVarintSrTeBsidStats(dAtA, i, uint64(*m.ColorIdentifier))
		i--
		dAtA[i] = 0x18
	}
	if m.ToIpPrefix == nil {
		return 0, github_com_gogo_protobuf_proto.NewRequiredNotSetError("to_ip_prefix")
	} else {
		i -= len(*m.ToIpPrefix)
		copy(dAtA[i:], *m.ToIpPrefix)
		i = encodeVarintSrTeBsidStats(dAtA, i, uint64(len(*m.ToIpPrefix)))
		i--
		dAtA[i] = 0x12
	}
	if m.BsidIdentifier == nil {
		return 0, github_com_gogo_protobuf_proto.NewRequiredNotSetError("bsid_identifier")
	} else {
		i = encodeVarintSrTeBsidStats(dAtA, i, uint64(*m.BsidIdentifier))
		i--
		dAtA[i] = 0x8
	}
	return len(dAtA) - i, nil
}

func (m *SegmentRoutingTeBsidStats) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *SegmentRoutingTeBsidStats) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *SegmentRoutingTeBsidStats) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.XXX_unrecognized != nil {
		i -= len(m.XXX_unrecognized)
		copy(dAtA[i:], m.XXX_unrecognized)
	}
	if m.ByteRate != nil {
		i = encodeVarintSrTeBsidStats(dAtA, i, uint64(*m.ByteRate))
		i--
		dAtA[i] = 0x20
	}
	if m.PacketRate != nil {
		i = encodeVarintSrTeBsidStats(dAtA, i, uint64(*m.PacketRate))
		i--
		dAtA[i] = 0x18
	}
	if m.Bytes != nil {
		i = encodeVarintSrTeBsidStats(dAtA, i, uint64(*m.Bytes))
		i--
		dAtA[i] = 0x10
	}
	if m.Packets != nil {
		i = encodeVarintSrTeBsidStats(dAtA, i, uint64(*m.Packets))
		i--
		dAtA[i] = 0x8
	}
	return len(dAtA) - i, nil
}

func encodeVarintSrTeBsidStats(dAtA []byte, offset int, v uint64) int {
	offset -= sovSrTeBsidStats(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *SrTeBsidStats) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if len(m.TeBsidStats) > 0 {
		for _, e := range m.TeBsidStats {
			l = e.Size()
			n += 1 + l + sovSrTeBsidStats(uint64(l))
		}
	}
	if m.XXX_unrecognized != nil {
		n += len(m.XXX_unrecognized)
	}
	return n
}

func (m *SegmentRoutingTeBsidRecord) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.BsidIdentifier != nil {
		n += 1 + sovSrTeBsidStats(uint64(*m.BsidIdentifier))
	}
	if m.ToIpPrefix != nil {
		l = len(*m.ToIpPrefix)
		n += 1 + l + sovSrTeBsidStats(uint64(l))
	}
	if m.ColorIdentifier != nil {
		n += 1 + sovSrTeBsidStats(uint64(*m.ColorIdentifier))
	}
	if m.InstanceIdentifier != nil {
		n += 1 + sovSrTeBsidStats(uint64(*m.InstanceIdentifier))
	}
	if m.CounterName != nil {
		l = len(*m.CounterName)
		n += 1 + l + sovSrTeBsidStats(uint64(l))
	}
	if m.Stats != nil {
		l = m.Stats.Size()
		n += 1 + l + sovSrTeBsidStats(uint64(l))
	}
	if m.XXX_unrecognized != nil {
		n += len(m.XXX_unrecognized)
	}
	return n
}

func (m *SegmentRoutingTeBsidStats) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.Packets != nil {
		n += 1 + sovSrTeBsidStats(uint64(*m.Packets))
	}
	if m.Bytes != nil {
		n += 1 + sovSrTeBsidStats(uint64(*m.Bytes))
	}
	if m.PacketRate != nil {
		n += 1 + sovSrTeBsidStats(uint64(*m.PacketRate))
	}
	if m.ByteRate != nil {
		n += 1 + sovSrTeBsidStats(uint64(*m.ByteRate))
	}
	if m.XXX_unrecognized != nil {
		n += len(m.XXX_unrecognized)
	}
	return n
}

func sovSrTeBsidStats(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozSrTeBsidStats(x uint64) (n int) {
	return sovSrTeBsidStats(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *SrTeBsidStats) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowSrTeBsidStats
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: SrTeBsidStats: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: SrTeBsidStats: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field TeBsidStats", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowSrTeBsidStats
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthSrTeBsidStats
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthSrTeBsidStats
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.TeBsidStats = append(m.TeBsidStats, &SegmentRoutingTeBsidRecord{})
			if err := m.TeBsidStats[len(m.TeBsidStats)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipSrTeBsidStats(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthSrTeBsidStats
			}
			if (iNdEx + skippy) < 0 {
				return ErrInvalidLengthSrTeBsidStats
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			m.XXX_unrecognized = append(m.XXX_unrecognized, dAtA[iNdEx:iNdEx+skippy]...)
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *SegmentRoutingTeBsidRecord) Unmarshal(dAtA []byte) error {
	var hasFields [1]uint64
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowSrTeBsidStats
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: SegmentRoutingTeBsidRecord: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: SegmentRoutingTeBsidRecord: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field BsidIdentifier", wireType)
			}
			var v uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowSrTeBsidStats
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				v |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			m.BsidIdentifier = &v
			hasFields[0] |= uint64(0x00000001)
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ToIpPrefix", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowSrTeBsidStats
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthSrTeBsidStats
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthSrTeBsidStats
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			s := string(dAtA[iNdEx:postIndex])
			m.ToIpPrefix = &s
			iNdEx = postIndex
			hasFields[0] |= uint64(0x00000002)
		case 3:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field ColorIdentifier", wireType)
			}
			var v uint32
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowSrTeBsidStats
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				v |= uint32(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			m.ColorIdentifier = &v
		case 4:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field InstanceIdentifier", wireType)
			}
			var v uint32
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowSrTeBsidStats
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				v |= uint32(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			m.InstanceIdentifier = &v
		case 5:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field CounterName", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowSrTeBsidStats
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthSrTeBsidStats
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthSrTeBsidStats
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			s := string(dAtA[iNdEx:postIndex])
			m.CounterName = &s
			iNdEx = postIndex
		case 6:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Stats", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowSrTeBsidStats
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthSrTeBsidStats
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthSrTeBsidStats
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.Stats == nil {
				m.Stats = &SegmentRoutingTeBsidStats{}
			}
			if err := m.Stats.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipSrTeBsidStats(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthSrTeBsidStats
			}
			if (iNdEx + skippy) < 0 {
				return ErrInvalidLengthSrTeBsidStats
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			m.XXX_unrecognized = append(m.XXX_unrecognized, dAtA[iNdEx:iNdEx+skippy]...)
			iNdEx += skippy
		}
	}
	if hasFields[0]&uint64(0x00000001) == 0 {
		return github_com_gogo_protobuf_proto.NewRequiredNotSetError("bsid_identifier")
	}
	if hasFields[0]&uint64(0x00000002) == 0 {
		return github_com_gogo_protobuf_proto.NewRequiredNotSetError("to_ip_prefix")
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *SegmentRoutingTeBsidStats) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowSrTeBsidStats
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: SegmentRoutingTeBsidStats: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: SegmentRoutingTeBsidStats: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Packets", wireType)
			}
			var v uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowSrTeBsidStats
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				v |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			m.Packets = &v
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Bytes", wireType)
			}
			var v uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowSrTeBsidStats
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				v |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			m.Bytes = &v
		case 3:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field PacketRate", wireType)
			}
			var v uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowSrTeBsidStats
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				v |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			m.PacketRate = &v
		case 4:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field ByteRate", wireType)
			}
			var v uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowSrTeBsidStats
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				v |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			m.ByteRate = &v
		default:
			iNdEx = preIndex
			skippy, err := skipSrTeBsidStats(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthSrTeBsidStats
			}
			if (iNdEx + skippy) < 0 {
				return ErrInvalidLengthSrTeBsidStats
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			m.XXX_unrecognized = append(m.XXX_unrecognized, dAtA[iNdEx:iNdEx+skippy]...)
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func skipSrTeBsidStats(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowSrTeBsidStats
			}
			if iNdEx >= l {
				return 0, io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		wireType := int(wire & 0x7)
		switch wireType {
		case 0:
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowSrTeBsidStats
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				iNdEx++
				if dAtA[iNdEx-1] < 0x80 {
					break
				}
			}
		case 1:
			iNdEx += 8
		case 2:
			var length int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowSrTeBsidStats
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				length |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if length < 0 {
				return 0, ErrInvalidLengthSrTeBsidStats
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupSrTeBsidStats
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthSrTeBsidStats
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthSrTeBsidStats        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowSrTeBsidStats          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupSrTeBsidStats = fmt.Errorf("proto: unexpected end of group")
)
