package httpt

import (
	"net"
)

// MsgNode transports node information
type MsgNode struct {
	IP   net.IP `json:"ip"`
	Port uint16 `json:"port"`
}

// MsgPushReq pushes information of a single node
type MsgPushReq struct{ MsgNode }

// MsgPullResp returns a list of nodes info
type MsgPullResp []MsgNode

// MsgProbeResp returns status info of a node
type MsgProbeResp struct {
	Active bool `json:"active"`
}

// MsgEmitReq requests a peer to emit data
type MsgEmitReq struct {
	Data []byte `json:"data"`
}
