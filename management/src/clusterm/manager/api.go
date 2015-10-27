package manager

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
)

func (m *Manager) apiLoop(errCh chan error) {
	r := mux.NewRouter()

	s := r.Headers("Content-Type", "application/json").Methods("Post").Subrouter()
	s.HandleFunc(fmt.Sprintf("/%s", postNodeCommission), post(m.nodeCommission))
	s.HandleFunc(fmt.Sprintf("/%s", postNodeDecommission), post(m.nodeDecommission))
	s.HandleFunc(fmt.Sprintf("/%s", postNodeMaintenance), post(m.nodeMaintenance))

	s = r.Methods("Get").Subrouter()
	s.HandleFunc(fmt.Sprintf("/%s", getNodeInfo), get(m.oneNode))
	s.HandleFunc(fmt.Sprintf("/%s", GetNodesInfo), get(m.allNodes))
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

func post(postCb func(tag string, extraVars string) error) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		tag := vars["tag"]
		extraVars := r.FormValue(ExtraVarsQuery)
		if err := postCb(tag, extraVars); err != nil {
			http.Error(w,
				err.Error(),
				http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		return
	}
}

func (m *Manager) nodeCommission(tag, extraVars string) error {
	if extraVars != "" {
		// extra vars string should be valid json.
		vars := &map[string]interface{}{}
		if err := json.Unmarshal([]byte(extraVars), vars); err != nil {
			return errInvalidJSON(ExtraVarsQuery, err)
		}
	}
	me := newWaitableEvent(newNodeCommissioned(m, tag, extraVars))
	m.reqQ <- me
	return me.waitForCompletion()
}

func (m *Manager) nodeDecommission(tag, extraVars string) error {
	if extraVars != "" {
		// extra vars string should be valid json.
		vars := &map[string]interface{}{}
		if err := json.Unmarshal([]byte(extraVars), vars); err != nil {
			return errInvalidJSON(ExtraVarsQuery, err)
		}
	}
	me := newWaitableEvent(newNodeDecommissioned(m, tag, extraVars))
	m.reqQ <- me
	return me.waitForCompletion()
}

func (m *Manager) nodeMaintenance(tag, extraVars string) error {
	if extraVars != "" {
		// extra vars string should be valid json.
		vars := &map[string]interface{}{}
		if err := json.Unmarshal([]byte(extraVars), vars); err != nil {
			return errInvalidJSON(ExtraVarsQuery, err)
		}
	}
	me := newWaitableEvent(newNodeInMaintenance(m, tag, extraVars))
	m.reqQ <- me
	return me.waitForCompletion()
}

func get(getCb func(tag string) ([]byte, error)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
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
	if a := m.inventory.GetAsset(tag); a != nil {
		var (
			out []byte
			err error
		)
		if out, err = json.Marshal(a); err != nil {
			return nil, err
		}
		return out, nil
	}
	return []byte{}, nil
}

func (m *Manager) allNodes(noop string) ([]byte, error) {
	var (
		out []byte
		err error
	)
	a := m.inventory.GetAllAssets()
	if out, err = json.Marshal(a); err != nil {
		return nil, err
	}
	return out, nil
}
