package manager

import (
	"encoding/json"
	"io/ioutil"
	"net"
	"net/http"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/contiv/cluster/management/src/configuration"
	"github.com/contiv/errored"
	"github.com/gorilla/mux"
)

// APIRequest is the general request body expected by clusterm from it's client
type APIRequest struct {
	Nodes     []string `json:"nodes,omitempty"`
	Addrs     []string `json:"addrs,omitempty"`
	HostGroup string   `json:"hostgroup,omitempty"`
	ExtraVars string   `json:"extravars,omitempty"`
	Job       string   `json:"job,omitempty"`
}

// errInvalidJSON is the error returned when an invalid json value is specified for
// the ansible extra variables configuration
func errInvalidJSON(name string, err error) error {
	return errored.Errorf("%q should be a valid json. Error: %s", name, err)
}

// errJobNotExist is the error returned when a job with specified label doesn't exists
func errJobNotExist(job string) error {
	return errored.Errorf("info for %q job doesn't exist", job)
}

// errInvalidJobLabel is the error returned when an invalid or empty job label
// is specified as part of job info request
func errInvalidJobLabel(job string) error {
	return errored.Errorf("Invalid or empty job label specified: %q", job)
}

func (m *Manager) apiLoop(errCh chan error) {
	//set following headers for requests expecting a body
	jsonContentHdrs := []string{"Content-Type", "application/json"}
	//set following headers for requests that don't expect a body like get node info.
	emptyHdrs := []string{}
	reqs := map[string][]struct {
		url  string
		hdrs []string
		hdlr http.HandlerFunc
	}{
		"GET": {
			{"/" + getNodeInfo, emptyHdrs, get(m.oneNode)},
			{"/" + GetNodesInfo, emptyHdrs, get(m.allNodes)},
			{"/" + GetGlobals, emptyHdrs, get(m.globalsGet)},
			{"/" + getJob, emptyHdrs, get(m.jobGet)},
		},
		"POST": {
			{"/" + postNodeCommission, emptyHdrs, post(m.nodesCommission)},
			{"/" + PostNodesCommission, jsonContentHdrs, post(m.nodesCommission)},
			{"/" + postNodeDecommission, emptyHdrs, post(m.nodesDecommission)},
			{"/" + PostNodesDecommission, jsonContentHdrs, post(m.nodesDecommission)},
			{"/" + postNodeMaintenance, emptyHdrs, post(m.nodesMaintenance)},
			{"/" + PostNodesMaintenance, jsonContentHdrs, post(m.nodesMaintenance)},
			{"/" + PostNodesDiscover, jsonContentHdrs, post(m.nodesDiscover)},
			{"/" + PostGlobals, emptyHdrs, post(m.globalsSet)},
		},
	}

	r := mux.NewRouter()
	for method, items := range reqs {
		for _, item := range items {
			r.Headers(item.hdrs...).Path(item.url).Methods(method).HandlerFunc(item.hdlr)
		}
	}

	l, err := net.Listen("tcp", m.addr)
	if err != nil {
		log.Errorf("Error setting up listener. Error: %s", err)
		errCh <- err
		return
	}
	if err := http.Serve(l, r); err != nil {
		log.Errorf("Error listening for http requests. Error: %s", err)
		errCh <- err
		return
	}
}

type postCallback func(req *APIRequest) error

func post(postCb postCallback) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// process data from request body, if any
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		req := APIRequest{}
		if len(body) > 0 {
			if err := json.Unmarshal(body, &req); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		// process data from url, if any
		vars := mux.Vars(r)
		if vars["tag"] != "" {
			req.Nodes = append(req.Nodes, vars["tag"])
		}
		if vars["addr"] != "" {
			req.Addrs = append(req.Addrs, vars["addr"])
		}

		// process query variables
		req.ExtraVars, err = validateAndSanitizeEmptyExtraVars(ExtraVarsQuery, req.ExtraVars)
		if err != nil {
			http.Error(w,
				err.Error(),
				http.StatusInternalServerError)
			return
		}

		// call the handler
		if err := postCb(&req); err != nil {
			http.Error(w,
				err.Error(),
				http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		return
	}
}

func validateAndSanitizeEmptyExtraVars(errorPrefix, extraVars string) (string, error) {
	if strings.TrimSpace(extraVars) == "" {
		return configuration.DefaultValidJSON, nil
	}

	// extra vars string should be valid json.
	vars := &map[string]interface{}{}
	if err := json.Unmarshal([]byte(extraVars), vars); err != nil {
		log.Errorf("failed to parse json: '%s'. Error: %v", extraVars, err)
		return "", errInvalidJSON(errorPrefix, err)
	}
	return extraVars, nil
}

func (m *Manager) nodesCommission(req *APIRequest) error {
	me := newWaitableEvent(newCommissionEvent(m, req.Nodes, req.ExtraVars, req.HostGroup))
	m.reqQ <- me
	return me.waitForCompletion()
}

func (m *Manager) nodesDecommission(req *APIRequest) error {
	me := newWaitableEvent(newDecommissionEvent(m, req.Nodes, req.ExtraVars))
	m.reqQ <- me
	return me.waitForCompletion()
}

func (m *Manager) nodesMaintenance(req *APIRequest) error {
	me := newWaitableEvent(newMaintenanceEvent(m, req.Nodes, req.ExtraVars))
	m.reqQ <- me
	return me.waitForCompletion()
}

func (m *Manager) nodesDiscover(req *APIRequest) error {
	me := newWaitableEvent(newDiscoverEvent(m, req.Addrs, req.ExtraVars))
	m.reqQ <- me
	return me.waitForCompletion()
}

func (m *Manager) globalsSet(req *APIRequest) error {
	me := newWaitableEvent(newSetGlobalsEvent(m, req.ExtraVars))
	m.reqQ <- me
	return me.waitForCompletion()
}

type getCallback func(req *APIRequest) ([]byte, error)

func get(getCb getCallback) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		req := &APIRequest{
			Nodes: []string{strings.TrimSpace(vars["tag"])},
			Job:   strings.TrimSpace(vars["job"]),
		}
		out, err := getCb(req)
		if err != nil {
			http.Error(w,
				err.Error(),
				http.StatusInternalServerError)
			return
		}
		w.Write(out)
	}
}

func (m *Manager) oneNode(req *APIRequest) ([]byte, error) {
	node, err := m.findNode(req.Nodes[0])
	if err != nil {
		return nil, err
	}

	out, err := json.Marshal(node)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (m *Manager) allNodes(noop *APIRequest) ([]byte, error) {
	out, err := json.Marshal(m.nodes)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (m *Manager) globalsGet(noop *APIRequest) ([]byte, error) {
	globals := m.configuration.GetGlobals()
	globalData := struct {
		ExtraVars map[string]interface{} `json:"extra-vars"`
	}{
		ExtraVars: make(map[string]interface{}),
	}
	if err := json.Unmarshal([]byte(globals), &globalData.ExtraVars); err != nil {
		return nil, err
	}
	out, err := json.Marshal(globalData)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (m *Manager) jobGet(req *APIRequest) ([]byte, error) {
	var j *Job
	switch req.Job {
	case jobLabelActive:
		j = m.activeJob
	case jobLabelLast:
		j = m.lastJob
	default:
		return nil, errInvalidJobLabel(req.Job)
	}

	if j == nil {
		return nil, errJobNotExist(req.Job)
	}

	out, err := json.Marshal(j)
	if err != nil {
		return nil, err
	}

	return out, nil
}
