// +build systemtest

package systemtests

import (
	"fmt"
	"time"

	. "gopkg.in/check.v1"
)

func (s *SystemTestSuite) TestJobStatusSuccess(c *C) {
	nodeName := validNodeNames[0]

	// launch commission on a node
	done := make(chan struct{})
	go func() {
		s.commissionNode(c, nodeName, ansibleMasterGroupName, s.tbn1)
		done <- struct{}{}
	}()

	// try to fetch the job logs
	time.Sleep(time.Second)
	cmdStr := fmt.Sprintf("clusterctl job get active")
	out, err := s.tbn1.RunCommandWithOutput(cmdStr)
	s.Assert(c, err, IsNil, Commentf("output: %s", out))

	// wait for job to finish
	select {
	case <-done:
	case <-time.After(30 * time.Second):
		s.Assert(c, false, Equals, true, Commentf("timeout waiting for job to finish"))
	}

	// try to fetch the job logs once it is done
	time.Sleep(time.Second)
	cmdStr = fmt.Sprintf("clusterctl job get last")
	out, err = s.tbn1.RunCommandWithOutput(cmdStr)
	s.Assert(c, err, IsNil, Commentf("output: %s", out))
}

func (s *SystemTestSuite) TestJobStatusFailureNonExistentJob(c *C) {
	cmdStr := fmt.Sprintf("clusterctl job get active")
	out, err := s.tbn1.RunCommandWithOutput(cmdStr)
	s.Assert(c, err, NotNil, Commentf("output: %s", out))
	exptdOut := fmt.Sprintf(`.*info for.*%s.*job doesn't exist.*`, "active")
	s.assertMatch(c, exptdOut, out)

	cmdStr = fmt.Sprintf("clusterctl job get last")
	out, err = s.tbn1.RunCommandWithOutput(cmdStr)
	s.Assert(c, err, NotNil, Commentf("output: %s", out))
	exptdOut = fmt.Sprintf(`.*info for.*%s.*job doesn't exist.*`, "last")
	s.assertMatch(c, exptdOut, out)
}

func (s *SystemTestSuite) TestJobStatusFailureInvalidJobLabel(c *C) {
	cmdStr := fmt.Sprintf("clusterctl job get foo")
	out, err := s.tbn1.RunCommandWithOutput(cmdStr)
	s.Assert(c, err, NotNil, Commentf("output: %s", out))
	exptdOut := fmt.Sprintf(`.*Invalid or empty job label specified:.*%s.*`, "foo")
	s.assertMatch(c, exptdOut, out)
}
