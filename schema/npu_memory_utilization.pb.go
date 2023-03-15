// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: npu_memory_utilization.proto

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

// Top level message NetworkProcessorMemoryUtilization
type NetworkProcessorMemoryUtilization struct {
	MemoryStats          []*NpuMemory `protobuf:"bytes,1,rep,name=memory_stats,json=memoryStats" json:"memory_stats,omitempty"`
	XXX_NoUnkeyedLiteral struct{}     `json:"-"`
	XXX_unrecognized     []byte       `json:"-"`
	XXX_sizecache        int32        `json:"-"`
}

func (m *NetworkProcessorMemoryUtilization) Reset()         { *m = NetworkProcessorMemoryUtilization{} }
func (m *NetworkProcessorMemoryUtilization) String() string { return proto.CompactTextString(m) }
func (*NetworkProcessorMemoryUtilization) ProtoMessage()    {}
func (*NetworkProcessorMemoryUtilization) Descriptor() ([]byte, []int) {
	return fileDescriptor_1c2c00ab8118d5b6, []int{0}
}
func (m *NetworkProcessorMemoryUtilization) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *NetworkProcessorMemoryUtilization) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_NetworkProcessorMemoryUtilization.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *NetworkProcessorMemoryUtilization) XXX_Merge(src proto.Message) {
	xxx_messageInfo_NetworkProcessorMemoryUtilization.Merge(m, src)
}
func (m *NetworkProcessorMemoryUtilization) XXX_Size() int {
	return m.Size()
}
func (m *NetworkProcessorMemoryUtilization) XXX_DiscardUnknown() {
	xxx_messageInfo_NetworkProcessorMemoryUtilization.DiscardUnknown(m)
}

var xxx_messageInfo_NetworkProcessorMemoryUtilization proto.InternalMessageInfo

func (m *NetworkProcessorMemoryUtilization) GetMemoryStats() []*NpuMemory {
	if m != nil {
		return m.MemoryStats
	}
	return nil
}

// Message that describes the memory utilization for each Network Processor
type NpuMemory struct {
	// Globally unique identifier for an NPU. This is of the form
	// "FPCX:NPUY", where X is the slot number of the line card and Y
	// is the index of the NPU on the linecard
	Identifier *string `protobuf:"bytes,1,req,name=identifier" json:"identifier,omitempty"`
	// NPU memory utilization statistics for different NPU memory types
	Summary []*NpuMemorySummary `protobuf:"bytes,2,rep,name=summary" json:"summary,omitempty"`
	// NPU memory utilization statistics for different NPU memory partitions
	Partition            []*NpuMemoryPartition `protobuf:"bytes,3,rep,name=partition" json:"partition,omitempty"`
	XXX_NoUnkeyedLiteral struct{}              `json:"-"`
	XXX_unrecognized     []byte                `json:"-"`
	XXX_sizecache        int32                 `json:"-"`
}

func (m *NpuMemory) Reset()         { *m = NpuMemory{} }
func (m *NpuMemory) String() string { return proto.CompactTextString(m) }
func (*NpuMemory) ProtoMessage()    {}
func (*NpuMemory) Descriptor() ([]byte, []int) {
	return fileDescriptor_1c2c00ab8118d5b6, []int{1}
}
func (m *NpuMemory) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *NpuMemory) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_NpuMemory.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *NpuMemory) XXX_Merge(src proto.Message) {
	xxx_messageInfo_NpuMemory.Merge(m, src)
}
func (m *NpuMemory) XXX_Size() int {
	return m.Size()
}
func (m *NpuMemory) XXX_DiscardUnknown() {
	xxx_messageInfo_NpuMemory.DiscardUnknown(m)
}

var xxx_messageInfo_NpuMemory proto.InternalMessageInfo

func (m *NpuMemory) GetIdentifier() string {
	if m != nil && m.Identifier != nil {
		return *m.Identifier
	}
	return ""
}

func (m *NpuMemory) GetSummary() []*NpuMemorySummary {
	if m != nil {
		return m.Summary
	}
	return nil
}

func (m *NpuMemory) GetPartition() []*NpuMemoryPartition {
	if m != nil {
		return m.Partition
	}
	return nil
}

