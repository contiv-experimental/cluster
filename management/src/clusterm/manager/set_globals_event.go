package manager

import "fmt"

// setGlobalsEvent triggers the update to global configuration
type setGlobalsEvent struct {
	mgr       *Manager
	extraVars string
}

// newSetGlobalsEvent creates and returns setGlobalsEvent
func newSetGlobalsEvent(mgr *Manager, extraVars string) *setGlobalsEvent {
	return &setGlobalsEvent{
		mgr:       mgr,
		extraVars: extraVars,
	}
}

func (e *setGlobalsEvent) String() string {
	return fmt.Sprintf("setGlobalsEvent: %s", e.extraVars)
}

func (e *setGlobalsEvent) process() error {
	if err := e.mgr.configuration.SetGlobals(e.extraVars); err != nil {
		return err
	}
	return nil
}
