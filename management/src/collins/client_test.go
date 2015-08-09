package collins

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

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

func TestCreateAsset(t *testing.T) {
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

	if err := client.CreateAsset(tag, status); err != nil {
		t.Fatalf("create asset failed. Error: %s", err)
	}
}

func TestCreateAssetStatusFailure(t *testing.T) {
	srvr, httpC := getHTTPTestClientAndServer(failureReturner)
	defer srvr.Close()
	client := &Client{
		config: defaultConfig(),
		client: httpC,
	}

	errStr := "unexpected. Response body"
	if err := client.CreateAsset("test", "status"); err == nil {
		t.Fatalf("create asset succeeded. expected to fail")
	} else if !strings.Contains(err.Error(), errStr) {
		t.Fatalf("create asset failed with unexpected error. rcvd: %s", err)
	}
}

func TestCreateAssetStatusAlreadyExists(t *testing.T) {
	srvr, httpC := getHTTPTestClientAndServer(alreadyExistsReturner)
	defer srvr.Close()
	client := &Client{
		config: defaultConfig(),
		client: httpC,
	}

	if err := client.CreateAsset("test", "status"); err != nil {
		t.Fatalf("create asset failed. Error: %s", err)
	}
}

func TestGetAsset(t *testing.T) {
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

	if rcvdAsset, err := client.GetAsset(tag); err != nil {
		t.Fatalf("get asset failed. Error: %s", err)
	} else if rcvdAsset.Tag != asset.Tag ||
		rcvdAsset.Status != asset.Status ||
		rcvdAsset.State.Name != asset.State.Name {
		t.Fatalf("mismatching asset value. expctd: %+v rcvd: %+v", asset, rcvdAsset)
	}
}

func TestGetAssetStatusFailure(t *testing.T) {
	srvr, httpC := getHTTPTestClientAndServer(failureReturner)
	defer srvr.Close()
	client := &Client{
		config: defaultConfig(),
		client: httpC,
	}

	errStr := "unexpected. Response body"
	if _, err := client.GetAsset("test"); err == nil {
		t.Fatalf("get asset succeeded. expected to fail")
	} else if !strings.Contains(err.Error(), errStr) {
		t.Fatalf("get asset failed with unexpected error. rcvd: %s", err)
	}
}

func TestGetAssetInvalidResponse(t *testing.T) {
	srvr, httpC := getHTTPTestClientAndServer(okReturner)
	defer srvr.Close()
	client := &Client{
		config: defaultConfig(),
		client: httpC,
	}

	errStr := "failed to unmarshal response"
	if _, err := client.GetAsset("test"); err == nil {
		t.Fatalf("get asset succeeded. expected to fail")
	} else if !strings.Contains(err.Error(), errStr) {
		t.Fatalf("get asset failed with unexpected error. rcvd: %s", err)
	}
}

func TestGetAllAssets(t *testing.T) {
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

	if rcvdAssets, err := client.GetAllAssets(); err != nil {
		t.Fatalf("get all assets failed. Error: %s", err)
	} else if len(rcvdAssets) != 1 {
		t.Fatalf("unexpected number of assets. exptd: 1, rcvd: %d", len(rcvdAssets))
	} else if rcvdAssets[0].Tag != asset.Tag ||
		rcvdAssets[0].Status != asset.Status ||
		rcvdAssets[0].State.Name != asset.State.Name {
		t.Fatalf("mismatching asset value. expctd: %+v rcvd: %+v", asset, rcvdAssets[0])
	}
}

func TestGetAllAssetsStatusFailure(t *testing.T) {
	srvr, httpC := getHTTPTestClientAndServer(failureReturner)
	defer srvr.Close()
	client := &Client{
		config: defaultConfig(),
		client: httpC,
	}

	errStr := "unexpected. Response body"
	if _, err := client.GetAllAssets(); err == nil {
		t.Fatalf("get all assets succeeded. expected to fail")
	} else if !strings.Contains(err.Error(), errStr) {
		t.Fatalf("get all assets failed with unexpected error. rcvd: %s", err)
	}
}

func TestGetAllAssetsInvalidResponse(t *testing.T) {
	srvr, httpC := getHTTPTestClientAndServer(okReturner)
	defer srvr.Close()
	client := &Client{
		config: defaultConfig(),
		client: httpC,
	}

	errStr := "failed to unmarshal response"
	if _, err := client.GetAllAssets(); err == nil {
		t.Fatalf("get all assets succeeded. expected to fail")
	} else if !strings.Contains(err.Error(), errStr) {
		t.Fatalf("get all assets failed with unexpected error. rcvd: %s", err)
	}
}

func TestCreateState(t *testing.T) {
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

	if err := client.CreateState(name, description, status); err != nil {
		t.Fatalf("create state failed. Error: %s", err)
	}
}

func TestCreateStateStatusFailure(t *testing.T) {
	srvr, httpC := getHTTPTestClientAndServer(failureReturner)
	defer srvr.Close()
	client := &Client{
		config: defaultConfig(),
		client: httpC,
	}

	errStr := "unexpected. Response body"
	if err := client.CreateState("test", "status", "description"); err == nil {
		t.Fatalf("create state succeeded. expected to fail")
	} else if !strings.Contains(err.Error(), errStr) {
		t.Fatalf("create state failed with unexpected error. rcvd: %s", err)
	}
}

func TestCreateStateStatusAlreadyExists(t *testing.T) {
	srvr, httpC := getHTTPTestClientAndServer(alreadyExistsReturner)
	defer srvr.Close()
	client := &Client{
		config: defaultConfig(),
		client: httpC,
	}

	if err := client.CreateState("test", "status", "description"); err != nil {
		t.Fatalf("create state failed. Error: %s", err)
	}
}

func TestSetAssetStatus(t *testing.T) {
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

	if err := client.SetAssetStatus(tag, status, state, reason); err != nil {
		t.Fatalf("set status failed. Error: %s", err)
	}
}

func TestSetAssetStatusStatusFailure(t *testing.T) {
	srvr, httpC := getHTTPTestClientAndServer(failureReturner)
	defer srvr.Close()
	client := &Client{
		config: defaultConfig(),
		client: httpC,
	}

	errStr := "unexpected. Response body"
	if err := client.SetAssetStatus("test", "status", "state", "reason"); err == nil {
		t.Fatalf("set status succeeded. expected to fail")
	} else if !strings.Contains(err.Error(), errStr) {
		t.Fatalf("set status failed with unexpected error. rcvd: %s", err)
	}
}