// Summary of NPU memory utilization for each type of memory
type NpuMemorySummary struct {
	// Name of the partition.
	ResourceName *string `protobuf:"bytes,1,opt,name=resource_name,json=resourceName" json:"resource_name,omitempty"`
	// Maximum memory size in bytes
	Size_ *uint64 `protobuf:"varint,2,opt,name=size" json:"size,omitempty"`
	// How much memory is used up
	Allocated *uint64 `protobuf:"varint,3,opt,name=allocated" json:"allocated,omitempty"`
	// Utilization in percent
	Utilization          *int32   `protobuf:"varint,4,opt,name=utilization" json:"utilization,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *NpuMemorySummary) Reset()         { *m = NpuMemorySummary{} }
func (m *NpuMemorySummary) String() string { return proto.CompactTextString(m) }
func (*NpuMemorySummary) ProtoMessage()    {}
func (*NpuMemorySummary) Descriptor() ([]byte, []int) {
	return fileDescriptor_1c2c00ab8118d5b6, []int{2}
}
func (m *NpuMemorySummary) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *NpuMemorySummary) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_NpuMemorySummary.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *NpuMemorySummary) XXX_Merge(src proto.Message) {
	xxx_messageInfo_NpuMemorySummary.Merge(m, src)
}
func (m *NpuMemorySummary) XXX_Size() int {
	return m.Size()
}
func (m *NpuMemorySummary) XXX_DiscardUnknown() {
	xxx_messageInfo_NpuMemorySummary.DiscardUnknown(m)
}

var xxx_messageInfo_NpuMemorySummary proto.InternalMessageInfo

func (m *NpuMemorySummary) GetResourceName() string {
	if m != nil && m.ResourceName != nil {
		return *m.ResourceName
	}
	return ""
}

func (m *NpuMemorySummary) GetSize_() uint64 {
	if m != nil && m.Size_ != nil {
		return *m.Size_
	}
	return 0
}

func (m *NpuMemorySummary) GetAllocated() uint64 {
	if m != nil && m.Allocated != nil {
		return *m.Allocated
	}
	return 0
}

func (m *NpuMemorySummary) GetUtilization() int32 {
	if m != nil && m.Utilization != nil {
		return *m.Utilization
	}
	return 0
}

// A set of detailed stats for NPU memory partition
type NpuMemoryPartition struct {
	// NPU memory Partition name
	Name *string `protobuf:"bytes,1,opt,name=name" json:"name,omitempty"`
	// Name of the Application for which npu memory is allocated
	ApplicationName *string `protobuf:"bytes,2,opt,name=application_name,json=applicationName" json:"application_name,omitempty"`
	// Number of bytes allocated for the application
	BytesAllocated *uint32 `protobuf:"varint,3,opt,name=bytes_allocated,json=bytesAllocated" json:"bytes_allocated,omitempty"`
	// number of allocations for the application
	AllocationCount *uint32 `protobuf:"varint,4,opt,name=allocation_count,json=allocationCount" json:"allocation_count,omitempty"`
	// number of frees for the application
	FreeCount            *uint32  `protobuf:"varint,5,opt,name=free_count,json=freeCount" json:"free_count,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *NpuMemoryPartition) Reset()         { *m = NpuMemoryPartition{} }
func (m *NpuMemoryPartition) String() string { return proto.CompactTextString(m) }
func (*NpuMemoryPartition) ProtoMessage()    {}
func (*NpuMemoryPartition) Descriptor() ([]byte, []int) {
	return fileDescriptor_1c2c00ab8118d5b6, []int{3}
}
func (m *NpuMemoryPartition) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *NpuMemoryPartition) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_NpuMemoryPartition.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *NpuMemoryPartition) XXX_Merge(src proto.Message) {
	xxx_messageInfo_NpuMemoryPartition.Merge(m, src)
}
func (m *NpuMemoryPartition) XXX_Size() int {
	return m.Size()
}
func (m *NpuMemoryPartition) XXX_DiscardUnknown() {
	xxx_messageInfo_NpuMemoryPartition.DiscardUnknown(m)
}

