// +build systemtest

package systemtests

import (
	"encoding/json"
	"fmt"

	"github.com/imdario/mergo"

	. "gopkg.in/check.v1"
)

func (s *SystemTestSuite) TestSetGetGlobalExtraVarsSuccess(c *C) {
	cmdStr := fmt.Sprintf(`clusterctl global set -e '{\\\"foo\\\":\\\"bar\\\"}'`)
	out, err := s.tbn1.RunCommandWithOutput(cmdStr)
	s.Assert(c, err, IsNil, Commentf("output: %s", out))

	cmdStr = fmt.Sprintf(`clusterctl global get --json`)
	out, err = s.tbn1.RunCommandWithOutput(cmdStr)
	s.Assert(c, err, IsNil, Commentf("output: %s", out))
	exptdOut := `.*"foo":.*"bar".*`
	s.assertMatch(c, exptdOut, out)
}

func (s *SystemTestSuite) TestSetGetGlobalExtraVarsFailureInvalidJSON(c *C) {
	cmdStr := fmt.Sprintf(`clusterctl global set -e '{\\\"foo\\\":}'`)
	out, err := s.tbn1.RunCommandWithOutput(cmdStr)
	s.Assert(c, err, NotNil, Commentf("output: %s", out))
	exptdOut := `.*Request URL: globals.*extra_vars.*should be a valid json.*`
	s.assertMatch(c, exptdOut, out)
}

func (s *SystemTestSuite) TestGetNodeInfoFailureNonExistentNode(c *C) {
	s.getNodeInfoFailureNonExistentNode(c, invalidNodeName)
}

func (s *SystemTestSuite) TestGetNodeInfoSuccess(c *C) {
	s.getNodeInfoSuccess(c, validNodeNames[0])
}

func (s *SystemTestSuite) TestGetNodesInfoSuccess(c *C) {
	cmdStr := `clusterctl nodes get --json`
	out, err := s.tbn1.RunCommandWithOutput(cmdStr)
	s.Assert(c, err, IsNil, Commentf("output: %s", out))
	exptdOut := `.*"monitoring_state":.*`
	s.assertMultiMatch(c, exptdOut, out, 2)
	exptdOut = `.*"inventory_state":.*`
	s.assertMultiMatch(c, exptdOut, out, 2)
	exptdOut = `.*"configuration_state".*`
	s.assertMultiMatch(c, exptdOut, out, 2)
}

func (s *SystemTestSuite) TestSetGetConfigSuccess(c *C) {
	// read current config, to prepare the test config with just ansible changes
	cmdStr := `clusterctl config get --json`
	out, err := s.tbn1.RunCommandWithOutput(cmdStr)
	s.Assert(c, err, IsNil, Commentf("output: %s", out))
	var config map[string]interface{}
	err = json.Unmarshal([]byte(out), &config)
	s.Assert(c, err, IsNil)
	var testConfig map[string]interface{}
	err = json.Unmarshal([]byte(`{
    "ansible": {
        "playbook_location":  "foo"
    }
}`), &testConfig)
	s.Assert(c, err, IsNil)
	err = mergo.MergeWithOverwrite(&config, &testConfig)
	s.Assert(c, err, IsNil)
	configStr, err := json.Marshal(config)
	s.Assert(c, err, IsNil)

	c.Logf("config: %+v", config)
	c.Logf("json: %q", configStr)
	cmdStr = fmt.Sprintf(`clusterctl config set - <<EOF
%s
EOF`, configStr)
	out, err = s.tbn1.RunCommandWithOutput(cmdStr)
	s.Assert(c, err, IsNil, Commentf("output: %s", out))
	cmdStr = `clusterctl config get --json`
	out, err = s.tbn1.RunCommandWithOutput(cmdStr)
	s.Assert(c, err, IsNil, Commentf("output: %s", out))
	exptdOut := `.*"playbook_location":.*"foo".*`
	s.assertMatch(c, exptdOut, out)
}

func (s *SystemTestSuite) TestSetGetConfigFailureNotPermitted(c *C) {
	cmdStr := `clusterctl config set - <<EOF
{
    "ansible": {
        "playbook_location":  "foo"
    },
    "inventory": {
        "collins": {
            "url": "foobar"
        }
    }
}
EOF
`
	out, err := s.tbn1.RunCommandWithOutput(cmdStr)
	s.Assert(c, err, NotNil, Commentf("output: %s", out))
	exptdOut := `.*Request URL: config.*Only changes to ansible configuration are allowed.*`
	s.assertMatch(c, exptdOut, out)
}
