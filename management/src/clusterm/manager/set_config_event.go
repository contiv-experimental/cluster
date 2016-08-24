package manager

import (
	"fmt"
	"io"
	"reflect"

	"github.com/contiv/errored"
)

func configChangeNotPermittedError(config string) error {
	return errored.Errorf("%q configuration can't be changed. Only changes to ansible configuration are allowed.", config)
}

// setConfigEvent triggers the update to global configuration
type setConfigEvent struct {
	mgr    *Manager
	config *Config
}

// newSetConfigEvent creates and returns setConfigEvent
func newSetConfigEvent(mgr *Manager, config *Config) *setConfigEvent {
	return &setConfigEvent{
		mgr:    mgr,
		config: config,
	}
}

func (e *setConfigEvent) String() string {
	return fmt.Sprintf("setConfigEvent: %+v", e.config)
}

func (e *setConfigEvent) process() error {
	// err shouldn't be redefined below
	var err error

	// we set a noop job to ensure that even for the short time this event is
	// run no other job get's enqueued and catches us in middle of things
	err = e.mgr.checkAndSetActiveJob(
		e.String(),
		e.noopRunner,
		func(status JobStatus, errRet error) { return })
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			e.mgr.resetActiveJob()
		}
	}()

	// merge the config with default and validate
	finalConfig := DefaultConfig()
	err = finalConfig.Merge(e.config)
	if err != nil {
		return err
	}
	e.config = finalConfig
	err = e.eventValidate()
	if err != nil {
		return err
	}

	// update manager's config
	e.mgr.config = e.config

	// trigger the noop job
	go e.mgr.runActiveJob()

	return nil
}

func (e *setConfigEvent) eventValidate() error {
	// make sure we are only changing ansible related config.
	// Changes to monitoring, inventory and manager config is not supported

	if !reflect.DeepEqual(e.config.Serf, e.mgr.config.Serf) {
		return configChangeNotPermittedError("serf")
	}
	if !reflect.DeepEqual(e.config.Inventory, e.mgr.config.Inventory) {
		return configChangeNotPermittedError("inventory")
	}
	if !reflect.DeepEqual(e.config.Manager, e.mgr.config.Manager) {
		return configChangeNotPermittedError("manager")
	}

	return nil
}

func (e *setConfigEvent) noopRunner(cancelCh CancelChannel, jobLogs io.Writer) error {
	return nil
}