var xxx_messageInfo_NpuMemoryPartition proto.InternalMessageInfo

func (m *NpuMemoryPartition) GetName() string {
	if m != nil && m.Name != nil {
		return *m.Name
	}
	return ""
}

func (m *NpuMemoryPartition) GetApplicationName() string {
	if m != nil && m.ApplicationName != nil {
		return *m.ApplicationName
	}
	return ""
}

func (m *NpuMemoryPartition) GetBytesAllocated() uint32 {
	if m != nil && m.BytesAllocated != nil {
		return *m.BytesAllocated
	}
	return 0
}

func (m *NpuMemoryPartition) GetAllocationCount() uint32 {
	if m != nil && m.AllocationCount != nil {
		return *m.AllocationCount
	}
	return 0
}

func (m *NpuMemoryPartition) GetFreeCount() uint32 {
	if m != nil && m.FreeCount != nil {
		return *m.FreeCount
	}
	return 0
}

var E_NpuMemoryExt = &proto.ExtensionDesc{
	ExtendedType:  (*JuniperNetworksSensors)(nil),
	ExtensionType: (*NetworkProcessorMemoryUtilization)(nil),
	Field:         11,
	Name:          "npu_memory_ext",
	Tag:           "bytes,11,opt,name=npu_memory_ext",
	Filename:      "npu_memory_utilization.proto",
}

func init() {
	proto.RegisterType((*NetworkProcessorMemoryUtilization)(nil), "NetworkProcessorMemoryUtilization")
	proto.RegisterType((*NpuMemory)(nil), "NpuMemory")
	proto.RegisterType((*NpuMemorySummary)(nil), "NpuMemorySummary")
	proto.RegisterType((*NpuMemoryPartition)(nil), "NpuMemoryPartition")
	proto.RegisterExtension(E_NpuMemoryExt)
}

func init() { proto.RegisterFile("npu_memory_utilization.proto", fileDescriptor_1c2c00ab8118d5b6) }

