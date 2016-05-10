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
	_                = Suite(&managerSuite{})
	baseURL          = "baseUrl.foo:1234"
	testNodeName     = "testNode"
	testGetData      = []byte("testdata123")
	testExtraVars    = "extraVars"
	testJobLabel     = "testjob"
	testReqNodesBody = APIRequest{
		Nodes: []string{testNodeName},
	}
	testReqEmptyBody = APIRequest{}

	testReqExtraVarsBody = APIRequest{
		ExtraVars: testExtraVars,
	}

	testReqHostGroupBody = APIRequest{
		HostGroup: ansibleMasterGroupName,
	}

	testReqNodesExtraVarsBody = APIRequest{
		Nodes:     []string{testNodeName},
		ExtraVars: testExtraVars,
	}

	testReqNodesHostGroupBody = APIRequest{
		Nodes:     []string{testNodeName},
		HostGroup: ansibleMasterGroupName,
	}

	testReqHostGroupExtraVarsBody = APIRequest{
		HostGroup: ansibleMasterGroupName,
		ExtraVars: testExtraVars,
	}

	testReqNodesHostGroupExtraVarsBody = APIRequest{
		Nodes:     []string{testNodeName},
		HostGroup: ansibleMasterGroupName,
		ExtraVars: testExtraVars,
	}

	testReqDiscoverBody = APIRequest{
		Addrs: []string{testNodeName},
	}

	testReqDiscoverExtraVarsBody = APIRequest{
		Addrs:     []string{testNodeName},
		ExtraVars: testExtraVars,
	}

	failureReturner = func(c *C, expURL *url.URL, expBody []byte) http.HandlerFunc {
		return http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				c.Assert(r.URL.Scheme, Equals, expURL.Scheme)
				c.Assert(r.URL.Host, Equals, expURL.Host)
				c.Assert(r.URL.Query(), DeepEquals, expURL.Query())
				body, err := ioutil.ReadAll(r.Body)
				c.Assert(err, IsNil)
				c.Assert(string(body), Equals, string(expBody))
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
				c.Assert(string(body), Equals, string(expBody))
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
	c.Assert(json.NewEncoder(&reqHostGroupBody).Encode(testReqHostGroupBody), IsNil)

	var reqHostGroupExtraVarsBody bytes.Buffer
	c.Assert(json.NewEncoder(&reqHostGroupExtraVarsBody).Encode(testReqHostGroupExtraVarsBody), IsNil)

	var reqExtraVarsBody bytes.Buffer
	c.Assert(json.NewEncoder(&reqExtraVarsBody).Encode(testReqExtraVarsBody), IsNil)

	var reqEmptyBody bytes.Buffer
	c.Assert(json.NewEncoder(&reqEmptyBody).Encode(testReqEmptyBody), IsNil)

	testsCommission := map[string]struct {
		expURLStr string
		nodeName  string
		extraVars string
		hostGroup string
		exptdBody []byte
		cb        func(name, extraVars, hostGroup string) error
	}{
		"commission": {
			expURLStr: fmt.Sprintf("http://%s/%s/%s", baseURL, PostNodeCommissionPrefix, testNodeName),
			nodeName:  testNodeName,
			extraVars: "",
			hostGroup: "",
			exptdBody: reqEmptyBody.Bytes(),
			cb:        clstrC.PostNodeCommission,
		},
		"commission-extra-vars": {
			expURLStr: fmt.Sprintf("http://%s/%s/%s", baseURL, PostNodeCommissionPrefix, testNodeName),
			nodeName:  testNodeName,
			extraVars: testExtraVars,
			hostGroup: "",
			exptdBody: reqExtraVarsBody.Bytes(),
			cb:        clstrC.PostNodeCommission,
		},
		"commission-host-group": {
			expURLStr: fmt.Sprintf("http://%s/%s/%s", baseURL, PostNodeCommissionPrefix, testNodeName),
			nodeName:  testNodeName,
			extraVars: "",
			hostGroup: ansibleMasterGroupName,
			exptdBody: reqHostGroupBody.Bytes(),
			cb:        clstrC.PostNodeCommission,
		},
		"commission-extra-vars-host-group": {
			expURLStr: fmt.Sprintf("http://%s/%s/%s", baseURL, PostNodeCommissionPrefix, testNodeName),
			nodeName:  testNodeName,
			extraVars: testExtraVars,
			hostGroup: ansibleMasterGroupName,
			exptdBody: reqHostGroupExtraVarsBody.Bytes(),
			cb:        clstrC.PostNodeCommission,
		},
	}
	for testname, test := range testsCommission {
		expURL, err := url.Parse(test.expURLStr)
		c.Assert(err, IsNil, Commentf("test: %s", testname))

		httpS, httpC := getHTTPTestClientAndServer(c, okReturner(c, expURL, test.exptdBody))
		defer httpS.Close()
		clstrC.httpC = httpC
		c.Assert(test.cb(test.nodeName, test.extraVars, test.hostGroup), IsNil, Commentf("test: %s", testname))
	}

	tests := map[string]struct {
		expURLStr string
		nodeName  string
		extraVars string
		exptdBody []byte
		cb        func(name string, extraVars string) error
	}{
		"decommission": {
			expURLStr: fmt.Sprintf("http://%s/%s/%s", baseURL, PostNodeDecommissionPrefix, testNodeName),
			nodeName:  testNodeName,
			extraVars: "",
			exptdBody: reqEmptyBody.Bytes(),
			cb:        clstrC.PostNodeDecommission,
		},
		"decommission-extra-vars": {
			expURLStr: fmt.Sprintf("http://%s/%s/%s", baseURL, PostNodeDecommissionPrefix, testNodeName),
			nodeName:  testNodeName,
			extraVars: testExtraVars,
			exptdBody: reqExtraVarsBody.Bytes(),
			cb:        clstrC.PostNodeDecommission,
		},
		"maintenance": {
			expURLStr: fmt.Sprintf("http://%s/%s/%s", baseURL, PostNodeMaintenancePrefix, testNodeName),
			nodeName:  testNodeName,
			extraVars: "",
			exptdBody: reqEmptyBody.Bytes(),
			cb:        clstrC.PostNodeDecommission,
		},
		"maintenance-extra-vars": {
			expURLStr: fmt.Sprintf("http://%s/%s/%s", baseURL, PostNodeMaintenancePrefix, testNodeName),
			nodeName:  testNodeName,
			extraVars: testExtraVars,
			exptdBody: reqExtraVarsBody.Bytes(),
			cb:        clstrC.PostNodeDecommission,
		},
	}
	for testname, test := range tests {
		expURL, err := url.Parse(test.expURLStr)
		c.Assert(err, IsNil, Commentf("test: %s", testname))

		httpS, httpC := getHTTPTestClientAndServer(c, okReturner(c, expURL, test.exptdBody))
		defer httpS.Close()
		clstrC.httpC = httpC
		c.Assert(test.cb(test.nodeName, test.extraVars), IsNil, Commentf("test: %s", testname))
	}
}

