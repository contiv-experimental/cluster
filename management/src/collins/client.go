package collins

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/contiv/errored"
)

// Config denotes the configuration for collins client
type Config struct {
	URL      string `json:"url"`
	User     string `json:"user"`
	Password string `json:"password"`
}

// DefaultConfig returns the default configuration values for the collins client
func DefaultConfig() Config {
	return Config{
		URL:      "http://localhost:9000",
		User:     "blake",
		Password: "admin:first",
	}
}

// Asset denotes the asset related information as read from collins. This is
// not all of the information.
type Asset struct {
	Tag    string `json:"TAG"`
	Status string `json:"STATUS"`
	State  struct {
		Name string `json:"NAME"`
	}
}

// Client denotes state for a collins client
type Client struct {
	client *http.Client
	config Config
}

// NewClientFromConfig initializes and return collins client using specified configuration
func NewClientFromConfig(config Config) *Client {
	return &Client{
		config: config,
		client: &http.Client{},
	}
}

// NewClient initializes and return collins client using default configuration
func NewClient() *Client {
	return NewClientFromConfig(DefaultConfig())
}

// CreateAsset creates an asset with specified tag, status and state
func (c *Client) CreateAsset(tag, status string) error {
	params := &url.Values{}
	params.Set("status", status)

	reqURL := c.config.URL + "/api/asset/" + tag + "?" + params.Encode()
	req, err := http.NewRequest("PUT", reqURL, nil)
	if err != nil {
		return err
	}
	req.SetBasicAuth(c.config.User, c.config.Password)

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusConflict {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			body = []byte{}
		}
		return errored.Errorf("status code %d unexpected. Response body: %q",
			resp.StatusCode, body)
	}

	if resp.StatusCode == http.StatusConflict {
		logrus.Warnf("asset %q already exists", tag)
	}

	return nil
}

// GetAsset queries and returns an asset with specified tag
func (c *Client) GetAsset(tag string) (Asset, error) {
	reqURL := c.config.URL + "/api/asset/" + tag
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return Asset{}, err
	}
	req.SetBasicAuth(c.config.User, c.config.Password)

	resp, err := c.client.Do(req)
	if err != nil {
		return Asset{}, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return Asset{}, errored.Errorf("failed to read response body. Error: %s", err)
	}

	if resp.StatusCode != http.StatusOK {
		return Asset{}, errored.Errorf("status code %d unexpected. Response body: %q",
			resp.StatusCode, body)
	}

	logrus.Debugf("response: %s", body)
	collinsResp := &struct {
		Data struct {
			Asset Asset `json:"ASSET"`
		} `json:"data"`
	}{}
	if err := json.Unmarshal(body, collinsResp); err != nil {
		return Asset{}, errored.Errorf("failed to unmarshal response. Error: %s", err)
	}

	logrus.Debugf("collins asset: %+v", collinsResp.Data.Asset)
	return collinsResp.Data.Asset, nil
}

// GetAllAssets queries and returns a all the assets
func (c *Client) GetAllAssets() (interface{}, error) {
	reqURL := c.config.URL + "/api/assets"
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(c.config.User, c.config.Password)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errored.Errorf("failed to read response body. Error: %s", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errored.Errorf("status code %d unexpected. Response body: %q",
			resp.StatusCode, body)
	}

	logrus.Debugf("response: %s", body)
	collinsResp := &struct {
		Data struct {
			Assets []struct {
				Asset Asset `json:"ASSET"`
			} `json:"Data"`
		} `json:"data"`
	}{}
	if err := json.Unmarshal(body, collinsResp); err != nil {
		return nil, errored.Errorf("failed to unmarshal response. Error: %s", err)
	}

	assets := []Asset{}
	for _, d := range collinsResp.Data.Assets {
		logrus.Debugf("collins asset: %+v", d.Asset)
		assets = append(assets, d.Asset)
	}
	return assets, nil
}

// CreateState creates a state with specified name, description and
// associated status
func (c *Client) CreateState(name, description, status string) error {
	params := &url.Values{}
	params.Set("name", strings.ToUpper(name))
	params.Set("label", strings.Title(name))
	params.Set("description", description)
	params.Set("status", status)

	reqURL := c.config.URL + "/api/state/" + name + "?" + params.Encode()
	req, err := http.NewRequest("PUT", reqURL, nil)
	if err != nil {
		return err
	}
	req.SetBasicAuth(c.config.User, c.config.Password)

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusConflict {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			body = []byte{}
		}
		return errored.Errorf("status code %d unexpected. Response body: %q",
			resp.StatusCode, body)
	}

	if resp.StatusCode == http.StatusConflict {
		logrus.Warnf("state %q already exists", name)
	}

	return nil
}

// AddAssetLog creates a log entry for an asset
func (c *Client) AddAssetLog(tag, mtype, message string) error {
	return errored.Errorf("not implemented")
}

// SetAssetStatus sets the status of an asset
func (c *Client) SetAssetStatus(tag, status, state, reason string) error {
	params := &url.Values{}
	params.Set("tag", tag)
	params.Set("status", status)
	params.Set("state", state)
	params.Set("reason", reason)

	reqURL := c.config.URL + "/api/asset/" + tag + "?" + params.Encode()
	req, err := http.NewRequest("POST", reqURL, nil)
	if err != nil {
		return err
	}
	req.SetBasicAuth(c.config.User, c.config.Password)

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			body = []byte{}
		}
		return errored.Errorf("status code %d unexpected. Response body: %q",
			resp.StatusCode, body)
	}

	return nil
}
