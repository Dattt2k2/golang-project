// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.35.2
// 	protoc        v5.28.3
// source: module/gRPC-Product/product_service.proto

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

// Messages for product information
type ProductRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"` // Product ID
}

func (x *ProductRequest) Reset() {
	*x = ProductRequest{}
	mi := &file_module_gRPC_Product_product_service_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ProductRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ProductRequest) ProtoMessage() {}

func (x *ProductRequest) ProtoReflect() protoreflect.Message {
	mi := &file_module_gRPC_Product_product_service_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ProductRequest.ProtoReflect.Descriptor instead.
func (*ProductRequest) Descriptor() ([]byte, []int) {
	return file_module_gRPC_Product_product_service_proto_rawDescGZIP(), []int{0}
}

func (x *ProductRequest) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

type BasicProductResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id    string  `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Name  string  `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
	Price float32 `protobuf:"fixed32,3,opt,name=price,proto3" json:"price,omitempty"`
}

func (x *BasicProductResponse) Reset() {
	*x = BasicProductResponse{}
	mi := &file_module_gRPC_Product_product_service_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *BasicProductResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BasicProductResponse) ProtoMessage() {}

func (x *BasicProductResponse) ProtoReflect() protoreflect.Message {
	mi := &file_module_gRPC_Product_product_service_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use BasicProductResponse.ProtoReflect.Descriptor instead.
func (*BasicProductResponse) Descriptor() ([]byte, []int) {
	return file_module_gRPC_Product_product_service_proto_rawDescGZIP(), []int{1}
}

func (x *BasicProductResponse) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *BasicProductResponse) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *BasicProductResponse) GetPrice() float32 {
	if x != nil {
		return x.Price
	}
	return 0
}

type ProductResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id          string  `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Name        string  `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
	Price       float32 `protobuf:"fixed32,3,opt,name=price,proto3" json:"price,omitempty"`
	Description string  `protobuf:"bytes,4,opt,name=description,proto3" json:"description,omitempty"`
	Quantity    int32   `protobuf:"varint,5,opt,name=quantity,proto3" json:"quantity,omitempty"`
	ImageUrl    string  `protobuf:"bytes,6,opt,name=image_url,json=imageUrl,proto3" json:"image_url,omitempty"`
}

func (x *ProductResponse) Reset() {
	*x = ProductResponse{}
	mi := &file_module_gRPC_Product_product_service_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ProductResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ProductResponse) ProtoMessage() {}

func (x *ProductResponse) ProtoReflect() protoreflect.Message {
	mi := &file_module_gRPC_Product_product_service_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ProductResponse.ProtoReflect.Descriptor instead.
func (*ProductResponse) Descriptor() ([]byte, []int) {
	return file_module_gRPC_Product_product_service_proto_rawDescGZIP(), []int{2}
}

func (x *ProductResponse) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *ProductResponse) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *ProductResponse) GetPrice() float32 {
	if x != nil {
		return x.Price
	}
	return 0
}

func (x *ProductResponse) GetDescription() string {
	if x != nil {
		return x.Description
	}
	return ""
}

func (x *ProductResponse) GetQuantity() int32 {
	if x != nil {
		return x.Quantity
	}
	return 0
}

func (x *ProductResponse) GetImageUrl() string {
	if x != nil {
		return x.ImageUrl
	}
	return ""
}

type StockResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	InStock           bool   `protobuf:"varint,1,opt,name=in_stock,json=inStock,proto3" json:"in_stock,omitempty"`
	AvailableQuantity int32  `protobuf:"varint,2,opt,name=available_quantity,json=availableQuantity,proto3" json:"available_quantity,omitempty"`
	Message           string `protobuf:"bytes,3,opt,name=message,proto3" json:"message,omitempty"`
}

func (x *StockResponse) Reset() {
	*x = StockResponse{}
	mi := &file_module_gRPC_Product_product_service_proto_msgTypes[3]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *StockResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*StockResponse) ProtoMessage() {}

