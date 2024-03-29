//
// Copyright 2019 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

// Version: 0.1

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.26.0
// 	protoc        v3.12.4
// source: proto/tunnel/tunnel.proto

// Package grpctunnel defines a service specification for a TCP over gRPC proxy
// interface. This interface is set so that the tunnel can act as a transparent
// TCP proxy between a pair of gRPC client and server endpoints.

package grpctunnel

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type TargetType int32

const (
	TargetType_UNKNOWN TargetType = 0
	// https://www.iana.org/assignments/service-names-port-numbers/service-names-port-numbers.xhtml?search=22
	TargetType_SSH TargetType = 22
	// https://www.iana.org/assignments/service-names-port-numbers/service-names-port-numbers.xhtml?search=6653
	TargetType_OPENFLOW TargetType = 6653
	// https://www.iana.org/assignments/service-names-port-numbers/service-names-port-numbers.xhtml?search=9339
	TargetType_GNMI_GNOI TargetType = 9339
	// https://www.iana.org/assignments/service-names-port-numbers/service-names-port-numbers.xhtml?search=9559
	TargetType_P4_RUNTIME TargetType = 9559
)

// Enum value maps for TargetType.
var (
	TargetType_name = map[int32]string{
		0:    "UNKNOWN",
		22:   "SSH",
		6653: "OPENFLOW",
		9339: "GNMI_GNOI",
		9559: "P4_RUNTIME",
	}
	TargetType_value = map[string]int32{
		"UNKNOWN":    0,
		"SSH":        22,
		"OPENFLOW":   6653,
		"GNMI_GNOI":  9339,
		"P4_RUNTIME": 9559,
	}
)

func (x TargetType) Enum() *TargetType {
	p := new(TargetType)
	*p = x
	return p
}

func (x TargetType) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (TargetType) Descriptor() protoreflect.EnumDescriptor {
	return file_proto_tunnel_tunnel_proto_enumTypes[0].Descriptor()
}

func (TargetType) Type() protoreflect.EnumType {
	return &file_proto_tunnel_tunnel_proto_enumTypes[0]
}

func (x TargetType) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use TargetType.Descriptor instead.
func (TargetType) EnumDescriptor() ([]byte, []int) {
	return file_proto_tunnel_tunnel_proto_rawDescGZIP(), []int{0}
}

type Target_TargetOp int32

const (
	Target_UNKNOWN Target_TargetOp = 0
	Target_ADD     Target_TargetOp = 1
	Target_REMOVE  Target_TargetOp = 2
)

// Enum value maps for Target_TargetOp.
var (
	Target_TargetOp_name = map[int32]string{
		0: "UNKNOWN",
		1: "ADD",
		2: "REMOVE",
	}
	Target_TargetOp_value = map[string]int32{
		"UNKNOWN": 0,
		"ADD":     1,
		"REMOVE":  2,
	}
)

func (x Target_TargetOp) Enum() *Target_TargetOp {
	p := new(Target_TargetOp)
	*p = x
	return p
}

func (x Target_TargetOp) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (Target_TargetOp) Descriptor() protoreflect.EnumDescriptor {
	return file_proto_tunnel_tunnel_proto_enumTypes[1].Descriptor()
}

func (Target_TargetOp) Type() protoreflect.EnumType {
	return &file_proto_tunnel_tunnel_proto_enumTypes[1]
}

func (x Target_TargetOp) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use Target_TargetOp.Descriptor instead.
func (Target_TargetOp) EnumDescriptor() ([]byte, []int) {
	return file_proto_tunnel_tunnel_proto_rawDescGZIP(), []int{2, 0}
}

type Subscription_SubscriptionOp int32

const (
	Subscription_UNKNOWN    Subscription_SubscriptionOp = 0
	Subscription_SUBCRIBE   Subscription_SubscriptionOp = 1
	Subscription_UNSUBCRIBE Subscription_SubscriptionOp = 2
)

// Enum value maps for Subscription_SubscriptionOp.
var (
	Subscription_SubscriptionOp_name = map[int32]string{
		0: "UNKNOWN",
		1: "SUBCRIBE",
		2: "UNSUBCRIBE",
	}
	Subscription_SubscriptionOp_value = map[string]int32{
		"UNKNOWN":    0,
		"SUBCRIBE":   1,
		"UNSUBCRIBE": 2,
	}
)

