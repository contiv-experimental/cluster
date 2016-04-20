package manager

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/contiv/cluster/management/src/configuration"
	"github.com/contiv/errored"
	"github.com/gorilla/mux"
)

// errInvalidJSON is the error returned when an invalid json value is specified for
// the ansible extra variables configuration
var errInvalidJSON = func(name string, err error) error {
	return errored.Errorf("%q should be a valid json. Error: %s", name, err)
}

func (m *Manager) apiLoop(errCh chan error) {
	r := mux.NewRouter()

	s := r.Headers("Content-Type", "application/json").Methods("Post").Subrouter()
	s.HandleFunc(fmt.Sprintf("/%s", postNodeCommission), post(m.nodeCommission))
	s.HandleFunc(fmt.Sprintf("/%s", postNodeDecommission), post(m.nodeDecommission))
	s.HandleFunc(fmt.Sprintf("/%s", postNodeMaintenance), post(m.nodeMaintenance))
	s.HandleFunc(fmt.Sprintf("/%s", postNodeDiscover), post(m.nodeDiscover))
	s.HandleFunc(fmt.Sprintf("/%s", PostGlobals), post(m.globalsSet))

	s = r.Methods("Get").Subrouter()
	s.HandleFunc(fmt.Sprintf("/%s", getNodeInfo), get(m.oneNode))
	s.HandleFunc(fmt.Sprintf("/%s", GetNodesInfo), get(m.allNodes))
	s.HandleFunc(fmt.Sprintf("/%s", GetGlobals), get(m.globalsGet))
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

func post(postCb func(tagOrAddr string, sanitizedExtraVars string) error) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		tagOrAddr := vars["tag"]
		if tagOrAddr == "" {
			tagOrAddr = vars["addr"]
		}
		extraVars := r.FormValue(ExtraVarsQuery)
		sanitzedExtraVars, err := validateAndSanitizeEmptyExtraVars(ExtraVarsQuery, extraVars)
		if err != nil {
			http.Error(w,
				err.Error(),
				http.StatusInternalServerError)
			return
		}
		if err := postCb(tagOrAddr, sanitzedExtraVars); err != nil {
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

func (m *Manager) nodeCommission(tag, sanitizedExtraVars string) error {
	me := newWaitableEvent(newNodeCommissioned(m, tag, sanitizedExtraVars))
	m.reqQ <- me
	return me.waitForCompletion()
}

func (m *Manager) nodeDecommission(tag, sanitizedExtraVars string) error {
	me := newWaitableEvent(newNodeDecommissioned(m, tag, sanitizedExtraVars))
	m.reqQ <- me
	return me.waitForCompletion()
}

func (m *Manager) nodeMaintenance(tag, sanitizedExtraVars string) error {
	me := newWaitableEvent(newNodeInMaintenance(m, tag, sanitizedExtraVars))
	m.reqQ <- me
	return me.waitForCompletion()
}

func (m *Manager) nodeDiscover(addr, sanitizedExtraVars string) error {
	me := newWaitableEvent(newNodeDiscover(m, addr, sanitizedExtraVars))
	m.reqQ <- me
	return me.waitForCompletion()
}

func (m *Manager) globalsSet(tag, extraVars string) error {
	extraVars, err := validateAndSanitizeEmptyExtraVars(ExtraVarsQuery, extraVars)
	if err != nil {
		return err
	}
	me := newWaitableEvent(newSetGlobals(m, extraVars))
	m.reqQ <- me
	return me.waitForCompletion()
}

func get(getCb func(tag string) ([]byte, error)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		tag := vars["tag"]
		var (
			out []byte
			err error
		)
		if out, err = getCb(tag); err != nil {
			http.Error(w,
				err.Error(),
				http.StatusInternalServerError)
			return
		}
		w.Write(out)
	}
}

func (m *Manager) oneNode(tag string) ([]byte, error) {
	node, err := m.findNode(tag)
	if err != nil {
		return nil, err
	}

	out, err := json.Marshal(node)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (m *Manager) allNodes(noop string) ([]byte, error) {
	out, err := json.Marshal(m.nodes)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (m *Manager) globalsGet(noop string) ([]byte, error) {
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