func (x *StockResponse) ProtoReflect() protoreflect.Message {
	mi := &file_module_gRPC_Product_product_service_proto_msgTypes[3]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use StockResponse.ProtoReflect.Descriptor instead.
func (*StockResponse) Descriptor() ([]byte, []int) {
	return file_module_gRPC_Product_product_service_proto_rawDescGZIP(), []int{3}
}

func (x *StockResponse) GetInStock() bool {
	if x != nil {
		return x.InStock
	}
	return false
}

func (x *StockResponse) GetAvailableQuantity() int32 {
	if x != nil {
		return x.AvailableQuantity
	}
	return 0
}

func (x *StockResponse) GetMessage() string {
	if x != nil {
		return x.Message
	}
	return ""
}

type StockStatus struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ProductId         string `protobuf:"bytes,1,opt,name=product_id,json=productId,proto3" json:"product_id,omitempty"`
	InStock           bool   `protobuf:"varint,2,opt,name=in_stock,json=inStock,proto3" json:"in_stock,omitempty"`
	AvailableQuantity int32  `protobuf:"varint,3,opt,name=available_quantity,json=availableQuantity,proto3" json:"available_quantity,omitempty"`
}

func (x *StockStatus) Reset() {
	*x = StockStatus{}
	mi := &file_module_gRPC_Product_product_service_proto_msgTypes[4]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *StockStatus) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*StockStatus) ProtoMessage() {}

func (x *StockStatus) ProtoReflect() protoreflect.Message {
	mi := &file_module_gRPC_Product_product_service_proto_msgTypes[4]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use StockStatus.ProtoReflect.Descriptor instead.
func (*StockStatus) Descriptor() ([]byte, []int) {
	return file_module_gRPC_Product_product_service_proto_rawDescGZIP(), []int{4}
}

func (x *StockStatus) GetProductId() string {
	if x != nil {
		return x.ProductId
	}
	return ""
}

func (x *StockStatus) GetInStock() bool {
	if x != nil {
		return x.InStock
	}
	return false
}

func (x *StockStatus) GetAvailableQuantity() int32 {
	if x != nil {
		return x.AvailableQuantity
	}
	return 0
}

type UpdateStockRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Items []*StockItem `protobuf:"bytes,1,rep,name=items,proto3" json:"items,omitempty"`
}

func (x *UpdateStockRequest) Reset() {
	*x = UpdateStockRequest{}
	mi := &file_module_gRPC_Product_product_service_proto_msgTypes[5]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *UpdateStockRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UpdateStockRequest) ProtoMessage() {}

func (x *UpdateStockRequest) ProtoReflect() protoreflect.Message {
	mi := &file_module_gRPC_Product_product_service_proto_msgTypes[5]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UpdateStockRequest.ProtoReflect.Descriptor instead.
func (*UpdateStockRequest) Descriptor() ([]byte, []int) {
	return file_module_gRPC_Product_product_service_proto_rawDescGZIP(), []int{5}
}

func (x *UpdateStockRequest) GetItems() []*StockItem {
	if x != nil {
		return x.Items
	}
	return nil
}

type StockItem struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ProductId string `protobuf:"bytes,1,opt,name=product_id,json=productId,proto3" json:"product_id,omitempty"`
	Quantity  int32  `protobuf:"varint,2,opt,name=quantity,proto3" json:"quantity,omitempty"`
}

func (x *StockItem) Reset() {
	*x = StockItem{}
	mi := &file_module_gRPC_Product_product_service_proto_msgTypes[6]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *StockItem) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*StockItem) ProtoMessage() {}

func (x *StockItem) ProtoReflect() protoreflect.Message {
	mi := &file_module_gRPC_Product_product_service_proto_msgTypes[6]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use StockItem.ProtoReflect.Descriptor instead.
func (*StockItem) Descriptor() ([]byte, []int) {
	return file_module_gRPC_Product_product_service_proto_rawDescGZIP(), []int{6}
}

func (x *StockItem) GetProductId() string {
	if x != nil {
		return x.ProductId
	}
	return ""
}

func (x *StockItem) GetQuantity() int32 {
	if x != nil {
		return x.Quantity
	}
	return 0
}

type UpdateStockResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	UpdateStatus []*StockUpdateStatus `protobuf:"bytes,1,rep,name=update_status,json=updateStatus,proto3" json:"update_status,omitempty"`
	Success      bool                 `protobuf:"varint,2,opt,name=success,proto3" json:"success,omitempty"`
	Message      string               `protobuf:"bytes,3,opt,name=message,proto3" json:"message,omitempty"`
}

func (x *UpdateStockResponse) Reset() {
	*x = UpdateStockResponse{}
	mi := &file_module_gRPC_Product_product_service_proto_msgTypes[7]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *UpdateStockResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UpdateStockResponse) ProtoMessage() {}

