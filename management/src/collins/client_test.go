// +build unittest

package collins

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	. "gopkg.in/check.v1"
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { TestingT(t) }

type collinsSuite struct {
}

var _ = Suite(&collinsSuite{})

var failureReturner = http.HandlerFunc(
	func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "test failure", http.StatusInternalServerError)
	})

var alreadyExistsReturner = http.HandlerFunc(
	func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "test already exists response", http.StatusConflict)
	})

var okReturner = http.HandlerFunc(
	func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

func getHTTPTestClientAndServer(handler http.HandlerFunc) (*httptest.Server, *http.Client) {
	srvr := httptest.NewServer(handler)

	transport := &http.Transport{
		Proxy: func(r *http.Request) (*url.URL, error) {
			return url.Parse(srvr.URL)
		},
	}
	httpC := &http.Client{Transport: transport}

	return srvr, httpC
}

func (s *collinsSuite) TestCreateAsset(c *C) {
	tag := "test"
	status := "somestatus"
	srvr, httpC := getHTTPTestClientAndServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			reqStr := "/api/asset/" + tag + "?" + "status=" + status
			if !strings.Contains(r.RequestURI, reqStr) {
				http.Error(w, "unexpected request", http.StatusInternalServerError)
			} else {
				w.WriteHeader(http.StatusCreated)
			}
		}))
	defer srvr.Close()
	client := &Client{
		config: defaultConfig(),
		client: httpC,
	}

	err := client.CreateAsset(tag, status)
	c.Assert(err, IsNil)
}

func (s *collinsSuite) TestCreateAssetStatusFailure(c *C) {
	srvr, httpC := getHTTPTestClientAndServer(failureReturner)
	defer srvr.Close()
	client := &Client{
		config: defaultConfig(),
		client: httpC,
	}

	errStr := ".*unexpected. Response body.*test failure.*"
	err := client.CreateAsset("test", "status")
	c.Assert(err, ErrorMatches, errStr)
}

func (s *collinsSuite) TestCreateAssetStatusAlreadyExists(c *C) {
	srvr, httpC := getHTTPTestClientAndServer(alreadyExistsReturner)
	defer srvr.Close()
	client := &Client{
		config: defaultConfig(),
		client: httpC,
	}

	err := client.CreateAsset("test", "status")
	c.Assert(err, IsNil)
}

func (s *collinsSuite) TestGetAsset(c *C) {
	tag := "test"
	status := "status"
	state := "state"
	asset := Asset{
		Tag:    tag,
		Status: status,
		State: struct {
			Name string `json:"NAME"`
		}{
			Name: state,
		},
	}
	srvr, httpC := getHTTPTestClientAndServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			reqStr := "/api/asset/" + tag
			if !strings.Contains(r.RequestURI, reqStr) {
				http.Error(w, "unexpected request", http.StatusInternalServerError)
			} else {
				resp := struct {
					Data struct {
						Asset Asset `json:"ASSET"`
					} `json:"data"`
				}{
					Data: struct {
						Asset Asset `json:"ASSET"`
					}{
						Asset: asset,
					},
				}
				if body, err := json.Marshal(resp); err == nil {
					w.Write(body)
					return
				}
				http.Error(w, "json failure", http.StatusInternalServerError)
			}
		}))
	defer srvr.Close()
	client := &Client{
		config: defaultConfig(),
		client: httpC,
	}

	rcvdAsset, err := client.GetAsset(tag)
	c.Assert(err, IsNil)
	c.Assert(rcvdAsset, DeepEquals, asset)
}

func (s *collinsSuite) TestGetAssetStatusFailure(c *C) {
	srvr, httpC := getHTTPTestClientAndServer(failureReturner)
	defer srvr.Close()
	client := &Client{
		config: defaultConfig(),
		client: httpC,
	}

	errStr := ".*unexpected. Response body.*test failure.*"
	_, err := client.GetAsset("test")
	c.Assert(err, ErrorMatches, errStr)
}

func (s *collinsSuite) TestGetAssetInvalidResponse(c *C) {
	srvr, httpC := getHTTPTestClientAndServer(okReturner)
	defer srvr.Close()
	client := &Client{
		config: defaultConfig(),
		client: httpC,
	}

	errStr := "failed to unmarshal response.*"
	_, err := client.GetAsset("test")
	c.Assert(err, ErrorMatches, errStr)
}

