package manager

import (
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/contiv/cluster/management/src/configuration"
	"github.com/contiv/errored"
)

// discoverEvent triggers the node discovery workflow
type discoverEvent struct {
	mgr       *Manager
	nodeAddrs []string
	extraVars string

	_hosts configuration.SubsysHosts
}

// newDiscoverEvent creates and returns discoverEvent
func newDiscoverEvent(mgr *Manager, nodeAddrs []string, extraVars string) *discoverEvent {
	return &discoverEvent{
		mgr:       mgr,
		nodeAddrs: nodeAddrs,
		extraVars: extraVars,
	}
}

func (e *discoverEvent) String() string {
	return fmt.Sprintf("discoverEvent: %v", e.nodeAddrs)
}

func (e *discoverEvent) process() error {
	if e.mgr.activeJob != nil {
		return errActiveJob(e.mgr.activeJob.String())
	}

	// validate
	existingNodes := []string{}
	for _, addr := range e.nodeAddrs {
		node, err := e.mgr.findNodeByMgmtAddr(addr)
		if err == nil {
			existingNodes = append(existingNodes, fmt.Sprintf("%s:%s", node.Inv.GetTag(), addr))
		}
	}
	if len(existingNodes) > 0 {
		return errored.Errorf("one or more nodes already exist with the specified management addresses. Existing nodes: %v", existingNodes)
	}

	// prepare inventory
	if err := e.pepareInventory(); err != nil {
		return err
	}

	// trigger node discovery provisioning
	e.mgr.activeJob = NewJob(
		e.discoverRunner,
		func(status JobStatus, errRet error) {
			if status == Errored {
				log.Errorf("provisioning discovery job failed. Error: %v", errRet)
			}
		})
	go e.mgr.activeJob.Run()
	return nil
}

// pepareInventory prepares the inventory
func (e *discoverEvent) pepareInventory() error {
	hosts := []*configuration.AnsibleHost{}
	for i, addr := range e.nodeAddrs {
		invName := fmt.Sprintf("node%d", i+1)
		hosts = append(hosts, configuration.NewAnsibleHost(
			invName, addr, ansibleDiscoverGroupName,
			map[string]string{
				ansibleNodeNameHostVar: invName,
				ansibleNodeAddrHostVar: addr,
			}))
	}
	e._hosts = hosts

	return nil
}

// discoverRunner is the job runner that runs configuration plabooks on one or more nodes
// It adds the node(s) to contiv-node hostgroup
func (e *discoverEvent) discoverRunner(cancelCh CancelChannel) error {
	// reset active job status once done
	defer func() { e.mgr.activeJob = nil }()

	outReader, cancelFunc, errCh := e.mgr.configuration.Configure(e._hosts, e.extraVars)
	if err := logOutputAndReturnStatus(outReader, errCh, cancelCh, cancelFunc); err != nil {
		log.Errorf("discover failed. Error: %s", err)
		return err
	}
	return nil
}
