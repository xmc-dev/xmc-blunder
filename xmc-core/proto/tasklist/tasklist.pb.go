// Code generated by protoc-gen-go. DO NOT EDIT.
// source: github.com/xmc-dev/xmc/xmc-core/proto/tasklist/tasklist.proto

/*
Package tasklist is a generated protocol buffer package.

It is generated from these files:
	github.com/xmc-dev/xmc/xmc-core/proto/tasklist/tasklist.proto

It has these top-level messages:
	TaskList
	CreateRequest
	CreateResponse
	ReadRequest
	ReadResponse
	GetRequest
	GetResponse
	UpdateRequest
	UpdateResponse
	DeleteRequest
	DeleteResponse
	SearchRequest
	SearchResponse
*/
package tasklist

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"
import google_protobuf "github.com/golang/protobuf/ptypes/wrappers"
import xmc_srv_core_tsrange "github.com/xmc-dev/xmc/xmc-core/proto/tsrange"
import xmc_srv_core_searchmeta "github.com/xmc-dev/xmc/xmc-core/proto/searchmeta"

import (
	client "github.com/micro/go-micro/client"
	server "github.com/micro/go-micro/server"
	context "golang.org/x/net/context"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type TaskList struct {
	Id          string                               `protobuf:"bytes,1,opt,name=id" json:"id,omitempty"`
	Name        string                               `protobuf:"bytes,2,opt,name=name" json:"name,omitempty"`
	Description string                               `protobuf:"bytes,3,opt,name=description" json:"description,omitempty"`
	TimeRange   *xmc_srv_core_tsrange.TimestampRange `protobuf:"bytes,4,opt,name=time_range,json=timeRange" json:"time_range,omitempty"`
	PageId      string                               `protobuf:"bytes,5,opt,name=page_id,json=pageId" json:"page_id,omitempty"`
	Title       string                               `protobuf:"bytes,6,opt,name=title" json:"title,omitempty"`
}

func (m *TaskList) Reset()                    { *m = TaskList{} }
func (m *TaskList) String() string            { return proto.CompactTextString(m) }
func (*TaskList) ProtoMessage()               {}
func (*TaskList) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *TaskList) GetId() string {
	if m != nil {
		return m.Id
	}
	return ""
}

func (m *TaskList) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *TaskList) GetDescription() string {
	if m != nil {
		return m.Description
	}
	return ""
}

func (m *TaskList) GetTimeRange() *xmc_srv_core_tsrange.TimestampRange {
	if m != nil {
		return m.TimeRange
	}
	return nil
}

func (m *TaskList) GetPageId() string {
	if m != nil {
		return m.PageId
	}
	return ""
}

func (m *TaskList) GetTitle() string {
	if m != nil {
		return m.Title
	}
	return ""
}

type CreateRequest struct {
	TaskList *TaskList `protobuf:"bytes,1,opt,name=task_list,json=taskList" json:"task_list,omitempty"`
}

func (m *CreateRequest) Reset()                    { *m = CreateRequest{} }
func (m *CreateRequest) String() string            { return proto.CompactTextString(m) }
func (*CreateRequest) ProtoMessage()               {}
func (*CreateRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *CreateRequest) GetTaskList() *TaskList {
	if m != nil {
		return m.TaskList
	}
	return nil
}

type CreateResponse struct {
	Id string `protobuf:"bytes,1,opt,name=id" json:"id,omitempty"`
}

