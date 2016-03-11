package monitor

import "encoding/json"

// Event is the state associate a node monitor event
type Event struct {
	Type EventType
	Node SubsysNode
}

// EventCb is the signature of the callback invoked when a monitor event occurs
type EventCb func(e []Event)

// Subsys provides the following services to the cluster manager:
// - Event interface to notify the manager when a node's operational status
//   changes like discovered, down etc
type Subsys interface {
	// RegisterCb registers the callback associated with pass monitor event type
	RegisterCb(e EventType, cb EventCb) error
	// Start triggers the monitor subsystems and delivers the node monitor
	// events to the client. Start should block and optionall returns error
	// when it encounters a non-revcoverable condition.
	Start() error
}

// SubsysNode provides node level info in a monitoring subsystem
type SubsysNode interface {
	// GetLabel returns the label associated with the node in the monitoring system.
	// This is usually the hostname but can be anything more descriptive.
	GetLabel() string
	// GetSerial returns the serial number of the node. This is usually the chassis' serial number
	GetSerial() string
	// GetAddress return the management address associated with the host. This address is
	// used for pushing configuration to provision a host with cluster level services.
	GetMgmtAddress() string
	// SubsysNode shall satisfy the json marshaller interface to encode node's info in json
	json.Marshaler
}
