// Parse the configuration and start the cluster manager service:
// Implement a basic event driven system that acts as follows:
// - get node discovery events and add to collins inventory (with state = new and status = discovered)
// - get node state change events (i.e. status != discovered) and trigger configuration management as below:
//   - status == added:
//   - status == deleted:
//   - status == down:
// - get user driven configuration push and trigger cluster wide configuration management as below:
//   - upgrade:
// - handle configuration changes

package manager

import (
	"github.com/contiv/cluster/management/src/boltdb"
	"github.com/contiv/cluster/management/src/configuration"
	"github.com/contiv/cluster/management/src/inventory"
	boltdbinv "github.com/contiv/cluster/management/src/inventory/boltdb"
	collinsinv "github.com/contiv/cluster/management/src/inventory/collins"
	"github.com/contiv/cluster/management/src/monitor"
	"github.com/contiv/errored"
	"golang.org/x/net/context"
	"golang.org/x/sync/errgroup"
)

// node is an aggregate structure that contains information about a cluster
// node as seen by cluster management subsystems.
type node struct {
	Mon monitor.SubsysNode       `json:"monitoring_state"`
	Inv inventory.SubsysAsset    `json:"inventory_state"`
	Cfg configuration.SubsysHost `json:"configuration_state"`
}

// Manager integrates the cluster infra services like node discovery, inventory
// and configuation management.
type Manager struct {
	inventory     inventory.Subsys
	configuration configuration.Subsys
	monitor       monitor.Subsys
	reqQ          chan event
	addr          string
	nodes         map[string]*node
	activeJob     *Job // there can be only one active job at a time
	lastJob       *Job
	config        *Config
	configFile    string // file containing clusterm config, when clusterm is started with a config file
}

// NewManager initializes and returns an instance of the Manager. It returns nil
// if a failure occurs as part of initialization.
func NewManager(config *Config, configFile string) (*Manager, error) {
	if config == nil {
		return nil, errored.Errorf("nil config passed")
	}

	var err error
	config.Ansible.ExtraVariables, err = validateAndSanitizeEmptyExtraVars(
		"ansible.ExtraVariables configuration", config.Ansible.ExtraVariables)
	if err != nil {
		return nil, err
	}

	m := &Manager{
		monitor:       monitor.NewSerfSubsys(&config.Serf),
		configuration: configuration.NewAnsibleSubsys(&config.Ansible),
		reqQ:          make(chan event, 100),
		addr:          config.Manager.Addr,
		nodes:         make(map[string]*node),
		config:        config,
		configFile:    configFile,
	}
	// We give priority to boltdb inventory if both are set in config
	if config.Inventory.BoltDB != nil {
		if m.inventory, err = boltdbinv.NewBoltdbSubsys(*config.Inventory.BoltDB); err != nil {
			return nil, err
		}
	} else if config.Inventory.Collins != nil {
		if m.inventory, err = collinsinv.NewCollinsSubsys(*config.Inventory.Collins); err != nil {
			return nil, err
		}
	} else {
		// if no inventory config was provided then we default to boltDb
		if m.inventory, err = boltdbinv.NewBoltdbSubsys(boltdb.DefaultConfig()); err != nil {
			return nil, err
		}
	}

	if err := m.monitor.RegisterCb(monitor.Discovered, m.enqueueMonitorEvent); err != nil {
		return nil, errored.Errorf("failed to register node discovery callback. Error: %s", err)
	}

	if err := m.monitor.RegisterCb(monitor.Disappeared, m.enqueueMonitorEvent); err != nil {
		return nil, errored.Errorf("failed to register node disappearance callback. Error: %s", err)
	}

	return m, nil
}

// Run triggers the manager loops
func (m *Manager) Run() error {

	eg, _ := errgroup.WithContext(context.Background())

	// start http server for servicing REST api endpoints. It feeds api/ux events.
	apiServingCh := make(chan struct{}, 1)
	eg.Go(func() error { return m.apiLoop(apiServingCh) })

	// start monitor subsystem. It feeds node state monitoring events.
	// It needs to be started after api loop as monitor subsystem post events through API endpoints.
	// Additionally, we wait for api loop to signal that it has setup socket to receive requests
	<-apiServingCh
	eg.Go(m.monitorLoop)

	// start signal handler loop.
	// It needs to be started after api loop as signal handler posts events through API endpoints.
	eg.Go(
		func() error {
			m.signalLoop()
			return nil
		})

	// start the event loop. It processes the events.
	eg.Go(
		func() error {
			m.eventLoop()
			return nil
		})

	return eg.Wait()
}
