// +build unittest

package manager

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	. "gopkg.in/check.v1"
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { TestingT(t) }

type managerSuite struct {
}

var (
	_             = Suite(&managerSuite{})
	baseURL       = "baseUrl.foo:1234"
	testNodeName  = "testNode"
	testGetData   = []byte("testdata123")
	testExtraVars = "extraVars"
	testReqBody   = APIRequest{
		Nodes: []string{testNodeName},
	}

	testReqHostGroupBody = APIRequest{
		Nodes:     []string{testNodeName},
		HostGroup: ansibleMasterGroupName,
	}

	testReq1HostGroupBody = APIRequest{
		Nodes:     []string{},
		HostGroup: ansibleMasterGroupName,
	}

	testReqDiscoverBody = APIRequest{
		Addrs: []string{testNodeName},
	}

	failureReturner = func(c *C, expURL *url.URL, expBody []byte) http.HandlerFunc {
		return http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				c.Assert(r.URL.Scheme, Equals, expURL.Scheme)
				c.Assert(r.URL.Host, Equals, expURL.Host)
				c.Assert(r.URL.Query(), DeepEquals, expURL.Query())
				body, err := ioutil.ReadAll(r.Body)
				c.Assert(err, IsNil)
				c.Assert(body, DeepEquals, expBody)
				http.Error(w, "test failure", http.StatusInternalServerError)
			})
	}

	okReturner = func(c *C, expURL *url.URL, expBody []byte) http.HandlerFunc {
		return http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				c.Assert(r.URL.Scheme, Equals, expURL.Scheme)
				c.Assert(r.URL.Host, Equals, expURL.Host)
				c.Assert(r.URL.Query(), DeepEquals, expURL.Query())
				body, err := ioutil.ReadAll(r.Body)
				c.Assert(err, IsNil)
				c.Assert(body, DeepEquals, expBody)
				w.WriteHeader(http.StatusOK)
			})
	}

	okGetReturner = func(c *C, expURL *url.URL) http.HandlerFunc {
		return http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				c.Assert(r.URL.Scheme, Equals, expURL.Scheme)
				c.Assert(r.URL.Host, Equals, expURL.Host)
				c.Assert(r.URL.Query(), DeepEquals, expURL.Query())
				w.Write(testGetData)
			})
	}
)

func getHTTPTestClientAndServer(c *C, handler http.HandlerFunc) (*httptest.Server, *http.Client) {
	srvr := httptest.NewServer(handler)

	transport := &http.Transport{
		Proxy: func(r *http.Request) (*url.URL, error) {
			return url.Parse(srvr.URL)
		},
	}
	httpC := &http.Client{Transport: transport}

	return srvr, httpC
}

