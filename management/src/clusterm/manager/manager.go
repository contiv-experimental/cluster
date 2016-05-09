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
//   - TBD

package manager

import (
	"github.com/contiv/cluster/management/src/boltdb"
	"github.com/contiv/cluster/management/src/collins"
	"github.com/contiv/cluster/management/src/configuration"
	"github.com/contiv/cluster/management/src/inventory"
	boltdbinv "github.com/contiv/cluster/management/src/inventory/boltdb"
	collinsinv "github.com/contiv/cluster/management/src/inventory/collins"
	"github.com/contiv/cluster/management/src/monitor"
	"github.com/contiv/errored"
	"github.com/mapuri/serf/client"
)

type clustermConfig struct {
	Addr string `json:"addr"`
}

type inventorySubsysConfig struct {
	Collins *collins.Config `json:"collins,omit-empty"`
	BoltDB  *boltdb.Config  `json:"boltdb,omit-empty"`
}

// Config is the configuration to cluster manager daemon
type Config struct {
	Serf      client.Config                     `json:"serf"`
	Inventory inventorySubsysConfig             `json:"inventory"`
	Ansible   configuration.AnsibleSubsysConfig `json:"ansible"`
	Manager   clustermConfig                    `json:"manager"`
}

// DefaultConfig returns the default configuration values for the cluster manager
// and it's sub-systems
func DefaultConfig() *Config {
	return &Config{
		Serf: client.Config{
			Addr: "127.0.0.1:7373",
		},
		Inventory: inventorySubsysConfig{
			BoltDB:  nil,
			Collins: nil,
		},
		Ansible: configuration.AnsibleSubsysConfig{
			ConfigurePlaybook: "site.yml",
			CleanupPlaybook:   "cleanup.yml",
			UpgradePlaybook:   "rolling-upgrade.yml",
			PlaybookLocation:  "/vagrant/vendor/ansible",
			User:              "vagrant",
			PrivKeyFile:       "/vagrant/management/src/demo/files/insecure_private_key",
		},
		Manager: clustermConfig{
			Addr: "0.0.0.0:9007",
		},
	}
}

// node is an aggregate structure that contains information about a cluster
// node as seen by cluster management subsystems.
type node struct {
	Mon monitor.SubsysNode       `json:"monitoring-state"`
	Inv inventory.SubsysAsset    `json:"inventory-state"`
	Cfg configuration.SubsysHost `json:"configuration-state"`
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
}

// NewManager initializes and returns an instance of the Manager. It returns nil
// if a failure occurs as part of initialization.
func NewManager(config *Config) (*Manager, error) {
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
func (m *Manager) Run(errCh chan error) {

	// start http server for service REST api endpoints. It feeds api/ux events.
	go m.apiLoop(errCh)

	// start monitor subsystem. It feeds node state monitoring events.
	// It needs to be started after api loop as monitor subsystem post events through API endpoints
	go m.monitorLoop(errCh)

	// start the event loop. It processes the events.
	go m.eventLoop()
}
