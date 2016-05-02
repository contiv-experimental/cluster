package manager

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/contiv/errored"
)

var httpErrorResp = func(rsrc string, req *APIRequest, status string, body []byte) error {
	return errored.Errorf("Request URL: %s Request Body: %+v Response status: %q. Response body: %s", rsrc, req, status, body)
}

// Client provides the methods for issuing post and get requests to cluster manager
type Client struct {
	url   string
	httpC *http.Client
}

// NewClient instantiates a REST based rpc client for cluster manager
func NewClient(url string) *Client {
	return &Client{url: url, httpC: http.DefaultClient}
}

func (c *Client) formURL(rsrc, extraVars string) string {
	if strings.TrimSpace(extraVars) == "" {
		return fmt.Sprintf("http://%s/%s", c.url, rsrc)
	}
	v := url.Values{}
	v.Set(ExtraVarsQuery, strings.TrimSpace(extraVars))
	return fmt.Sprintf("http://%s/%s?%s", c.url, rsrc, v.Encode())
}

func (c *Client) doPost(rsrc, extraVars string, req *APIRequest) error {

	var reqJSON *bytes.Buffer
	if req != nil {
		reqJSON = new(bytes.Buffer)
		if err := json.NewEncoder(reqJSON).Encode(req); err != nil {
			return err
		}
	}

	// XXX: http.NewRequest (that http.Post()) calls panics when a reqJSON
	// variable is nil, hence doing this explicit check here.
	// golang issue: https://github.com/golang/go/issues/15455
	var (
		resp *http.Response
		err  error
	)
	if reqJSON == nil {
		resp, err = c.httpC.Post(c.formURL(rsrc, extraVars), "application/json", nil)
	} else {
		resp, err = c.httpC.Post(c.formURL(rsrc, extraVars), "application/json", reqJSON)
	}
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			body = []byte{}
		}
		return httpErrorResp(rsrc, req, resp.Status, body)
	}

	return nil
}

// XXX: we should have a well defined structure for the info that is resturned
func (c *Client) doGet(rsrc string) ([]byte, error) {
	resp, err := c.httpC.Get(c.formURL(rsrc, ""))
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, httpErrorResp(rsrc, nil, resp.Status, body)
	}

	return body, nil
}

// PostNodeCommission posts the request to commission a node
func (c *Client) PostNodeCommission(nodeName string, extraVars string, hostGroup string) error {
	req := &APIRequest{
		HostGroup: hostGroup,
	}
	return c.doPost(fmt.Sprintf("%s/%s", PostNodeCommissionPrefix, nodeName), extraVars, req)
}

// PostNodesCommission posts the request to commission a set of nodes
func (c *Client) PostNodesCommission(nodeNames []string, extraVars string, hostGroup string) error {
	req := &APIRequest{
		Nodes:     nodeNames,
		HostGroup: hostGroup,
	}
	return c.doPost(PostNodesCommission, extraVars, req)
}

// PostNodeDecommission posts the request to decommission a node
func (c *Client) PostNodeDecommission(nodeName, extraVars string) error {
	return c.doPost(fmt.Sprintf("%s/%s", PostNodeDecommissionPrefix, nodeName), extraVars, nil)
}

// PostNodesDecommission posts the request to decommission a set of nodes
func (c *Client) PostNodesDecommission(nodeNames []string, extraVars string) error {
	req := &APIRequest{
		Nodes: nodeNames,
	}
	return c.doPost(PostNodesDecommission, extraVars, req)
}

// PostNodeInMaintenance posts the request to put a node in-maintenance
func (c *Client) PostNodeInMaintenance(nodeName, extraVars string) error {
	return c.doPost(fmt.Sprintf("%s/%s", PostNodeMaintenancePrefix, nodeName), extraVars, nil)
}

// PostNodesInMaintenance posts the request to put a set of nodes in-maintenance
func (c *Client) PostNodesInMaintenance(nodeNames []string, extraVars string) error {
	req := &APIRequest{
		Nodes: nodeNames,
	}
	return c.doPost(PostNodesMaintenance, extraVars, req)
}

// PostNodesDiscover posts the request to provision a set of nodes for discovery
func (c *Client) PostNodesDiscover(nodeAddrs []string, extraVars string) error {
	req := &APIRequest{
		Addrs: nodeAddrs,
	}
	return c.doPost(PostNodesDiscover, extraVars, req)
}

// PostGlobals posts the request to set global extra vars
func (c *Client) PostGlobals(extraVars string) error {
	return c.doPost(PostGlobals, extraVars, nil)
}

// GetNode requests info of a specified node
func (c *Client) GetNode(nodeName string) ([]byte, error) {
	return c.doGet(fmt.Sprintf("%s/%s", GetNodeInfoPrefix, nodeName))
}

// GetAllNodes requests info of all known nodes
func (c *Client) GetAllNodes() ([]byte, error) {
	return c.doGet(GetNodesInfo)
}

// GetGlobals requests the value global extra vars
func (c *Client) GetGlobals() ([]byte, error) {
	return c.doGet(GetGlobals)
}
