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

	node, err := e.mgr.findNodeByMgmtAddr(e.nodeAddrs[0])
	if err == nil {
		return errored.Errorf("a node %q already exists with the management address %q",
			node.Inv.GetTag(), e.nodeAddrs[0])
	}

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

// discoverRunner is the job runner that runs configuration plabooks on one or more nodes
// It adds the node(s) to contiv-node hostgroup
func (e *discoverEvent) discoverRunner(cancelCh CancelChannel) error {
	// reset active job status once done
	defer func() { e.mgr.activeJob = nil }()

	// create a temporary ansible host config to provision the host in discover host-group
	hostCfg := configuration.NewAnsibleHost("node1", e.nodeAddrs[0],
		ansibleDiscoverGroupName, map[string]string{
			ansibleNodeNameHostVar: "node1",
			ansibleNodeAddrHostVar: e.nodeAddrs[0],
		})

	outReader, cancelFunc, errCh := e.mgr.configuration.Configure(
		configuration.SubsysHosts([]*configuration.AnsibleHost{hostCfg}), e.extraVars)
	if err := logOutputAndReturnStatus(outReader, errCh, cancelCh, cancelFunc); err != nil {
		log.Errorf("discover failed. Error: %s", err)
		return err
	}
	return nil
}