var fileDescriptor_1c2c00ab8118d5b6 = []byte{
	// 480 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x84, 0x92, 0xc1, 0x8e, 0xd3, 0x3c,
	0x14, 0x85, 0xe5, 0xb6, 0xa3, 0x5f, 0x71, 0xda, 0xe9, 0xfc, 0x9e, 0x05, 0x11, 0x1a, 0xaa, 0x50,
	0x16, 0x04, 0xa1, 0x26, 0x62, 0x96, 0xec, 0x00, 0xb1, 0x41, 0xa2, 0x1a, 0xb9, 0x62, 0xc3, 0x26,
	0xb8, 0xe9, 0x9d, 0xd6, 0x9d, 0xc4, 0x8e, 0xec, 0x1b, 0x98, 0xce, 0x2b, 0xf0, 0x5e, 0x88, 0x25,
	0x8f, 0x80, 0xfa, 0x24, 0x28, 0x4e, 0x3b, 0x09, 0xb0, 0x60, 0x67, 0x7f, 0xe7, 0xf8, 0xf8, 0xe8,
	0xea, 0xd2, 0x0b, 0x55, 0x56, 0x69, 0x01, 0x85, 0x36, 0xbb, 0xb4, 0x42, 0x99, 0xcb, 0x3b, 0x81,
	0x52, 0xab, 0xb8, 0x34, 0x1a, 0xf5, 0xc3, 0x73, 0x84, 0x1c, 0x0a, 0x40, 0xb3, 0x4b, 0x51, 0x97,
	0x0d, 0x9c, 0x72, 0xfa, 0x78, 0x0e, 0xf8, 0x45, 0x9b, 0x9b, 0x2b, 0xa3, 0x33, 0xb0, 0x56, 0x9b,
	0xf7, 0x2e, 0xe0, 0x43, 0xfb, 0x9e, 0xcd, 0xe8, 0xf0, 0x90, 0x6a, 0x51, 0xa0, 0x0d, 0x48, 0xd8,
	0x8f, 0xfc, 0x4b, 0x1a, 0xcf, 0xcb, 0xaa, 0x31, 0x73, 0xbf, 0xd1, 0x17, 0xb5, 0x3c, 0xfd, 0x4a,
	0xa8, 0x77, 0x2f, 0xb1, 0x09, 0xa5, 0x72, 0x05, 0x0a, 0xe5, 0xb5, 0x04, 0x13, 0x90, 0xb0, 0x17,
	0x79, 0xbc, 0x43, 0xd8, 0x73, 0xfa, 0x9f, 0xad, 0x8a, 0x42, 0x98, 0x5d, 0xd0, 0x73, 0xb9, 0xff,
	0xb7, 0xb9, 0x8b, 0x46, 0xe0, 0x47, 0x07, 0x7b, 0x41, 0xbd, 0x52, 0x18, 0x94, 0x75, 0xad, 0xa0,
	0xef, 0xec, 0xe7, 0xad, 0xfd, 0xea, 0x28, 0xf1, 0xd6, 0x55, 0xb7, 0x39, 0xfb, 0x33, 0x90, 0x3d,
	0xa1, 0x23, 0x03, 0x56, 0x57, 0x26, 0x83, 0x54, 0x89, 0x02, 0x02, 0x12, 0x92, 0xc8, 0xe3, 0xc3,
	0x23, 0x9c, 0x8b, 0x02, 0x18, 0xa3, 0x03, 0x2b, 0xef, 0x20, 0xe8, 0x85, 0x24, 0x1a, 0x70, 0x77,
	0x66, 0x17, 0xd4, 0x13, 0x79, 0xae, 0x33, 0x81, 0xb0, 0x0a, 0xfa, 0x4e, 0x68, 0x01, 0x0b, 0xa9,
	0xdf, 0x99, 0x7b, 0x30, 0x08, 0x49, 0x74, 0xc2, 0xbb, 0x68, 0xfa, 0x8d, 0x50, 0xf6, 0x77, 0xdf,
	0xfa, 0xab, 0x4e, 0x0d, 0x77, 0x66, 0xcf, 0xe8, 0x99, 0x28, 0xcb, 0x5c, 0x66, 0xee, 0x65, 0x53,
	0xb3, 0xe7, 0xf4, 0x71, 0x87, 0xbb, 0xa6, 0x4f, 0xe9, 0x78, 0xb9, 0x43, 0xb0, 0xe9, 0xef, 0xdd,
	0x46, 0xfc, 0xd4, 0xe1, 0x57, 0xf7, 0x05, 0xeb, 0xcc, 0xe6, 0x52, 0x47, 0x66, 0xba, 0x52, 0xe8,
	0x5a, 0x8e, 0xf8, 0xb8, 0xe5, 0x6f, 0x6a, 0xcc, 0x1e, 0x51, 0x7a, 0x6d, 0x00, 0x0e, 0xa6, 0x13,
	0x67, 0xf2, 0x6a, 0xe2, 0xe4, 0x97, 0x82, 0x9e, 0x76, 0xb6, 0x0d, 0x6e, 0x91, 0x3d, 0x88, 0xdf,
	0x55, 0x4a, 0x96, 0x60, 0x0e, 0x0b, 0x65, 0x17, 0xa0, 0xac, 0x36, 0x36, 0xf0, 0x43, 0x12, 0xf9,
	0x97, 0xd3, 0xf8, 0x9f, 0x8b, 0xc6, 0x87, 0xea, 0x38, 0x9a, 0xb7, 0xb7, 0xf8, 0xfa, 0xd3, 0xf7,
	0xfd, 0x84, 0xfc, 0xd8, 0x4f, 0xc8, 0xcf, 0xfd, 0x84, 0x7c, 0xe4, 0x6b, 0x89, 0xf1, 0xb6, 0xf9,
	0x21, 0x56, 0x80, 0x89, 0xcc, 0x60, 0x09, 0x66, 0x9d, 0x6c, 0x40, 0xe4, 0xb8, 0x59, 0x6a, 0x9c,
	0x49, 0xb5, 0x06, 0x8b, 0xb3, 0x2d, 0xca, 0x99, 0x12, 0x28, 0x3f, 0x43, 0x52, 0xde, 0xac, 0x93,
	0x2d, 0xca, 0x44, 0xac, 0x44, 0x89, 0x60, 0x6c, 0x92, 0x6e, 0x2b, 0xa5, 0x6d, 0x62, 0xb3, 0x0d,
	0x14, 0xe2, 0x57, 0x00, 0x00, 0x00, 0xff, 0xff, 0xad, 0x77, 0x97, 0x18, 0x31, 0x03, 0x00, 0x00,
}

