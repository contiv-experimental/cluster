package monitor

import "encoding/json"

// Node denotes the common information about the node
type Node struct {
	label  string
	serial string
	addr   string
}

// NewNode returns an instamce of node in monitoring subsystem
func NewNode(label, serial, addr string) *Node {
	return &Node{
		label:  label,
		serial: serial,
		addr:   addr,
	}
}

// GetLabel returns the label associated with the node in the monitoring system.
// This is usually the hostname but can be anything more descriptive.
func (n *Node) GetLabel() string {
	return n.label
}

// GetSerial returns the serial number of the node. This is usually the chassis' serial number
func (n *Node) GetSerial() string {
	return n.serial
}

// GetMgmtAddress returns the management address associated with the host. This address is
// used for pushing configuration to provision a host with cluster level services.
func (n *Node) GetMgmtAddress() string {
	return n.addr
}

// MarshalJSON satisfies the json marshaller interface and shall encode asset info in json
func (n *Node) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Label       string `json:"label"`
		Serial      string `json:"serial-number"`
		MgmtAddress string `json:"management-address"`
	}{
		Label:       n.label,
		Serial:      n.serial,
		MgmtAddress: n.addr,
	})
}
