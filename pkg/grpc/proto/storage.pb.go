// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        v3.21.12
// source: proto/storage.proto

package gophkeeper

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

type ErrorCode int32

const (
	ErrorCode_ERROR_CODE_OK ErrorCode = 0
)

// Enum value maps for ErrorCode.
var (
	ErrorCode_name = map[int32]string{
		0: "ERROR_CODE_OK",
	}
	ErrorCode_value = map[string]int32{
		"ERROR_CODE_OK": 0,
	}
)

func (x ErrorCode) Enum() *ErrorCode {
	p := new(ErrorCode)
	*p = x
	return p
}

func (x ErrorCode) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (ErrorCode) Descriptor() protoreflect.EnumDescriptor {
	return file_proto_storage_proto_enumTypes[0].Descriptor()
}

func (ErrorCode) Type() protoreflect.EnumType {
	return &file_proto_storage_proto_enumTypes[0]
}

func (x ErrorCode) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use ErrorCode.Descriptor instead.
func (ErrorCode) EnumDescriptor() ([]byte, []int) {
	return file_proto_storage_proto_rawDescGZIP(), []int{0}
}

type Resource struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id        *string `protobuf:"bytes,1,opt,name=id,proto3,oneof" json:"id,omitempty"`
	Data      []byte  `protobuf:"bytes,2,opt,name=data,proto3,oneof" json:"data,omitempty"`
	IsDeleted *bool   `protobuf:"varint,3,opt,name=is_deleted,json=isDeleted,proto3,oneof" json:"is_deleted,omitempty"`
}

func (x *Resource) Reset() {
	*x = Resource{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_storage_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Resource) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Resource) ProtoMessage() {}

func (x *Resource) ProtoReflect() protoreflect.Message {
	mi := &file_proto_storage_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Resource.ProtoReflect.Descriptor instead.
func (*Resource) Descriptor() ([]byte, []int) {
	return file_proto_storage_proto_rawDescGZIP(), []int{0}
}

func (x *Resource) GetId() string {
	if x != nil && x.Id != nil {
		return *x.Id
	}
	return ""
}

func (x *Resource) GetData() []byte {
	if x != nil {
		return x.Data
	}
	return nil
}

func (x *Resource) GetIsDeleted() bool {
	if x != nil && x.IsDeleted != nil {
		return *x.IsDeleted
	}
	return false
}

type DataUpdate struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
}

func (x *DataUpdate) Reset() {
	*x = DataUpdate{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_storage_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DataUpdate) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DataUpdate) ProtoMessage() {}

func (x *DataUpdate) ProtoReflect() protoreflect.Message {
	mi := &file_proto_storage_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DataUpdate.ProtoReflect.Descriptor instead.
func (*DataUpdate) Descriptor() ([]byte, []int) {
	return file_proto_storage_proto_rawDescGZIP(), []int{1}
}

func (x *DataUpdate) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

type ListRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *ListRequest) Reset() {
	*x = ListRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_storage_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ListRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ListRequest) ProtoMessage() {}

func (x *ListRequest) ProtoReflect() protoreflect.Message {
	mi := &file_proto_storage_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ListRequest.ProtoReflect.Descriptor instead.
func (*ListRequest) Descriptor() ([]byte, []int) {
	return file_proto_storage_proto_rawDescGZIP(), []int{2}
}

type ResourceOperationData struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Types that are assignable to Data:
	//
	//	*ResourceOperationData_Meta
	//	*ResourceOperationData_Chunk
	Data isResourceOperationData_Data `protobuf_oneof:"data"`
}

func (x *ResourceOperationData) Reset() {
	*x = ResourceOperationData{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_storage_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ResourceOperationData) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ResourceOperationData) ProtoMessage() {}