func (m *CreateResponse) Reset()                    { *m = CreateResponse{} }
func (m *CreateResponse) String() string            { return proto.CompactTextString(m) }
func (*CreateResponse) ProtoMessage()               {}
func (*CreateResponse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

func (m *CreateResponse) GetId() string {
	if m != nil {
		return m.Id
	}
	return ""
}

type ReadRequest struct {
	Id string `protobuf:"bytes,1,opt,name=id" json:"id,omitempty"`
}

func (m *ReadRequest) Reset()                    { *m = ReadRequest{} }
func (m *ReadRequest) String() string            { return proto.CompactTextString(m) }
func (*ReadRequest) ProtoMessage()               {}
func (*ReadRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{3} }

func (m *ReadRequest) GetId() string {
	if m != nil {
		return m.Id
	}
	return ""
}

type ReadResponse struct {
	TaskList *TaskList `protobuf:"bytes,1,opt,name=task_list,json=taskList" json:"task_list,omitempty"`
}

func (m *ReadResponse) Reset()                    { *m = ReadResponse{} }
func (m *ReadResponse) String() string            { return proto.CompactTextString(m) }
func (*ReadResponse) ProtoMessage()               {}
func (*ReadResponse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{4} }

func (m *ReadResponse) GetTaskList() *TaskList {
	if m != nil {
		return m.TaskList
	}
	return nil
}

type GetRequest struct {
	Name string `protobuf:"bytes,1,opt,name=name" json:"name,omitempty"`
}

func (m *GetRequest) Reset()                    { *m = GetRequest{} }
func (m *GetRequest) String() string            { return proto.CompactTextString(m) }
func (*GetRequest) ProtoMessage()               {}
func (*GetRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{5} }

func (m *GetRequest) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

type GetResponse struct {
	TaskList *TaskList `protobuf:"bytes,1,opt,name=task_list,json=taskList" json:"task_list,omitempty"`
}

func (m *GetResponse) Reset()                    { *m = GetResponse{} }
func (m *GetResponse) String() string            { return proto.CompactTextString(m) }
func (*GetResponse) ProtoMessage()               {}
func (*GetResponse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{6} }

func (m *GetResponse) GetTaskList() *TaskList {
	if m != nil {
		return m.TaskList
	}
	return nil
}

type UpdateRequest struct {
	Id          string                               `protobuf:"bytes,1,opt,name=id" json:"id,omitempty"`
	Name        string                               `protobuf:"bytes,2,opt,name=name" json:"name,omitempty"`
	Description string                               `protobuf:"bytes,3,opt,name=description" json:"description,omitempty"`
	TimeRange   *xmc_srv_core_tsrange.TimestampRange `protobuf:"bytes,4,opt,name=time_range,json=timeRange" json:"time_range,omitempty"`
	SetNullTime bool                                 `protobuf:"varint,5,opt,name=set_null_time,json=setNullTime" json:"set_null_time,omitempty"`
	Title       string                               `protobuf:"bytes,6,opt,name=title" json:"title,omitempty"`
}

func (m *UpdateRequest) Reset()                    { *m = UpdateRequest{} }
func (m *UpdateRequest) String() string            { return proto.CompactTextString(m) }
func (*UpdateRequest) ProtoMessage()               {}
func (*UpdateRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{7} }

func (m *UpdateRequest) GetId() string {
	if m != nil {
		return m.Id
	}
	return ""
}

func (m *UpdateRequest) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *UpdateRequest) GetDescription() string {
	if m != nil {
		return m.Description
	}
	return ""
}

func (m *UpdateRequest) GetTimeRange() *xmc_srv_core_tsrange.TimestampRange {
	if m != nil {
		return m.TimeRange
	}
	return nil
}

func (m *UpdateRequest) GetSetNullTime() bool {
	if m != nil {
		return m.SetNullTime
	}
	return false
}

func (m *UpdateRequest) GetTitle() string {
	if m != nil {
		return m.Title
	}
	return ""
}

type UpdateResponse struct {
}

func (m *UpdateResponse) Reset()                    { *m = UpdateResponse{} }
func (m *UpdateResponse) String() string            { return proto.CompactTextString(m) }
func (*UpdateResponse) ProtoMessage()               {}
func (*UpdateResponse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{8} }

type DeleteRequest struct {
	Id         string `protobuf:"bytes,1,opt,name=id" json:"id,omitempty"`
	LeaveTasks bool   `protobuf:"varint,2,opt,name=leave_tasks,json=leaveTasks" json:"leave_tasks,omitempty"`
}