func (x *UpdateStockResponse) ProtoReflect() protoreflect.Message {
	mi := &file_module_gRPC_Product_product_service_proto_msgTypes[7]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UpdateStockResponse.ProtoReflect.Descriptor instead.
func (*UpdateStockResponse) Descriptor() ([]byte, []int) {
	return file_module_gRPC_Product_product_service_proto_rawDescGZIP(), []int{7}
}

func (x *UpdateStockResponse) GetUpdateStatus() []*StockUpdateStatus {
	if x != nil {
		return x.UpdateStatus
	}
	return nil
}

func (x *UpdateStockResponse) GetSuccess() bool {
	if x != nil {
		return x.Success
	}
	return false
}

func (x *UpdateStockResponse) GetMessage() string {
	if x != nil {
		return x.Message
	}
	return ""
}

type StockUpdateStatus struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ProductId string `protobuf:"bytes,1,opt,name=product_id,json=productId,proto3" json:"product_id,omitempty"`
	Updated   bool   `protobuf:"varint,2,opt,name=updated,proto3" json:"updated,omitempty"`
	Message   string `protobuf:"bytes,3,opt,name=message,proto3" json:"message,omitempty"`
}

func (x *StockUpdateStatus) Reset() {
	*x = StockUpdateStatus{}
	mi := &file_module_gRPC_Product_product_service_proto_msgTypes[8]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *StockUpdateStatus) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*StockUpdateStatus) ProtoMessage() {}