func (x Subscription_SubscriptionOp) Enum() *Subscription_SubscriptionOp {
	p := new(Subscription_SubscriptionOp)
	*p = x
	return p
}

func (x Subscription_SubscriptionOp) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (Subscription_SubscriptionOp) Descriptor() protoreflect.EnumDescriptor {
	return file_proto_tunnel_tunnel_proto_enumTypes[2].Descriptor()
}

func (Subscription_SubscriptionOp) Type() protoreflect.EnumType {
	return &file_proto_tunnel_tunnel_proto_enumTypes[2]
}

func (x Subscription_SubscriptionOp) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use Subscription_SubscriptionOp.Descriptor instead.
func (Subscription_SubscriptionOp) EnumDescriptor() ([]byte, []int) {
	return file_proto_tunnel_tunnel_proto_rawDescGZIP(), []int{4, 0}
}

type Data struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Tag   int32  `protobuf:"varint,1,opt,name=tag,proto3" json:"tag,omitempty"`     // Tag associated with the initial TCP stream setup.
	Data  []byte `protobuf:"bytes,2,opt,name=data,proto3" json:"data,omitempty"`    // Bytes received from client connection.
	Close bool   `protobuf:"varint,3,opt,name=close,proto3" json:"close,omitempty"` // Connection has reached EOF.
}

func (x *Data) Reset() {
	*x = Data{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_tunnel_tunnel_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Data) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Data) ProtoMessage() {}

func (x *Data) ProtoReflect() protoreflect.Message {
	mi := &file_proto_tunnel_tunnel_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Data.ProtoReflect.Descriptor instead.
func (*Data) Descriptor() ([]byte, []int) {
	return file_proto_tunnel_tunnel_proto_rawDescGZIP(), []int{0}
}

func (x *Data) GetTag() int32 {
	if x != nil {
		return x.Tag
	}
	return 0
}

func (x *Data) GetData() []byte {
	if x != nil {
		return x.Data
	}
	return nil
}

func (x *Data) GetClose() bool {
	if x != nil {
		return x.Close
	}
	return false
}

type RegisterOp struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Types that are assignable to Registration:
	//	*RegisterOp_Target
	//	*RegisterOp_Session
	//	*RegisterOp_Subscription
	Registration isRegisterOp_Registration `protobuf_oneof:"Registration"`
}

func (x *RegisterOp) Reset() {
	*x = RegisterOp{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_tunnel_tunnel_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RegisterOp) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RegisterOp) ProtoMessage() {}

func (x *RegisterOp) ProtoReflect() protoreflect.Message {
	mi := &file_proto_tunnel_tunnel_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RegisterOp.ProtoReflect.Descriptor instead.
func (*RegisterOp) Descriptor() ([]byte, []int) {
	return file_proto_tunnel_tunnel_proto_rawDescGZIP(), []int{1}
}

func (m *RegisterOp) GetRegistration() isRegisterOp_Registration {
	if m != nil {
		return m.Registration
	}
	return nil
}

func (x *RegisterOp) GetTarget() *Target {
	if x, ok := x.GetRegistration().(*RegisterOp_Target); ok {
		return x.Target
	}
	return nil
}

func (x *RegisterOp) GetSession() *Session {
	if x, ok := x.GetRegistration().(*RegisterOp_Session); ok {
		return x.Session
	}
	return nil
}

func (x *RegisterOp) GetSubscription() *Subscription {
	if x, ok := x.GetRegistration().(*RegisterOp_Subscription); ok {
		return x.Subscription
	}
	return nil
}

type isRegisterOp_Registration interface {
	isRegisterOp_Registration()
}

type RegisterOp_Target struct {
	Target *Target `protobuf:"bytes,1,opt,name=target,proto3,oneof"`
}

type RegisterOp_Session struct {
	Session *Session `protobuf:"bytes,2,opt,name=session,proto3,oneof"`
}

type RegisterOp_Subscription struct {
	Subscription *Subscription `protobuf:"bytes,3,opt,name=subscription,proto3,oneof"`
}

func (*RegisterOp_Target) isRegisterOp_Registration() {}

func (*RegisterOp_Session) isRegisterOp_Registration() {}

func (*RegisterOp_Subscription) isRegisterOp_Registration() {}

type Target struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Op Target_TargetOp `protobuf:"varint,1,opt,name=op,proto3,enum=grpctunnel.Target_TargetOp" json:"op,omitempty"`
	// Used to ack the registration of target and target_type.
	Accept bool `protobuf:"varint,2,opt,name=accept,proto3" json:"accept,omitempty"`
	// Target identifies which handler to use for a tunnel stream.
	Target string `protobuf:"bytes,3,opt,name=target,proto3" json:"target,omitempty"`
	// String value of the corresponding TargetType for a standard protocol.
	// A non-enumerated protocol is supported so long as both tunnel client and
	// server are in agreement on a particular value.
	TargetType string `protobuf:"bytes,4,opt,name=target_type,json=targetType,proto3" json:"target_type,omitempty"`
	Error      string `protobuf:"bytes,5,opt,name=error,proto3" json:"error,omitempty"`
}

