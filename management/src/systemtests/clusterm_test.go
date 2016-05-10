// +build systemtest

package systemtests

import (
	"fmt"
	"strings"
	"time"

	. "gopkg.in/check.v1"
)

func (s *SystemTestSuite) TestClustermRestart(c *C) {
	if strings.Contains(testDataDir, "/collins") {
		c.Skip("skipping clusterm restart test with collins, due to collins issue: https://github.com/tumblr/collins/issues/436")
	}

	nodeName1 := validNodeNames[0]
	nodeName2 := validNodeNames[1]

	// commission the nodes. First node is master, second node is worker
	s.commissionNode(c, nodeName1, ansibleMasterGroupName, s.tbn1)
	s.commissionNode(c, nodeName2, ansibleWorkerGroupName, s.tbn2)

	// decommission one node
	s.decommissionNode(c, nodeName2, s.tbn2)

	// restart clusterm
	s.restartClusterm(c, s.tbn1, 30)

	// verify that the last provision state of nodes is restored
	s.checkProvisionStatus(c, s.tbn1, nodeName1, "Allocated")
	s.checkProvisionStatus(c, s.tbn1, nodeName2, "Decommissioned")
}

func (s *SystemTestSuite) TestClustermFailureActiveJob(c *C) {
	nodeName1 := validNodeNames[0]
	nodeName2 := validNodeNames[1]

	// launch commission on a node
	done := make(chan struct{})
	go func() {
		s.commissionNode(c, nodeName1, ansibleMasterGroupName, s.tbn1)
		done <- struct{}{}
	}()

	// try to start another commission job
	time.Sleep(time.Second)
	cmdStr := fmt.Sprintf("clusterctl node commission %s --host-group %s", nodeName2, ansibleWorkerGroupName)
	out, err := s.tbn1.RunCommandWithOutput(cmdStr)
	s.Assert(c, err, NotNil, Commentf("output: %s", out))
	exptStr := ".*there is already an active job.*"
	s.assertMatch(c, exptStr, out)

	// wait for job to finish
	select {
	case <-done:
	case <-time.After(30 * time.Second):
		s.Assert(c, false, Equals, true, Commentf("timeout waiting for job to finish"))
	}

	// try a decommission job once previous job is done
	s.decommissionNode(c, nodeName1, s.tbn2)
}

func (s *SystemTestSuite) TestSerfFailureOnClustermHost(c *C) {
	nodeName1 := validNodeNames[0]
	nodeName2 := validNodeNames[1]

	// make sure test node is visible in inventory
	s.getNodeInfoSuccess(c, nodeName1)

	// stop serf discovery on test node
	s.stopSerf(c, s.tbn1)

	// wait for serf membership to update
	s.waitForSerfMembership(c, s.tbn2, nodeName1, "failed")

	// make sure clusterm is up and running
	s.checkClustermState(c, s.tbn1, "active")
	s.commissionNode(c, nodeName2, ansibleMasterGroupName, s.tbn2)

	// start serf discovery on test node
	s.startSerf(c, s.tbn1)

	// wait for serf membership to update
	s.waitForSerfMembership(c, s.tbn2, nodeName1, "alive")

	// make sure clusterm stays up and running
	s.checkClustermState(c, s.tbn1, "active")
}