func (x *StockUpdateStatus) ProtoReflect() protoreflect.Message {
	mi := &file_module_gRPC_Product_product_service_proto_msgTypes[8]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use StockUpdateStatus.ProtoReflect.Descriptor instead.
func (*StockUpdateStatus) Descriptor() ([]byte, []int) {
	return file_module_gRPC_Product_product_service_proto_rawDescGZIP(), []int{8}
}

func (x *StockUpdateStatus) GetProductId() string {
	if x != nil {
		return x.ProductId
	}
	return ""
}

func (x *StockUpdateStatus) GetUpdated() bool {
	if x != nil {
		return x.Updated
	}
	return false
}

func (x *StockUpdateStatus) GetMessage() string {
	if x != nil {
		return x.Message
	}
	return ""
}

var File_module_gRPC_Product_product_service_proto protoreflect.FileDescriptor

var file_module_gRPC_Product_product_service_proto_rawDesc = []byte{
	0x0a, 0x29, 0x6d, 0x6f, 0x64, 0x75, 0x6c, 0x65, 0x2f, 0x67, 0x52, 0x50, 0x43, 0x2d, 0x50, 0x72,
	0x6f, 0x64, 0x75, 0x63, 0x74, 0x2f, 0x70, 0x72, 0x6f, 0x64, 0x75, 0x63, 0x74, 0x5f, 0x73, 0x65,
	0x72, 0x76, 0x69, 0x63, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x07, 0x70, 0x72, 0x6f,
	0x64, 0x75, 0x63, 0x74, 0x22, 0x20, 0x0a, 0x0e, 0x50, 0x72, 0x6f, 0x64, 0x75, 0x63, 0x74, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x02, 0x69, 0x64, 0x22, 0x50, 0x0a, 0x14, 0x42, 0x61, 0x73, 0x69, 0x63, 0x50,
	0x72, 0x6f, 0x64, 0x75, 0x63, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x0e,
	0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x64, 0x12, 0x12,
	0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61,
	0x6d, 0x65, 0x12, 0x14, 0x0a, 0x05, 0x70, 0x72, 0x69, 0x63, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28,
	0x02, 0x52, 0x05, 0x70, 0x72, 0x69, 0x63, 0x65, 0x22, 0xa6, 0x01, 0x0a, 0x0f, 0x50, 0x72, 0x6f,
	0x64, 0x75, 0x63, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x0e, 0x0a, 0x02,
	0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x64, 0x12, 0x12, 0x0a, 0x04,
	0x6e, 0x61, 0x6d, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65,
	0x12, 0x14, 0x0a, 0x05, 0x70, 0x72, 0x69, 0x63, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x02, 0x52,
	0x05, 0x70, 0x72, 0x69, 0x63, 0x65, 0x12, 0x20, 0x0a, 0x0b, 0x64, 0x65, 0x73, 0x63, 0x72, 0x69,
	0x70, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x64, 0x65, 0x73,
	0x63, 0x72, 0x69, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x1a, 0x0a, 0x08, 0x71, 0x75, 0x61, 0x6e,
	0x74, 0x69, 0x74, 0x79, 0x18, 0x05, 0x20, 0x01, 0x28, 0x05, 0x52, 0x08, 0x71, 0x75, 0x61, 0x6e,
	0x74, 0x69, 0x74, 0x79, 0x12, 0x1b, 0x0a, 0x09, 0x69, 0x6d, 0x61, 0x67, 0x65, 0x5f, 0x75, 0x72,
	0x6c, 0x18, 0x06, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x69, 0x6d, 0x61, 0x67, 0x65, 0x55, 0x72,
	0x6c, 0x22, 0x73, 0x0a, 0x0d, 0x53, 0x74, 0x6f, 0x63, 0x6b, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e,
	0x73, 0x65, 0x12, 0x19, 0x0a, 0x08, 0x69, 0x6e, 0x5f, 0x73, 0x74, 0x6f, 0x63, 0x6b, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x08, 0x52, 0x07, 0x69, 0x6e, 0x53, 0x74, 0x6f, 0x63, 0x6b, 0x12, 0x2d, 0x0a,
	0x12, 0x61, 0x76, 0x61, 0x69, 0x6c, 0x61, 0x62, 0x6c, 0x65, 0x5f, 0x71, 0x75, 0x61, 0x6e, 0x74,
	0x69, 0x74, 0x79, 0x18, 0x02, 0x20, 0x01, 0x28, 0x05, 0x52, 0x11, 0x61, 0x76, 0x61, 0x69, 0x6c,
	0x61, 0x62, 0x6c, 0x65, 0x51, 0x75, 0x61, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x12, 0x18, 0x0a, 0x07,
	0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x6d,
	0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x22, 0x76, 0x0a, 0x0b, 0x53, 0x74, 0x6f, 0x63, 0x6b, 0x53,
	0x74, 0x61, 0x74, 0x75, 0x73, 0x12, 0x1d, 0x0a, 0x0a, 0x70, 0x72, 0x6f, 0x64, 0x75, 0x63, 0x74,
	0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x70, 0x72, 0x6f, 0x64, 0x75,
	0x63, 0x74, 0x49, 0x64, 0x12, 0x19, 0x0a, 0x08, 0x69, 0x6e, 0x5f, 0x73, 0x74, 0x6f, 0x63, 0x6b,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x08, 0x52, 0x07, 0x69, 0x6e, 0x53, 0x74, 0x6f, 0x63, 0x6b, 0x12,
	0x2d, 0x0a, 0x12, 0x61, 0x76, 0x61, 0x69, 0x6c, 0x61, 0x62, 0x6c, 0x65, 0x5f, 0x71, 0x75, 0x61,
	0x6e, 0x74, 0x69, 0x74, 0x79, 0x18, 0x03, 0x20, 0x01, 0x28, 0x05, 0x52, 0x11, 0x61, 0x76, 0x61,
	0x69, 0x6c, 0x61, 0x62, 0x6c, 0x65, 0x51, 0x75, 0x61, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x22, 0x3e,
	0x0a, 0x12, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x53, 0x74, 0x6f, 0x63, 0x6b, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x12, 0x28, 0x0a, 0x05, 0x69, 0x74, 0x65, 0x6d, 0x73, 0x18, 0x01, 0x20,
	0x03, 0x28, 0x0b, 0x32, 0x12, 0x2e, 0x70, 0x72, 0x6f, 0x64, 0x75, 0x63, 0x74, 0x2e, 0x53, 0x74,
	0x6f, 0x63, 0x6b, 0x49, 0x74, 0x65, 0x6d, 0x52, 0x05, 0x69, 0x74, 0x65, 0x6d, 0x73, 0x22, 0x46,
	0x0a, 0x09, 0x53, 0x74, 0x6f, 0x63, 0x6b, 0x49, 0x74, 0x65, 0x6d, 0x12, 0x1d, 0x0a, 0x0a, 0x70,
	0x72, 0x6f, 0x64, 0x75, 0x63, 0x74, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x09, 0x70, 0x72, 0x6f, 0x64, 0x75, 0x63, 0x74, 0x49, 0x64, 0x12, 0x1a, 0x0a, 0x08, 0x71, 0x75,
	0x61, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x18, 0x02, 0x20, 0x01, 0x28, 0x05, 0x52, 0x08, 0x71, 0x75,
	0x61, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x22, 0x8a, 0x01, 0x0a, 0x13, 0x55, 0x70, 0x64, 0x61, 0x74,
	0x65, 0x53, 0x74, 0x6f, 0x63, 0x6b, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x3f,
	0x0a, 0x0d, 0x75, 0x70, 0x64, 0x61, 0x74, 0x65, 0x5f, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x18,
	0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x70, 0x72, 0x6f, 0x64, 0x75, 0x63, 0x74, 0x2e,
	0x53, 0x74, 0x6f, 0x63, 0x6b, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x53, 0x74, 0x61, 0x74, 0x75,
	0x73, 0x52, 0x0c, 0x75, 0x70, 0x64, 0x61, 0x74, 0x65, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x12,
	0x18, 0x0a, 0x07, 0x73, 0x75, 0x63, 0x63, 0x65, 0x73, 0x73, 0x18, 0x02, 0x20, 0x01, 0x28, 0x08,
	0x52, 0x07, 0x73, 0x75, 0x63, 0x63, 0x65, 0x73, 0x73, 0x12, 0x18, 0x0a, 0x07, 0x6d, 0x65, 0x73,
	0x73, 0x61, 0x67, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x6d, 0x65, 0x73, 0x73,
	0x61, 0x67, 0x65, 0x22, 0x66, 0x0a, 0x11, 0x53, 0x74, 0x6f, 0x63, 0x6b, 0x55, 0x70, 0x64, 0x61,
	0x74, 0x65, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x12, 0x1d, 0x0a, 0x0a, 0x70, 0x72, 0x6f, 0x64,
	0x75, 0x63, 0x74, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x70, 0x72,
	0x6f, 0x64, 0x75, 0x63, 0x74, 0x49, 0x64, 0x12, 0x18, 0x0a, 0x07, 0x75, 0x70, 0x64, 0x61, 0x74,
	0x65, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x08, 0x52, 0x07, 0x75, 0x70, 0x64, 0x61, 0x74, 0x65,
	0x64, 0x12, 0x18, 0x0a, 0x07, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x18, 0x03, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x07, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x32, 0xa6, 0x02, 0x0a, 0x0e,
	0x50, 0x72, 0x6f, 0x64, 0x75, 0x63, 0x74, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x46,
	0x0a, 0x0c, 0x47, 0x65, 0x74, 0x42, 0x61, 0x73, 0x69, 0x63, 0x49, 0x6e, 0x66, 0x6f, 0x12, 0x17,
	0x2e, 0x70, 0x72, 0x6f, 0x64, 0x75, 0x63, 0x74, 0x2e, 0x50, 0x72, 0x6f, 0x64, 0x75, 0x63, 0x74,
	0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1d, 0x2e, 0x70, 0x72, 0x6f, 0x64, 0x75, 0x63,
	0x74, 0x2e, 0x42, 0x61, 0x73, 0x69, 0x63, 0x50, 0x72, 0x6f, 0x64, 0x75, 0x63, 0x74, 0x52, 0x65,
	0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x43, 0x0a, 0x0e, 0x47, 0x65, 0x74, 0x50, 0x72, 0x6f,
	0x64, 0x75, 0x63, 0x74, 0x49, 0x6e, 0x66, 0x6f, 0x12, 0x17, 0x2e, 0x70, 0x72, 0x6f, 0x64, 0x75,
	0x63, 0x74, 0x2e, 0x50, 0x72, 0x6f, 0x64, 0x75, 0x63, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x1a, 0x18, 0x2e, 0x70, 0x72, 0x6f, 0x64, 0x75, 0x63, 0x74, 0x2e, 0x50, 0x72, 0x6f, 0x64,
	0x75, 0x63, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x3d, 0x0a, 0x0a, 0x43,
	0x68, 0x65, 0x63, 0x6b, 0x53, 0x74, 0x6f, 0x63, 0x6b, 0x12, 0x17, 0x2e, 0x70, 0x72, 0x6f, 0x64,
	0x75, 0x63, 0x74, 0x2e, 0x50, 0x72, 0x6f, 0x64, 0x75, 0x63, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x1a, 0x16, 0x2e, 0x70, 0x72, 0x6f, 0x64, 0x75, 0x63, 0x74, 0x2e, 0x53, 0x74, 0x6f,
	0x63, 0x6b, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x48, 0x0a, 0x0b, 0x55, 0x70,
	0x64, 0x61, 0x74, 0x65, 0x53, 0x74, 0x6f, 0x63, 0x6b, 0x12, 0x1b, 0x2e, 0x70, 0x72, 0x6f, 0x64,
	0x75, 0x63, 0x74, 0x2e, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x53, 0x74, 0x6f, 0x63, 0x6b, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1c, 0x2e, 0x70, 0x72, 0x6f, 0x64, 0x75, 0x63, 0x74,
	0x2e, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x53, 0x74, 0x6f, 0x63, 0x6b, 0x52, 0x65, 0x73, 0x70,
	0x6f, 0x6e, 0x73, 0x65, 0x42, 0x1d, 0x5a, 0x1b, 0x6d, 0x6f, 0x64, 0x75, 0x6c, 0x65, 0x2f, 0x67,
	0x52, 0x50, 0x43, 0x2d, 0x50, 0x72, 0x6f, 0x64, 0x75, 0x63, 0x74, 0x2f, 0x73, 0x65, 0x72, 0x76,
	0x69, 0x63, 0x65, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_module_gRPC_Product_product_service_proto_rawDescOnce sync.Once
	file_module_gRPC_Product_product_service_proto_rawDescData = file_module_gRPC_Product_product_service_proto_rawDesc
)

func file_module_gRPC_Product_product_service_proto_rawDescGZIP() []byte {
	file_module_gRPC_Product_product_service_proto_rawDescOnce.Do(func() {
		file_module_gRPC_Product_product_service_proto_rawDescData = protoimpl.X.CompressGZIP(file_module_gRPC_Product_product_service_proto_rawDescData)
	})
	return file_module_gRPC_Product_product_service_proto_rawDescData
}

var file_module_gRPC_Product_product_service_proto_msgTypes = make([]protoimpl.MessageInfo, 9)
var file_module_gRPC_Product_product_service_proto_goTypes = []any{
	(*ProductRequest)(nil),       // 0: product.ProductRequest
	(*BasicProductResponse)(nil), // 1: product.BasicProductResponse
	(*ProductResponse)(nil),      // 2: product.ProductResponse
	(*StockResponse)(nil),        // 3: product.StockResponse
	(*StockStatus)(nil),          // 4: product.StockStatus
	(*UpdateStockRequest)(nil),   // 5: product.UpdateStockRequest
	(*StockItem)(nil),            // 6: product.StockItem
	(*UpdateStockResponse)(nil),  // 7: product.UpdateStockResponse
	(*StockUpdateStatus)(nil),    // 8: product.StockUpdateStatus
}
var file_module_gRPC_Product_product_service_proto_depIdxs = []int32{
	6, // 0: product.UpdateStockRequest.items:type_name -> product.StockItem
	8, // 1: product.UpdateStockResponse.update_status:type_name -> product.StockUpdateStatus
	0, // 2: product.ProductService.GetBasicInfo:input_type -> product.ProductRequest
	0, // 3: product.ProductService.GetProductInfo:input_type -> product.ProductRequest
	0, // 4: product.ProductService.CheckStock:input_type -> product.ProductRequest
	5, // 5: product.ProductService.UpdateStock:input_type -> product.UpdateStockRequest
	1, // 6: product.ProductService.GetBasicInfo:output_type -> product.BasicProductResponse
	2, // 7: product.ProductService.GetProductInfo:output_type -> product.ProductResponse
	3, // 8: product.ProductService.CheckStock:output_type -> product.StockResponse
	7, // 9: product.ProductService.UpdateStock:output_type -> product.UpdateStockResponse
	6, // [6:10] is the sub-list for method output_type
	2, // [2:6] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_module_gRPC_Product_product_service_proto_init() }
func file_module_gRPC_Product_product_service_proto_init() {
	if File_module_gRPC_Product_product_service_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_module_gRPC_Product_product_service_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   9,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_module_gRPC_Product_product_service_proto_goTypes,
		DependencyIndexes: file_module_gRPC_Product_product_service_proto_depIdxs,
		MessageInfos:      file_module_gRPC_Product_product_service_proto_msgTypes,
	}.Build()
	File_module_gRPC_Product_product_service_proto = out.File
	file_module_gRPC_Product_product_service_proto_rawDesc = nil
	file_module_gRPC_Product_product_service_proto_goTypes = nil
	file_module_gRPC_Product_product_service_proto_depIdxs = nil
}