func (x *Target) Reset() {
	*x = Target{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_tunnel_tunnel_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Target) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Target) ProtoMessage() {}

func (x *Target) ProtoReflect() protoreflect.Message {
	mi := &file_proto_tunnel_tunnel_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Target.ProtoReflect.Descriptor instead.
func (*Target) Descriptor() ([]byte, []int) {
	return file_proto_tunnel_tunnel_proto_rawDescGZIP(), []int{2}
}

func (x *Target) GetOp() Target_TargetOp {
	if x != nil {
		return x.Op
	}
	return Target_UNKNOWN
}

func (x *Target) GetAccept() bool {
	if x != nil {
		return x.Accept
	}
	return false
}

func (x *Target) GetTarget() string {
	if x != nil {
		return x.Target
	}
	return ""
}

func (x *Target) GetTargetType() string {
	if x != nil {
		return x.TargetType
	}
	return ""
}

func (x *Target) GetError() string {
	if x != nil {
		return x.Error
	}
	return ""
}

type Session struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The tag associated with the initial TCP stream setup.
	Tag int32 `protobuf:"varint,1,opt,name=tag,proto3" json:"tag,omitempty"`
	// Used to ack the connection tag.
	Accept bool `protobuf:"varint,2,opt,name=accept,proto3" json:"accept,omitempty"`
	// Target identifies which handler to use for a tunnel stream.
	Target string `protobuf:"bytes,3,opt,name=target,proto3" json:"target,omitempty"`
	// String value of the corresponding TargetType for a standard protocol.
	// A non-enumerated protocol is supported so long as both tunnel client and
	// server are in agreement on a particular value.
	TargetType string `protobuf:"bytes,4,opt,name=target_type,json=targetType,proto3" json:"target_type,omitempty"`
	// Error allows the register stream to return an error without breaking the
	// stream.
	Error string `protobuf:"bytes,5,opt,name=error,proto3" json:"error,omitempty"`
}

func (x *Session) Reset() {
	*x = Session{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_tunnel_tunnel_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Session) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Session) ProtoMessage() {}

func (x *Session) ProtoReflect() protoreflect.Message {
	mi := &file_proto_tunnel_tunnel_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Session.ProtoReflect.Descriptor instead.
func (*Session) Descriptor() ([]byte, []int) {
	return file_proto_tunnel_tunnel_proto_rawDescGZIP(), []int{3}
}

func (x *Session) GetTag() int32 {
	if x != nil {
		return x.Tag
	}
	return 0
}

func (x *Session) GetAccept() bool {
	if x != nil {
		return x.Accept
	}
	return false
}

func (x *Session) GetTarget() string {
	if x != nil {
		return x.Target
	}
	return ""
}

func (x *Session) GetTargetType() string {
	if x != nil {
		return x.TargetType
	}
	return ""
}

func (x *Session) GetError() string {
	if x != nil {
		return x.Error
	}
	return ""
}

type Subscription struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Op Subscription_SubscriptionOp `protobuf:"varint,1,opt,name=op,proto3,enum=grpctunnel.Subscription_SubscriptionOp" json:"op,omitempty"`
	// Used to ack the registration of (un)subscription.
	Accept bool `protobuf:"varint,2,opt,name=accept,proto3" json:"accept,omitempty"`
	// String value of the corresponding TargetType for a standard protocol.
	// Used to filter targets for subscription. If empty, it will subscribe all.
	TargetType string `protobuf:"bytes,3,opt,name=target_type,json=targetType,proto3" json:"target_type,omitempty"`
	Error      string `protobuf:"bytes,4,opt,name=error,proto3" json:"error,omitempty"`
}