func (s *managerSuite) TestPostMultiNodesSuccess(c *C) {
	clstrC := Client{
		url: baseURL,
	}

	var reqBody bytes.Buffer
	c.Assert(json.NewEncoder(&reqBody).Encode(testReqNodesBody), IsNil)

	var reqNodesExtraVarsBody bytes.Buffer
	c.Assert(json.NewEncoder(&reqNodesExtraVarsBody).Encode(testReqNodesExtraVarsBody), IsNil)

	var reqNodesHostGroupBody bytes.Buffer
	c.Assert(json.NewEncoder(&reqNodesHostGroupBody).Encode(testReqNodesHostGroupBody), IsNil)

	var reqNodesHostGroupExtraVarsBody bytes.Buffer
	c.Assert(json.NewEncoder(&reqNodesHostGroupExtraVarsBody).Encode(testReqNodesHostGroupExtraVarsBody), IsNil)

	var reqDiscoverBody bytes.Buffer
	c.Assert(json.NewEncoder(&reqDiscoverBody).Encode(testReqDiscoverBody), IsNil)

	var reqDiscoverExtraVarsBody bytes.Buffer
	c.Assert(json.NewEncoder(&reqDiscoverExtraVarsBody).Encode(testReqDiscoverExtraVarsBody), IsNil)

	testsCommission := map[string]struct {
		expURLStr string
		nodeNames []string
		extraVars string
		hostGroup string
		exptdBody []byte
		cb        func(names []string, extraVars string, hostGroup string) error
	}{
		"commission": {
			expURLStr: fmt.Sprintf("http://%s/%s", baseURL, PostNodesCommission),
			nodeNames: []string{testNodeName},
			extraVars: "",
			hostGroup: "",
			exptdBody: reqBody.Bytes(),
			cb:        clstrC.PostNodesCommission,
		},
		"commission-extra-vars": {
			expURLStr: fmt.Sprintf("http://%s/%s", baseURL, PostNodesCommission),
			nodeNames: []string{testNodeName},
			extraVars: testExtraVars,
			hostGroup: "",
			exptdBody: reqNodesExtraVarsBody.Bytes(),
			cb:        clstrC.PostNodesCommission,
		},

		"commission-host-group": {
			expURLStr: fmt.Sprintf("http://%s/%s", baseURL, PostNodesCommission),
			nodeNames: []string{testNodeName},
			extraVars: "",
			hostGroup: ansibleMasterGroupName,
			exptdBody: reqNodesHostGroupBody.Bytes(),
			cb:        clstrC.PostNodesCommission,
		},
		"commission-extra-vars-host-group": {
			expURLStr: fmt.Sprintf("http://%s/%s", baseURL, PostNodesCommission),
			nodeNames: []string{testNodeName},
			extraVars: testExtraVars,
			hostGroup: ansibleMasterGroupName,
			exptdBody: reqNodesHostGroupExtraVarsBody.Bytes(),
			cb:        clstrC.PostNodesCommission,
		},
	}
	for testname, test := range testsCommission {
		expURL, err := url.Parse(test.expURLStr)
		c.Assert(err, IsNil, Commentf("test: %s", testname))

		httpS, httpC := getHTTPTestClientAndServer(c, okReturner(c, expURL, test.exptdBody))
		defer httpS.Close()
		clstrC.httpC = httpC
		c.Assert(test.cb(test.nodeNames, test.extraVars, test.hostGroup), IsNil, Commentf("test: %s", testname))
	}

	tests := map[string]struct {
		expURLStr string
		nodeNames []string
		extraVars string
		exptdBody []byte
		cb        func(names []string, extraVars string) error
	}{
		"decommission": {
			expURLStr: fmt.Sprintf("http://%s/%s", baseURL, PostNodesDecommission),
			nodeNames: []string{testNodeName},
			extraVars: "",
			exptdBody: reqBody.Bytes(),
			cb:        clstrC.PostNodesDecommission,
		},
		"decommission-extra-vars": {
			expURLStr: fmt.Sprintf("http://%s/%s", baseURL, PostNodesDecommission),
			nodeNames: []string{testNodeName},
			extraVars: testExtraVars,
			exptdBody: reqNodesExtraVarsBody.Bytes(),
			cb:        clstrC.PostNodesDecommission,
		},
		"maintenance": {
			expURLStr: fmt.Sprintf("http://%s/%s", baseURL, PostNodesMaintenance),
			nodeNames: []string{testNodeName},
			extraVars: "",
			exptdBody: reqBody.Bytes(),
			cb:        clstrC.PostNodesDecommission,
		},
		"maintenance-extra-vars": {
			expURLStr: fmt.Sprintf("http://%s/%s", baseURL, PostNodesMaintenance),
			nodeNames: []string{testNodeName},
			extraVars: testExtraVars,
			exptdBody: reqNodesExtraVarsBody.Bytes(),
			cb:        clstrC.PostNodesDecommission,
		},
		"discover": {
			expURLStr: fmt.Sprintf("http://%s/%s", baseURL, PostNodesDiscover),
			nodeNames: []string{testNodeName},
			extraVars: "",
			exptdBody: reqDiscoverBody.Bytes(),
			cb:        clstrC.PostNodesDiscover,
		},
		"discover-extra-vars": {
			expURLStr: fmt.Sprintf("http://%s/%s", baseURL, PostNodesDiscover),
			nodeNames: []string{testNodeName},
			extraVars: testExtraVars,
			exptdBody: reqDiscoverExtraVarsBody.Bytes(),
			cb:        clstrC.PostNodesDiscover,
		},
	}
	for testname, test := range tests {
		expURL, err := url.Parse(test.expURLStr)
		c.Assert(err, IsNil, Commentf("test: %s", testname))

		httpS, httpC := getHTTPTestClientAndServer(c, okReturner(c, expURL, test.exptdBody))
		defer httpS.Close()
		clstrC.httpC = httpC
		c.Assert(test.cb(test.nodeNames, test.extraVars), IsNil, Commentf("test: %s", testname))
	}
}