func (s *managerSuite) TestPostSuccess(c *C) {
	clstrC := Client{
		url: baseURL,
	}

	var reqHostGroupBody bytes.Buffer
	c.Assert(json.NewEncoder(&reqHostGroupBody).Encode(testReq1HostGroupBody), IsNil)

	var testFlags ActionFlags
	tests := map[string]struct {
		expURLStr string
		nodeName  string
		flags     ActionFlags
		exptdBody []byte
		cb        func(name string, flags ActionFlags) error
	}{
		"commission": {
			expURLStr: fmt.Sprintf("http://%s/%s/%s", baseURL, PostNodeCommissionPrefix, testNodeName),
			nodeName:  testNodeName,
			flags:     ActionFlags{"", ansibleMasterGroupName},
			exptdBody: reqHostGroupBody.Bytes(),
			cb:        clstrC.PostNodeCommission,
		},
		"commission-extra-vars": {
			expURLStr: fmt.Sprintf("http://%s/%s/%s?%s=%s",
				baseURL, PostNodeCommissionPrefix, testNodeName, ExtraVarsQuery, testExtraVars),
			nodeName:  testNodeName,
			flags:     ActionFlags{testExtraVars, ansibleMasterGroupName},
			exptdBody: reqHostGroupBody.Bytes(),
			cb:        clstrC.PostNodeCommission,
		},
		"decommission": {
			expURLStr: fmt.Sprintf("http://%s/%s/%s", baseURL, PostNodeDecommissionPrefix, testNodeName),
			nodeName:  testNodeName,
			flags:     testFlags,
			exptdBody: []byte{},
			cb:        clstrC.PostNodeDecommission,
		},
		"decommission-extra-vars": {
			expURLStr: fmt.Sprintf("http://%s/%s/%s?%s=%s",
				baseURL, PostNodeDecommissionPrefix, testNodeName, ExtraVarsQuery, testExtraVars),
			nodeName:  testNodeName,
			flags:     ActionFlags{testExtraVars, ""},
			exptdBody: []byte{},
			cb:        clstrC.PostNodeDecommission,
		},
		"maintenance": {
			expURLStr: fmt.Sprintf("http://%s/%s/%s", baseURL, PostNodeMaintenancePrefix, testNodeName),
			nodeName:  testNodeName,
			flags:     testFlags,
			exptdBody: []byte{},
			cb:        clstrC.PostNodeDecommission,
		},
		"maintenance-extra-vars": {
			expURLStr: fmt.Sprintf("http://%s/%s/%s?%s=%s",
				baseURL, PostNodeMaintenancePrefix, testNodeName, ExtraVarsQuery, testExtraVars),
			nodeName:  testNodeName,
			flags:     ActionFlags{testExtraVars, ""},
			exptdBody: []byte{},
			cb:        clstrC.PostNodeDecommission,
		},
	}
	for testname, test := range tests {
		expURL, err := url.Parse(test.expURLStr)
		c.Assert(err, IsNil, Commentf("test: %s", testname))

		httpS, httpC := getHTTPTestClientAndServer(c, okReturner(c, expURL, test.exptdBody))
		defer httpS.Close()
		clstrC.httpC = httpC
		c.Assert(test.cb(test.nodeName, test.flags), IsNil, Commentf("test: %s", testname))
	}
}

