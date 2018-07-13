// Code generated by protoc-gen-go. DO NOT EDIT.
// source: github.com/xmc-dev/xmc/xmc-core/proto/page/page.proto

/*
Package page is a generated protocol buffer package.

It is generated from these files:
	github.com/xmc-dev/xmc/xmc-core/proto/page/page.proto

It has these top-level messages:
	Version
	Page
	CreateRequest
	CreateResponse
	ReadRequest
	ReadResponse
	GetRequest
	GetResponse
	GetVersionsRequest
	GetVersionsResponse
	GetFirstChildrenRequest
	GetFirstChildrenResponse
	UpdateRequest
	UpdateResponse
	DeleteRequest
	DeleteResponse
	SearchRequest
	SearchResponse
*/
package page

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"
import google_protobuf "github.com/golang/protobuf/ptypes/timestamp"
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

type Version struct {
	PageId    string                     `protobuf:"bytes,1,opt,name=page_id,json=pageId" json:"page_id,omitempty"`
	Timestamp *google_protobuf.Timestamp `protobuf:"bytes,2,opt,name=timestamp" json:"timestamp,omitempty"`
	Title     string                     `protobuf:"bytes,4,opt,name=title" json:"title,omitempty"`
	Contents  string                     `protobuf:"bytes,5,opt,name=contents" json:"contents,omitempty"`
}

func (m *Version) Reset()                    { *m = Version{} }
func (m *Version) String() string            { return proto.CompactTextString(m) }
func (*Version) ProtoMessage()               {}
func (*Version) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *Version) GetPageId() string {
	if m != nil {
		return m.PageId
	}
	return ""
}

func (m *Version) GetTimestamp() *google_protobuf.Timestamp {
	if m != nil {
		return m.Timestamp
	}
	return nil
}

func (m *Version) GetTitle() string {
	if m != nil {
		return m.Title
	}
	return ""
}

func (m *Version) GetContents() string {
	if m != nil {
		return m.Contents
	}
	return ""
}

type Page struct {
	Id              string                     `protobuf:"bytes,1,opt,name=id" json:"id,omitempty"`
	Path            string                     `protobuf:"bytes,2,opt,name=path" json:"path,omitempty"`
	LatestTimestamp *google_protobuf.Timestamp `protobuf:"bytes,3,opt,name=latest_timestamp,json=latestTimestamp" json:"latest_timestamp,omitempty"`
	Version         *Version                   `protobuf:"bytes,4,opt,name=version" json:"version,omitempty"`
}

func (m *Page) Reset()                    { *m = Page{} }
func (m *Page) String() string            { return proto.CompactTextString(m) }
func (*Page) ProtoMessage()               {}
func (*Page) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *Page) GetId() string {
	if m != nil {
		return m.Id
	}
	return ""
}

func (m *Page) GetPath() string {
	if m != nil {
		return m.Path
	}
	return ""
}

func (m *Page) GetLatestTimestamp() *google_protobuf.Timestamp {
	if m != nil {
		return m.LatestTimestamp
	}
	return nil
}

func (m *Page) GetVersion() *Version {
	if m != nil {
		return m.Version
	}
	return nil
}

type CreateRequest struct {
	Page     *Page  `protobuf:"bytes,1,opt,name=page" json:"page,omitempty"`
	Contents string `protobuf:"bytes,2,opt,name=contents" json:"contents,omitempty"`
	Title    string `protobuf:"bytes,3,opt,name=title" json:"title,omitempty"`
}

