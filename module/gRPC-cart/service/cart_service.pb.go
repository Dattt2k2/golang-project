// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.35.2
// 	protoc        v5.28.3
// source: cart-service/gRPC/cart_service.proto

package service

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

type CartRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	UserId string `protobuf:"bytes,1,opt,name=user_id,json=userId,proto3" json:"user_id,omitempty"`
}

func (x *CartRequest) Reset() {
	*x = CartRequest{}
	mi := &file_cart_service_gRPC_cart_service_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *CartRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CartRequest) ProtoMessage() {}

func (x *CartRequest) ProtoReflect() protoreflect.Message {
	mi := &file_cart_service_gRPC_cart_service_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CartRequest.ProtoReflect.Descriptor instead.
func (*CartRequest) Descriptor() ([]byte, []int) {
	return file_cart_service_gRPC_cart_service_proto_rawDescGZIP(), []int{0}
}

func (x *CartRequest) GetUserId() string {
	if x != nil {
		return x.UserId
	}
	return ""
}

type CartResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Items []*CartItem `protobuf:"bytes,1,rep,name=items,proto3" json:"items,omitempty"`
}

func (x *CartResponse) Reset() {
	*x = CartResponse{}
	mi := &file_cart_service_gRPC_cart_service_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *CartResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CartResponse) ProtoMessage() {}

func (x *CartResponse) ProtoReflect() protoreflect.Message {
	mi := &file_cart_service_gRPC_cart_service_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CartResponse.ProtoReflect.Descriptor instead.
func (*CartResponse) Descriptor() ([]byte, []int) {
	return file_cart_service_gRPC_cart_service_proto_rawDescGZIP(), []int{1}
}

func (x *CartResponse) GetItems() []*CartItem {
	if x != nil {
		return x.Items
	}
	return nil
}

type CartItem struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ProductId string  `protobuf:"bytes,1,opt,name=product_id,json=productId,proto3" json:"product_id,omitempty"`
	Quantity  int32   `protobuf:"varint,2,opt,name=quantity,proto3" json:"quantity,omitempty"`
	Price     float32 `protobuf:"fixed32,3,opt,name=price,proto3" json:"price,omitempty"`
}

func (x *CartItem) Reset() {
	*x = CartItem{}
	mi := &file_cart_service_gRPC_cart_service_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *CartItem) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CartItem) ProtoMessage() {}

