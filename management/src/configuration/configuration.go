package configuration

import "io"

// Subsys provides the following services to the cluster manager:
// - Interface to trigger configuration action on one or more nodes, with
//   possible actions being configure, cleanup and upgrade.
type Subsys interface {
	// Configure triggers the configuration logic on specified set of nodes.
	// It return a error channel that the caller can wait on to get completion status.
	Configure(nodes SubsysHosts) (io.Reader, io.Reader, chan error)
	Cleanup(nodes SubsysHosts) (io.Reader, io.Reader, chan error)
	Upgrade(nodes SubsysHosts) (io.Reader, io.Reader, chan error)
}

// SubsysHost denotes a host in configuration subsystem
type SubsysHost interface{}

// SubsysHosts denotes a collection of hosts in configuration subsystem
type SubsysHosts interface{}