func (s *managerSuite) TestPostMultiNodesSuccess(c *C) {
	clstrC := Client{
		url: baseURL,
	}

	var reqBody bytes.Buffer
	c.Assert(json.NewEncoder(&reqBody).Encode(testReqBody), IsNil)

	var reqHostGroupBody bytes.Buffer
	c.Assert(json.NewEncoder(&reqHostGroupBody).Encode(testReqHostGroupBody), IsNil)

	var reqDiscoverBody bytes.Buffer
	c.Assert(json.NewEncoder(&reqDiscoverBody).Encode(testReqDiscoverBody), IsNil)

	tests := map[string]struct {
		expURLStr string
		nodeNames []string
		flags     ActionFlags
		exptdBody []byte
		cb        func(names []string, flags ActionFlags) error
	}{
		"commission": {
			expURLStr: fmt.Sprintf("http://%s/%s", baseURL, PostNodesCommission),
			nodeNames: []string{testNodeName},
			flags:     ActionFlags{"", "service-master"},
			exptdBody: reqHostGroupBody.Bytes(),
			cb:        clstrC.PostNodesCommission,
		},
		"commission-extra-vars": {
			expURLStr: fmt.Sprintf("http://%s/%s?%s=%s",
				baseURL, PostNodesCommission, ExtraVarsQuery, testExtraVars),
			nodeNames: []string{testNodeName},
			flags:     ActionFlags{testExtraVars, "service-master"},
			exptdBody: reqHostGroupBody.Bytes(),
			cb:        clstrC.PostNodesCommission,
		},
		"decommission": {
			expURLStr: fmt.Sprintf("http://%s/%s", baseURL, PostNodesDecommission),
			nodeNames: []string{testNodeName},
			flags:     ActionFlags{"", ""},
			exptdBody: reqBody.Bytes(),
			cb:        clstrC.PostNodesDecommission,
		},
		"decommission-extra-vars": {
			expURLStr: fmt.Sprintf("http://%s/%s?%s=%s",
				baseURL, PostNodesDecommission, ExtraVarsQuery, testExtraVars),
			nodeNames: []string{testNodeName},
			flags:     ActionFlags{testExtraVars, ""},
			exptdBody: reqBody.Bytes(),
			cb:        clstrC.PostNodesDecommission,
		},
		"maintenance": {
			expURLStr: fmt.Sprintf("http://%s/%s", baseURL, PostNodesMaintenance),
			nodeNames: []string{testNodeName},
			flags:     ActionFlags{"", ""},
			exptdBody: reqBody.Bytes(),
			cb:        clstrC.PostNodesDecommission,
		},
		"maintenance-extra-vars": {
			expURLStr: fmt.Sprintf("http://%s/%s?%s=%s",
				baseURL, PostNodesMaintenance, ExtraVarsQuery, testExtraVars),
			nodeNames: []string{testNodeName},
			flags:     ActionFlags{testExtraVars, ""},
			exptdBody: reqBody.Bytes(),
			cb:        clstrC.PostNodesDecommission,
		},
		"discover": {
			expURLStr: fmt.Sprintf("http://%s/%s", baseURL, PostNodesDiscover),
			nodeNames: []string{testNodeName},
			flags:     ActionFlags{"", ""},
			exptdBody: reqDiscoverBody.Bytes(),
			cb:        clstrC.PostNodesDiscover,
		},
		"discover-extra-vars": {
			expURLStr: fmt.Sprintf("http://%s/%s?%s=%s",
				baseURL, PostNodesDiscover, ExtraVarsQuery, testExtraVars),
			nodeNames: []string{testNodeName},
			flags:     ActionFlags{testExtraVars, ""},
			exptdBody: reqDiscoverBody.Bytes(),
			cb:        clstrC.PostNodesDiscover,
		},
	}
	for testname, test := range tests {
		expURL, err := url.Parse(test.expURLStr)
		c.Assert(err, IsNil, Commentf("test: %s", testname))

		httpS, httpC := getHTTPTestClientAndServer(c, okReturner(c, expURL, test.exptdBody))
		defer httpS.Close()
		clstrC.httpC = httpC
		c.Assert(test.cb(test.nodeNames, test.flags), IsNil, Commentf("test: %s", testname))
	}
}

func (s *managerSuite) TestPostGlobalsWithVarsSuccess(c *C) {
	expURLStr := fmt.Sprintf("http://%s/%s?%s=%s",
		baseURL, PostGlobals, ExtraVarsQuery, testExtraVars)
	expURL, err := url.Parse(expURLStr)
	c.Assert(err, IsNil)
	httpS, httpC := getHTTPTestClientAndServer(c, okReturner(c, expURL, []byte{}))
	defer httpS.Close()
	clstrC := Client{
		url:   baseURL,
		httpC: httpC,
	}

	var flags ActionFlags
	flags.ExtraVars = testExtraVars
	err = clstrC.PostGlobals(flags)
	c.Assert(err, IsNil)
}

func (s *managerSuite) TestPostGlobalsWithEmptyVarsSuccess(c *C) {
	expURLStr := fmt.Sprintf("http://%s/%s", baseURL, PostGlobals)
	expURL, err := url.Parse(expURLStr)
	c.Assert(err, IsNil)
	httpS, httpC := getHTTPTestClientAndServer(c, okReturner(c, expURL, []byte{}))
	defer httpS.Close()
	clstrC := Client{
		url:   baseURL,
		httpC: httpC,
	}

	var flags ActionFlags
	err = clstrC.PostGlobals(flags)
	c.Assert(err, IsNil)
}

