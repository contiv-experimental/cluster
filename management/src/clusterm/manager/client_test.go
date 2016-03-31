// +build unittest

package manager

import (
	"fmt"
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
	nodeName      = "testNode"
	testGetData   = []byte("testdata123")
	testExtraVars = "extraVars"

	failureReturner = func(c *C, expURL *url.URL) http.HandlerFunc {
		return http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				c.Assert(r.URL.Scheme, Equals, expURL.Scheme)
				c.Assert(r.URL.Host, Equals, expURL.Host)
				c.Assert(r.URL.Query(), DeepEquals, expURL.Query())
				http.Error(w, "test failure", http.StatusInternalServerError)
			})
	}

	okReturner = func(c *C, expURL *url.URL) http.HandlerFunc {
		return http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				c.Assert(r.URL.Scheme, Equals, expURL.Scheme)
				c.Assert(r.URL.Host, Equals, expURL.Host)
				c.Assert(r.URL.Query(), DeepEquals, expURL.Query())
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

func (s *managerSuite) TestPostCommissionSuccess(c *C) {
	expURLStr := fmt.Sprintf("http://%s/%s/%s", baseURL, PostNodeCommissionPrefix, nodeName)
	expURL, err := url.Parse(expURLStr)
	c.Assert(err, IsNil)
	httpS, httpC := getHTTPTestClientAndServer(c, okReturner(c, expURL))
	defer httpS.Close()
	clstrC := Client{
		url:   baseURL,
		httpC: httpC,
	}

	err = clstrC.PostNodeCommission(nodeName, "")
	c.Assert(err, IsNil)
}

func (s *managerSuite) TestPostCommissionWithVarsSuccess(c *C) {
	expURLStr := fmt.Sprintf("http://%s/%s/%s?%s=%s",
		baseURL, PostNodeCommissionPrefix, nodeName, ExtraVarsQuery, testExtraVars)
	expURL, err := url.Parse(expURLStr)
	c.Assert(err, IsNil)
	httpS, httpC := getHTTPTestClientAndServer(c, okReturner(c, expURL))
	defer httpS.Close()
	clstrC := Client{
		url:   baseURL,
		httpC: httpC,
	}

	err = clstrC.PostNodeCommission(nodeName, testExtraVars)
	c.Assert(err, IsNil)
}

func (s *managerSuite) TestPostDecommissionSuccess(c *C) {
	expURLStr := fmt.Sprintf("http://%s/%s/%s", baseURL, PostNodeDecommissionPrefix, nodeName)
	expURL, err := url.Parse(expURLStr)
	c.Assert(err, IsNil)
	httpS, httpC := getHTTPTestClientAndServer(c, okReturner(c, expURL))
	defer httpS.Close()
	clstrC := Client{
		url:   baseURL,
		httpC: httpC,
	}

	err = clstrC.PostNodeDecommission(nodeName, "")
	c.Assert(err, IsNil)
}

func (s *managerSuite) TestPostDecommissionWithVarsSuccess(c *C) {
	expURLStr := fmt.Sprintf("http://%s/%s/%s?%s=%s",
		baseURL, PostNodeDecommissionPrefix, nodeName, ExtraVarsQuery, testExtraVars)
	expURL, err := url.Parse(expURLStr)
	c.Assert(err, IsNil)
	httpS, httpC := getHTTPTestClientAndServer(c, okReturner(c, expURL))
	defer httpS.Close()
	clstrC := Client{
		url:   baseURL,
		httpC: httpC,
	}

	err = clstrC.PostNodeDecommission(nodeName, testExtraVars)
	c.Assert(err, IsNil)
}

func (s *managerSuite) TestPostDiscoverSuccess(c *C) {
	expURLStr := fmt.Sprintf("http://%s/%s/%s", baseURL, PostNodeDiscoverPrefix, nodeName)
	expURL, err := url.Parse(expURLStr)
	c.Assert(err, IsNil)
	httpS, httpC := getHTTPTestClientAndServer(c, okReturner(c, expURL))
	defer httpS.Close()
	clstrC := Client{
		url:   baseURL,
		httpC: httpC,
	}

	err = clstrC.PostNodeDiscover(nodeName, "")
	c.Assert(err, IsNil)
}

func (s *managerSuite) TestPostDiscoverWithVarsSuccess(c *C) {
	expURLStr := fmt.Sprintf("http://%s/%s/%s?%s=%s",
		baseURL, PostNodeDiscoverPrefix, nodeName, ExtraVarsQuery, testExtraVars)
	expURL, err := url.Parse(expURLStr)
	c.Assert(err, IsNil)
	httpS, httpC := getHTTPTestClientAndServer(c, okReturner(c, expURL))
	defer httpS.Close()
	clstrC := Client{
		url:   baseURL,
		httpC: httpC,
	}

	err = clstrC.PostNodeDiscover(nodeName, testExtraVars)
	c.Assert(err, IsNil)
}

func (s *managerSuite) TestPostInMaintenance(c *C) {
	expURLStr := fmt.Sprintf("http://%s/%s/%s", baseURL, PostNodeMaintenancePrefix, nodeName)
	expURL, err := url.Parse(expURLStr)
	c.Assert(err, IsNil)
	httpS, httpC := getHTTPTestClientAndServer(c, okReturner(c, expURL))
	defer httpS.Close()
	clstrC := Client{
		url:   baseURL,
		httpC: httpC,
	}

	err = clstrC.PostNodeInMaintenance(nodeName, "")
	c.Assert(err, IsNil)
}

func (s *managerSuite) TestPostInMaintenanceWithVarsSuccess(c *C) {
	expURLStr := fmt.Sprintf("http://%s/%s/%s?%s=%s",
		baseURL, PostNodeMaintenancePrefix, nodeName, ExtraVarsQuery, testExtraVars)
	expURL, err := url.Parse(expURLStr)
	c.Assert(err, IsNil)
	httpS, httpC := getHTTPTestClientAndServer(c, okReturner(c, expURL))
	defer httpS.Close()
	clstrC := Client{
		url:   baseURL,
		httpC: httpC,
	}

	err = clstrC.PostNodeInMaintenance(nodeName, testExtraVars)
	c.Assert(err, IsNil)
}

func (s *managerSuite) TestPostGlobalsWithVarsSuccess(c *C) {
	expURLStr := fmt.Sprintf("http://%s/%s?%s=%s",
		baseURL, PostGlobals, ExtraVarsQuery, testExtraVars)
	expURL, err := url.Parse(expURLStr)
	c.Assert(err, IsNil)
	httpS, httpC := getHTTPTestClientAndServer(c, okReturner(c, expURL))
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
	httpS, httpC := getHTTPTestClientAndServer(c, okReturner(c, expURL))
	defer httpS.Close()
	clstrC := Client{
		url:   baseURL,
		httpC: httpC,
	}

	err = clstrC.PostGlobals("")
	c.Assert(err, IsNil)
}

func (s *managerSuite) TestPostError(c *C) {
	expURLStr := fmt.Sprintf("http://%s/%s/%s", baseURL, PostNodeMaintenancePrefix, nodeName)
	expURL, err := url.Parse(expURLStr)
	c.Assert(err, IsNil)
	httpS, httpC := getHTTPTestClientAndServer(c, failureReturner(c, expURL))
	defer httpS.Close()
	clstrC := Client{
		url:   baseURL,
		httpC: httpC,
	}

	err = clstrC.PostNodeInMaintenance(nodeName, "")
	c.Assert(err, ErrorMatches, ".*test failure\n")
}

func (s *managerSuite) TestGetNodeSuccess(c *C) {
	expURLStr := fmt.Sprintf("http://%s/%s/%s", baseURL, GetNodeInfoPrefix, nodeName)
	expURL, err := url.Parse(expURLStr)
	c.Assert(err, IsNil)
	httpS, httpC := getHTTPTestClientAndServer(c, okGetReturner(c, expURL))
	defer httpS.Close()
	clstrC := Client{
		url:   baseURL,
		httpC: httpC,
	}

	resp, err := clstrC.GetNode(nodeName)
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

	resp, err := clstrC.GetNode(nodeName)
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
	expURLStr := fmt.Sprintf("http://%s/%s/%s", baseURL, GetNodeInfoPrefix, nodeName)
	expURL, err := url.Parse(expURLStr)
	c.Assert(err, IsNil)
	httpS, httpC := getHTTPTestClientAndServer(c, failureReturner(c, expURL))
	defer httpS.Close()
	clstrC := Client{
		url:   baseURL,
		httpC: httpC,
	}

	_, err = clstrC.GetNode(nodeName)
	c.Assert(err, ErrorMatches, ".*test failure\n")
}