func (m *NetworkProcessorMemoryUtilization) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *NetworkProcessorMemoryUtilization) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *NetworkProcessorMemoryUtilization) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.XXX_unrecognized != nil {
		i -= len(m.XXX_unrecognized)
		copy(dAtA[i:], m.XXX_unrecognized)
	}
	if len(m.MemoryStats) > 0 {
		for iNdEx := len(m.MemoryStats) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.MemoryStats[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintNpuMemoryUtilization(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0xa
		}
	}
	return len(dAtA) - i, nil
}

func (m *NpuMemory) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *NpuMemory) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *NpuMemory) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.XXX_unrecognized != nil {
		i -= len(m.XXX_unrecognized)
		copy(dAtA[i:], m.XXX_unrecognized)
	}
	if len(m.Partition) > 0 {
		for iNdEx := len(m.Partition) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.Partition[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintNpuMemoryUtilization(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0x1a
		}
	}
	if len(m.Summary) > 0 {
		for iNdEx := len(m.Summary) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.Summary[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintNpuMemoryUtilization(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0x12
		}
	}
	if m.Identifier == nil {
		return 0, github_com_gogo_protobuf_proto.NewRequiredNotSetError("identifier")
	} else {
		i -= len(*m.Identifier)
		copy(dAtA[i:], *m.Identifier)
		i = encodeVarintNpuMemoryUtilization(dAtA, i, uint64(len(*m.Identifier)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *NpuMemorySummary) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *NpuMemorySummary) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *NpuMemorySummary) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.XXX_unrecognized != nil {
		i -= len(m.XXX_unrecognized)
		copy(dAtA[i:], m.XXX_unrecognized)
	}
	if m.Utilization != nil {
		i = encodeVarintNpuMemoryUtilization(dAtA, i, uint64(*m.Utilization))
		i--
		dAtA[i] = 0x20
	}
	if m.Allocated != nil {
		i = encodeVarintNpuMemoryUtilization(dAtA, i, uint64(*m.Allocated))
		i--
		dAtA[i] = 0x18
	}
	if m.Size_ != nil {
		i = encodeVarintNpuMemoryUtilization(dAtA, i, uint64(*m.Size_))
		i--
		dAtA[i] = 0x10
	}
	if m.ResourceName != nil {
		i -= len(*m.ResourceName)
		copy(dAtA[i:], *m.ResourceName)
		i = encodeVarintNpuMemoryUtilization(dAtA, i, uint64(len(*m.ResourceName)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *NpuMemoryPartition) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *NpuMemoryPartition) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *NpuMemoryPartition) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.XXX_unrecognized != nil {
		i -= len(m.XXX_unrecognized)
		copy(dAtA[i:], m.XXX_unrecognized)
	}
	if m.FreeCount != nil {
		i = encodeVarintNpuMemoryUtilization(dAtA, i, uint64(*m.FreeCount))
		i--
		dAtA[i] = 0x28
	}
	if m.AllocationCount != nil {
		i = encodeVarintNpuMemoryUtilization(dAtA, i, uint64(*m.AllocationCount))
		i--
		dAtA[i] = 0x20
	}
	if m.BytesAllocated != nil {
		i = encodeVarintNpuMemoryUtilization(dAtA, i, uint64(*m.BytesAllocated))
		i--
		dAtA[i] = 0x18
	}
	if m.ApplicationName != nil {
		i -= len(*m.ApplicationName)
		copy(dAtA[i:], *m.ApplicationName)
		i = encodeVarintNpuMemoryUtilization(dAtA, i, uint64(len(*m.ApplicationName)))
		i--
		dAtA[i] = 0x12
	}
	if m.Name != nil {
		i -= len(*m.Name)
		copy(dAtA[i:], *m.Name)
		i = encodeVarintNpuMemoryUtilization(dAtA, i, uint64(len(*m.Name)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func encodeVarintNpuMemoryUtilization(dAtA []byte, offset int, v uint64) int {
	offset -= sovNpuMemoryUtilization(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *NetworkProcessorMemoryUtilization) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if len(m.MemoryStats) > 0 {
		for _, e := range m.MemoryStats {
			l = e.Size()
			n += 1 + l + sovNpuMemoryUtilization(uint64(l))
		}
	}
	if m.XXX_unrecognized != nil {
		n += len(m.XXX_unrecognized)
	}
	return n
}

func (m *NpuMemory) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.Identifier != nil {
		l = len(*m.Identifier)
		n += 1 + l + sovNpuMemoryUtilization(uint64(l))
	}
	if len(m.Summary) > 0 {
		for _, e := range m.Summary {
			l = e.Size()
			n += 1 + l + sovNpuMemoryUtilization(uint64(l))
		}
	}
	if len(m.Partition) > 0 {
		for _, e := range m.Partition {
			l = e.Size()
			n += 1 + l + sovNpuMemoryUtilization(uint64(l))
		}
	}
	if m.XXX_unrecognized != nil {
		n += len(m.XXX_unrecognized)
	}
	return n
}

func (m *NpuMemorySummary) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.ResourceName != nil {
		l = len(*m.ResourceName)
		n += 1 + l + sovNpuMemoryUtilization(uint64(l))
	}
	if m.Size_ != nil {
		n += 1 + sovNpuMemoryUtilization(uint64(*m.Size_))
	}
	if m.Allocated != nil {
		n += 1 + sovNpuMemoryUtilization(uint64(*m.Allocated))
	}
	if m.Utilization != nil {
		n += 1 + sovNpuMemoryUtilization(uint64(*m.Utilization))
	}
	if m.XXX_unrecognized != nil {
		n += len(m.XXX_unrecognized)
	}
	return n
}

func (m *NpuMemoryPartition) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.Name != nil {
		l = len(*m.Name)
		n += 1 + l + sovNpuMemoryUtilization(uint64(l))
	}
	if m.ApplicationName != nil {
		l = len(*m.ApplicationName)
		n += 1 + l + sovNpuMemoryUtilization(uint64(l))
	}
	if m.BytesAllocated != nil {
		n += 1 + sovNpuMemoryUtilization(uint64(*m.BytesAllocated))
	}
	if m.AllocationCount != nil {
		n += 1 + sovNpuMemoryUtilization(uint64(*m.AllocationCount))
	}
	if m.FreeCount != nil {
		n += 1 + sovNpuMemoryUtilization(uint64(*m.FreeCount))
	}
	if m.XXX_unrecognized != nil {
		n += len(m.XXX_unrecognized)
	}
	return n
}

func sovNpuMemoryUtilization(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozNpuMemoryUtilization(x uint64) (n int) {
	return sovNpuMemoryUtilization(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *NetworkProcessorMemoryUtilization) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowNpuMemoryUtilization
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
			return fmt.Errorf("proto: NetworkProcessorMemoryUtilization: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: NetworkProcessorMemoryUtilization: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field MemoryStats", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowNpuMemoryUtilization
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
				return ErrInvalidLengthNpuMemoryUtilization
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthNpuMemoryUtilization
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.MemoryStats = append(m.MemoryStats, &NpuMemory{})
			if err := m.MemoryStats[len(m.MemoryStats)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipNpuMemoryUtilization(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthNpuMemoryUtilization
			}
			if (iNdEx + skippy) < 0 {
				return ErrInvalidLengthNpuMemoryUtilization
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
func (m *NpuMemory) Unmarshal(dAtA []byte) error {
	var hasFields [1]uint64
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowNpuMemoryUtilization
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
			return fmt.Errorf("proto: NpuMemory: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: NpuMemory: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Identifier", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowNpuMemoryUtilization
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
				return ErrInvalidLengthNpuMemoryUtilization
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthNpuMemoryUtilization
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			s := string(dAtA[iNdEx:postIndex])
			m.Identifier = &s
			iNdEx = postIndex
			hasFields[0] |= uint64(0x00000001)
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Summary", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowNpuMemoryUtilization
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
				return ErrInvalidLengthNpuMemoryUtilization
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthNpuMemoryUtilization
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Summary = append(m.Summary, &NpuMemorySummary{})
			if err := m.Summary[len(m.Summary)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Partition", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowNpuMemoryUtilization
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
				return ErrInvalidLengthNpuMemoryUtilization
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthNpuMemoryUtilization
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Partition = append(m.Partition, &NpuMemoryPartition{})
			if err := m.Partition[len(m.Partition)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipNpuMemoryUtilization(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthNpuMemoryUtilization
			}
			if (iNdEx + skippy) < 0 {
				return ErrInvalidLengthNpuMemoryUtilization
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			m.XXX_unrecognized = append(m.XXX_unrecognized, dAtA[iNdEx:iNdEx+skippy]...)
			iNdEx += skippy
		}
	}
	if hasFields[0]&uint64(0x00000001) == 0 {
		return github_com_gogo_protobuf_proto.NewRequiredNotSetError("identifier")
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *NpuMemorySummary) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowNpuMemoryUtilization
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
			return fmt.Errorf("proto: NpuMemorySummary: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: NpuMemorySummary: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ResourceName", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowNpuMemoryUtilization
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
				return ErrInvalidLengthNpuMemoryUtilization
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthNpuMemoryUtilization
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			s := string(dAtA[iNdEx:postIndex])
			m.ResourceName = &s
			iNdEx = postIndex
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Size_", wireType)
			}
			var v uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowNpuMemoryUtilization
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
			m.Size_ = &v
		case 3:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Allocated", wireType)
			}
			var v uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowNpuMemoryUtilization
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
			m.Allocated = &v
		case 4:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Utilization", wireType)
			}
			var v int32
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowNpuMemoryUtilization
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				v |= int32(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			m.Utilization = &v
		default:
			iNdEx = preIndex
			skippy, err := skipNpuMemoryUtilization(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthNpuMemoryUtilization
			}
			if (iNdEx + skippy) < 0 {
				return ErrInvalidLengthNpuMemoryUtilization
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
func (m *NpuMemoryPartition) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowNpuMemoryUtilization
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
			return fmt.Errorf("proto: NpuMemoryPartition: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: NpuMemoryPartition: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Name", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowNpuMemoryUtilization
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
				return ErrInvalidLengthNpuMemoryUtilization
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthNpuMemoryUtilization
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			s := string(dAtA[iNdEx:postIndex])
			m.Name = &s
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ApplicationName", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowNpuMemoryUtilization
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
				return ErrInvalidLengthNpuMemoryUtilization
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthNpuMemoryUtilization
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			s := string(dAtA[iNdEx:postIndex])
			m.ApplicationName = &s
			iNdEx = postIndex
		case 3:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field BytesAllocated", wireType)
			}
			var v uint32
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowNpuMemoryUtilization
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
			m.BytesAllocated = &v
		case 4:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field AllocationCount", wireType)
			}
			var v uint32
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowNpuMemoryUtilization
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
			m.AllocationCount = &v
		case 5:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field FreeCount", wireType)
			}
			var v uint32
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowNpuMemoryUtilization
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
			m.FreeCount = &v
		default:
			iNdEx = preIndex
			skippy, err := skipNpuMemoryUtilization(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthNpuMemoryUtilization
			}
			if (iNdEx + skippy) < 0 {
				return ErrInvalidLengthNpuMemoryUtilization
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
func skipNpuMemoryUtilization(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowNpuMemoryUtilization
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
					return 0, ErrIntOverflowNpuMemoryUtilization
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
					return 0, ErrIntOverflowNpuMemoryUtilization
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
				return 0, ErrInvalidLengthNpuMemoryUtilization
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupNpuMemoryUtilization
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthNpuMemoryUtilization
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthNpuMemoryUtilization        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowNpuMemoryUtilization          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupNpuMemoryUtilization = fmt.Errorf("proto: unexpected end of group")
)