func (x *ResourceOperationData) ProtoReflect() protoreflect.Message {
	mi := &file_proto_storage_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ResourceOperationData.ProtoReflect.Descriptor instead.
func (*ResourceOperationData) Descriptor() ([]byte, []int) {
	return file_proto_storage_proto_rawDescGZIP(), []int{3}
}

func (m *ResourceOperationData) GetData() isResourceOperationData_Data {
	if m != nil {
		return m.Data
	}
	return nil
}

func (x *ResourceOperationData) GetMeta() *ResourceOperationData_ResourceMeta {
	if x, ok := x.GetData().(*ResourceOperationData_Meta); ok {
		return x.Meta
	}
	return nil
}

func (x *ResourceOperationData) GetChunk() *ResourceOperationData_DataChunk {
	if x, ok := x.GetData().(*ResourceOperationData_Chunk); ok {
		return x.Chunk
	}
	return nil
}

type isResourceOperationData_Data interface {
	isResourceOperationData_Data()
}

type ResourceOperationData_Meta struct {
	Meta *ResourceOperationData_ResourceMeta `protobuf:"bytes,1,opt,name=meta,proto3,oneof"`
}

type ResourceOperationData_Chunk struct {
	Chunk *ResourceOperationData_DataChunk `protobuf:"bytes,2,opt,name=chunk,proto3,oneof"`
}

func (*ResourceOperationData_Meta) isResourceOperationData_Data() {}

func (*ResourceOperationData_Chunk) isResourceOperationData_Data() {}

type ResourceOperationResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Types that are assignable to Result:
	//
	//	*ResourceOperationResponse_ErrorCode
	//	*ResourceOperationResponse_Resource
	Result isResourceOperationResponse_Result `protobuf_oneof:"result"`
}

func (x *ResourceOperationResponse) Reset() {
	*x = ResourceOperationResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_storage_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ResourceOperationResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ResourceOperationResponse) ProtoMessage() {}

func (x *ResourceOperationResponse) ProtoReflect() protoreflect.Message {
	mi := &file_proto_storage_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ResourceOperationResponse.ProtoReflect.Descriptor instead.
func (*ResourceOperationResponse) Descriptor() ([]byte, []int) {
	return file_proto_storage_proto_rawDescGZIP(), []int{4}
}

func (m *ResourceOperationResponse) GetResult() isResourceOperationResponse_Result {
	if m != nil {
		return m.Result
	}
	return nil
}

func (x *ResourceOperationResponse) GetErrorCode() int32 {
	if x, ok := x.GetResult().(*ResourceOperationResponse_ErrorCode); ok {
		return x.ErrorCode
	}
	return 0
}

func (x *ResourceOperationResponse) GetResource() *Resource {
	if x, ok := x.GetResult().(*ResourceOperationResponse_Resource); ok {
		return x.Resource
	}
	return nil
}

type isResourceOperationResponse_Result interface {
	isResourceOperationResponse_Result()
}

type ResourceOperationResponse_ErrorCode struct {
	ErrorCode int32 `protobuf:"varint,1,opt,name=error_code,json=errorCode,proto3,oneof"`
}

type ResourceOperationResponse_Resource struct {
	Resource *Resource `protobuf:"bytes,2,opt,name=resource,proto3,oneof"`
}

func (*ResourceOperationResponse_ErrorCode) isResourceOperationResponse_Result() {}

func (*ResourceOperationResponse_Resource) isResourceOperationResponse_Result() {}

type ResourceOperationData_ResourceMeta struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Salt             []byte  `protobuf:"bytes,1,opt,name=salt,proto3,oneof" json:"salt,omitempty"`
	ResourceByteSize *uint64 `protobuf:"varint,2,opt,name=resource_byte_size,json=resourceByteSize,proto3,oneof" json:"resource_byte_size,omitempty"`
}

func (x *ResourceOperationData_ResourceMeta) Reset() {
	*x = ResourceOperationData_ResourceMeta{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_storage_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ResourceOperationData_ResourceMeta) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ResourceOperationData_ResourceMeta) ProtoMessage() {}

