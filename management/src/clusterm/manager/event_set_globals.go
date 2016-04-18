package manager

import "fmt"

// setGlobals triggers the update to global configuration
type setGlobals struct {
	mgr       *Manager
	extraVars string
}

// newSetGlobals creates and returns setGlobals event
func newSetGlobals(mgr *Manager, extraVars string) *setGlobals {
	return &setGlobals{
		mgr:       mgr,
		extraVars: extraVars,
	}
}

func (e *setGlobals) String() string {
	return fmt.Sprintf("setGlobals")
}

func (e *setGlobals) process() error {
	if err := e.mgr.configuration.SetGlobals(e.extraVars); err != nil {
		return err
	}
	return nil
}
