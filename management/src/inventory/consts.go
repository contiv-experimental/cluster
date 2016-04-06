//go:generate stringer -type=AssetStatus,AssetState $GOFILE

package inventory

// AssetStatus enumerates all the possible lifecycle status of asset in collins
type AssetStatus int

const (
	// Incomplete status in collins implies that host is not yet ready for use. It has been powered on and
	// entered in Collins but burn-in is likely being run. This is not yet used in contiv cluster.
	Incomplete AssetStatus = iota
	// New status in collins implies that host has completed the burn-in process and is waiting for an onsite
	// tech to complete physical intake. This is not yet used in contiv cluster.
	New
	// Unallocated status in collins implies that host has completed intake process and is ready for use.
	// In contiv cluster this status is set when the host is discovered for the first time.
	Unallocated
	// Provisioning status in collins implies that host has started provisioning process but has not yet
	// completed it. In contiv cluster this status is set when a host is signalled to be commissioned by the
	// admin or automatically. The configuration for infrastructure is pushed at this status.
	Provisioning
	// Provisioned status in collins implies that Host has finished provisioning and is awaiting final
	// automated verification. This is not yet used in contiv cluster.
	Provisioned
	// Allocated status in collins implies that this asset is in what should likely be considered a production
	// state. In contiv cluster this status is set when the host configuration was successful.
	Allocated
	// Cancelled status in collins implies that asset is no longer needed and is awaiting decommissioning.
	// In contiv cluster this status is set when a host is signalled to be decommissioned by the
	// admin or automatically. The configuration for infrastructure is cleaned up at this status.
	Cancelled
	// Decommissioned status in collins implies that sset has completed the outtake process and can no
	// longer be managed. In contiv cluster this status is set when the host configuration has been cleaned up.
	Decommissioned
	// Maintenance status in collins implies that asset is undergoing some kind of maintenance and should
	// not be considered for production use. In contiv cluster this status is set when a host is signalled
	// for maintenance by the user or automatically. The host's configuration is cleaned up at this state.
	Maintenance
	// Any status can't be set, it is used for binding states to all status' in collins
	Any
)

// AssetState enumerates the custome lifecycle states of an asset in collins
type AssetState int

const (
	// Unknown state denotes the first state before initialization is complete
	Unknown AssetState = iota
	// Discovered state denotes that host is alive in monitoring susbsystem
	Discovered
	// Disappeared state denotes that host has disappeared from monitoring subsystem.
	Disappeared
)
