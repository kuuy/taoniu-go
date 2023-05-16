// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        v3.12.4
// source: binance/spot/margin/isolated/tradings/fishers/grids/grids.proto

package grids

import (
	timestamp "github.com/golang/protobuf/ptypes/timestamp"
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

type GridInfo struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id           string               `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Symbol       string               `protobuf:"bytes,2,opt,name=symbol,proto3" json:"symbol,omitempty"`
	BuyPrice     float32              `protobuf:"fixed32,3,opt,name=buyPrice,proto3" json:"buyPrice,omitempty"`
	BuyQuantity  float32              `protobuf:"fixed32,4,opt,name=buyQuantity,proto3" json:"buyQuantity,omitempty"`
	SellPrice    float32              `protobuf:"fixed32,5,opt,name=sellPrice,proto3" json:"sellPrice,omitempty"`
	SellQuantity float32              `protobuf:"fixed32,6,opt,name=sellQuantity,proto3" json:"sellQuantity,omitempty"`
	Status       int32                `protobuf:"varint,7,opt,name=status,proto3" json:"status,omitempty"`
	CreatedAt    *timestamp.Timestamp `protobuf:"bytes,8,opt,name=createdAt,proto3" json:"createdAt,omitempty"`
	UpdatedAt    *timestamp.Timestamp `protobuf:"bytes,9,opt,name=updatedAt,proto3" json:"updatedAt,omitempty"`
}

func (x *GridInfo) Reset() {
	*x = GridInfo{}
	if protoimpl.UnsafeEnabled {
		mi := &file_binance_spot_margin_isolated_tradings_fishers_grids_grids_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GridInfo) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GridInfo) ProtoMessage() {}

func (x *GridInfo) ProtoReflect() protoreflect.Message {
	mi := &file_binance_spot_margin_isolated_tradings_fishers_grids_grids_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GridInfo.ProtoReflect.Descriptor instead.
func (*GridInfo) Descriptor() ([]byte, []int) {
	return file_binance_spot_margin_isolated_tradings_fishers_grids_grids_proto_rawDescGZIP(), []int{0}
}

func (x *GridInfo) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *GridInfo) GetSymbol() string {
	if x != nil {
		return x.Symbol
	}
	return ""
}

func (x *GridInfo) GetBuyPrice() float32 {
	if x != nil {
		return x.BuyPrice
	}
	return 0
}

func (x *GridInfo) GetBuyQuantity() float32 {
	if x != nil {
		return x.BuyQuantity
	}
	return 0
}

func (x *GridInfo) GetSellPrice() float32 {
	if x != nil {
		return x.SellPrice
	}
	return 0
}

func (x *GridInfo) GetSellQuantity() float32 {
	if x != nil {
		return x.SellQuantity
	}
	return 0
}

func (x *GridInfo) GetStatus() int32 {
	if x != nil {
		return x.Status
	}
	return 0
}

func (x *GridInfo) GetCreatedAt() *timestamp.Timestamp {
	if x != nil {
		return x.CreatedAt
	}
	return nil
}

func (x *GridInfo) GetUpdatedAt() *timestamp.Timestamp {
	if x != nil {
		return x.UpdatedAt
	}
	return nil
}

type PagenateRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Symbol   string   `protobuf:"bytes,1,opt,name=symbol,proto3" json:"symbol,omitempty"`
	Status   []uint32 `protobuf:"varint,2,rep,packed,name=status,proto3" json:"status,omitempty"`
	Page     int32    `protobuf:"varint,3,opt,name=page,proto3" json:"page,omitempty"`
	PageSize int32    `protobuf:"varint,4,opt,name=pageSize,proto3" json:"pageSize,omitempty"`
}

func (x *PagenateRequest) Reset() {
	*x = PagenateRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_binance_spot_margin_isolated_tradings_fishers_grids_grids_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PagenateRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PagenateRequest) ProtoMessage() {}

func (x *PagenateRequest) ProtoReflect() protoreflect.Message {
	mi := &file_binance_spot_margin_isolated_tradings_fishers_grids_grids_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PagenateRequest.ProtoReflect.Descriptor instead.
func (*PagenateRequest) Descriptor() ([]byte, []int) {
	return file_binance_spot_margin_isolated_tradings_fishers_grids_grids_proto_rawDescGZIP(), []int{1}
}

func (x *PagenateRequest) GetSymbol() string {
	if x != nil {
		return x.Symbol
	}
	return ""
}

func (x *PagenateRequest) GetStatus() []uint32 {
	if x != nil {
		return x.Status
	}
	return nil
}

func (x *PagenateRequest) GetPage() int32 {
	if x != nil {
		return x.Page
	}
	return 0
}

func (x *PagenateRequest) GetPageSize() int32 {
	if x != nil {
		return x.PageSize
	}
	return 0
}

type PagenateReply struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Total int64       `protobuf:"varint,1,opt,name=total,proto3" json:"total,omitempty"`
	Data  []*GridInfo `protobuf:"bytes,2,rep,name=data,proto3" json:"data,omitempty"`
}

func (x *PagenateReply) Reset() {
	*x = PagenateReply{}
	if protoimpl.UnsafeEnabled {
		mi := &file_binance_spot_margin_isolated_tradings_fishers_grids_grids_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PagenateReply) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PagenateReply) ProtoMessage() {}

func (x *PagenateReply) ProtoReflect() protoreflect.Message {
	mi := &file_binance_spot_margin_isolated_tradings_fishers_grids_grids_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PagenateReply.ProtoReflect.Descriptor instead.
func (*PagenateReply) Descriptor() ([]byte, []int) {
	return file_binance_spot_margin_isolated_tradings_fishers_grids_grids_proto_rawDescGZIP(), []int{2}
}

func (x *PagenateReply) GetTotal() int64 {
	if x != nil {
		return x.Total
	}
	return 0
}

func (x *PagenateReply) GetData() []*GridInfo {
	if x != nil {
		return x.Data
	}
	return nil
}

var File_binance_spot_margin_isolated_tradings_fishers_grids_grids_proto protoreflect.FileDescriptor

var file_binance_spot_margin_isolated_tradings_fishers_grids_grids_proto_rawDesc = []byte{
	0x0a, 0x3f, 0x62, 0x69, 0x6e, 0x61, 0x6e, 0x63, 0x65, 0x2f, 0x73, 0x70, 0x6f, 0x74, 0x2f, 0x6d,
	0x61, 0x72, 0x67, 0x69, 0x6e, 0x2f, 0x69, 0x73, 0x6f, 0x6c, 0x61, 0x74, 0x65, 0x64, 0x2f, 0x74,
	0x72, 0x61, 0x64, 0x69, 0x6e, 0x67, 0x73, 0x2f, 0x66, 0x69, 0x73, 0x68, 0x65, 0x72, 0x73, 0x2f,
	0x67, 0x72, 0x69, 0x64, 0x73, 0x2f, 0x67, 0x72, 0x69, 0x64, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x12, 0x4d, 0x74, 0x61, 0x6f, 0x6e, 0x69, 0x75, 0x2e, 0x6c, 0x6f, 0x63, 0x61, 0x6c, 0x2e,
	0x63, 0x72, 0x79, 0x70, 0x74, 0x6f, 0x73, 0x2e, 0x67, 0x72, 0x70, 0x63, 0x2e, 0x62, 0x69, 0x6e,
	0x61, 0x6e, 0x63, 0x65, 0x2e, 0x73, 0x70, 0x6f, 0x74, 0x2e, 0x6d, 0x61, 0x72, 0x67, 0x69, 0x6e,
	0x2e, 0x69, 0x73, 0x6f, 0x6c, 0x61, 0x74, 0x65, 0x64, 0x2e, 0x74, 0x72, 0x61, 0x64, 0x69, 0x6e,
	0x67, 0x73, 0x2e, 0x66, 0x69, 0x73, 0x68, 0x65, 0x72, 0x73, 0x2e, 0x67, 0x72, 0x69, 0x64, 0x73,
	0x1a, 0x1f, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75,
	0x66, 0x2f, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x22, 0xbe, 0x02, 0x0a, 0x08, 0x47, 0x72, 0x69, 0x64, 0x49, 0x6e, 0x66, 0x6f, 0x12, 0x0e,
	0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x64, 0x12, 0x16,
	0x0a, 0x06, 0x73, 0x79, 0x6d, 0x62, 0x6f, 0x6c, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06,
	0x73, 0x79, 0x6d, 0x62, 0x6f, 0x6c, 0x12, 0x1a, 0x0a, 0x08, 0x62, 0x75, 0x79, 0x50, 0x72, 0x69,
	0x63, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x02, 0x52, 0x08, 0x62, 0x75, 0x79, 0x50, 0x72, 0x69,
	0x63, 0x65, 0x12, 0x20, 0x0a, 0x0b, 0x62, 0x75, 0x79, 0x51, 0x75, 0x61, 0x6e, 0x74, 0x69, 0x74,
	0x79, 0x18, 0x04, 0x20, 0x01, 0x28, 0x02, 0x52, 0x0b, 0x62, 0x75, 0x79, 0x51, 0x75, 0x61, 0x6e,
	0x74, 0x69, 0x74, 0x79, 0x12, 0x1c, 0x0a, 0x09, 0x73, 0x65, 0x6c, 0x6c, 0x50, 0x72, 0x69, 0x63,
	0x65, 0x18, 0x05, 0x20, 0x01, 0x28, 0x02, 0x52, 0x09, 0x73, 0x65, 0x6c, 0x6c, 0x50, 0x72, 0x69,
	0x63, 0x65, 0x12, 0x22, 0x0a, 0x0c, 0x73, 0x65, 0x6c, 0x6c, 0x51, 0x75, 0x61, 0x6e, 0x74, 0x69,
	0x74, 0x79, 0x18, 0x06, 0x20, 0x01, 0x28, 0x02, 0x52, 0x0c, 0x73, 0x65, 0x6c, 0x6c, 0x51, 0x75,
	0x61, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x12, 0x16, 0x0a, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73,
	0x18, 0x07, 0x20, 0x01, 0x28, 0x05, 0x52, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x12, 0x38,
	0x0a, 0x09, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65, 0x64, 0x41, 0x74, 0x18, 0x08, 0x20, 0x01, 0x28,
	0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x62, 0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x52, 0x09, 0x63,
	0x72, 0x65, 0x61, 0x74, 0x65, 0x64, 0x41, 0x74, 0x12, 0x38, 0x0a, 0x09, 0x75, 0x70, 0x64, 0x61,
	0x74, 0x65, 0x64, 0x41, 0x74, 0x18, 0x09, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f,
	0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54, 0x69,
	0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x52, 0x09, 0x75, 0x70, 0x64, 0x61, 0x74, 0x65, 0x64,
	0x41, 0x74, 0x22, 0x71, 0x0a, 0x0f, 0x50, 0x61, 0x67, 0x65, 0x6e, 0x61, 0x74, 0x65, 0x52, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x16, 0x0a, 0x06, 0x73, 0x79, 0x6d, 0x62, 0x6f, 0x6c, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x73, 0x79, 0x6d, 0x62, 0x6f, 0x6c, 0x12, 0x16, 0x0a,
	0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0d, 0x52, 0x06, 0x73,
	0x74, 0x61, 0x74, 0x75, 0x73, 0x12, 0x12, 0x0a, 0x04, 0x70, 0x61, 0x67, 0x65, 0x18, 0x03, 0x20,
	0x01, 0x28, 0x05, 0x52, 0x04, 0x70, 0x61, 0x67, 0x65, 0x12, 0x1a, 0x0a, 0x08, 0x70, 0x61, 0x67,
	0x65, 0x53, 0x69, 0x7a, 0x65, 0x18, 0x04, 0x20, 0x01, 0x28, 0x05, 0x52, 0x08, 0x70, 0x61, 0x67,
	0x65, 0x53, 0x69, 0x7a, 0x65, 0x22, 0x92, 0x01, 0x0a, 0x0d, 0x50, 0x61, 0x67, 0x65, 0x6e, 0x61,
	0x74, 0x65, 0x52, 0x65, 0x70, 0x6c, 0x79, 0x12, 0x14, 0x0a, 0x05, 0x74, 0x6f, 0x74, 0x61, 0x6c,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x03, 0x52, 0x05, 0x74, 0x6f, 0x74, 0x61, 0x6c, 0x12, 0x6b, 0x0a,
	0x04, 0x64, 0x61, 0x74, 0x61, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x57, 0x2e, 0x74, 0x61,
	0x6f, 0x6e, 0x69, 0x75, 0x2e, 0x6c, 0x6f, 0x63, 0x61, 0x6c, 0x2e, 0x63, 0x72, 0x79, 0x70, 0x74,
	0x6f, 0x73, 0x2e, 0x67, 0x72, 0x70, 0x63, 0x2e, 0x62, 0x69, 0x6e, 0x61, 0x6e, 0x63, 0x65, 0x2e,
	0x73, 0x70, 0x6f, 0x74, 0x2e, 0x6d, 0x61, 0x72, 0x67, 0x69, 0x6e, 0x2e, 0x69, 0x73, 0x6f, 0x6c,
	0x61, 0x74, 0x65, 0x64, 0x2e, 0x74, 0x72, 0x61, 0x64, 0x69, 0x6e, 0x67, 0x73, 0x2e, 0x66, 0x69,
	0x73, 0x68, 0x65, 0x72, 0x73, 0x2e, 0x67, 0x72, 0x69, 0x64, 0x73, 0x2e, 0x47, 0x72, 0x69, 0x64,
	0x49, 0x6e, 0x66, 0x6f, 0x52, 0x04, 0x64, 0x61, 0x74, 0x61, 0x32, 0xd4, 0x01, 0x0a, 0x05, 0x47,
	0x72, 0x69, 0x64, 0x73, 0x12, 0xca, 0x01, 0x0a, 0x08, 0x50, 0x61, 0x67, 0x65, 0x6e, 0x61, 0x74,
	0x65, 0x12, 0x5e, 0x2e, 0x74, 0x61, 0x6f, 0x6e, 0x69, 0x75, 0x2e, 0x6c, 0x6f, 0x63, 0x61, 0x6c,
	0x2e, 0x63, 0x72, 0x79, 0x70, 0x74, 0x6f, 0x73, 0x2e, 0x67, 0x72, 0x70, 0x63, 0x2e, 0x62, 0x69,
	0x6e, 0x61, 0x6e, 0x63, 0x65, 0x2e, 0x73, 0x70, 0x6f, 0x74, 0x2e, 0x6d, 0x61, 0x72, 0x67, 0x69,
	0x6e, 0x2e, 0x69, 0x73, 0x6f, 0x6c, 0x61, 0x74, 0x65, 0x64, 0x2e, 0x74, 0x72, 0x61, 0x64, 0x69,
	0x6e, 0x67, 0x73, 0x2e, 0x66, 0x69, 0x73, 0x68, 0x65, 0x72, 0x73, 0x2e, 0x67, 0x72, 0x69, 0x64,
	0x73, 0x2e, 0x50, 0x61, 0x67, 0x65, 0x6e, 0x61, 0x74, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x1a, 0x5c, 0x2e, 0x74, 0x61, 0x6f, 0x6e, 0x69, 0x75, 0x2e, 0x6c, 0x6f, 0x63, 0x61, 0x6c,
	0x2e, 0x63, 0x72, 0x79, 0x70, 0x74, 0x6f, 0x73, 0x2e, 0x67, 0x72, 0x70, 0x63, 0x2e, 0x62, 0x69,
	0x6e, 0x61, 0x6e, 0x63, 0x65, 0x2e, 0x73, 0x70, 0x6f, 0x74, 0x2e, 0x6d, 0x61, 0x72, 0x67, 0x69,
	0x6e, 0x2e, 0x69, 0x73, 0x6f, 0x6c, 0x61, 0x74, 0x65, 0x64, 0x2e, 0x74, 0x72, 0x61, 0x64, 0x69,
	0x6e, 0x67, 0x73, 0x2e, 0x66, 0x69, 0x73, 0x68, 0x65, 0x72, 0x73, 0x2e, 0x67, 0x72, 0x69, 0x64,
	0x73, 0x2e, 0x50, 0x61, 0x67, 0x65, 0x6e, 0x61, 0x74, 0x65, 0x52, 0x65, 0x70, 0x6c, 0x79, 0x22,
	0x00, 0x42, 0x4f, 0x5a, 0x4d, 0x74, 0x61, 0x6f, 0x6e, 0x69, 0x75, 0x2e, 0x6c, 0x6f, 0x63, 0x61,
	0x6c, 0x2f, 0x63, 0x72, 0x79, 0x70, 0x74, 0x6f, 0x73, 0x2f, 0x67, 0x72, 0x70, 0x63, 0x2f, 0x62,
	0x69, 0x6e, 0x61, 0x6e, 0x63, 0x65, 0x2f, 0x73, 0x70, 0x6f, 0x74, 0x2f, 0x6d, 0x61, 0x72, 0x67,
	0x69, 0x6e, 0x2f, 0x69, 0x73, 0x6f, 0x6c, 0x61, 0x74, 0x65, 0x64, 0x2f, 0x74, 0x72, 0x61, 0x64,
	0x69, 0x6e, 0x67, 0x73, 0x2f, 0x66, 0x69, 0x73, 0x68, 0x65, 0x72, 0x73, 0x2f, 0x67, 0x72, 0x69,
	0x64, 0x73, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_binance_spot_margin_isolated_tradings_fishers_grids_grids_proto_rawDescOnce sync.Once
	file_binance_spot_margin_isolated_tradings_fishers_grids_grids_proto_rawDescData = file_binance_spot_margin_isolated_tradings_fishers_grids_grids_proto_rawDesc
)

func file_binance_spot_margin_isolated_tradings_fishers_grids_grids_proto_rawDescGZIP() []byte {
	file_binance_spot_margin_isolated_tradings_fishers_grids_grids_proto_rawDescOnce.Do(func() {
		file_binance_spot_margin_isolated_tradings_fishers_grids_grids_proto_rawDescData = protoimpl.X.CompressGZIP(file_binance_spot_margin_isolated_tradings_fishers_grids_grids_proto_rawDescData)
	})
	return file_binance_spot_margin_isolated_tradings_fishers_grids_grids_proto_rawDescData
}

var file_binance_spot_margin_isolated_tradings_fishers_grids_grids_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_binance_spot_margin_isolated_tradings_fishers_grids_grids_proto_goTypes = []interface{}{
	(*GridInfo)(nil),            // 0: taoniu.local.cryptos.grpc.binance.spot.margin.isolated.tradings.fishers.grids.GridInfo
	(*PagenateRequest)(nil),     // 1: taoniu.local.cryptos.grpc.binance.spot.margin.isolated.tradings.fishers.grids.PagenateRequest
	(*PagenateReply)(nil),       // 2: taoniu.local.cryptos.grpc.binance.spot.margin.isolated.tradings.fishers.grids.PagenateReply
	(*timestamp.Timestamp)(nil), // 3: google.protobuf.Timestamp
}
var file_binance_spot_margin_isolated_tradings_fishers_grids_grids_proto_depIdxs = []int32{
	3, // 0: taoniu.local.cryptos.grpc.binance.spot.margin.isolated.tradings.fishers.grids.GridInfo.createdAt:type_name -> google.protobuf.Timestamp
	3, // 1: taoniu.local.cryptos.grpc.binance.spot.margin.isolated.tradings.fishers.grids.GridInfo.updatedAt:type_name -> google.protobuf.Timestamp
	0, // 2: taoniu.local.cryptos.grpc.binance.spot.margin.isolated.tradings.fishers.grids.PagenateReply.data:type_name -> taoniu.local.cryptos.grpc.binance.spot.margin.isolated.tradings.fishers.grids.GridInfo
	1, // 3: taoniu.local.cryptos.grpc.binance.spot.margin.isolated.tradings.fishers.grids.Grids.Pagenate:input_type -> taoniu.local.cryptos.grpc.binance.spot.margin.isolated.tradings.fishers.grids.PagenateRequest
	2, // 4: taoniu.local.cryptos.grpc.binance.spot.margin.isolated.tradings.fishers.grids.Grids.Pagenate:output_type -> taoniu.local.cryptos.grpc.binance.spot.margin.isolated.tradings.fishers.grids.PagenateReply
	4, // [4:5] is the sub-list for method output_type
	3, // [3:4] is the sub-list for method input_type
	3, // [3:3] is the sub-list for extension type_name
	3, // [3:3] is the sub-list for extension extendee
	0, // [0:3] is the sub-list for field type_name
}

func init() { file_binance_spot_margin_isolated_tradings_fishers_grids_grids_proto_init() }
func file_binance_spot_margin_isolated_tradings_fishers_grids_grids_proto_init() {
	if File_binance_spot_margin_isolated_tradings_fishers_grids_grids_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_binance_spot_margin_isolated_tradings_fishers_grids_grids_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GridInfo); i {
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
		file_binance_spot_margin_isolated_tradings_fishers_grids_grids_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PagenateRequest); i {
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
		file_binance_spot_margin_isolated_tradings_fishers_grids_grids_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PagenateReply); i {
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
			RawDescriptor: file_binance_spot_margin_isolated_tradings_fishers_grids_grids_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_binance_spot_margin_isolated_tradings_fishers_grids_grids_proto_goTypes,
		DependencyIndexes: file_binance_spot_margin_isolated_tradings_fishers_grids_grids_proto_depIdxs,
		MessageInfos:      file_binance_spot_margin_isolated_tradings_fishers_grids_grids_proto_msgTypes,
	}.Build()
	File_binance_spot_margin_isolated_tradings_fishers_grids_grids_proto = out.File
	file_binance_spot_margin_isolated_tradings_fishers_grids_grids_proto_rawDesc = nil
	file_binance_spot_margin_isolated_tradings_fishers_grids_grids_proto_goTypes = nil
	file_binance_spot_margin_isolated_tradings_fishers_grids_grids_proto_depIdxs = nil
}