func (m *DeleteRequest) Reset()                    { *m = DeleteRequest{} }
func (m *DeleteRequest) String() string            { return proto.CompactTextString(m) }
func (*DeleteRequest) ProtoMessage()               {}
func (*DeleteRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{9} }

func (m *DeleteRequest) GetId() string {
	if m != nil {
		return m.Id
	}
	return ""
}

func (m *DeleteRequest) GetLeaveTasks() bool {
	if m != nil {
		return m.LeaveTasks
	}
	return false
}

type DeleteResponse struct {
}

func (m *DeleteResponse) Reset()                    { *m = DeleteResponse{} }
func (m *DeleteResponse) String() string            { return proto.CompactTextString(m) }
func (*DeleteResponse) ProtoMessage()               {}
func (*DeleteResponse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{10} }

type SearchRequest struct {
	Limit       uint32                               `protobuf:"varint,1,opt,name=limit" json:"limit,omitempty"`
	Offset      uint32                               `protobuf:"varint,2,opt,name=offset" json:"offset,omitempty"`
	Name        string                               `protobuf:"bytes,3,opt,name=name" json:"name,omitempty"`
	Description string                               `protobuf:"bytes,4,opt,name=description" json:"description,omitempty"`
	TimeRange   *xmc_srv_core_tsrange.TimestampRange `protobuf:"bytes,5,opt,name=time_range,json=timeRange" json:"time_range,omitempty"`
	Title       string                               `protobuf:"bytes,6,opt,name=title" json:"title,omitempty"`
	IsPermanent *google_protobuf.BoolValue           `protobuf:"bytes,7,opt,name=is_permanent,json=isPermanent" json:"is_permanent,omitempty"`
}

func (m *SearchRequest) Reset()                    { *m = SearchRequest{} }
func (m *SearchRequest) String() string            { return proto.CompactTextString(m) }
func (*SearchRequest) ProtoMessage()               {}
func (*SearchRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{11} }

func (m *SearchRequest) GetLimit() uint32 {
	if m != nil {
		return m.Limit
	}
	return 0
}

func (m *SearchRequest) GetOffset() uint32 {
	if m != nil {
		return m.Offset
	}
	return 0
}

func (m *SearchRequest) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *SearchRequest) GetDescription() string {
	if m != nil {
		return m.Description
	}
	return ""
}

func (m *SearchRequest) GetTimeRange() *xmc_srv_core_tsrange.TimestampRange {
	if m != nil {
		return m.TimeRange
	}
	return nil
}

func (m *SearchRequest) GetTitle() string {
	if m != nil {
		return m.Title
	}
	return ""
}

func (m *SearchRequest) GetIsPermanent() *google_protobuf.BoolValue {
	if m != nil {
		return m.IsPermanent
	}
	return nil
}

type SearchResponse struct {
	TaskLists []*TaskList                   `protobuf:"bytes,1,rep,name=task_lists,json=taskLists" json:"task_lists,omitempty"`
	Meta      *xmc_srv_core_searchmeta.Meta `protobuf:"bytes,2,opt,name=meta" json:"meta,omitempty"`
}

func (m *SearchResponse) Reset()                    { *m = SearchResponse{} }
func (m *SearchResponse) String() string            { return proto.CompactTextString(m) }
func (*SearchResponse) ProtoMessage()               {}
func (*SearchResponse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{12} }

func (m *SearchResponse) GetTaskLists() []*TaskList {
	if m != nil {
		return m.TaskLists
	}
	return nil
}

func (m *SearchResponse) GetMeta() *xmc_srv_core_searchmeta.Meta {
	if m != nil {
		return m.Meta
	}
	return nil
}