func (s *managerSuite) TestPostGlobalsWithVarsSuccess(c *C) {
	expURLStr := fmt.Sprintf("http://%s/%s", baseURL, PostGlobals)
	expURL, err := url.Parse(expURLStr)
	c.Assert(err, IsNil)
	var reqExtraVarsBody bytes.Buffer
	c.Assert(json.NewEncoder(&reqExtraVarsBody).Encode(testReqExtraVarsBody), IsNil)
	httpS, httpC := getHTTPTestClientAndServer(c, okReturner(c, expURL, reqExtraVarsBody.Bytes()))
	defer httpS.Close()
	clstrC := Client{
		url:   baseURL,
		httpC: httpC,
	}

	err = clstrC.PostGlobals(testExtraVars)
	c.Assert(err, IsNil)
}

func (s *managerSuite) TestPostGlobalsWithEmptyVarsSuccess(c *C) {
	expURLStr := fmt.Sprintf("http://%s/%s", baseURL, PostGlobals)
	expURL, err := url.Parse(expURLStr)
	c.Assert(err, IsNil)
	var reqEmptyBody bytes.Buffer
	c.Assert(json.NewEncoder(&reqEmptyBody).Encode(testReqEmptyBody), IsNil)
	httpS, httpC := getHTTPTestClientAndServer(c, okReturner(c, expURL, reqEmptyBody.Bytes()))
	defer httpS.Close()
	clstrC := Client{
		url:   baseURL,
		httpC: httpC,
	}

	err = clstrC.PostGlobals("")
	c.Assert(err, IsNil)
}

