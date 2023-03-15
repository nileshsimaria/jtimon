// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: port_exp.proto

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
type PortExp struct {
	InterfaceExpStats    []*InterfaceExpInfos `protobuf:"bytes,1,rep,name=interfaceExp_stats,json=interfaceExpStats" json:"interfaceExp_stats,omitempty"`
	XXX_NoUnkeyedLiteral struct{}             `json:"-"`
	XXX_unrecognized     []byte               `json:"-"`
	XXX_sizecache        int32                `json:"-"`
}

func (m *PortExp) Reset()         { *m = PortExp{} }
func (m *PortExp) String() string { return proto.CompactTextString(m) }
func (*PortExp) ProtoMessage()    {}
func (*PortExp) Descriptor() ([]byte, []int) {
	return fileDescriptor_dbd3f2d611fd16ff, []int{0}
}
func (m *PortExp) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *PortExp) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_PortExp.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *PortExp) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PortExp.Merge(m, src)
}
func (m *PortExp) XXX_Size() int {
	return m.Size()
}
func (m *PortExp) XXX_DiscardUnknown() {
	xxx_messageInfo_PortExp.DiscardUnknown(m)
}

var xxx_messageInfo_PortExp proto.InternalMessageInfo

func (m *PortExp) GetInterfaceExpStats() []*InterfaceExpInfos {
	if m != nil {
		return m.InterfaceExpStats
	}
	return nil
}