func init() {
	proto.RegisterType((*TaskList)(nil), "xmc.srv.core.tasklist.TaskList")
	proto.RegisterType((*CreateRequest)(nil), "xmc.srv.core.tasklist.CreateRequest")
	proto.RegisterType((*CreateResponse)(nil), "xmc.srv.core.tasklist.CreateResponse")
	proto.RegisterType((*ReadRequest)(nil), "xmc.srv.core.tasklist.ReadRequest")
	proto.RegisterType((*ReadResponse)(nil), "xmc.srv.core.tasklist.ReadResponse")
	proto.RegisterType((*GetRequest)(nil), "xmc.srv.core.tasklist.GetRequest")
	proto.RegisterType((*GetResponse)(nil), "xmc.srv.core.tasklist.GetResponse")
	proto.RegisterType((*UpdateRequest)(nil), "xmc.srv.core.tasklist.UpdateRequest")
	proto.RegisterType((*UpdateResponse)(nil), "xmc.srv.core.tasklist.UpdateResponse")
	proto.RegisterType((*DeleteRequest)(nil), "xmc.srv.core.tasklist.DeleteRequest")
	proto.RegisterType((*DeleteResponse)(nil), "xmc.srv.core.tasklist.DeleteResponse")
	proto.RegisterType((*SearchRequest)(nil), "xmc.srv.core.tasklist.SearchRequest")
	proto.RegisterType((*SearchResponse)(nil), "xmc.srv.core.tasklist.SearchResponse")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ client.Option
var _ server.Option

// Client API for TaskListService service

type TaskListServiceClient interface {
	Create(ctx context.Context, in *CreateRequest, opts ...client.CallOption) (*CreateResponse, error)
	Read(ctx context.Context, in *ReadRequest, opts ...client.CallOption) (*ReadResponse, error)
	Get(ctx context.Context, in *GetRequest, opts ...client.CallOption) (*GetResponse, error)
	Update(ctx context.Context, in *UpdateRequest, opts ...client.CallOption) (*UpdateResponse, error)
	Delete(ctx context.Context, in *DeleteRequest, opts ...client.CallOption) (*DeleteResponse, error)
	Search(ctx context.Context, in *SearchRequest, opts ...client.CallOption) (*SearchResponse, error)
}

type taskListServiceClient struct {
	c           client.Client
	serviceName string
}

func NewTaskListServiceClient(serviceName string, c client.Client) TaskListServiceClient {
	if c == nil {
		c = client.NewClient()
	}
	if len(serviceName) == 0 {
		serviceName = "xmc.srv.core.tasklist"
	}
	return &taskListServiceClient{
		c:           c,
		serviceName: serviceName,
	}
}

func (c *taskListServiceClient) Create(ctx context.Context, in *CreateRequest, opts ...client.CallOption) (*CreateResponse, error) {
	req := c.c.NewRequest(c.serviceName, "TaskListService.Create", in)
	out := new(CreateResponse)
	err := c.c.Call(ctx, req, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *taskListServiceClient) Read(ctx context.Context, in *ReadRequest, opts ...client.CallOption) (*ReadResponse, error) {
	req := c.c.NewRequest(c.serviceName, "TaskListService.Read", in)
	out := new(ReadResponse)
	err := c.c.Call(ctx, req, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *taskListServiceClient) Get(ctx context.Context, in *GetRequest, opts ...client.CallOption) (*GetResponse, error) {
	req := c.c.NewRequest(c.serviceName, "TaskListService.Get", in)
	out := new(GetResponse)
	err := c.c.Call(ctx, req, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *taskListServiceClient) Update(ctx context.Context, in *UpdateRequest, opts ...client.CallOption) (*UpdateResponse, error) {
	req := c.c.NewRequest(c.serviceName, "TaskListService.Update", in)
	out := new(UpdateResponse)
	err := c.c.Call(ctx, req, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *taskListServiceClient) Delete(ctx context.Context, in *DeleteRequest, opts ...client.CallOption) (*DeleteResponse, error) {
	req := c.c.NewRequest(c.serviceName, "TaskListService.Delete", in)
	out := new(DeleteResponse)
	err := c.c.Call(ctx, req, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *taskListServiceClient) Search(ctx context.Context, in *SearchRequest, opts ...client.CallOption) (*SearchResponse, error) {
	req := c.c.NewRequest(c.serviceName, "TaskListService.Search", in)
	out := new(SearchResponse)
	err := c.c.Call(ctx, req, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for TaskListService service

type TaskListServiceHandler interface {
	Create(context.Context, *CreateRequest, *CreateResponse) error
	Read(context.Context, *ReadRequest, *ReadResponse) error
	Get(context.Context, *GetRequest, *GetResponse) error
	Update(context.Context, *UpdateRequest, *UpdateResponse) error
	Delete(context.Context, *DeleteRequest, *DeleteResponse) error
	Search(context.Context, *SearchRequest, *SearchResponse) error
}

func RegisterTaskListServiceHandler(s server.Server, hdlr TaskListServiceHandler, opts ...server.HandlerOption) {
	s.Handle(s.NewHandler(&TaskListService{hdlr}, opts...))
}

type TaskListService struct {
	TaskListServiceHandler
}

func (h *TaskListService) Create(ctx context.Context, in *CreateRequest, out *CreateResponse) error {
	return h.TaskListServiceHandler.Create(ctx, in, out)
}

func (h *TaskListService) Read(ctx context.Context, in *ReadRequest, out *ReadResponse) error {
	return h.TaskListServiceHandler.Read(ctx, in, out)
}

func (h *TaskListService) Get(ctx context.Context, in *GetRequest, out *GetResponse) error {
	return h.TaskListServiceHandler.Get(ctx, in, out)
}

func (h *TaskListService) Update(ctx context.Context, in *UpdateRequest, out *UpdateResponse) error {
	return h.TaskListServiceHandler.Update(ctx, in, out)
}

func (h *TaskListService) Delete(ctx context.Context, in *DeleteRequest, out *DeleteResponse) error {
	return h.TaskListServiceHandler.Delete(ctx, in, out)
}

func (h *TaskListService) Search(ctx context.Context, in *SearchRequest, out *SearchResponse) error {
	return h.TaskListServiceHandler.Search(ctx, in, out)
}

func init() {
	proto.RegisterFile("github.com/xmc-dev/xmc/xmc-core/proto/tasklist/tasklist.proto", fileDescriptor0)
}

var fileDescriptor0 = []byte{
	// 689 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xcc, 0x54, 0x4d, 0x6f, 0xd3, 0x4c,
	0x10, 0xae, 0xf3, 0xd5, 0x64, 0xdc, 0xe4, 0x7d, 0xb5, 0x2a, 0x60, 0x45, 0x2a, 0x0d, 0xa6, 0x48,
	0xbd, 0xe0, 0x88, 0x70, 0x84, 0x22, 0x68, 0x91, 0x2a, 0x44, 0x5b, 0x81, 0x5b, 0xa8, 0xc4, 0xc5,
	0xda, 0xda, 0xd3, 0x74, 0x55, 0x7f, 0xe1, 0xdd, 0x84, 0xde, 0x39, 0x73, 0xe7, 0x07, 0xf1, 0x13,
	0xf8, 0x41, 0x68, 0xd7, 0xde, 0x34, 0x8e, 0xea, 0xb6, 0x02, 0x0e, 0x9c, 0xbc, 0x3b, 0x3b, 0xf3,
	0x8c, 0xe7, 0x99, 0x67, 0x06, 0xb6, 0xc6, 0x4c, 0x9c, 0x4d, 0x4e, 0x1c, 0x3f, 0x89, 0x86, 0x17,
	0x91, 0xff, 0x38, 0xc0, 0xa9, 0xfc, 0xaa, 0xb3, 0x9f, 0x64, 0x38, 0x4c, 0xb3, 0x44, 0x24, 0x43,
	0x41, 0xf9, 0x79, 0xc8, 0xb8, 0x98, 0x1d, 0x1c, 0x65, 0x27, 0x77, 0x2e, 0x22, 0xdf, 0xe1, 0xd9,
	0xd4, 0x91, 0xbe, 0x8e, 0x7e, 0xec, 0xdf, 0x1f, 0x27, 0xc9, 0x38, 0x2c, 0x82, 0x4f, 0x26, 0xa7,
	0xc3, 0x2f, 0x19, 0x4d, 0x53, 0xcc, 0x78, 0x1e, 0xd6, 0x7f, 0x76, 0xcb, 0xac, 0x3c, 0xa3, 0xf1,
	0x18, 0xf5, 0xb7, 0x08, 0x7e, 0x75, 0xbb, 0x60, 0x8e, 0x34, 0xf3, 0xcf, 0x22, 0x14, 0x74, 0xee,
	0x98, 0x43, 0xd8, 0x3f, 0x0c, 0x68, 0x1f, 0x51, 0x7e, 0xbe, 0xc7, 0xb8, 0x20, 0x3d, 0xa8, 0xb1,
	0xc0, 0x32, 0x06, 0xc6, 0x66, 0xc7, 0xad, 0xb1, 0x80, 0x10, 0x68, 0xc4, 0x34, 0x42, 0xab, 0xa6,
	0x2c, 0xea, 0x4c, 0x06, 0x60, 0x06, 0xc8, 0xfd, 0x8c, 0xa5, 0x82, 0x25, 0xb1, 0x55, 0x57, 0x4f,
	0xf3, 0x26, 0xb2, 0x03, 0x20, 0x58, 0x84, 0x9e, 0xfa, 0x53, 0xab, 0x31, 0x30, 0x36, 0xcd, 0xd1,
	0x86, 0x53, 0xa6, 0xa7, 0x28, 0xe3, 0x88, 0x45, 0xc8, 0x05, 0x8d, 0x52, 0x57, 0x5e, 0xdd, 0x8e,
	0x8c, 0x53, 0x47, 0x72, 0x0f, 0x96, 0x53, 0x3a, 0x46, 0x8f, 0x05, 0x56, 0x53, 0xa5, 0x68, 0xc9,
	0xeb, 0x9b, 0x80, 0xac, 0x42, 0x53, 0x30, 0x11, 0xa2, 0xd5, 0x52, 0xe6, 0xfc, 0x62, 0xef, 0x43,
	0x77, 0x27, 0x43, 0x2a, 0xd0, 0xc5, 0xcf, 0x13, 0xe4, 0x82, 0x3c, 0x87, 0x8e, 0xec, 0x81, 0x27,
	0x9b, 0xa0, 0x2a, 0x32, 0x47, 0xeb, 0xce, 0x95, 0x2d, 0x72, 0x74, 0xf9, 0x6e, 0x5b, 0x14, 0x27,
	0x7b, 0x00, 0x3d, 0x0d, 0xc7, 0xd3, 0x24, 0xe6, 0xb8, 0x48, 0x8d, 0xbd, 0x06, 0xa6, 0x8b, 0x34,
	0xd0, 0xe9, 0x16, 0x9f, 0xf7, 0x60, 0x25, 0x7f, 0x2e, 0xc2, 0xff, 0xf4, 0x77, 0x60, 0x17, 0x85,
	0xce, 0xa5, 0xbb, 0x62, 0x5c, 0x76, 0xc5, 0x7e, 0x0b, 0xa6, 0xf2, 0xf8, 0x2b, 0xe9, 0x7e, 0x1a,
	0xd0, 0xfd, 0x90, 0x06, 0x73, 0x6c, 0xfe, 0x43, 0xc2, 0xb0, 0xa1, 0xcb, 0x51, 0x78, 0xf1, 0x24,
	0x0c, 0x3d, 0x69, 0x55, 0xf2, 0x68, 0xbb, 0x26, 0x47, 0x71, 0x30, 0x09, 0x43, 0x19, 0x58, 0xa1,
	0x91, 0xff, 0xa1, 0xa7, 0xab, 0xca, 0x69, 0xb2, 0x5f, 0x42, 0xf7, 0x35, 0x86, 0x58, 0x5d, 0xe7,
	0x3a, 0x98, 0x21, 0xd2, 0x29, 0x7a, 0x92, 0x1b, 0xae, 0xca, 0x6d, 0xbb, 0xa0, 0x4c, 0x92, 0x37,
	0x2e, 0x31, 0x35, 0x42, 0x81, 0xf9, 0xad, 0x06, 0xdd, 0x43, 0x35, 0x65, 0x1a, 0x74, 0x15, 0x9a,
	0x21, 0x8b, 0x58, 0xde, 0x88, 0xae, 0x9b, 0x5f, 0xc8, 0x5d, 0x68, 0x25, 0xa7, 0xa7, 0x1c, 0x85,
	0x42, 0xed, 0xba, 0xc5, 0x6d, 0x46, 0x6d, 0xbd, 0x9a, 0xda, 0xc6, 0x4d, 0xd4, 0x36, 0x7f, 0x8f,
	0xda, 0x2b, 0x69, 0x23, 0x5b, 0xb0, 0xc2, 0xb8, 0x97, 0x62, 0x16, 0xd1, 0x18, 0x63, 0x61, 0x2d,
	0x2b, 0xf0, 0xbe, 0x93, 0x2f, 0x36, 0x47, 0x2f, 0x36, 0x67, 0x3b, 0x49, 0xc2, 0x8f, 0x34, 0x9c,
	0xa0, 0x6b, 0x32, 0xfe, 0x4e, 0xbb, 0xdb, 0x5f, 0x0d, 0xe8, 0x69, 0x3e, 0x0a, 0x75, 0xbe, 0x00,
	0x98, 0xa9, 0x93, 0x5b, 0xc6, 0xa0, 0x7e, 0x1b, 0x79, 0x76, 0xb4, 0x3c, 0x39, 0x79, 0x02, 0x0d,
	0xb9, 0xc1, 0x14, 0x71, 0xe6, 0x68, 0xad, 0x1c, 0x39, 0xb7, 0xe1, 0xf6, 0x51, 0x50, 0x57, 0xb9,
	0x8e, 0xbe, 0x37, 0xe0, 0x3f, 0x0d, 0x75, 0x88, 0xd9, 0x94, 0xf9, 0x48, 0x8e, 0xa1, 0x95, 0x0f,
	0x39, 0xd9, 0xa8, 0x48, 0x5e, 0x5a, 0x29, 0xfd, 0x47, 0x37, 0x78, 0x15, 0x02, 0x58, 0x22, 0xef,
	0xa1, 0x21, 0x87, 0x9f, 0xd8, 0x15, 0x01, 0x73, 0x8b, 0xa3, 0xff, 0xf0, 0x5a, 0x9f, 0x19, 0xe4,
	0x01, 0xd4, 0x77, 0x51, 0x90, 0x07, 0x15, 0xde, 0x97, 0xdb, 0xa1, 0x6f, 0x5f, 0xe7, 0x32, 0xc3,
	0x3b, 0x86, 0x56, 0x3e, 0x0b, 0x95, 0xb5, 0x97, 0x16, 0x40, 0x65, 0xed, 0x0b, 0x03, 0xa5, 0x80,
	0xf3, 0x81, 0xa8, 0x04, 0x2e, 0x4d, 0x5c, 0x25, 0xf0, 0xc2, 0x54, 0x29, 0xe0, 0x5c, 0x46, 0x95,
	0xc0, 0xa5, 0xa9, 0xab, 0x04, 0x2e, 0x6b, 0xd1, 0x5e, 0xda, 0x86, 0x4f, 0x6d, 0xfd, 0x78, 0xd2,
	0x52, 0x6a, 0x7e, 0xfa, 0x2b, 0x00, 0x00, 0xff, 0xff, 0xaa, 0xf9, 0x97, 0x7a, 0x0c, 0x08, 0x00,
	0x00,
}