func (s *managerSuite) TestPostMonitorEvent(c *C) {
	expURLStr := fmt.Sprintf("http://%s/%s", baseURL, PostMonitorEvent)
	expURL, err := url.Parse(expURLStr)
	c.Assert(err, IsNil)
	testEvent := "fooEvent"
	testNode := MonitorNode{
		Label:    "foo",
		Serial:   "bar",
		MgmtAddr: "1234",
	}
	reqBody := &APIRequest{
		Event: MonitorEvent{
			Name:  testEvent,
			Nodes: []MonitorNode{testNode},
		},
	}
	var reqJSON bytes.Buffer
	c.Assert(json.NewEncoder(&reqJSON).Encode(reqBody), IsNil)
	httpS, httpC := getHTTPTestClientAndServer(c, okReturner(c, expURL, reqJSON.Bytes()))
	defer httpS.Close()
	clstrC := Client{
		url:   baseURL,
		httpC: httpC,
	}

	err = clstrC.PostMonitorEvent(testEvent, []MonitorNode{testNode})
	c.Assert(err, IsNil)
}

func (s *managerSuite) TestPostError(c *C) {
	var reqEmptyBody bytes.Buffer
	c.Assert(json.NewEncoder(&reqEmptyBody).Encode(testReqEmptyBody), IsNil)
	expURLStr := fmt.Sprintf("http://%s/%s/%s", baseURL, PostNodeMaintenancePrefix, testNodeName)
	expURL, err := url.Parse(expURLStr)
	c.Assert(err, IsNil)
	httpS, httpC := getHTTPTestClientAndServer(c, failureReturner(c, expURL, reqEmptyBody.Bytes()))
	defer httpS.Close()
	clstrC := Client{
		url:   baseURL,
		httpC: httpC,
	}
	err = clstrC.PostNodeInMaintenance(testNodeName, "")
	c.Assert(err, ErrorMatches, ".*test failure\n")

	expURLStr = fmt.Sprintf("http://%s/%s", baseURL, PostNodesMaintenance)
	expURL, err = url.Parse(expURLStr)
	c.Assert(err, IsNil)
	var reqBody bytes.Buffer
	c.Assert(json.NewEncoder(&reqBody).Encode(testReqNodesBody), IsNil)
	httpS, httpC = getHTTPTestClientAndServer(c, failureReturner(c, expURL, reqBody.Bytes()))
	defer httpS.Close()
	clstrC = Client{
		url:   baseURL,
		httpC: httpC,
	}
	err = clstrC.PostNodesInMaintenance([]string{testNodeName}, "")
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

func (s *managerSuite) TestGetJobSuccess(c *C) {
	expURLStr := fmt.Sprintf("http://%s/%s/%s", baseURL, GetJobPrefix, testJobLabel)
	expURL, err := url.Parse(expURLStr)
	c.Assert(err, IsNil)
	httpS, httpC := getHTTPTestClientAndServer(c, okGetReturner(c, expURL))
	defer httpS.Close()
	clstrC := Client{
		url:   baseURL,
		httpC: httpC,
	}

	resp, err := clstrC.GetJob(testJobLabel)
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