func (s *collinsSuite) TestGetAllAssets(c *C) {
	tag := "test"
	status := "status"
	state := "state"
	asset := Asset{
		Tag:    tag,
		Status: status,
		State: struct {
			Name string `json:"NAME"`
		}{
			Name: state,
		},
	}
	srvr, httpC := getHTTPTestClientAndServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			reqStr := "/api/assets"
			if !strings.Contains(r.RequestURI, reqStr) {
				http.Error(w, "unexpected request", http.StatusInternalServerError)
			} else {
				resp := struct {
					Data struct {
						Assets []struct {
							Asset Asset `json:"ASSET"`
						} `json:"Data"`
					} `json:"data"`
				}{
					Data: struct {
						Assets []struct {
							Asset Asset `json:"ASSET"`
						} `json:"Data"`
					}{
						Assets: []struct {
							Asset Asset `json:"ASSET"`
						}{
							struct {
								Asset Asset `json:"ASSET"`
							}{
								Asset: asset,
							},
						},
					},
				}
				if body, err := json.Marshal(resp); err == nil {
					w.Write(body)
					return
				}
				http.Error(w, "json failure", http.StatusInternalServerError)
			}
		}))
	defer srvr.Close()
	client := &Client{
		config: defaultConfig(),
		client: httpC,
	}

	rcvdAssets, err := client.GetAllAssets()
	c.Assert(err, IsNil)
	c.Assert(len(rcvdAssets), Equals, 1)
	c.Assert(rcvdAssets[0], DeepEquals, asset)
}

func (s *collinsSuite) TestGetAllAssetsStatusFailure(c *C) {
	srvr, httpC := getHTTPTestClientAndServer(failureReturner)
	defer srvr.Close()
	client := &Client{
		config: defaultConfig(),
		client: httpC,
	}

	errStr := ".*unexpected. Response body.*test failure.*"
	_, err := client.GetAllAssets()
	c.Assert(err, ErrorMatches, errStr)
}

func (s *collinsSuite) TestGetAllAssetsInvalidResponse(c *C) {
	srvr, httpC := getHTTPTestClientAndServer(okReturner)
	defer srvr.Close()
	client := &Client{
		config: defaultConfig(),
		client: httpC,
	}

	errStr := "failed to unmarshal response.*"
	_, err := client.GetAllAssets()
	c.Assert(err, ErrorMatches, errStr)
}

func (s *collinsSuite) TestCreateState(c *C) {
	name := "test"
	description := "somedescription"
	status := "somestatus"
	srvr, httpC := getHTTPTestClientAndServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			reqStr := "/api/state/" + name
			if !strings.Contains(r.RequestURI, reqStr) ||
				!strings.Contains(r.RequestURI, "name="+strings.ToUpper(name)) ||
				!strings.Contains(r.RequestURI, "label="+strings.Title(name)) ||
				!strings.Contains(r.RequestURI, "description="+description) ||
				!strings.Contains(r.RequestURI, "status="+status) {
				http.Error(w, "unexpected request", http.StatusInternalServerError)
			} else {
				w.WriteHeader(http.StatusCreated)
			}
		}))
	defer srvr.Close()
	client := &Client{
		config: defaultConfig(),
		client: httpC,
	}

	err := client.CreateState(name, description, status)
	c.Assert(err, IsNil)
}

func (s *collinsSuite) TestCreateStateStatusFailure(c *C) {
	srvr, httpC := getHTTPTestClientAndServer(failureReturner)
	defer srvr.Close()
	client := &Client{
		config: defaultConfig(),
		client: httpC,
	}

	errStr := ".*unexpected. Response body.*test failure.*"
	err := client.CreateState("test", "status", "description")
	c.Assert(err, ErrorMatches, errStr)
}

func (s *collinsSuite) TestCreateStateStatusAlreadyExists(c *C) {
	srvr, httpC := getHTTPTestClientAndServer(alreadyExistsReturner)
	defer srvr.Close()
	client := &Client{
		config: defaultConfig(),
		client: httpC,
	}

	err := client.CreateState("test", "status", "description")
	c.Assert(err, IsNil)
}

func (s *collinsSuite) TestSetAssetStatus(c *C) {
	tag := "test"
	status := "status"
	state := "state"
	reason := "reason"
	srvr, httpC := getHTTPTestClientAndServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			reqStr := "/api/asset/" + tag
			if !strings.Contains(r.RequestURI, reqStr) ||
				!strings.Contains(r.RequestURI, "tag="+tag) ||
				!strings.Contains(r.RequestURI, "status="+status) ||
				!strings.Contains(r.RequestURI, "state="+state) ||
				!strings.Contains(r.RequestURI, "reason="+reason) {
				http.Error(w, "unexpected request", http.StatusInternalServerError)
			} else {
				w.WriteHeader(http.StatusOK)
			}
		}))
	defer srvr.Close()
	client := &Client{
		config: defaultConfig(),
		client: httpC,
	}

	err := client.SetAssetStatus(tag, status, state, reason)
	c.Assert(err, IsNil)
}

func (s *collinsSuite) TestSetAssetStatusStatusFailure(c *C) {
	srvr, httpC := getHTTPTestClientAndServer(failureReturner)
	defer srvr.Close()
	client := &Client{
		config: defaultConfig(),
		client: httpC,
	}

	errStr := ".*unexpected. Response body.*test failure.*"
	err := client.SetAssetStatus("test", "status", "state", "reason")
	c.Assert(err, ErrorMatches, errStr)
}