//
// Interface information
//
type InterfaceExpInfos struct {
	// Interface name, e.g., et-0/0/0
	IfName *string `protobuf:"bytes,1,req,name=if_name,json=ifName" json:"if_name,omitempty"`
	// Interface operational status
	IfOperationalStatus  *string  `protobuf:"bytes,2,opt,name=if_operational_status,json=ifOperationalStatus" json:"if_operational_status,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *InterfaceExpInfos) Reset()         { *m = InterfaceExpInfos{} }
func (m *InterfaceExpInfos) String() string { return proto.CompactTextString(m) }
func (*InterfaceExpInfos) ProtoMessage()    {}
func (*InterfaceExpInfos) Descriptor() ([]byte, []int) {
	return fileDescriptor_dbd3f2d611fd16ff, []int{1}
}
func (m *InterfaceExpInfos) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *InterfaceExpInfos) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_InterfaceExpInfos.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *InterfaceExpInfos) XXX_Merge(src proto.Message) {
	xxx_messageInfo_InterfaceExpInfos.Merge(m, src)
}
func (m *InterfaceExpInfos) XXX_Size() int {
	return m.Size()
}
func (m *InterfaceExpInfos) XXX_DiscardUnknown() {
	xxx_messageInfo_InterfaceExpInfos.DiscardUnknown(m)
}

var xxx_messageInfo_InterfaceExpInfos proto.InternalMessageInfo

func (m *InterfaceExpInfos) GetIfName() string {
	if m != nil && m.IfName != nil {
		return *m.IfName
	}
	return ""
}

func (m *InterfaceExpInfos) GetIfOperationalStatus() string {
	if m != nil && m.IfOperationalStatus != nil {
		return *m.IfOperationalStatus
	}
	return ""
}

var E_JnprInterfaceExpExt = &proto.ExtensionDesc{
	ExtendedType:  (*JuniperNetworksSensors)(nil),
	ExtensionType: (*PortExp)(nil),
	Field:         18,
	Name:          "jnpr_interface_exp_ext",
	Tag:           "bytes,18,opt,name=jnpr_interface_exp_ext",
	Filename:      "port_exp.proto",
}

func init() {
	proto.RegisterType((*PortExp)(nil), "Port_exp")
	proto.RegisterType((*InterfaceExpInfos)(nil), "InterfaceExpInfos")
	proto.RegisterExtension(E_JnprInterfaceExpExt)
}

func init() { proto.RegisterFile("port_exp.proto", fileDescriptor_dbd3f2d611fd16ff) }

var fileDescriptor_dbd3f2d611fd16ff = []byte{
	// 318 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x64, 0x8f, 0xc1, 0x4a, 0xf3, 0x40,
	0x14, 0x85, 0x99, 0xfe, 0xfc, 0xda, 0x4e, 0x41, 0xe8, 0x14, 0x35, 0xb8, 0x08, 0xa1, 0xab, 0x6c,
	0x9a, 0x81, 0x2e, 0x5d, 0x69, 0xa1, 0x8b, 0x0a, 0x56, 0x49, 0x17, 0x82, 0x9b, 0x71, 0x5a, 0x6f,
	0xd2, 0x49, 0x9b, 0x99, 0x61, 0xe6, 0x56, 0xe3, 0xd6, 0xa7, 0x73, 0xe9, 0x23, 0x48, 0x9f, 0x44,
	0x5a, 0xab, 0x06, 0xdc, 0xde, 0xef, 0xe3, 0xdc, 0x73, 0xe8, 0x91, 0x35, 0x0e, 0x05, 0x54, 0x36,
	0xb1, 0xce, 0xa0, 0x39, 0xeb, 0x22, 0xac, 0xa0, 0x04, 0x74, 0x2f, 0x02, 0xcd, 0xfe, 0xd8, 0xbb,
	0xa6, 0xcd, 0xdb, 0xbd, 0xc6, 0x2e, 0x29, 0x53, 0x1a, 0xc1, 0x65, 0x72, 0x0e, 0xa3, 0xca, 0x0a,
	0x8f, 0x12, 0x7d, 0x40, 0xa2, 0x7f, 0x71, 0x7b, 0xc0, 0x92, 0x71, 0x0d, 0x8d, 0x75, 0x66, 0x7c,
	0xda, 0xa9, 0xdb, 0xd3, 0xad, 0xdc, 0xcb, 0x69, 0xe7, 0x8f, 0xc7, 0x42, 0x7a, 0xa8, 0x32, 0xa1,
	0x65, 0x09, 0x01, 0x89, 0x1a, 0x71, 0x6b, 0xf8, 0xff, 0xf5, 0xa2, 0xd1, 0x24, 0xe9, 0x81, 0xca,
	0x26, 0xb2, 0x04, 0x36, 0xa0, 0xc7, 0x2a, 0x13, 0xc6, 0x82, 0x93, 0xa8, 0x8c, 0x96, 0xab, 0xdd,
	0xe7, 0xb5, 0x0f, 0x1a, 0x11, 0x89, 0x5b, 0x69, 0x57, 0x65, 0x37, 0xbf, 0x6c, 0xba, 0x43, 0xe7,
	0x77, 0xf4, 0xa4, 0xd0, 0xd6, 0x89, 0x9f, 0x0a, 0xdb, 0x05, 0x02, 0x2a, 0x64, 0xa7, 0xc9, 0xd5,
	0x5a, 0x2b, 0x0b, 0x6e, 0x02, 0xf8, 0x6c, 0xdc, 0xd2, 0x4f, 0x41, 0x7b, 0xe3, 0x7c, 0xc0, 0x22,
	0x12, 0xb7, 0x07, 0xad, 0xe4, 0x7b, 0x6f, 0xda, 0xdd, 0x26, 0xd4, 0xeb, 0x8e, 0x2a, 0x1c, 0x3e,
	0xbc, 0x6d, 0x42, 0xf2, 0xbe, 0x09, 0xc9, 0xc7, 0x26, 0x24, 0xf7, 0x69, 0xae, 0x30, 0x29, 0xbe,
	0xf2, 0x12, 0x0d, 0xc8, 0xd5, 0x1c, 0x66, 0xe0, 0x72, 0xbe, 0x00, 0xb9, 0xc2, 0xc5, 0xcc, 0x60,
	0x5f, 0xe9, 0x1c, 0x3c, 0xf6, 0x0b, 0x54, 0x7d, 0x2d, 0x51, 0x3d, 0x01, 0xb7, 0xcb, 0x9c, 0x17,
	0xa8, 0xb8, 0x7c, 0x94, 0x16, 0xc1, 0x79, 0x2e, 0x8a, 0xb5, 0x36, 0x9e, 0xfb, 0xf9, 0x02, 0x4a,
	0xf9, 0x19, 0x00, 0x00, 0xff, 0xff, 0xba, 0xea, 0x78, 0x5c, 0x98, 0x01, 0x00, 0x00,
}

func (m *PortExp) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *PortExp) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *PortExp) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.XXX_unrecognized != nil {
		i -= len(m.XXX_unrecognized)
		copy(dAtA[i:], m.XXX_unrecognized)
	}
	if len(m.InterfaceExpStats) > 0 {
		for iNdEx := len(m.InterfaceExpStats) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.InterfaceExpStats[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintPortExp(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0xa
		}
	}
	return len(dAtA) - i, nil
}

func (m *InterfaceExpInfos) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *InterfaceExpInfos) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *InterfaceExpInfos) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.XXX_unrecognized != nil {
		i -= len(m.XXX_unrecognized)
		copy(dAtA[i:], m.XXX_unrecognized)
	}
	if m.IfOperationalStatus != nil {
		i -= len(*m.IfOperationalStatus)
		copy(dAtA[i:], *m.IfOperationalStatus)
		i = encodeVarintPortExp(dAtA, i, uint64(len(*m.IfOperationalStatus)))
		i--
		dAtA[i] = 0x12
	}
	if m.IfName == nil {
		return 0, github_com_gogo_protobuf_proto.NewRequiredNotSetError("if_name")
	} else {
		i -= len(*m.IfName)
		copy(dAtA[i:], *m.IfName)
		i = encodeVarintPortExp(dAtA, i, uint64(len(*m.IfName)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func encodeVarintPortExp(dAtA []byte, offset int, v uint64) int {
	offset -= sovPortExp(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *PortExp) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if len(m.InterfaceExpStats) > 0 {
		for _, e := range m.InterfaceExpStats {
			l = e.Size()
			n += 1 + l + sovPortExp(uint64(l))
		}
	}
	if m.XXX_unrecognized != nil {
		n += len(m.XXX_unrecognized)
	}
	return n
}

func (m *InterfaceExpInfos) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.IfName != nil {
		l = len(*m.IfName)
		n += 1 + l + sovPortExp(uint64(l))
	}
	if m.IfOperationalStatus != nil {
		l = len(*m.IfOperationalStatus)
		n += 1 + l + sovPortExp(uint64(l))
	}
	if m.XXX_unrecognized != nil {
		n += len(m.XXX_unrecognized)
	}
	return n
}

func sovPortExp(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozPortExp(x uint64) (n int) {
	return sovPortExp(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *PortExp) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowPortExp
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
			return fmt.Errorf("proto: Port_exp: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Port_exp: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field InterfaceExpStats", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowPortExp
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
				return ErrInvalidLengthPortExp
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthPortExp
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.InterfaceExpStats = append(m.InterfaceExpStats, &InterfaceExpInfos{})
			if err := m.InterfaceExpStats[len(m.InterfaceExpStats)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipPortExp(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthPortExp
			}
			if (iNdEx + skippy) < 0 {
				return ErrInvalidLengthPortExp
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
func (m *InterfaceExpInfos) Unmarshal(dAtA []byte) error {
	var hasFields [1]uint64
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowPortExp
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
			return fmt.Errorf("proto: InterfaceExpInfos: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: InterfaceExpInfos: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field IfName", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowPortExp
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
				return ErrInvalidLengthPortExp
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthPortExp
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			s := string(dAtA[iNdEx:postIndex])
			m.IfName = &s
			iNdEx = postIndex
			hasFields[0] |= uint64(0x00000001)
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field IfOperationalStatus", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowPortExp
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
				return ErrInvalidLengthPortExp
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthPortExp
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			s := string(dAtA[iNdEx:postIndex])
			m.IfOperationalStatus = &s
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipPortExp(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthPortExp
			}
			if (iNdEx + skippy) < 0 {
				return ErrInvalidLengthPortExp
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			m.XXX_unrecognized = append(m.XXX_unrecognized, dAtA[iNdEx:iNdEx+skippy]...)
			iNdEx += skippy
		}
	}
	if hasFields[0]&uint64(0x00000001) == 0 {
		return github_com_gogo_protobuf_proto.NewRequiredNotSetError("if_name")
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func skipPortExp(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowPortExp
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
					return 0, ErrIntOverflowPortExp
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
					return 0, ErrIntOverflowPortExp
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
				return 0, ErrInvalidLengthPortExp
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupPortExp
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthPortExp
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthPortExp        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowPortExp          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupPortExp = fmt.Errorf("proto: unexpected end of group")
)
