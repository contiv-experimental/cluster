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
	"encoding/json"
	"fmt"

	"github.com/contiv/cluster/management/src/collins"
	"github.com/contiv/cluster/management/src/configuration"
	"github.com/contiv/cluster/management/src/inventory"
	"github.com/contiv/cluster/management/src/monitor"
	"github.com/mapuri/serf/client"
)

// Config is the configuration to cluster manager daemon
type Config struct {
	Serf    client.Config                     `json:"serf"`
	Collins collins.Config                    `json:"collins"`
	Ansible configuration.AnsibleSubsysConfig `json:"ansible"`
	Manager struct {
		Addr string `json:"addr"`
	} `json:"manager"`
}

// node is an aggregate structure that contains information about a cluster
// node as seen by cluster management subsystems.
type node struct {
	mInfo monitor.SubsysNode
	iInfo inventory.SubsysAsset
	cInfo configuration.SubsysHost
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
}

// NewManager initializes and returns an instance of the Manager. It returns nil
// if a failure occurs as part of initialization.
func NewManager(config []byte) (*Manager, error) {
	mgrConfig := &Config{}
	if err := json.Unmarshal(config, mgrConfig); err != nil {
		return nil, fmt.Errorf("failed to parse configuration. Error: %s", err)
	}

	m := &Manager{
		monitor:       monitor.NewSerfSubsys(&mgrConfig.Serf),
		configuration: configuration.NewAnsibleSubsys(&mgrConfig.Ansible),
		reqQ:          make(chan event, 100),
		addr:          mgrConfig.Manager.Addr,
		nodes:         make(map[string]*node),
	}

	var err error
	if m.inventory, err = inventory.NewCollinsSubsys(&mgrConfig.Collins); err != nil {
		return nil, err
	}

	if err := m.monitor.RegisterCb(monitor.Discovered, m.enqueueMonitorEvent); err != nil {
		return nil, fmt.Errorf("failed to register node discovery callback. Error: %s", err)
	}

	if err := m.monitor.RegisterCb(monitor.Disappeared, m.enqueueMonitorEvent); err != nil {
		return nil, fmt.Errorf("failed to register node disappearance callback. Error: %s", err)
	}

	return m, nil
}

// Run triggers the manager loops
func (m *Manager) Run(errCh chan error) {

	// start monitor subsystem. It feeds node state monitoring events.
	go m.monitorLoop(errCh)

	// start http server for service REST api endpoints. It feeds api/ux events.
	go m.apiLoop(errCh)

	// start the event loop. It processes the events.
	go m.eventLoop()
}
