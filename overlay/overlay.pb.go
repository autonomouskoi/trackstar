// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.34.2
// 	protoc        v4.25.3
// source: overlay.proto

package overlay

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

type BusTopic int32

const (
	BusTopic_TRACKSTAR_OVERLAY_EVENT   BusTopic = 0
	BusTopic_TRACKSTAR_OVERLAY_REQUEST BusTopic = 1
)

// Enum value maps for BusTopic.
var (
	BusTopic_name = map[int32]string{
		0: "TRACKSTAR_OVERLAY_EVENT",
		1: "TRACKSTAR_OVERLAY_REQUEST",
	}
	BusTopic_value = map[string]int32{
		"TRACKSTAR_OVERLAY_EVENT":   0,
		"TRACKSTAR_OVERLAY_REQUEST": 1,
	}
)

func (x BusTopic) Enum() *BusTopic {
	p := new(BusTopic)
	*p = x
	return p
}

func (x BusTopic) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (BusTopic) Descriptor() protoreflect.EnumDescriptor {
	return file_overlay_proto_enumTypes[0].Descriptor()
}

func (BusTopic) Type() protoreflect.EnumType {
	return &file_overlay_proto_enumTypes[0]
}

func (x BusTopic) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use BusTopic.Descriptor instead.
func (BusTopic) EnumDescriptor() ([]byte, []int) {
	return file_overlay_proto_rawDescGZIP(), []int{0}
}

type MessageType int32

const (
	MessageType_UNSPECIFIED         MessageType = 0
	MessageType_CONFIG_UPDATED      MessageType = 1
	MessageType_SET_STYLE           MessageType = 2
	MessageType_GET_CONFIG_REQUEST  MessageType = 3
	MessageType_GET_CONFIG_RESPONSE MessageType = 4
	MessageType_SET_TRACK_COUNT     MessageType = 5
	MessageType_TRACK_COUNT_UPDATE  MessageType = 6
	MessageType_CONFIG_SET_REQ      MessageType = 7
	MessageType_CONFIG_SET_RESP     MessageType = 8
)

// Enum value maps for MessageType.
var (
	MessageType_name = map[int32]string{
		0: "UNSPECIFIED",
		1: "CONFIG_UPDATED",
		2: "SET_STYLE",
		3: "GET_CONFIG_REQUEST",
		4: "GET_CONFIG_RESPONSE",
		5: "SET_TRACK_COUNT",
		6: "TRACK_COUNT_UPDATE",
		7: "CONFIG_SET_REQ",
		8: "CONFIG_SET_RESP",
	}
	MessageType_value = map[string]int32{
		"UNSPECIFIED":         0,
		"CONFIG_UPDATED":      1,
		"SET_STYLE":           2,
		"GET_CONFIG_REQUEST":  3,
		"GET_CONFIG_RESPONSE": 4,
		"SET_TRACK_COUNT":     5,
		"TRACK_COUNT_UPDATE":  6,
		"CONFIG_SET_REQ":      7,
		"CONFIG_SET_RESP":     8,
	}
)

func (x MessageType) Enum() *MessageType {
	p := new(MessageType)
	*p = x
	return p
}

func (x MessageType) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (MessageType) Descriptor() protoreflect.EnumDescriptor {
	return file_overlay_proto_enumTypes[1].Descriptor()
}

func (MessageType) Type() protoreflect.EnumType {
	return &file_overlay_proto_enumTypes[1]
}

func (x MessageType) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use MessageType.Descriptor instead.
func (MessageType) EnumDescriptor() ([]byte, []int) {
	return file_overlay_proto_rawDescGZIP(), []int{1}
}

type StyleUpdate struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Selector string `protobuf:"bytes,1,opt,name=selector,proto3" json:"selector,omitempty"`
	Property string `protobuf:"bytes,2,opt,name=property,proto3" json:"property,omitempty"`
	Value    string `protobuf:"bytes,3,opt,name=value,proto3" json:"value,omitempty"`
}

func (x *StyleUpdate) Reset() {
	*x = StyleUpdate{}
	if protoimpl.UnsafeEnabled {
		mi := &file_overlay_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *StyleUpdate) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*StyleUpdate) ProtoMessage() {}