func (x *ResourceOperationData_ResourceMeta) ProtoReflect() protoreflect.Message {
	mi := &file_proto_storage_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ResourceOperationData_ResourceMeta.ProtoReflect.Descriptor instead.
func (*ResourceOperationData_ResourceMeta) Descriptor() ([]byte, []int) {
	return file_proto_storage_proto_rawDescGZIP(), []int{3, 0}
}

func (x *ResourceOperationData_ResourceMeta) GetSalt() []byte {
	if x != nil {
		return x.Salt
	}
	return nil
}

func (x *ResourceOperationData_ResourceMeta) GetResourceByteSize() uint64 {
	if x != nil && x.ResourceByteSize != nil {
		return *x.ResourceByteSize
	}
	return 0
}

type ResourceOperationData_DataChunk struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Data []byte `protobuf:"bytes,1,opt,name=data,proto3" json:"data,omitempty"`
}

func (x *ResourceOperationData_DataChunk) Reset() {
	*x = ResourceOperationData_DataChunk{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_storage_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ResourceOperationData_DataChunk) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ResourceOperationData_DataChunk) ProtoMessage() {}

func (x *ResourceOperationData_DataChunk) ProtoReflect() protoreflect.Message {
	mi := &file_proto_storage_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ResourceOperationData_DataChunk.ProtoReflect.Descriptor instead.
func (*ResourceOperationData_DataChunk) Descriptor() ([]byte, []int) {
	return file_proto_storage_proto_rawDescGZIP(), []int{3, 1}
}

func (x *ResourceOperationData_DataChunk) GetData() []byte {
	if x != nil {
		return x.Data
	}
	return nil
}

var File_proto_storage_proto protoreflect.FileDescriptor

var file_proto_storage_proto_rawDesc = []byte{
	0x0a, 0x13, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x73, 0x74, 0x6f, 0x72, 0x61, 0x67, 0x65, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0a, 0x67, 0x6f, 0x70, 0x68, 0x6b, 0x65, 0x65, 0x70, 0x65,
	0x72, 0x22, 0x7b, 0x0a, 0x08, 0x52, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x12, 0x13, 0x0a,
	0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x48, 0x00, 0x52, 0x02, 0x69, 0x64, 0x88,
	0x01, 0x01, 0x12, 0x17, 0x0a, 0x04, 0x64, 0x61, 0x74, 0x61, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0c,
	0x48, 0x01, 0x52, 0x04, 0x64, 0x61, 0x74, 0x61, 0x88, 0x01, 0x01, 0x12, 0x22, 0x0a, 0x0a, 0x69,
	0x73, 0x5f, 0x64, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x64, 0x18, 0x03, 0x20, 0x01, 0x28, 0x08, 0x48,
	0x02, 0x52, 0x09, 0x69, 0x73, 0x44, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x64, 0x88, 0x01, 0x01, 0x42,
	0x05, 0x0a, 0x03, 0x5f, 0x69, 0x64, 0x42, 0x07, 0x0a, 0x05, 0x5f, 0x64, 0x61, 0x74, 0x61, 0x42,
	0x0d, 0x0a, 0x0b, 0x5f, 0x69, 0x73, 0x5f, 0x64, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x64, 0x22, 0x1c,
	0x0a, 0x0a, 0x44, 0x61, 0x74, 0x61, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x12, 0x0e, 0x0a, 0x02,
	0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x64, 0x22, 0x0d, 0x0a, 0x0b,
	0x4c, 0x69, 0x73, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x22, 0xc7, 0x02, 0x0a, 0x15,
	0x52, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x4f, 0x70, 0x65, 0x72, 0x61, 0x74, 0x69, 0x6f,
	0x6e, 0x44, 0x61, 0x74, 0x61, 0x12, 0x44, 0x0a, 0x04, 0x6d, 0x65, 0x74, 0x61, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x2e, 0x2e, 0x67, 0x6f, 0x70, 0x68, 0x6b, 0x65, 0x65, 0x70, 0x65, 0x72,
	0x2e, 0x52, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x4f, 0x70, 0x65, 0x72, 0x61, 0x74, 0x69,
	0x6f, 0x6e, 0x44, 0x61, 0x74, 0x61, 0x2e, 0x52, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x4d,
	0x65, 0x74, 0x61, 0x48, 0x00, 0x52, 0x04, 0x6d, 0x65, 0x74, 0x61, 0x12, 0x43, 0x0a, 0x05, 0x63,
	0x68, 0x75, 0x6e, 0x6b, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x2b, 0x2e, 0x67, 0x6f, 0x70,
	0x68, 0x6b, 0x65, 0x65, 0x70, 0x65, 0x72, 0x2e, 0x52, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65,
	0x4f, 0x70, 0x65, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x44, 0x61, 0x74, 0x61, 0x2e, 0x44, 0x61,
	0x74, 0x61, 0x43, 0x68, 0x75, 0x6e, 0x6b, 0x48, 0x00, 0x52, 0x05, 0x63, 0x68, 0x75, 0x6e, 0x6b,
	0x1a, 0x7a, 0x0a, 0x0c, 0x52, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x4d, 0x65, 0x74, 0x61,
	0x12, 0x17, 0x0a, 0x04, 0x73, 0x61, 0x6c, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x48, 0x00,
	0x52, 0x04, 0x73, 0x61, 0x6c, 0x74, 0x88, 0x01, 0x01, 0x12, 0x31, 0x0a, 0x12, 0x72, 0x65, 0x73,
	0x6f, 0x75, 0x72, 0x63, 0x65, 0x5f, 0x62, 0x79, 0x74, 0x65, 0x5f, 0x73, 0x69, 0x7a, 0x65, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x04, 0x48, 0x01, 0x52, 0x10, 0x72, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63,
	0x65, 0x42, 0x79, 0x74, 0x65, 0x53, 0x69, 0x7a, 0x65, 0x88, 0x01, 0x01, 0x42, 0x07, 0x0a, 0x05,
	0x5f, 0x73, 0x61, 0x6c, 0x74, 0x42, 0x15, 0x0a, 0x13, 0x5f, 0x72, 0x65, 0x73, 0x6f, 0x75, 0x72,
	0x63, 0x65, 0x5f, 0x62, 0x79, 0x74, 0x65, 0x5f, 0x73, 0x69, 0x7a, 0x65, 0x1a, 0x1f, 0x0a, 0x09,
	0x44, 0x61, 0x74, 0x61, 0x43, 0x68, 0x75, 0x6e, 0x6b, 0x12, 0x12, 0x0a, 0x04, 0x64, 0x61, 0x74,
	0x61, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x04, 0x64, 0x61, 0x74, 0x61, 0x42, 0x06, 0x0a,
	0x04, 0x64, 0x61, 0x74, 0x61, 0x22, 0x7a, 0x0a, 0x19, 0x52, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63,
	0x65, 0x4f, 0x70, 0x65, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e,
	0x73, 0x65, 0x12, 0x1f, 0x0a, 0x0a, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x5f, 0x63, 0x6f, 0x64, 0x65,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x05, 0x48, 0x00, 0x52, 0x09, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x43,
	0x6f, 0x64, 0x65, 0x12, 0x32, 0x0a, 0x08, 0x72, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x14, 0x2e, 0x67, 0x6f, 0x70, 0x68, 0x6b, 0x65, 0x65, 0x70,
	0x65, 0x72, 0x2e, 0x52, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x48, 0x00, 0x52, 0x08, 0x72,
	0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x42, 0x08, 0x0a, 0x06, 0x72, 0x65, 0x73, 0x75, 0x6c,
	0x74, 0x2a, 0x1e, 0x0a, 0x09, 0x45, 0x72, 0x72, 0x6f, 0x72, 0x43, 0x6f, 0x64, 0x65, 0x12, 0x11,
	0x0a, 0x0d, 0x45, 0x52, 0x52, 0x4f, 0x52, 0x5f, 0x43, 0x4f, 0x44, 0x45, 0x5f, 0x4f, 0x4b, 0x10,
	0x00, 0x32, 0x9e, 0x02, 0x0a, 0x07, 0x53, 0x74, 0x6f, 0x72, 0x61, 0x67, 0x65, 0x12, 0x37, 0x0a,
	0x04, 0x4c, 0x69, 0x73, 0x74, 0x12, 0x17, 0x2e, 0x67, 0x6f, 0x70, 0x68, 0x6b, 0x65, 0x65, 0x70,
	0x65, 0x72, 0x2e, 0x4c, 0x69, 0x73, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x14,
	0x2e, 0x67, 0x6f, 0x70, 0x68, 0x6b, 0x65, 0x65, 0x70, 0x65, 0x72, 0x2e, 0x52, 0x65, 0x73, 0x6f,
	0x75, 0x72, 0x63, 0x65, 0x30, 0x01, 0x12, 0x51, 0x0a, 0x03, 0x41, 0x64, 0x64, 0x12, 0x21, 0x2e,
	0x67, 0x6f, 0x70, 0x68, 0x6b, 0x65, 0x65, 0x70, 0x65, 0x72, 0x2e, 0x52, 0x65, 0x73, 0x6f, 0x75,
	0x72, 0x63, 0x65, 0x4f, 0x70, 0x65, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x44, 0x61, 0x74, 0x61,
	0x1a, 0x25, 0x2e, 0x67, 0x6f, 0x70, 0x68, 0x6b, 0x65, 0x65, 0x70, 0x65, 0x72, 0x2e, 0x52, 0x65,
	0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x4f, 0x70, 0x65, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x52,
	0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x28, 0x01, 0x12, 0x40, 0x0a, 0x03, 0x47, 0x65, 0x74,
	0x12, 0x14, 0x2e, 0x67, 0x6f, 0x70, 0x68, 0x6b, 0x65, 0x65, 0x70, 0x65, 0x72, 0x2e, 0x52, 0x65,
	0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x1a, 0x21, 0x2e, 0x67, 0x6f, 0x70, 0x68, 0x6b, 0x65, 0x65,
	0x70, 0x65, 0x72, 0x2e, 0x52, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x4f, 0x70, 0x65, 0x72,
	0x61, 0x74, 0x69, 0x6f, 0x6e, 0x44, 0x61, 0x74, 0x61, 0x30, 0x01, 0x12, 0x45, 0x0a, 0x06, 0x44,
	0x65, 0x6c, 0x65, 0x74, 0x65, 0x12, 0x14, 0x2e, 0x67, 0x6f, 0x70, 0x68, 0x6b, 0x65, 0x65, 0x70,
	0x65, 0x72, 0x2e, 0x52, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x1a, 0x25, 0x2e, 0x67, 0x6f,
	0x70, 0x68, 0x6b, 0x65, 0x65, 0x70, 0x65, 0x72, 0x2e, 0x52, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63,
	0x65, 0x4f, 0x70, 0x65, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e,
	0x73, 0x65, 0x42, 0x15, 0x5a, 0x13, 0x70, 0x6b, 0x67, 0x2f, 0x67, 0x72, 0x70, 0x63, 0x2f, 0x67,
	0x6f, 0x70, 0x68, 0x6b, 0x65, 0x65, 0x70, 0x65, 0x72, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x33,
}

var (
	file_proto_storage_proto_rawDescOnce sync.Once
	file_proto_storage_proto_rawDescData = file_proto_storage_proto_rawDesc
)

func file_proto_storage_proto_rawDescGZIP() []byte {
	file_proto_storage_proto_rawDescOnce.Do(func() {
		file_proto_storage_proto_rawDescData = protoimpl.X.CompressGZIP(file_proto_storage_proto_rawDescData)
	})
	return file_proto_storage_proto_rawDescData
}

var file_proto_storage_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_proto_storage_proto_msgTypes = make([]protoimpl.MessageInfo, 7)
var file_proto_storage_proto_goTypes = []interface{}{
	(ErrorCode)(0),                             // 0: gophkeeper.ErrorCode
	(*Resource)(nil),                           // 1: gophkeeper.Resource
	(*DataUpdate)(nil),                         // 2: gophkeeper.DataUpdate
	(*ListRequest)(nil),                        // 3: gophkeeper.ListRequest
	(*ResourceOperationData)(nil),              // 4: gophkeeper.ResourceOperationData
	(*ResourceOperationResponse)(nil),          // 5: gophkeeper.ResourceOperationResponse
	(*ResourceOperationData_ResourceMeta)(nil), // 6: gophkeeper.ResourceOperationData.ResourceMeta
	(*ResourceOperationData_DataChunk)(nil),    // 7: gophkeeper.ResourceOperationData.DataChunk
}
var file_proto_storage_proto_depIdxs = []int32{
	6, // 0: gophkeeper.ResourceOperationData.meta:type_name -> gophkeeper.ResourceOperationData.ResourceMeta
	7, // 1: gophkeeper.ResourceOperationData.chunk:type_name -> gophkeeper.ResourceOperationData.DataChunk
	1, // 2: gophkeeper.ResourceOperationResponse.resource:type_name -> gophkeeper.Resource
	3, // 3: gophkeeper.Storage.List:input_type -> gophkeeper.ListRequest
	4, // 4: gophkeeper.Storage.Add:input_type -> gophkeeper.ResourceOperationData
	1, // 5: gophkeeper.Storage.Get:input_type -> gophkeeper.Resource
	1, // 6: gophkeeper.Storage.Delete:input_type -> gophkeeper.Resource
	1, // 7: gophkeeper.Storage.List:output_type -> gophkeeper.Resource
	5, // 8: gophkeeper.Storage.Add:output_type -> gophkeeper.ResourceOperationResponse
	4, // 9: gophkeeper.Storage.Get:output_type -> gophkeeper.ResourceOperationData
	5, // 10: gophkeeper.Storage.Delete:output_type -> gophkeeper.ResourceOperationResponse
	7, // [7:11] is the sub-list for method output_type
	3, // [3:7] is the sub-list for method input_type
	3, // [3:3] is the sub-list for extension type_name
	3, // [3:3] is the sub-list for extension extendee
	0, // [0:3] is the sub-list for field type_name
}

func init() { file_proto_storage_proto_init() }
func file_proto_storage_proto_init() {
	if File_proto_storage_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_proto_storage_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Resource); i {
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
		file_proto_storage_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DataUpdate); i {
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
		file_proto_storage_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ListRequest); i {
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
		file_proto_storage_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ResourceOperationData); i {
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
		file_proto_storage_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ResourceOperationResponse); i {
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
		file_proto_storage_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ResourceOperationData_ResourceMeta); i {
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
		file_proto_storage_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ResourceOperationData_DataChunk); i {
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
	file_proto_storage_proto_msgTypes[0].OneofWrappers = []interface{}{}
	file_proto_storage_proto_msgTypes[3].OneofWrappers = []interface{}{
		(*ResourceOperationData_Meta)(nil),
		(*ResourceOperationData_Chunk)(nil),
	}
	file_proto_storage_proto_msgTypes[4].OneofWrappers = []interface{}{
		(*ResourceOperationResponse_ErrorCode)(nil),
		(*ResourceOperationResponse_Resource)(nil),
	}
	file_proto_storage_proto_msgTypes[5].OneofWrappers = []interface{}{}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_proto_storage_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   7,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_proto_storage_proto_goTypes,
		DependencyIndexes: file_proto_storage_proto_depIdxs,
		EnumInfos:         file_proto_storage_proto_enumTypes,
		MessageInfos:      file_proto_storage_proto_msgTypes,
	}.Build()
	File_proto_storage_proto = out.File
	file_proto_storage_proto_rawDesc = nil
	file_proto_storage_proto_goTypes = nil
	file_proto_storage_proto_depIdxs = nil
}
