//go:generate stringer -type=JobStatus $GOFILE

package manager

const (
	// PostNodeCommissionPrefix is the prefix for the POST REST endpoint
	// to commission an asset
	PostNodeCommissionPrefix = "commission/node"
	postNodeCommission       = PostNodeCommissionPrefix + "/{tag}"

	// PostNodesCommission is the prefix for the POST REST endpoint
	// to commission one or more assets
	PostNodesCommission = "commission/nodes"

	// PostNodeDecommissionPrefix is the prefix for the POST REST endpoint
	// to decommission an asset
	PostNodeDecommissionPrefix = "decommission/node"
	postNodeDecommission       = PostNodeDecommissionPrefix + "/{tag}"

	// PostNodesDecommission is the prefix for the POST REST endpoint
	// to decommission one or more assets
	PostNodesDecommission = "decommission/nodes"

	// PostNodeMaintenancePrefix is the prefix for the POST REST endpoint
	// to put an asset in maintenance
	PostNodeMaintenancePrefix = "maintenance/node"
	postNodeMaintenance       = PostNodeMaintenancePrefix + "/{tag}"

	// PostNodesMaintenance is the prefix for the POST REST endpoint
	// to put one or more assets in maintenance
	PostNodesMaintenance = "maintenance/nodes"

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

	// ExtraVarsQuery is the key for the query variable used to specify the ansible extra
	// variables for configuration actions. The variables shall be specified as a json map.
	ExtraVarsQuery = "extra_vars"
)

const (
	ansibleMasterGroupName       = "service-master"
	ansibleWorkerGroupName       = "service-worker"
	ansibleDiscoverGroupName     = "cluster-node"
	ansibleNodeNameHostVar       = "node_name"
	ansibleNodeAddrHostVar       = "node_addr"
	ansibleEtcdMasterAddrHostVar = "etcd_master_addr"
	ansibleEtcdMasterNameHostVar = "etcd_master_name"

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
