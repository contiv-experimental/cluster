package manager

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

var httpErrorResp = func(rsrc, status string, body []byte) error {
	return fmt.Errorf("Request: %s Response status: %q. Response body: %s", rsrc, status, body)
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

func (c *Client) formURL(rsrc string) string {
	return fmt.Sprintf("http://%s/%s", c.url, rsrc)
}

func (c *Client) doPost(rsrc string) error {
	var (
		err  error
		resp *http.Response
	)

	if resp, err = c.httpC.Post(c.formURL(rsrc), "application/json", nil); err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			body = []byte{}
		}
		return httpErrorResp(rsrc, resp.Status, body)
	}

	return nil
}

// XXX: we should have a well defined structure for the info that is resturned
func (c *Client) doGet(rsrc string) ([]byte, error) {
	var (
		body []byte
		err  error
		resp *http.Response
	)

	if resp, err = c.httpC.Get(c.formURL(rsrc)); err != nil {
		return nil, err
	}

	if body, err = ioutil.ReadAll(resp.Body); err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, httpErrorResp(rsrc, resp.Status, body)
	}

	return body, nil
}

// PostNodeCommission posts the request to commission a node
func (c *Client) PostNodeCommission(nodeName string) error {
	return c.doPost(fmt.Sprintf("%s/%s", PostNodeCommissionPrefix, nodeName))
}

// PostNodeDecommission posts the request to decommission a node
func (c *Client) PostNodeDecommission(nodeName string) error {
	return c.doPost(fmt.Sprintf("%s/%s", PostNodeDecommissionPrefix, nodeName))
}

// PostNodeInMaintenance posts the request to put a node in-maintenance
func (c *Client) PostNodeInMaintenance(nodeName string) error {
	return c.doPost(fmt.Sprintf("%s/%s", PostNodeMaintenancePrefix, nodeName))
}

// GetNode requests info of a specified node
func (c *Client) GetNode(nodeName string) ([]byte, error) {
	return c.doGet(fmt.Sprintf("%s/%s", GetNodeInfoPrefix, nodeName))
}

// GetAllNodes requests info of all known nodes
func (c *Client) GetAllNodes() ([]byte, error) {
	return c.doGet(GetNodesInfo)
}