func (x *StyleUpdate) ProtoReflect() protoreflect.Message {
	mi := &file_overlay_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use StyleUpdate.ProtoReflect.Descriptor instead.
func (*StyleUpdate) Descriptor() ([]byte, []int) {
	return file_overlay_proto_rawDescGZIP(), []int{0}
}

func (x *StyleUpdate) GetSelector() string {
	if x != nil {
		return x.Selector
	}
	return ""
}

func (x *StyleUpdate) GetProperty() string {
	if x != nil {
		return x.Property
	}
	return ""
}

func (x *StyleUpdate) GetValue() string {
	if x != nil {
		return x.Value
	}
	return ""
}

type Config struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Styles     []*StyleUpdate `protobuf:"bytes,1,rep,name=styles,proto3" json:"styles,omitempty"`
	TrackCount uint32         `protobuf:"varint,2,opt,name=track_count,json=trackCount,proto3" json:"track_count,omitempty"`
	CustomCss  string         `protobuf:"bytes,3,opt,name=custom_css,json=customCss,proto3" json:"custom_css,omitempty"`
}

func (x *Config) Reset() {
	*x = Config{}
	if protoimpl.UnsafeEnabled {
		mi := &file_overlay_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Config) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Config) ProtoMessage() {}

func (x *Config) ProtoReflect() protoreflect.Message {
	mi := &file_overlay_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Config.ProtoReflect.Descriptor instead.
func (*Config) Descriptor() ([]byte, []int) {
	return file_overlay_proto_rawDescGZIP(), []int{1}
}

func (x *Config) GetStyles() []*StyleUpdate {
	if x != nil {
		return x.Styles
	}
	return nil
}

func (x *Config) GetTrackCount() uint32 {
	if x != nil {
		return x.TrackCount
	}
	return 0
}

func (x *Config) GetCustomCss() string {
	if x != nil {
		return x.CustomCss
	}
	return ""
}

type GetConfigRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *GetConfigRequest) Reset() {
	*x = GetConfigRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_overlay_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetConfigRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetConfigRequest) ProtoMessage() {}

func (x *GetConfigRequest) ProtoReflect() protoreflect.Message {
	mi := &file_overlay_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetConfigRequest.ProtoReflect.Descriptor instead.
func (*GetConfigRequest) Descriptor() ([]byte, []int) {
	return file_overlay_proto_rawDescGZIP(), []int{2}
}

type GetConfigResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Config *Config `protobuf:"bytes,1,opt,name=config,proto3" json:"config,omitempty"`
}

func (x *GetConfigResponse) Reset() {
	*x = GetConfigResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_overlay_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetConfigResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetConfigResponse) ProtoMessage() {}

func (x *GetConfigResponse) ProtoReflect() protoreflect.Message {
	mi := &file_overlay_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetConfigResponse.ProtoReflect.Descriptor instead.
func (*GetConfigResponse) Descriptor() ([]byte, []int) {
	return file_overlay_proto_rawDescGZIP(), []int{3}
}

func (x *GetConfigResponse) GetConfig() *Config {
	if x != nil {
		return x.Config
	}
	return nil
}

type SetTrackCount struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Count uint32 `protobuf:"varint,1,opt,name=count,proto3" json:"count,omitempty"`
}

func (x *SetTrackCount) Reset() {
	*x = SetTrackCount{}
	if protoimpl.UnsafeEnabled {
		mi := &file_overlay_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SetTrackCount) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SetTrackCount) ProtoMessage() {}

func (x *SetTrackCount) ProtoReflect() protoreflect.Message {
	mi := &file_overlay_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SetTrackCount.ProtoReflect.Descriptor instead.
func (*SetTrackCount) Descriptor() ([]byte, []int) {
	return file_overlay_proto_rawDescGZIP(), []int{4}
}

func (x *SetTrackCount) GetCount() uint32 {
	if x != nil {
		return x.Count
	}
	return 0
}

type TrackCountUpdate struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Count uint32 `protobuf:"varint,1,opt,name=count,proto3" json:"count,omitempty"`
}

func (x *TrackCountUpdate) Reset() {
	*x = TrackCountUpdate{}
	if protoimpl.UnsafeEnabled {
		mi := &file_overlay_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TrackCountUpdate) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TrackCountUpdate) ProtoMessage() {}

func (x *TrackCountUpdate) ProtoReflect() protoreflect.Message {
	mi := &file_overlay_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TrackCountUpdate.ProtoReflect.Descriptor instead.
func (*TrackCountUpdate) Descriptor() ([]byte, []int) {
	return file_overlay_proto_rawDescGZIP(), []int{5}
}

func (x *TrackCountUpdate) GetCount() uint32 {
	if x != nil {
		return x.Count
	}
	return 0
}

type ConfigSetRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Config *Config `protobuf:"bytes,1,opt,name=config,proto3" json:"config,omitempty"`
}