func (s *managerSuite) TestPostError(c *C) {
	expURLStr := fmt.Sprintf("http://%s/%s/%s", baseURL, PostNodeMaintenancePrefix, testNodeName)
	expURL, err := url.Parse(expURLStr)
	c.Assert(err, IsNil)
	httpS, httpC := getHTTPTestClientAndServer(c, failureReturner(c, expURL, []byte{}))
	defer httpS.Close()
	clstrC := Client{
		url:   baseURL,
		httpC: httpC,
	}
	var flags ActionFlags
	err = clstrC.PostNodeInMaintenance(testNodeName, flags)
	c.Assert(err, ErrorMatches, ".*test failure\n")

	expURLStr = fmt.Sprintf("http://%s/%s", baseURL, PostNodesMaintenance)
	expURL, err = url.Parse(expURLStr)
	c.Assert(err, IsNil)
	var reqBody bytes.Buffer
	c.Assert(json.NewEncoder(&reqBody).Encode(testReqBody), IsNil)
	httpS, httpC = getHTTPTestClientAndServer(c, failureReturner(c, expURL, reqBody.Bytes()))
	defer httpS.Close()
	clstrC = Client{
		url:   baseURL,
		httpC: httpC,
	}
	err = clstrC.PostNodesInMaintenance([]string{testNodeName}, flags)
	c.Assert(err, ErrorMatches, ".*test failure\n")
}

func (s *managerSuite) TestGetNodeSuccess(c *C) {
	expURLStr := fmt.Sprintf("http://%s/%s/%s", baseURL, GetNodeInfoPrefix, testNodeName)
	expURL, err := url.Parse(expURLStr)
	c.Assert(err, IsNil)
	httpS, httpC := getHTTPTestClientAndServer(c, okGetReturner(c, expURL))
	defer httpS.Close()
	clstrC := Client{
		url:   baseURL,
		httpC: httpC,
	}

	resp, err := clstrC.GetNode(testNodeName)
	c.Assert(err, IsNil)
	c.Assert(resp, DeepEquals, testGetData)
}

func (s *managerSuite) TestGetNodesSuccess(c *C) {
	expURLStr := fmt.Sprintf("http://%s/%s", baseURL, GetNodesInfo)
	expURL, err := url.Parse(expURLStr)
	c.Assert(err, IsNil)
	httpS, httpC := getHTTPTestClientAndServer(c, okGetReturner(c, expURL))
	defer httpS.Close()
	clstrC := Client{
		url:   baseURL,
		httpC: httpC,
	}

	resp, err := clstrC.GetNode(testNodeName)
	c.Assert(err, IsNil)
	c.Assert(resp, DeepEquals, testGetData)
}

func (s *managerSuite) TestGetGlobalsSuccess(c *C) {
	expURLStr := fmt.Sprintf("http://%s/%s", baseURL, GetGlobals)
	expURL, err := url.Parse(expURLStr)
	c.Assert(err, IsNil)
	httpS, httpC := getHTTPTestClientAndServer(c, okGetReturner(c, expURL))
	defer httpS.Close()
	clstrC := Client{
		url:   baseURL,
		httpC: httpC,
	}

	resp, err := clstrC.GetGlobals()
	c.Assert(err, IsNil)
	c.Assert(resp, DeepEquals, testGetData)
}

func (s *managerSuite) TestGetError(c *C) {
	expURLStr := fmt.Sprintf("http://%s/%s/%s", baseURL, GetNodeInfoPrefix, testNodeName)
	expURL, err := url.Parse(expURLStr)
	c.Assert(err, IsNil)
	httpS, httpC := getHTTPTestClientAndServer(c, failureReturner(c, expURL, []byte{}))
	defer httpS.Close()
	clstrC := Client{
		url:   baseURL,
		httpC: httpC,
	}

	_, err = clstrC.GetNode(testNodeName)
	c.Assert(err, ErrorMatches, ".*test failure\n")
}