func (m *CreateRequest) Reset()                    { *m = CreateRequest{} }
func (m *CreateRequest) String() string            { return proto.CompactTextString(m) }
func (*CreateRequest) ProtoMessage()               {}
func (*CreateRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

func (m *CreateRequest) GetPage() *Page {
	if m != nil {
		return m.Page
	}
	return nil
}

func (m *CreateRequest) GetContents() string {
	if m != nil {
		return m.Contents
	}
	return ""
}

func (m *CreateRequest) GetTitle() string {
	if m != nil {
		return m.Title
	}
	return ""
}

type CreateResponse struct {
	Id string `protobuf:"bytes,1,opt,name=id" json:"id,omitempty"`
}

func (m *CreateResponse) Reset()                    { *m = CreateResponse{} }
func (m *CreateResponse) String() string            { return proto.CompactTextString(m) }
func (*CreateResponse) ProtoMessage()               {}
func (*CreateResponse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{3} }

func (m *CreateResponse) GetId() string {
	if m != nil {
		return m.Id
	}
	return ""
}

type ReadRequest struct {
	Id        string                     `protobuf:"bytes,1,opt,name=id" json:"id,omitempty"`
	Timestamp *google_protobuf.Timestamp `protobuf:"bytes,2,opt,name=timestamp" json:"timestamp,omitempty"`
}

func (m *ReadRequest) Reset()                    { *m = ReadRequest{} }
func (m *ReadRequest) String() string            { return proto.CompactTextString(m) }
func (*ReadRequest) ProtoMessage()               {}
func (*ReadRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{4} }

func (m *ReadRequest) GetId() string {
	if m != nil {
		return m.Id
	}
	return ""
}

func (m *ReadRequest) GetTimestamp() *google_protobuf.Timestamp {
	if m != nil {
		return m.Timestamp
	}
	return nil
}

type ReadResponse struct {
	Page *Page `protobuf:"bytes,1,opt,name=page" json:"page,omitempty"`
}

func (m *ReadResponse) Reset()                    { *m = ReadResponse{} }
func (m *ReadResponse) String() string            { return proto.CompactTextString(m) }
func (*ReadResponse) ProtoMessage()               {}
func (*ReadResponse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{5} }

func (m *ReadResponse) GetPage() *Page {
	if m != nil {
		return m.Page
	}
	return nil
}

type GetRequest struct {
	Path string `protobuf:"bytes,1,opt,name=path" json:"path,omitempty"`
}

func (m *GetRequest) Reset()                    { *m = GetRequest{} }
func (m *GetRequest) String() string            { return proto.CompactTextString(m) }
func (*GetRequest) ProtoMessage()               {}
func (*GetRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{6} }

func (m *GetRequest) GetPath() string {
	if m != nil {
		return m.Path
	}
	return ""
}

type GetResponse struct {
	Page *Page `protobuf:"bytes,1,opt,name=page" json:"page,omitempty"`
}

func (m *GetResponse) Reset()                    { *m = GetResponse{} }
func (m *GetResponse) String() string            { return proto.CompactTextString(m) }
func (*GetResponse) ProtoMessage()               {}
func (*GetResponse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{7} }

func (m *GetResponse) GetPage() *Page {
	if m != nil {
		return m.Page
	}
	return nil
}

type GetVersionsRequest struct {
	Limit  uint32 `protobuf:"varint,1,opt,name=limit" json:"limit,omitempty"`
	Offset uint32 `protobuf:"varint,2,opt,name=offset" json:"offset,omitempty"`
	Id     string `protobuf:"bytes,3,opt,name=id" json:"id,omitempty"`
}

func (m *GetVersionsRequest) Reset()                    { *m = GetVersionsRequest{} }
func (m *GetVersionsRequest) String() string            { return proto.CompactTextString(m) }
func (*GetVersionsRequest) ProtoMessage()               {}
func (*GetVersionsRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{8} }

func (m *GetVersionsRequest) GetLimit() uint32 {
	if m != nil {
		return m.Limit
	}
	return 0
}

func (m *GetVersionsRequest) GetOffset() uint32 {
	if m != nil {
		return m.Offset
	}
	return 0
}

func (m *GetVersionsRequest) GetId() string {
	if m != nil {
		return m.Id
	}
	return ""
}

type GetVersionsResponse struct {
	Versions []*Version                    `protobuf:"bytes,1,rep,name=versions" json:"versions,omitempty"`
	Meta     *xmc_srv_core_searchmeta.Meta `protobuf:"bytes,2,opt,name=meta" json:"meta,omitempty"`
}

func (m *GetVersionsResponse) Reset()                    { *m = GetVersionsResponse{} }
func (m *GetVersionsResponse) String() string            { return proto.CompactTextString(m) }
func (*GetVersionsResponse) ProtoMessage()               {}
func (*GetVersionsResponse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{9} }

func (m *GetVersionsResponse) GetVersions() []*Version {
	if m != nil {
		return m.Versions
	}
	return nil
}

func (m *GetVersionsResponse) GetMeta() *xmc_srv_core_searchmeta.Meta {
	if m != nil {
		return m.Meta
	}
	return nil
}

type GetFirstChildrenRequest struct {
	Limit  uint32 `protobuf:"varint,1,opt,name=limit" json:"limit,omitempty"`
	Offset uint32 `protobuf:"varint,2,opt,name=offset" json:"offset,omitempty"`
	Id     string `protobuf:"bytes,3,opt,name=id" json:"id,omitempty"`
}

func (m *GetFirstChildrenRequest) Reset()                    { *m = GetFirstChildrenRequest{} }
func (m *GetFirstChildrenRequest) String() string            { return proto.CompactTextString(m) }
func (*GetFirstChildrenRequest) ProtoMessage()               {}
func (*GetFirstChildrenRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{10} }

func (m *GetFirstChildrenRequest) GetLimit() uint32 {
	if m != nil {
		return m.Limit
	}
	return 0
}

func (m *GetFirstChildrenRequest) GetOffset() uint32 {
	if m != nil {
		return m.Offset
	}
	return 0
}

func (m *GetFirstChildrenRequest) GetId() string {
	if m != nil {
		return m.Id
	}
	return ""
}

type GetFirstChildrenResponse struct {
	Pages []*Page                       `protobuf:"bytes,1,rep,name=pages" json:"pages,omitempty"`
	Meta  *xmc_srv_core_searchmeta.Meta `protobuf:"bytes,2,opt,name=meta" json:"meta,omitempty"`
}

func (m *GetFirstChildrenResponse) Reset()                    { *m = GetFirstChildrenResponse{} }
func (m *GetFirstChildrenResponse) String() string            { return proto.CompactTextString(m) }
func (*GetFirstChildrenResponse) ProtoMessage()               {}
func (*GetFirstChildrenResponse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{11} }

func (m *GetFirstChildrenResponse) GetPages() []*Page {
	if m != nil {
		return m.Pages
	}
	return nil
}

func (m *GetFirstChildrenResponse) GetMeta() *xmc_srv_core_searchmeta.Meta {
	if m != nil {
		return m.Meta
	}
	return nil
}

type UpdateRequest struct {
	Id       string `protobuf:"bytes,1,opt,name=id" json:"id,omitempty"`
	Path     string `protobuf:"bytes,2,opt,name=path" json:"path,omitempty"`
	Contents string `protobuf:"bytes,3,opt,name=contents" json:"contents,omitempty"`
	Title    string `protobuf:"bytes,4,opt,name=title" json:"title,omitempty"`
}

func (m *UpdateRequest) Reset()                    { *m = UpdateRequest{} }
func (m *UpdateRequest) String() string            { return proto.CompactTextString(m) }
func (*UpdateRequest) ProtoMessage()               {}
func (*UpdateRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{12} }

func (m *UpdateRequest) GetId() string {
	if m != nil {
		return m.Id
	}
	return ""
}

func (m *UpdateRequest) GetPath() string {
	if m != nil {
		return m.Path
	}
	return ""
}

func (m *UpdateRequest) GetContents() string {
	if m != nil {
		return m.Contents
	}
	return ""
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
func (*UpdateResponse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{13} }

type DeleteRequest struct {
	Id   string `protobuf:"bytes,1,opt,name=id" json:"id,omitempty"`
	Hard bool   `protobuf:"varint,2,opt,name=hard" json:"hard,omitempty"`
}

func (m *DeleteRequest) Reset()                    { *m = DeleteRequest{} }
func (m *DeleteRequest) String() string            { return proto.CompactTextString(m) }
func (*DeleteRequest) ProtoMessage()               {}
func (*DeleteRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{14} }

func (m *DeleteRequest) GetId() string {
	if m != nil {
		return m.Id
	}
	return ""
}

func (m *DeleteRequest) GetHard() bool {
	if m != nil {
		return m.Hard
	}
	return false
}

type DeleteResponse struct {
}

func (m *DeleteResponse) Reset()                    { *m = DeleteResponse{} }
func (m *DeleteResponse) String() string            { return proto.CompactTextString(m) }
func (*DeleteResponse) ProtoMessage()               {}
func (*DeleteResponse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{15} }

type SearchRequest struct {
	Limit  uint32 `protobuf:"varint,1,opt,name=limit" json:"limit,omitempty"`
	Offset uint32 `protobuf:"varint,2,opt,name=offset" json:"offset,omitempty"`
	Path   string `protobuf:"bytes,3,opt,name=path" json:"path,omitempty"`
	Title  string `protobuf:"bytes,4,opt,name=title" json:"title,omitempty"`
}

func (m *SearchRequest) Reset()                    { *m = SearchRequest{} }
func (m *SearchRequest) String() string            { return proto.CompactTextString(m) }
func (*SearchRequest) ProtoMessage()               {}
func (*SearchRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{16} }

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

func (m *SearchRequest) GetPath() string {
	if m != nil {
		return m.Path
	}
	return ""
}

func (m *SearchRequest) GetTitle() string {
	if m != nil {
		return m.Title
	}
	return ""
}

type SearchResponse struct {
	Pages []*Page                       `protobuf:"bytes,1,rep,name=pages" json:"pages,omitempty"`
	Meta  *xmc_srv_core_searchmeta.Meta `protobuf:"bytes,2,opt,name=meta" json:"meta,omitempty"`
}

func (m *SearchResponse) Reset()                    { *m = SearchResponse{} }
func (m *SearchResponse) String() string            { return proto.CompactTextString(m) }
func (*SearchResponse) ProtoMessage()               {}
func (*SearchResponse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{17} }

func (m *SearchResponse) GetPages() []*Page {
	if m != nil {
		return m.Pages
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
	proto.RegisterType((*Version)(nil), "xmc.srv.core.page.Version")
	proto.RegisterType((*Page)(nil), "xmc.srv.core.page.Page")
	proto.RegisterType((*CreateRequest)(nil), "xmc.srv.core.page.CreateRequest")
	proto.RegisterType((*CreateResponse)(nil), "xmc.srv.core.page.CreateResponse")
	proto.RegisterType((*ReadRequest)(nil), "xmc.srv.core.page.ReadRequest")
	proto.RegisterType((*ReadResponse)(nil), "xmc.srv.core.page.ReadResponse")
	proto.RegisterType((*GetRequest)(nil), "xmc.srv.core.page.GetRequest")
	proto.RegisterType((*GetResponse)(nil), "xmc.srv.core.page.GetResponse")
	proto.RegisterType((*GetVersionsRequest)(nil), "xmc.srv.core.page.GetVersionsRequest")
	proto.RegisterType((*GetVersionsResponse)(nil), "xmc.srv.core.page.GetVersionsResponse")
	proto.RegisterType((*GetFirstChildrenRequest)(nil), "xmc.srv.core.page.GetFirstChildrenRequest")
	proto.RegisterType((*GetFirstChildrenResponse)(nil), "xmc.srv.core.page.GetFirstChildrenResponse")
	proto.RegisterType((*UpdateRequest)(nil), "xmc.srv.core.page.UpdateRequest")
	proto.RegisterType((*UpdateResponse)(nil), "xmc.srv.core.page.UpdateResponse")
	proto.RegisterType((*DeleteRequest)(nil), "xmc.srv.core.page.DeleteRequest")
	proto.RegisterType((*DeleteResponse)(nil), "xmc.srv.core.page.DeleteResponse")
	proto.RegisterType((*SearchRequest)(nil), "xmc.srv.core.page.SearchRequest")
	proto.RegisterType((*SearchResponse)(nil), "xmc.srv.core.page.SearchResponse")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ client.Option
var _ server.Option

// Client API for PageService service

type PageServiceClient interface {
	Create(ctx context.Context, in *CreateRequest, opts ...client.CallOption) (*CreateResponse, error)
	Read(ctx context.Context, in *ReadRequest, opts ...client.CallOption) (*ReadResponse, error)
	Get(ctx context.Context, in *GetRequest, opts ...client.CallOption) (*GetResponse, error)
	GetVersions(ctx context.Context, in *GetVersionsRequest, opts ...client.CallOption) (*GetVersionsResponse, error)
	GetFirstChildren(ctx context.Context, in *GetFirstChildrenRequest, opts ...client.CallOption) (*GetFirstChildrenResponse, error)
	Update(ctx context.Context, in *UpdateRequest, opts ...client.CallOption) (*UpdateResponse, error)
	Delete(ctx context.Context, in *DeleteRequest, opts ...client.CallOption) (*DeleteResponse, error)
	Search(ctx context.Context, in *SearchRequest, opts ...client.CallOption) (*SearchResponse, error)
}

type pageServiceClient struct {
	c           client.Client
	serviceName string
}

func NewPageServiceClient(serviceName string, c client.Client) PageServiceClient {
	if c == nil {
		c = client.NewClient()
	}
	if len(serviceName) == 0 {
		serviceName = "xmc.srv.core.page"
	}
	return &pageServiceClient{
		c:           c,
		serviceName: serviceName,
	}
}

func (c *pageServiceClient) Create(ctx context.Context, in *CreateRequest, opts ...client.CallOption) (*CreateResponse, error) {
	req := c.c.NewRequest(c.serviceName, "PageService.Create", in)
	out := new(CreateResponse)
	err := c.c.Call(ctx, req, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *pageServiceClient) Read(ctx context.Context, in *ReadRequest, opts ...client.CallOption) (*ReadResponse, error) {
	req := c.c.NewRequest(c.serviceName, "PageService.Read", in)
	out := new(ReadResponse)
	err := c.c.Call(ctx, req, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *pageServiceClient) Get(ctx context.Context, in *GetRequest, opts ...client.CallOption) (*GetResponse, error) {
	req := c.c.NewRequest(c.serviceName, "PageService.Get", in)
	out := new(GetResponse)
	err := c.c.Call(ctx, req, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *pageServiceClient) GetVersions(ctx context.Context, in *GetVersionsRequest, opts ...client.CallOption) (*GetVersionsResponse, error) {
	req := c.c.NewRequest(c.serviceName, "PageService.GetVersions", in)
	out := new(GetVersionsResponse)
	err := c.c.Call(ctx, req, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *pageServiceClient) GetFirstChildren(ctx context.Context, in *GetFirstChildrenRequest, opts ...client.CallOption) (*GetFirstChildrenResponse, error) {
	req := c.c.NewRequest(c.serviceName, "PageService.GetFirstChildren", in)
	out := new(GetFirstChildrenResponse)
	err := c.c.Call(ctx, req, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *pageServiceClient) Update(ctx context.Context, in *UpdateRequest, opts ...client.CallOption) (*UpdateResponse, error) {
	req := c.c.NewRequest(c.serviceName, "PageService.Update", in)
	out := new(UpdateResponse)
	err := c.c.Call(ctx, req, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *pageServiceClient) Delete(ctx context.Context, in *DeleteRequest, opts ...client.CallOption) (*DeleteResponse, error) {
	req := c.c.NewRequest(c.serviceName, "PageService.Delete", in)
	out := new(DeleteResponse)
	err := c.c.Call(ctx, req, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *pageServiceClient) Search(ctx context.Context, in *SearchRequest, opts ...client.CallOption) (*SearchResponse, error) {
	req := c.c.NewRequest(c.serviceName, "PageService.Search", in)
	out := new(SearchResponse)
	err := c.c.Call(ctx, req, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for PageService service

type PageServiceHandler interface {
	Create(context.Context, *CreateRequest, *CreateResponse) error
	Read(context.Context, *ReadRequest, *ReadResponse) error
	Get(context.Context, *GetRequest, *GetResponse) error
	GetVersions(context.Context, *GetVersionsRequest, *GetVersionsResponse) error
	GetFirstChildren(context.Context, *GetFirstChildrenRequest, *GetFirstChildrenResponse) error
	Update(context.Context, *UpdateRequest, *UpdateResponse) error
	Delete(context.Context, *DeleteRequest, *DeleteResponse) error
	Search(context.Context, *SearchRequest, *SearchResponse) error
}

func RegisterPageServiceHandler(s server.Server, hdlr PageServiceHandler, opts ...server.HandlerOption) {
	s.Handle(s.NewHandler(&PageService{hdlr}, opts...))
}

type PageService struct {
	PageServiceHandler
}

func (h *PageService) Create(ctx context.Context, in *CreateRequest, out *CreateResponse) error {
	return h.PageServiceHandler.Create(ctx, in, out)
}

func (h *PageService) Read(ctx context.Context, in *ReadRequest, out *ReadResponse) error {
	return h.PageServiceHandler.Read(ctx, in, out)
}

func (h *PageService) Get(ctx context.Context, in *GetRequest, out *GetResponse) error {
	return h.PageServiceHandler.Get(ctx, in, out)
}

func (h *PageService) GetVersions(ctx context.Context, in *GetVersionsRequest, out *GetVersionsResponse) error {
	return h.PageServiceHandler.GetVersions(ctx, in, out)
}

func (h *PageService) GetFirstChildren(ctx context.Context, in *GetFirstChildrenRequest, out *GetFirstChildrenResponse) error {
	return h.PageServiceHandler.GetFirstChildren(ctx, in, out)
}

func (h *PageService) Update(ctx context.Context, in *UpdateRequest, out *UpdateResponse) error {
	return h.PageServiceHandler.Update(ctx, in, out)
}

func (h *PageService) Delete(ctx context.Context, in *DeleteRequest, out *DeleteResponse) error {
	return h.PageServiceHandler.Delete(ctx, in, out)
}

func (h *PageService) Search(ctx context.Context, in *SearchRequest, out *SearchResponse) error {
	return h.PageServiceHandler.Search(ctx, in, out)
}

func init() {
	proto.RegisterFile("github.com/xmc-dev/xmc/xmc-core/proto/page/page.proto", fileDescriptor0)
}

var fileDescriptor0 = []byte{
	// 740 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xbc, 0x55, 0xd1, 0x6e, 0xd3, 0x4a,
	0x10, 0xad, 0x63, 0x27, 0x4d, 0x27, 0x37, 0xb9, 0xb9, 0x7b, 0x2b, 0x1a, 0x59, 0x6a, 0x1b, 0x2c,
	0x81, 0x2a, 0xaa, 0x3a, 0xa2, 0x05, 0x84, 0xe0, 0x09, 0x0a, 0x8d, 0x8a, 0x84, 0x8a, 0x5c, 0xa0,
	0x12, 0x0f, 0x54, 0xae, 0x3d, 0x75, 0x2c, 0xc5, 0x71, 0xb0, 0xb7, 0x51, 0x1f, 0x78, 0xe0, 0x23,
	0xf8, 0x07, 0x3e, 0x80, 0x1f, 0x44, 0xbb, 0xeb, 0x75, 0xec, 0x74, 0xd3, 0x96, 0x08, 0xf1, 0xd2,
	0x7a, 0xb3, 0x67, 0xce, 0xcc, 0x9c, 0xf1, 0x19, 0xc3, 0xe3, 0x20, 0xa4, 0x83, 0x8b, 0x33, 0xdb,
	0x8b, 0xa3, 0xde, 0x65, 0xe4, 0xed, 0xf8, 0x38, 0x61, 0xff, 0xf9, 0xb3, 0x17, 0x27, 0xd8, 0x1b,
	0x27, 0x31, 0x8d, 0x7b, 0x63, 0x37, 0x40, 0xfe, 0xc7, 0xe6, 0x67, 0xf2, 0xdf, 0x65, 0xe4, 0xd9,
	0x69, 0x32, 0xb1, 0x19, 0xc6, 0x66, 0x17, 0xe6, 0x66, 0x10, 0xc7, 0xc1, 0x30, 0x0b, 0x38, 0xbb,
	0x38, 0xef, 0xd1, 0x30, 0xc2, 0x94, 0xba, 0xd1, 0x58, 0xc4, 0x98, 0x2f, 0x6e, 0x97, 0x2a, 0x45,
	0x37, 0xf1, 0x06, 0x11, 0x52, 0xb7, 0xf0, 0x28, 0x28, 0xac, 0xef, 0x1a, 0x2c, 0x7f, 0xc4, 0x24,
	0x0d, 0xe3, 0x11, 0x59, 0x83, 0x65, 0x96, 0xf7, 0x34, 0xf4, 0x3b, 0x5a, 0x57, 0xdb, 0x5a, 0x71,
	0x6a, 0xec, 0x78, 0xe8, 0x93, 0xa7, 0xb0, 0x92, 0xa7, 0xee, 0x54, 0xba, 0xda, 0x56, 0x63, 0xd7,
	0xb4, 0x45, 0x71, 0xb6, 0x2c, 0xce, 0x7e, 0x2f, 0x11, 0xce, 0x14, 0x4c, 0x56, 0xa1, 0x4a, 0x43,
	0x3a, 0xc4, 0x8e, 0xc1, 0x09, 0xc5, 0x81, 0x98, 0x50, 0xf7, 0xe2, 0x11, 0xc5, 0x11, 0x4d, 0x3b,
	0x55, 0x7e, 0x91, 0x9f, 0xdf, 0x18, 0x75, 0xbd, 0x6d, 0x58, 0x3f, 0x34, 0x30, 0xde, 0xb9, 0x01,
	0x92, 0x16, 0x54, 0xf2, 0x72, 0x2a, 0xa1, 0x4f, 0x08, 0x18, 0x63, 0x97, 0x0e, 0x78, 0x15, 0x2b,
	0x0e, 0x7f, 0x26, 0xaf, 0xa1, 0x3d, 0x74, 0x29, 0xa6, 0xf4, 0x74, 0x5a, 0xa5, 0x7e, 0x63, 0x95,
	0xff, 0x8a, 0x98, 0xfc, 0x07, 0xf2, 0x08, 0x96, 0x27, 0x42, 0x09, 0x5e, 0x2d, 0x8b, 0xbe, 0x32,
	0x13, 0x3b, 0xd3, 0xca, 0x91, 0x50, 0x6b, 0x04, 0xcd, 0xfd, 0x04, 0x5d, 0x8a, 0x0e, 0x7e, 0xb9,
	0xc0, 0x94, 0x92, 0x6d, 0x56, 0x61, 0x80, 0xbc, 0xe6, 0xc6, 0xee, 0x9a, 0x82, 0x83, 0x35, 0xe6,
	0x70, 0x50, 0x49, 0x89, 0x4a, 0x59, 0x89, 0xa9, 0x76, 0x7a, 0x41, 0x3b, 0xab, 0x0b, 0x2d, 0x99,
	0x2f, 0x1d, 0xc7, 0xa3, 0xf4, 0x8a, 0x44, 0xd6, 0x09, 0x34, 0x1c, 0x74, 0x7d, 0x59, 0xcf, 0xac,
	0x82, 0x0b, 0x0f, 0xd3, 0x7a, 0x0e, 0xff, 0x08, 0xe2, 0x2c, 0xf1, 0xef, 0x74, 0x6a, 0x75, 0x01,
	0xfa, 0x48, 0x65, 0x51, 0x72, 0x8c, 0xda, 0x74, 0x8c, 0xd6, 0x33, 0x68, 0x70, 0xc4, 0x22, 0xec,
	0x0e, 0x90, 0x3e, 0xd2, 0x6c, 0x38, 0xa9, 0xcc, 0xb2, 0x0a, 0xd5, 0x61, 0x18, 0x85, 0x94, 0x73,
	0x34, 0x1d, 0x71, 0x20, 0x77, 0xa0, 0x16, 0x9f, 0x9f, 0xa7, 0x48, 0x79, 0xf7, 0x4d, 0x27, 0x3b,
	0x65, 0x42, 0xe9, 0xb9, 0x8e, 0xdf, 0x34, 0xf8, 0xbf, 0x44, 0x9a, 0x15, 0xf6, 0x04, 0xea, 0xd9,
	0xf0, 0xd3, 0x8e, 0xd6, 0xd5, 0x6f, 0x78, 0x51, 0x72, 0x2c, 0x79, 0x08, 0x06, 0x33, 0x5e, 0xa6,
	0xf9, 0x7a, 0x39, 0xa6, 0x60, 0xcc, 0xb7, 0x48, 0x5d, 0x87, 0x43, 0xad, 0x13, 0x58, 0xeb, 0x23,
	0x3d, 0x08, 0x93, 0x94, 0xee, 0x0f, 0xc2, 0xa1, 0x9f, 0xe0, 0xe8, 0xcf, 0xf4, 0xf6, 0x15, 0x3a,
	0x57, 0x89, 0xb3, 0xfe, 0x76, 0xa0, 0xca, 0x3a, 0x90, 0xcd, 0xcd, 0x55, 0x5e, 0xa0, 0x16, 0x69,
	0x0b, 0xa1, 0xf9, 0x61, 0xec, 0x17, 0x3c, 0x73, 0x1b, 0x97, 0x17, 0xad, 0xa2, 0xcf, 0xb3, 0x4a,
	0x71, 0xcd, 0x58, 0x6d, 0x68, 0xc9, 0x34, 0xa2, 0x35, 0x6b, 0x0f, 0x9a, 0xaf, 0x70, 0x88, 0xd7,
	0x26, 0x1e, 0xb8, 0x89, 0xcf, 0x13, 0xd7, 0x1d, 0xfe, 0xcc, 0x68, 0x64, 0x50, 0x46, 0x13, 0x40,
	0xf3, 0x98, 0x37, 0xb6, 0xd8, 0x30, 0x64, 0x77, 0x7a, 0xa1, 0x3b, 0x75, 0x07, 0x09, 0xb4, 0x64,
	0xa2, 0xbf, 0x35, 0x9c, 0xdd, 0x9f, 0x55, 0x68, 0x30, 0x8a, 0x63, 0x4c, 0x26, 0xa1, 0x87, 0xe4,
	0x08, 0x6a, 0x62, 0xe1, 0x90, 0xae, 0x22, 0x59, 0x69, 0xf7, 0x99, 0x77, 0xaf, 0x41, 0x64, 0xda,
	0x2d, 0x91, 0x43, 0x30, 0xd8, 0x1a, 0x21, 0x1b, 0x0a, 0x70, 0x61, 0x71, 0x99, 0x9b, 0x73, 0xef,
	0x73, 0xaa, 0x03, 0xd0, 0xfb, 0x48, 0xc9, 0xba, 0x02, 0x39, 0x5d, 0x36, 0xe6, 0xc6, 0xbc, 0xeb,
	0x9c, 0xe7, 0x33, 0x5f, 0x3d, 0xd2, 0xe9, 0xe4, 0x9e, 0x3a, 0x60, 0x66, 0xbd, 0x98, 0xf7, 0x6f,
	0x82, 0xe5, 0xfc, 0x11, 0xb4, 0x67, 0xed, 0x46, 0x1e, 0xa8, 0xa3, 0x55, 0x66, 0x37, 0xb7, 0x6f,
	0x85, 0xcd, 0xd3, 0x1d, 0x41, 0x4d, 0xbc, 0xf8, 0xca, 0x91, 0x95, 0xac, 0xa7, 0x1c, 0xd9, 0x8c,
	0x6b, 0x38, 0xa1, 0xb0, 0x80, 0x92, 0xb0, 0x64, 0x29, 0x25, 0xe1, 0x8c, 0x7f, 0x38, 0xa1, 0x78,
	0xb1, 0x95, 0x84, 0x25, 0x73, 0x29, 0x09, 0xcb, 0xae, 0xb0, 0x96, 0x5e, 0xd6, 0x3e, 0xf1, 0x0f,
	0xc1, 0x59, 0x8d, 0x7f, 0xc2, 0xf6, 0x7e, 0x05, 0x00, 0x00, 0xff, 0xff, 0xf5, 0x72, 0x13, 0x7d,
	0x86, 0x09, 0x00, 0x00,
}
