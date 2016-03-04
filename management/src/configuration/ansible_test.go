package configuration

import (
	"encoding/json"
	"testing"

	. "gopkg.in/check.v1"
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { TestingT(t) }

type ansibleSuite struct {
}

var _ = Suite(&ansibleSuite{})

func (s *ansibleSuite) TestMergeExtraVarsSuccess(c *C) {
	dst := `{
		"foo": "bar",
		"fooMap": {
			"key1": "val1"
		},
		"keyReplace": "val"
	}`
	src := `{
		"fooMap": {
			"key2": "val2"
		},
		"keyReplace": "valReplace"
	}`
	//XXX: note that below the "fooMap" value is replaced than being merged. Do we
	// need a merge here as well?
	exptd := `{
		"foo": "bar",
		"fooMap": {
			"key2": "val2"
		},
		"keyReplace": "valReplace"
	}`

	out, err := mergeExtraVars(dst, src)
	c.Assert(err, IsNil)
	var (
		outMap   map[string]interface{}
		exptdMap map[string]interface{}
	)
	c.Assert(json.Unmarshal([]byte(out), &outMap), IsNil)
	c.Assert(json.Unmarshal([]byte(exptd), &exptdMap), IsNil)
	c.Assert(outMap, DeepEquals, exptdMap)
}

func (s *ansibleSuite) TestMergeExtraVarsInvalidJSON(c *C) {
	dst := `{
		"foo": 
	}`
	src := `{}`
	out, err := mergeExtraVars(dst, src)
	c.Assert(err, ErrorMatches, "failed to unmarshal dest extra vars.*",
		Commentf("output string: %s", out))

	dst = `{}`
	src = `{
		"foo": 
	}`
	out, err = mergeExtraVars(dst, src)
	c.Assert(err, ErrorMatches, "failed to unmarshal src extra vars.*",
		Commentf("output string: %s", out))
}
