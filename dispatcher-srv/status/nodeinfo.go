package status

import (
	"github.com/xmc-dev/xmc/eval-srv/proto/eval"
)

type NodeInfo struct {
	ID          string
	Name        string
	Description string
	Address     string
	Idle        bool
	Disabled    bool
}

func NewNodeInfo(pni *eval.NodeInfo, address string, id string) *NodeInfo {
	pi := &NodeInfo{
		ID:          id,
		Name:        pni.Name,
		Description: pni.Description,
		Address:     address,
		Idle:        pni.Idle,
		Disabled:    pni.Disabled,
	}

	return pi
}

func (ni *NodeInfo) Update(pni *eval.NodeInfo, address string) {
	ni.Address = address
	ni.Idle = pni.Idle
}

func (ni *NodeInfo) ToProto() *eval.NodeInfo {
	pni := &eval.NodeInfo{
		Id:          ni.ID,
		Name:        ni.Name,
		Description: ni.Description,
		Address:     ni.Address,
		Idle:        ni.Idle,
		Disabled:    ni.Disabled,
	}

	return pni
}
