//go:generate stringer -type=EventType $GOFILE

package monitor

// EventType denotes the possible events associated with node monitoring
// viz. discovery and disappearance
type EventType int

const (
	// Discovered is constant for the node discovery event
	Discovered EventType = iota

	// Disappeared is constant for the node disappearance event
	Disappeared
)
