// +build systemtest

package systemtests

import (
	"strings"

	. "gopkg.in/check.v1"
)

func (s *CliTestSuite) TestClustermRestart(c *C) {
	if strings.Contains(testDataDir, "/collins") {
		c.Skip("skipping clusterm restart test with collins, due to collins issue: https://github.com/tumblr/collins/issues/436")
	}

	nodeName1 := validNodeNames[0]
	nodeName2 := validNodeNames[1]

	// commision the nodes. First node is master, second node is worker
	s.commissionNode(c, nodeName1, s.tbn1)
	s.commissionNode(c, nodeName2, s.tbn2)

	// decommission one node
	s.decommissionNode(c, nodeName2, s.tbn2)

	// restart clusterm
	s.restartClusterm(c, s.tbn1, 30)

	// verify that the last provision state of nodes is restored
	s.checkProvisionStatus(c, s.tbn1, nodeName1, "Allocated")
	s.checkProvisionStatus(c, s.tbn1, nodeName2, "Decommissioned")
}