func (x *Subscription) Reset() {
	*x = Subscription{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_tunnel_tunnel_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Subscription) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Subscription) ProtoMessage() {}

func (x *Subscription) ProtoReflect() protoreflect.Message {
	mi := &file_proto_tunnel_tunnel_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Subscription.ProtoReflect.Descriptor instead.
func (*Subscription) Descriptor() ([]byte, []int) {
	return file_proto_tunnel_tunnel_proto_rawDescGZIP(), []int{4}
}

func (x *Subscription) GetOp() Subscription_SubscriptionOp {
	if x != nil {
		return x.Op
	}
	return Subscription_UNKNOWN
}

func (x *Subscription) GetAccept() bool {
	if x != nil {
		return x.Accept
	}
	return false
}

func (x *Subscription) GetTargetType() string {
	if x != nil {
		return x.TargetType
	}
	return ""
}

func (x *Subscription) GetError() string {
	if x != nil {
		return x.Error
	}
	return ""
}

var File_proto_tunnel_tunnel_proto protoreflect.FileDescriptor

var file_proto_tunnel_tunnel_proto_rawDesc = []byte{
	0x0a, 0x19, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x74, 0x75, 0x6e, 0x6e, 0x65, 0x6c, 0x2f, 0x74,
	0x75, 0x6e, 0x6e, 0x65, 0x6c, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0a, 0x67, 0x72, 0x70,
	0x63, 0x74, 0x75, 0x6e, 0x6e, 0x65, 0x6c, 0x22, 0x42, 0x0a, 0x04, 0x44, 0x61, 0x74, 0x61, 0x12,
	0x10, 0x0a, 0x03, 0x74, 0x61, 0x67, 0x18, 0x01, 0x20, 0x01, 0x28, 0x05, 0x52, 0x03, 0x74, 0x61,
	0x67, 0x12, 0x12, 0x0a, 0x04, 0x64, 0x61, 0x74, 0x61, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0c, 0x52,
	0x04, 0x64, 0x61, 0x74, 0x61, 0x12, 0x14, 0x0a, 0x05, 0x63, 0x6c, 0x6f, 0x73, 0x65, 0x18, 0x03,
	0x20, 0x01, 0x28, 0x08, 0x52, 0x05, 0x63, 0x6c, 0x6f, 0x73, 0x65, 0x22, 0xbb, 0x01, 0x0a, 0x0a,
	0x52, 0x65, 0x67, 0x69, 0x73, 0x74, 0x65, 0x72, 0x4f, 0x70, 0x12, 0x2c, 0x0a, 0x06, 0x74, 0x61,
	0x72, 0x67, 0x65, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x12, 0x2e, 0x67, 0x72, 0x70,
	0x63, 0x74, 0x75, 0x6e, 0x6e, 0x65, 0x6c, 0x2e, 0x54, 0x61, 0x72, 0x67, 0x65, 0x74, 0x48, 0x00,
	0x52, 0x06, 0x74, 0x61, 0x72, 0x67, 0x65, 0x74, 0x12, 0x2f, 0x0a, 0x07, 0x73, 0x65, 0x73, 0x73,
	0x69, 0x6f, 0x6e, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x13, 0x2e, 0x67, 0x72, 0x70, 0x63,
	0x74, 0x75, 0x6e, 0x6e, 0x65, 0x6c, 0x2e, 0x53, 0x65, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x48, 0x00,
	0x52, 0x07, 0x73, 0x65, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x12, 0x3e, 0x0a, 0x0c, 0x73, 0x75, 0x62,
	0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x18, 0x2e, 0x67, 0x72, 0x70, 0x63, 0x74, 0x75, 0x6e, 0x6e, 0x65, 0x6c, 0x2e, 0x53, 0x75, 0x62,
	0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x48, 0x00, 0x52, 0x0c, 0x73, 0x75, 0x62,
	0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x42, 0x0e, 0x0a, 0x0c, 0x52, 0x65, 0x67,
	0x69, 0x73, 0x74, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x22, 0xca, 0x01, 0x0a, 0x06, 0x54, 0x61,
	0x72, 0x67, 0x65, 0x74, 0x12, 0x2b, 0x0a, 0x02, 0x6f, 0x70, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0e,
	0x32, 0x1b, 0x2e, 0x67, 0x72, 0x70, 0x63, 0x74, 0x75, 0x6e, 0x6e, 0x65, 0x6c, 0x2e, 0x54, 0x61,
	0x72, 0x67, 0x65, 0x74, 0x2e, 0x54, 0x61, 0x72, 0x67, 0x65, 0x74, 0x4f, 0x70, 0x52, 0x02, 0x6f,
	0x70, 0x12, 0x16, 0x0a, 0x06, 0x61, 0x63, 0x63, 0x65, 0x70, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x08, 0x52, 0x06, 0x61, 0x63, 0x63, 0x65, 0x70, 0x74, 0x12, 0x16, 0x0a, 0x06, 0x74, 0x61, 0x72,
	0x67, 0x65, 0x74, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x74, 0x61, 0x72, 0x67, 0x65,
	0x74, 0x12, 0x1f, 0x0a, 0x0b, 0x74, 0x61, 0x72, 0x67, 0x65, 0x74, 0x5f, 0x74, 0x79, 0x70, 0x65,
	0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x74, 0x61, 0x72, 0x67, 0x65, 0x74, 0x54, 0x79,
	0x70, 0x65, 0x12, 0x14, 0x0a, 0x05, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x18, 0x05, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x05, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x22, 0x2c, 0x0a, 0x08, 0x54, 0x61, 0x72, 0x67,
	0x65, 0x74, 0x4f, 0x70, 0x12, 0x0b, 0x0a, 0x07, 0x55, 0x4e, 0x4b, 0x4e, 0x4f, 0x57, 0x4e, 0x10,
	0x00, 0x12, 0x07, 0x0a, 0x03, 0x41, 0x44, 0x44, 0x10, 0x01, 0x12, 0x0a, 0x0a, 0x06, 0x52, 0x45,
	0x4d, 0x4f, 0x56, 0x45, 0x10, 0x02, 0x22, 0x82, 0x01, 0x0a, 0x07, 0x53, 0x65, 0x73, 0x73, 0x69,
	0x6f, 0x6e, 0x12, 0x10, 0x0a, 0x03, 0x74, 0x61, 0x67, 0x18, 0x01, 0x20, 0x01, 0x28, 0x05, 0x52,
	0x03, 0x74, 0x61, 0x67, 0x12, 0x16, 0x0a, 0x06, 0x61, 0x63, 0x63, 0x65, 0x70, 0x74, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x08, 0x52, 0x06, 0x61, 0x63, 0x63, 0x65, 0x70, 0x74, 0x12, 0x16, 0x0a, 0x06,
	0x74, 0x61, 0x72, 0x67, 0x65, 0x74, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x74, 0x61,
	0x72, 0x67, 0x65, 0x74, 0x12, 0x1f, 0x0a, 0x0b, 0x74, 0x61, 0x72, 0x67, 0x65, 0x74, 0x5f, 0x74,
	0x79, 0x70, 0x65, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x74, 0x61, 0x72, 0x67, 0x65,
	0x74, 0x54, 0x79, 0x70, 0x65, 0x12, 0x14, 0x0a, 0x05, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x18, 0x05,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x22, 0xd3, 0x01, 0x0a, 0x0c,
	0x53, 0x75, 0x62, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x37, 0x0a, 0x02,
	0x6f, 0x70, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x27, 0x2e, 0x67, 0x72, 0x70, 0x63, 0x74,
	0x75, 0x6e, 0x6e, 0x65, 0x6c, 0x2e, 0x53, 0x75, 0x62, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x69,
	0x6f, 0x6e, 0x2e, 0x53, 0x75, 0x62, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x4f,
	0x70, 0x52, 0x02, 0x6f, 0x70, 0x12, 0x16, 0x0a, 0x06, 0x61, 0x63, 0x63, 0x65, 0x70, 0x74, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x08, 0x52, 0x06, 0x61, 0x63, 0x63, 0x65, 0x70, 0x74, 0x12, 0x1f, 0x0a,
	0x0b, 0x74, 0x61, 0x72, 0x67, 0x65, 0x74, 0x5f, 0x74, 0x79, 0x70, 0x65, 0x18, 0x03, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x0a, 0x74, 0x61, 0x72, 0x67, 0x65, 0x74, 0x54, 0x79, 0x70, 0x65, 0x12, 0x14,
	0x0a, 0x05, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x65,
	0x72, 0x72, 0x6f, 0x72, 0x22, 0x3b, 0x0a, 0x0e, 0x53, 0x75, 0x62, 0x73, 0x63, 0x72, 0x69, 0x70,
	0x74, 0x69, 0x6f, 0x6e, 0x4f, 0x70, 0x12, 0x0b, 0x0a, 0x07, 0x55, 0x4e, 0x4b, 0x4e, 0x4f, 0x57,
	0x4e, 0x10, 0x00, 0x12, 0x0c, 0x0a, 0x08, 0x53, 0x55, 0x42, 0x43, 0x52, 0x49, 0x42, 0x45, 0x10,
	0x01, 0x12, 0x0e, 0x0a, 0x0a, 0x55, 0x4e, 0x53, 0x55, 0x42, 0x43, 0x52, 0x49, 0x42, 0x45, 0x10,
	0x02, 0x2a, 0x52, 0x0a, 0x0a, 0x54, 0x61, 0x72, 0x67, 0x65, 0x74, 0x54, 0x79, 0x70, 0x65, 0x12,
	0x0b, 0x0a, 0x07, 0x55, 0x4e, 0x4b, 0x4e, 0x4f, 0x57, 0x4e, 0x10, 0x00, 0x12, 0x07, 0x0a, 0x03,
	0x53, 0x53, 0x48, 0x10, 0x16, 0x12, 0x0d, 0x0a, 0x08, 0x4f, 0x50, 0x45, 0x4e, 0x46, 0x4c, 0x4f,
	0x57, 0x10, 0xfd, 0x33, 0x12, 0x0e, 0x0a, 0x09, 0x47, 0x4e, 0x4d, 0x49, 0x5f, 0x47, 0x4e, 0x4f,
	0x49, 0x10, 0xfb, 0x48, 0x12, 0x0f, 0x0a, 0x0a, 0x50, 0x34, 0x5f, 0x52, 0x55, 0x4e, 0x54, 0x49,
	0x4d, 0x45, 0x10, 0xd7, 0x4a, 0x32, 0x7a, 0x0a, 0x06, 0x54, 0x75, 0x6e, 0x6e, 0x65, 0x6c, 0x12,
	0x3e, 0x0a, 0x08, 0x52, 0x65, 0x67, 0x69, 0x73, 0x74, 0x65, 0x72, 0x12, 0x16, 0x2e, 0x67, 0x72,
	0x70, 0x63, 0x74, 0x75, 0x6e, 0x6e, 0x65, 0x6c, 0x2e, 0x52, 0x65, 0x67, 0x69, 0x73, 0x74, 0x65,
	0x72, 0x4f, 0x70, 0x1a, 0x16, 0x2e, 0x67, 0x72, 0x70, 0x63, 0x74, 0x75, 0x6e, 0x6e, 0x65, 0x6c,
	0x2e, 0x52, 0x65, 0x67, 0x69, 0x73, 0x74, 0x65, 0x72, 0x4f, 0x70, 0x28, 0x01, 0x30, 0x01, 0x12,
	0x30, 0x0a, 0x06, 0x54, 0x75, 0x6e, 0x6e, 0x65, 0x6c, 0x12, 0x10, 0x2e, 0x67, 0x72, 0x70, 0x63,
	0x74, 0x75, 0x6e, 0x6e, 0x65, 0x6c, 0x2e, 0x44, 0x61, 0x74, 0x61, 0x1a, 0x10, 0x2e, 0x67, 0x72,
	0x70, 0x63, 0x74, 0x75, 0x6e, 0x6e, 0x65, 0x6c, 0x2e, 0x44, 0x61, 0x74, 0x61, 0x28, 0x01, 0x30,
	0x01, 0x42, 0x19, 0x5a, 0x17, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x74, 0x75, 0x6e, 0x6e, 0x65,
	0x6c, 0x3b, 0x67, 0x72, 0x70, 0x63, 0x74, 0x75, 0x6e, 0x6e, 0x65, 0x6c, 0x62, 0x06, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_proto_tunnel_tunnel_proto_rawDescOnce sync.Once
	file_proto_tunnel_tunnel_proto_rawDescData = file_proto_tunnel_tunnel_proto_rawDesc
)

func file_proto_tunnel_tunnel_proto_rawDescGZIP() []byte {
	file_proto_tunnel_tunnel_proto_rawDescOnce.Do(func() {
		file_proto_tunnel_tunnel_proto_rawDescData = protoimpl.X.CompressGZIP(file_proto_tunnel_tunnel_proto_rawDescData)
	})
	return file_proto_tunnel_tunnel_proto_rawDescData
}

var file_proto_tunnel_tunnel_proto_enumTypes = make([]protoimpl.EnumInfo, 3)
var file_proto_tunnel_tunnel_proto_msgTypes = make([]protoimpl.MessageInfo, 5)
var file_proto_tunnel_tunnel_proto_goTypes = []interface{}{
	(TargetType)(0),                  // 0: grpctunnel.TargetType
	(Target_TargetOp)(0),             // 1: grpctunnel.Target.TargetOp
	(Subscription_SubscriptionOp)(0), // 2: grpctunnel.Subscription.SubscriptionOp
	(*Data)(nil),                     // 3: grpctunnel.Data
	(*RegisterOp)(nil),               // 4: grpctunnel.RegisterOp
	(*Target)(nil),                   // 5: grpctunnel.Target
	(*Session)(nil),                  // 6: grpctunnel.Session
	(*Subscription)(nil),             // 7: grpctunnel.Subscription
}
var file_proto_tunnel_tunnel_proto_depIdxs = []int32{
	5, // 0: grpctunnel.RegisterOp.target:type_name -> grpctunnel.Target
	6, // 1: grpctunnel.RegisterOp.session:type_name -> grpctunnel.Session
	7, // 2: grpctunnel.RegisterOp.subscription:type_name -> grpctunnel.Subscription
	1, // 3: grpctunnel.Target.op:type_name -> grpctunnel.Target.TargetOp
	2, // 4: grpctunnel.Subscription.op:type_name -> grpctunnel.Subscription.SubscriptionOp
	4, // 5: grpctunnel.Tunnel.Register:input_type -> grpctunnel.RegisterOp
	3, // 6: grpctunnel.Tunnel.Tunnel:input_type -> grpctunnel.Data
	4, // 7: grpctunnel.Tunnel.Register:output_type -> grpctunnel.RegisterOp
	3, // 8: grpctunnel.Tunnel.Tunnel:output_type -> grpctunnel.Data
	7, // [7:9] is the sub-list for method output_type
	5, // [5:7] is the sub-list for method input_type
	5, // [5:5] is the sub-list for extension type_name
	5, // [5:5] is the sub-list for extension extendee
	0, // [0:5] is the sub-list for field type_name
}

func init() { file_proto_tunnel_tunnel_proto_init() }
func file_proto_tunnel_tunnel_proto_init() {
	if File_proto_tunnel_tunnel_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_proto_tunnel_tunnel_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Data); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_proto_tunnel_tunnel_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RegisterOp); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_proto_tunnel_tunnel_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Target); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_proto_tunnel_tunnel_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Session); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_proto_tunnel_tunnel_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Subscription); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	file_proto_tunnel_tunnel_proto_msgTypes[1].OneofWrappers = []interface{}{
		(*RegisterOp_Target)(nil),
		(*RegisterOp_Session)(nil),
		(*RegisterOp_Subscription)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_proto_tunnel_tunnel_proto_rawDesc,
			NumEnums:      3,
			NumMessages:   5,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_proto_tunnel_tunnel_proto_goTypes,
		DependencyIndexes: file_proto_tunnel_tunnel_proto_depIdxs,
		EnumInfos:         file_proto_tunnel_tunnel_proto_enumTypes,
		MessageInfos:      file_proto_tunnel_tunnel_proto_msgTypes,
	}.Build()
	File_proto_tunnel_tunnel_proto = out.File
	file_proto_tunnel_tunnel_proto_rawDesc = nil
	file_proto_tunnel_tunnel_proto_goTypes = nil
	file_proto_tunnel_tunnel_proto_depIdxs = nil
}
