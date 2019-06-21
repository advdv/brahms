package brahms

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"net"
	"strconv"
)

// NID is a node id
type NID [32]byte

func (id NID) String() string { return hex.EncodeToString(id[:2]) }

// Bytes returns the id as a byte slice
func (id NID) Bytes() []byte {
	return id[:]
}

// IsNil returns whether the is its zero value
func (id NID) IsNil() bool {
	return id == NID{}
}

// N describes a node by its ip info
func N(ip string, port uint16) (n *Node) {
	return &Node{IP: net.ParseIP(ip), Port: port}
}

// Node describes how to reach another peer in the network
type Node struct {
	IP   net.IP
	Port uint16
}

// Hash a node description into an ip
func (n *Node) Hash() (id NID) {
	pb := make([]byte, 2)
	binary.BigEndian.PutUint16(pb, n.Port)
	return NID(sha256.Sum256(append(n.IP, pb...)))
}

func (n *Node) String() string {
	return n.IP.String() + ":" + strconv.Itoa(int(n.Port))
}

// IsZero returns whether the node is a zero value
func (n Node) IsZero() bool {
	return (n.IP == nil && n.Port == 0)
}
