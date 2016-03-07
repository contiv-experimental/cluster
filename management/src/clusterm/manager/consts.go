package manager

const (
	// PostNodeCommissionPrefix is the prefix for the POST REST endpoint
	// to commission an asset
	PostNodeCommissionPrefix = "commission/node"
	postNodeCommission       = PostNodeCommissionPrefix + "/{tag}"
	// PostNodeDecommissionPrefix is the prefix for the POST REST endpoint
	// to decommission an asset
	PostNodeDecommissionPrefix = "decommission/node"
	postNodeDecommission       = PostNodeDecommissionPrefix + "/{tag}"
	// PostNodeMaintenancePrefix is the prefix for the POST REST endpoint
	// to put an asset in maintenance
	PostNodeMaintenancePrefix = "maintenance/node"
	postNodeMaintenance       = PostNodeMaintenancePrefix + "/{tag}"
	// PostGlobals is the prefix for the POST REST endpoint
	// to set global configuration values
	PostGlobals = "globals"
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
	// ExtraVarsQuery is the key for the query variable used to specify the ansible extra
	// variables for configuration actions. The variables shall be specified as a json map.
	ExtraVarsQuery = "extra_vars"
)

const (
	ansibleMasterGroupName       = "service-master"
	ansibleWorkerGroupName       = "service-worker"
	ansibleNodeNameHostVar       = "node_name"
	ansibleNodeAddrHostVar       = "node_addr"
	ansibleEtcdMasterAddrHostVar = "etcd_master_addr"
	ansibleEtcdMasterNameHostVar = "etcd_master_name"
)