func (x *CartItem) ProtoReflect() protoreflect.Message {
	mi := &file_cart_service_gRPC_cart_service_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CartItem.ProtoReflect.Descriptor instead.
func (*CartItem) Descriptor() ([]byte, []int) {
	return file_cart_service_gRPC_cart_service_proto_rawDescGZIP(), []int{2}
}

func (x *CartItem) GetProductId() string {
	if x != nil {
		return x.ProductId
	}
	return ""
}

func (x *CartItem) GetQuantity() int32 {
	if x != nil {
		return x.Quantity
	}
	return 0
}

func (x *CartItem) GetPrice() float32 {
	if x != nil {
		return x.Price
	}
	return 0
}

var File_cart_service_gRPC_cart_service_proto protoreflect.FileDescriptor

var file_cart_service_gRPC_cart_service_proto_rawDesc = []byte{
	0x0a, 0x24, 0x63, 0x61, 0x72, 0x74, 0x2d, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x2f, 0x67,
	0x52, 0x50, 0x43, 0x2f, 0x63, 0x61, 0x72, 0x74, 0x5f, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x04, 0x63, 0x61, 0x72, 0x74, 0x22, 0x26, 0x0a, 0x0b,
	0x43, 0x61, 0x72, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x17, 0x0a, 0x07, 0x75,
	0x73, 0x65, 0x72, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x75, 0x73,
	0x65, 0x72, 0x49, 0x64, 0x22, 0x34, 0x0a, 0x0c, 0x43, 0x61, 0x72, 0x74, 0x52, 0x65, 0x73, 0x70,
	0x6f, 0x6e, 0x73, 0x65, 0x12, 0x24, 0x0a, 0x05, 0x69, 0x74, 0x65, 0x6d, 0x73, 0x18, 0x01, 0x20,
	0x03, 0x28, 0x0b, 0x32, 0x0e, 0x2e, 0x63, 0x61, 0x72, 0x74, 0x2e, 0x43, 0x61, 0x72, 0x74, 0x49,
	0x74, 0x65, 0x6d, 0x52, 0x05, 0x69, 0x74, 0x65, 0x6d, 0x73, 0x22, 0x5b, 0x0a, 0x08, 0x43, 0x61,
	0x72, 0x74, 0x49, 0x74, 0x65, 0x6d, 0x12, 0x1d, 0x0a, 0x0a, 0x70, 0x72, 0x6f, 0x64, 0x75, 0x63,
	0x74, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x70, 0x72, 0x6f, 0x64,
	0x75, 0x63, 0x74, 0x49, 0x64, 0x12, 0x1a, 0x0a, 0x08, 0x71, 0x75, 0x61, 0x6e, 0x74, 0x69, 0x74,
	0x79, 0x18, 0x02, 0x20, 0x01, 0x28, 0x05, 0x52, 0x08, 0x71, 0x75, 0x61, 0x6e, 0x74, 0x69, 0x74,
	0x79, 0x12, 0x14, 0x0a, 0x05, 0x70, 0x72, 0x69, 0x63, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x02,
	0x52, 0x05, 0x70, 0x72, 0x69, 0x63, 0x65, 0x32, 0x44, 0x0a, 0x0b, 0x43, 0x61, 0x72, 0x74, 0x53,
	0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x35, 0x0a, 0x0c, 0x47, 0x65, 0x74, 0x43, 0x61, 0x72,
	0x74, 0x49, 0x74, 0x65, 0x6d, 0x73, 0x12, 0x11, 0x2e, 0x63, 0x61, 0x72, 0x74, 0x2e, 0x43, 0x61,
	0x72, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x12, 0x2e, 0x63, 0x61, 0x72, 0x74,
	0x2e, 0x43, 0x61, 0x72, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x42, 0x1d, 0x5a,
	0x1b, 0x2e, 0x2f, 0x63, 0x61, 0x72, 0x74, 0x2d, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x2f,
	0x67, 0x52, 0x50, 0x43, 0x2f, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x62, 0x06, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_cart_service_gRPC_cart_service_proto_rawDescOnce sync.Once
	file_cart_service_gRPC_cart_service_proto_rawDescData = file_cart_service_gRPC_cart_service_proto_rawDesc
)

func file_cart_service_gRPC_cart_service_proto_rawDescGZIP() []byte {
	file_cart_service_gRPC_cart_service_proto_rawDescOnce.Do(func() {
		file_cart_service_gRPC_cart_service_proto_rawDescData = protoimpl.X.CompressGZIP(file_cart_service_gRPC_cart_service_proto_rawDescData)
	})
	return file_cart_service_gRPC_cart_service_proto_rawDescData
}

var file_cart_service_gRPC_cart_service_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_cart_service_gRPC_cart_service_proto_goTypes = []any{
	(*CartRequest)(nil),  // 0: cart.CartRequest
	(*CartResponse)(nil), // 1: cart.CartResponse
	(*CartItem)(nil),     // 2: cart.CartItem
}
var file_cart_service_gRPC_cart_service_proto_depIdxs = []int32{
	2, // 0: cart.CartResponse.items:type_name -> cart.CartItem
	0, // 1: cart.CartService.GetCartItems:input_type -> cart.CartRequest
	1, // 2: cart.CartService.GetCartItems:output_type -> cart.CartResponse
	2, // [2:3] is the sub-list for method output_type
	1, // [1:2] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_cart_service_gRPC_cart_service_proto_init() }
func file_cart_service_gRPC_cart_service_proto_init() {
	if File_cart_service_gRPC_cart_service_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_cart_service_gRPC_cart_service_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_cart_service_gRPC_cart_service_proto_goTypes,
		DependencyIndexes: file_cart_service_gRPC_cart_service_proto_depIdxs,
		MessageInfos:      file_cart_service_gRPC_cart_service_proto_msgTypes,
	}.Build()
	File_cart_service_gRPC_cart_service_proto = out.File
	file_cart_service_gRPC_cart_service_proto_rawDesc = nil
	file_cart_service_gRPC_cart_service_proto_goTypes = nil
	file_cart_service_gRPC_cart_service_proto_depIdxs = nil
}