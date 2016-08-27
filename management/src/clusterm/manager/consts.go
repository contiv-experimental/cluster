//go:generate stringer -type=JobStatus $GOFILE

package manager

const (
	// PostNodesCommission is the prefix for the POST REST endpoint
	// to commission one or more assets
	PostNodesCommission = "commission/nodes"

	// PostNodesDecommission is the prefix for the POST REST endpoint
	// to decommission one or more assets
	PostNodesDecommission = "decommission/nodes"

	// PostNodesUpdate is the prefix for the POST REST endpoint
	// to update configuration of one or more assets
	PostNodesUpdate = "update/nodes"

	// PostNodesDiscover is the prefix for the POST REST endpoint
	// to provision one or more specified nodes for discovery
	PostNodesDiscover = "discover/nodes"

	// PostGlobals is the prefix for the POST REST endpoint
	// to set global configuration values
	PostGlobals = "globals"

	// PostMonitorEvent is the prefix for the POST REST endpoint
	// to post a monitor event for one or more nodes.
	PostMonitorEvent = "monitor/event"

	// GetNodeInfoPrefix is the prefix for the GET REST endpoint
	// to fetch info for an asset
	GetNodeInfoPrefix = "info/node"
	getNodeInfo       = GetNodeInfoPrefix + "/{tag}"

	// GetNodesInfo is the prefix for the GET REST endpoint
	// to fetch info for all know assets
	GetNodesInfo = "info/nodes"

	// GetGlobals is the prefix for the GET REST endpoint
	// to fetch the global configuration values
	GetGlobals = "info/globals"

	// GetJobPrefix is the prefix for the GET REST endpoint
	// to fetch the status and logs of a provisioning job. {job} value can be
	// 'active' or 'last'
	GetJobPrefix = "info/job"
	getJob       = GetJobPrefix + "/{job}"

	// GetPostConfig is the prefix for the REST endpoint
	// to GET current or POST updated clusterm's configuration
	GetPostConfig = "config"
)

const (
	ansibleMasterGroupName   = "service-master"
	ansibleWorkerGroupName   = "service-worker"
	ansibleDiscoverGroupName = "cluster-node"
	ansibleNodeNameHostVar   = "node_name"
	ansibleNodeAddrHostVar   = "node_addr"

	jobLabelActive = "active"
	jobLabelLast   = "last"
)

// JobStatus corresponds to possible status values of a job
type JobStatus int

const (
	// Queued is the initial status of a job when it is created
	Queued JobStatus = iota
	// Running is the status of a running job
	Running
	// Complete is the status of the job that ends with success
	Complete
	// Errored is the status of the job that ends with error including user triggered cancellation
	Errored
)
