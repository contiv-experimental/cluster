package manager

import (
	"fmt"
	"io"

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
	return fmt.Sprintf("discoverEvent: addr: %v extra-vars: %v", e.nodeAddrs, e.extraVars)
}

func (e *discoverEvent) process() error {
	// err shouldn't be redefined below
	var err error

	err = e.mgr.checkAndSetActiveJob(
		e.String(),
		e.discoverRunner,
		func(status JobStatus, errRet error) {
			if status == Errored {
				log.Errorf("provisioning discovery job failed. Error: %v", errRet)
			}
		})
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			e.mgr.resetActiveJob()
		}
	}()

	// validate
	existingNodes := []string{}
	for _, addr := range e.nodeAddrs {
		node, err := e.mgr.findNodeByMgmtAddr(addr)
		if err == nil {
			existingNodes = append(existingNodes, fmt.Sprintf("%s:%s", node.Inv.GetTag(), addr))
		}
	}
	if len(existingNodes) > 0 {
		err = errored.Errorf("one or more nodes already exist with the specified management addresses. Existing nodes: %v", existingNodes)
		return err
	}

	// prepare inventory
	if err = e.pepareInventory(); err != nil {
		return err
	}

	// trigger node discovery provisioning
	go e.mgr.runActiveJob()

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
func (e *discoverEvent) discoverRunner(cancelCh CancelChannel, jobLogs io.Writer) error {
	outReader, cancelFunc, errCh := e.mgr.configuration.Configure(e._hosts, e.extraVars)
	if err := logOutputAndReturnStatus(outReader, errCh, cancelCh, cancelFunc, jobLogs); err != nil {
		log.Errorf("discover failed. Error: %s", err)
		return err
	}
	return nil
}