func (x *ConfigSetRequest) Reset() {
	*x = ConfigSetRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_overlay_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ConfigSetRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ConfigSetRequest) ProtoMessage() {}

func (x *ConfigSetRequest) ProtoReflect() protoreflect.Message {
	mi := &file_overlay_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ConfigSetRequest.ProtoReflect.Descriptor instead.
func (*ConfigSetRequest) Descriptor() ([]byte, []int) {
	return file_overlay_proto_rawDescGZIP(), []int{6}
}

func (x *ConfigSetRequest) GetConfig() *Config {
	if x != nil {
		return x.Config
	}
	return nil
}

type ConfigSetResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Config *Config `protobuf:"bytes,1,opt,name=config,proto3" json:"config,omitempty"`
}

func (x *ConfigSetResponse) Reset() {
	*x = ConfigSetResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_overlay_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ConfigSetResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ConfigSetResponse) ProtoMessage() {}

func (x *ConfigSetResponse) ProtoReflect() protoreflect.Message {
	mi := &file_overlay_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ConfigSetResponse.ProtoReflect.Descriptor instead.
func (*ConfigSetResponse) Descriptor() ([]byte, []int) {
	return file_overlay_proto_rawDescGZIP(), []int{7}
}

func (x *ConfigSetResponse) GetConfig() *Config {
	if x != nil {
		return x.Config
	}
	return nil
}

var File_overlay_proto protoreflect.FileDescriptor

var file_overlay_proto_rawDesc = []byte{
	0x0a, 0x0d, 0x6f, 0x76, 0x65, 0x72, 0x6c, 0x61, 0x79, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12,
	0x07, 0x6f, 0x76, 0x65, 0x72, 0x6c, 0x61, 0x79, 0x22, 0x5b, 0x0a, 0x0b, 0x53, 0x74, 0x79, 0x6c,
	0x65, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x12, 0x1a, 0x0a, 0x08, 0x73, 0x65, 0x6c, 0x65, 0x63,
	0x74, 0x6f, 0x72, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x73, 0x65, 0x6c, 0x65, 0x63,
	0x74, 0x6f, 0x72, 0x12, 0x1a, 0x0a, 0x08, 0x70, 0x72, 0x6f, 0x70, 0x65, 0x72, 0x74, 0x79, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x70, 0x72, 0x6f, 0x70, 0x65, 0x72, 0x74, 0x79, 0x12,
	0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05,
	0x76, 0x61, 0x6c, 0x75, 0x65, 0x22, 0x76, 0x0a, 0x06, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x12,
	0x2c, 0x0a, 0x06, 0x73, 0x74, 0x79, 0x6c, 0x65, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32,
	0x14, 0x2e, 0x6f, 0x76, 0x65, 0x72, 0x6c, 0x61, 0x79, 0x2e, 0x53, 0x74, 0x79, 0x6c, 0x65, 0x55,
	0x70, 0x64, 0x61, 0x74, 0x65, 0x52, 0x06, 0x73, 0x74, 0x79, 0x6c, 0x65, 0x73, 0x12, 0x1f, 0x0a,
	0x0b, 0x74, 0x72, 0x61, 0x63, 0x6b, 0x5f, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x0d, 0x52, 0x0a, 0x74, 0x72, 0x61, 0x63, 0x6b, 0x43, 0x6f, 0x75, 0x6e, 0x74, 0x12, 0x1d,
	0x0a, 0x0a, 0x63, 0x75, 0x73, 0x74, 0x6f, 0x6d, 0x5f, 0x63, 0x73, 0x73, 0x18, 0x03, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x09, 0x63, 0x75, 0x73, 0x74, 0x6f, 0x6d, 0x43, 0x73, 0x73, 0x22, 0x12, 0x0a,
	0x10, 0x47, 0x65, 0x74, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x22, 0x3c, 0x0a, 0x11, 0x47, 0x65, 0x74, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x52, 0x65,
	0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x27, 0x0a, 0x06, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0f, 0x2e, 0x6f, 0x76, 0x65, 0x72, 0x6c, 0x61, 0x79,
	0x2e, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x52, 0x06, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x22,
	0x25, 0x0a, 0x0d, 0x53, 0x65, 0x74, 0x54, 0x72, 0x61, 0x63, 0x6b, 0x43, 0x6f, 0x75, 0x6e, 0x74,
	0x12, 0x14, 0x0a, 0x05, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0d, 0x52,
	0x05, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x22, 0x28, 0x0a, 0x10, 0x54, 0x72, 0x61, 0x63, 0x6b, 0x43,
	0x6f, 0x75, 0x6e, 0x74, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x12, 0x14, 0x0a, 0x05, 0x63, 0x6f,
	0x75, 0x6e, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x05, 0x63, 0x6f, 0x75, 0x6e, 0x74,
	0x22, 0x3b, 0x0a, 0x10, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x53, 0x65, 0x74, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x12, 0x27, 0x0a, 0x06, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x0b, 0x32, 0x0f, 0x2e, 0x6f, 0x76, 0x65, 0x72, 0x6c, 0x61, 0x79, 0x2e, 0x43,
	0x6f, 0x6e, 0x66, 0x69, 0x67, 0x52, 0x06, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x22, 0x3c, 0x0a,
	0x11, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x53, 0x65, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e,
	0x73, 0x65, 0x12, 0x27, 0x0a, 0x06, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x0b, 0x32, 0x0f, 0x2e, 0x6f, 0x76, 0x65, 0x72, 0x6c, 0x61, 0x79, 0x2e, 0x43, 0x6f, 0x6e,
	0x66, 0x69, 0x67, 0x52, 0x06, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x2a, 0x46, 0x0a, 0x08, 0x42,
	0x75, 0x73, 0x54, 0x6f, 0x70, 0x69, 0x63, 0x12, 0x1b, 0x0a, 0x17, 0x54, 0x52, 0x41, 0x43, 0x4b,
	0x53, 0x54, 0x41, 0x52, 0x5f, 0x4f, 0x56, 0x45, 0x52, 0x4c, 0x41, 0x59, 0x5f, 0x45, 0x56, 0x45,
	0x4e, 0x54, 0x10, 0x00, 0x12, 0x1d, 0x0a, 0x19, 0x54, 0x52, 0x41, 0x43, 0x4b, 0x53, 0x54, 0x41,
	0x52, 0x5f, 0x4f, 0x56, 0x45, 0x52, 0x4c, 0x41, 0x59, 0x5f, 0x52, 0x45, 0x51, 0x55, 0x45, 0x53,
	0x54, 0x10, 0x01, 0x2a, 0xc8, 0x01, 0x0a, 0x0b, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x54,
	0x79, 0x70, 0x65, 0x12, 0x0f, 0x0a, 0x0b, 0x55, 0x4e, 0x53, 0x50, 0x45, 0x43, 0x49, 0x46, 0x49,
	0x45, 0x44, 0x10, 0x00, 0x12, 0x12, 0x0a, 0x0e, 0x43, 0x4f, 0x4e, 0x46, 0x49, 0x47, 0x5f, 0x55,
	0x50, 0x44, 0x41, 0x54, 0x45, 0x44, 0x10, 0x01, 0x12, 0x0d, 0x0a, 0x09, 0x53, 0x45, 0x54, 0x5f,
	0x53, 0x54, 0x59, 0x4c, 0x45, 0x10, 0x02, 0x12, 0x16, 0x0a, 0x12, 0x47, 0x45, 0x54, 0x5f, 0x43,
	0x4f, 0x4e, 0x46, 0x49, 0x47, 0x5f, 0x52, 0x45, 0x51, 0x55, 0x45, 0x53, 0x54, 0x10, 0x03, 0x12,
	0x17, 0x0a, 0x13, 0x47, 0x45, 0x54, 0x5f, 0x43, 0x4f, 0x4e, 0x46, 0x49, 0x47, 0x5f, 0x52, 0x45,
	0x53, 0x50, 0x4f, 0x4e, 0x53, 0x45, 0x10, 0x04, 0x12, 0x13, 0x0a, 0x0f, 0x53, 0x45, 0x54, 0x5f,
	0x54, 0x52, 0x41, 0x43, 0x4b, 0x5f, 0x43, 0x4f, 0x55, 0x4e, 0x54, 0x10, 0x05, 0x12, 0x16, 0x0a,
	0x12, 0x54, 0x52, 0x41, 0x43, 0x4b, 0x5f, 0x43, 0x4f, 0x55, 0x4e, 0x54, 0x5f, 0x55, 0x50, 0x44,
	0x41, 0x54, 0x45, 0x10, 0x06, 0x12, 0x12, 0x0a, 0x0e, 0x43, 0x4f, 0x4e, 0x46, 0x49, 0x47, 0x5f,
	0x53, 0x45, 0x54, 0x5f, 0x52, 0x45, 0x51, 0x10, 0x07, 0x12, 0x13, 0x0a, 0x0f, 0x43, 0x4f, 0x4e,
	0x46, 0x49, 0x47, 0x5f, 0x53, 0x45, 0x54, 0x5f, 0x52, 0x45, 0x53, 0x50, 0x10, 0x08, 0x42, 0x2c,
	0x5a, 0x2a, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x61, 0x75, 0x74,
	0x6f, 0x6e, 0x6f, 0x6d, 0x6f, 0x75, 0x73, 0x6b, 0x6f, 0x69, 0x2f, 0x74, 0x72, 0x61, 0x63, 0x6b,
	0x73, 0x74, 0x61, 0x72, 0x2f, 0x6f, 0x76, 0x65, 0x72, 0x6c, 0x61, 0x79, 0x62, 0x06, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_overlay_proto_rawDescOnce sync.Once
	file_overlay_proto_rawDescData = file_overlay_proto_rawDesc
)

func file_overlay_proto_rawDescGZIP() []byte {
	file_overlay_proto_rawDescOnce.Do(func() {
		file_overlay_proto_rawDescData = protoimpl.X.CompressGZIP(file_overlay_proto_rawDescData)
	})
	return file_overlay_proto_rawDescData
}

var file_overlay_proto_enumTypes = make([]protoimpl.EnumInfo, 2)
var file_overlay_proto_msgTypes = make([]protoimpl.MessageInfo, 8)
var file_overlay_proto_goTypes = []any{
	(BusTopic)(0),             // 0: overlay.BusTopic
	(MessageType)(0),          // 1: overlay.MessageType
	(*StyleUpdate)(nil),       // 2: overlay.StyleUpdate
	(*Config)(nil),            // 3: overlay.Config
	(*GetConfigRequest)(nil),  // 4: overlay.GetConfigRequest
	(*GetConfigResponse)(nil), // 5: overlay.GetConfigResponse
	(*SetTrackCount)(nil),     // 6: overlay.SetTrackCount
	(*TrackCountUpdate)(nil),  // 7: overlay.TrackCountUpdate
	(*ConfigSetRequest)(nil),  // 8: overlay.ConfigSetRequest
	(*ConfigSetResponse)(nil), // 9: overlay.ConfigSetResponse
}
var file_overlay_proto_depIdxs = []int32{
	2, // 0: overlay.Config.styles:type_name -> overlay.StyleUpdate
	3, // 1: overlay.GetConfigResponse.config:type_name -> overlay.Config
	3, // 2: overlay.ConfigSetRequest.config:type_name -> overlay.Config
	3, // 3: overlay.ConfigSetResponse.config:type_name -> overlay.Config
	4, // [4:4] is the sub-list for method output_type
	4, // [4:4] is the sub-list for method input_type
	4, // [4:4] is the sub-list for extension type_name
	4, // [4:4] is the sub-list for extension extendee
	0, // [0:4] is the sub-list for field type_name
}

func init() { file_overlay_proto_init() }
func file_overlay_proto_init() {
	if File_overlay_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_overlay_proto_msgTypes[0].Exporter = func(v any, i int) any {
			switch v := v.(*StyleUpdate); i {
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
		file_overlay_proto_msgTypes[1].Exporter = func(v any, i int) any {
			switch v := v.(*Config); i {
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
		file_overlay_proto_msgTypes[2].Exporter = func(v any, i int) any {
			switch v := v.(*GetConfigRequest); i {
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
		file_overlay_proto_msgTypes[3].Exporter = func(v any, i int) any {
			switch v := v.(*GetConfigResponse); i {
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
		file_overlay_proto_msgTypes[4].Exporter = func(v any, i int) any {
			switch v := v.(*SetTrackCount); i {
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
		file_overlay_proto_msgTypes[5].Exporter = func(v any, i int) any {
			switch v := v.(*TrackCountUpdate); i {
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
		file_overlay_proto_msgTypes[6].Exporter = func(v any, i int) any {
			switch v := v.(*ConfigSetRequest); i {
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
		file_overlay_proto_msgTypes[7].Exporter = func(v any, i int) any {
			switch v := v.(*ConfigSetResponse); i {
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
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_overlay_proto_rawDesc,
			NumEnums:      2,
			NumMessages:   8,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_overlay_proto_goTypes,
		DependencyIndexes: file_overlay_proto_depIdxs,
		EnumInfos:         file_overlay_proto_enumTypes,
		MessageInfos:      file_overlay_proto_msgTypes,
	}.Build()
	File_overlay_proto = out.File
	file_overlay_proto_rawDesc = nil
	file_overlay_proto_goTypes = nil
	file_overlay_proto_depIdxs = nil
}
