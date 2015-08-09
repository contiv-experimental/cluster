package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/contiv/cluster/management/src/clusterm/manager"
)

func main() {
	//XXX: set log level debug for now, later take it from cli
	log.SetLevel(log.DebugLevel)

	//XXX: start with a hardcoded config for now. This shall be read from commandline or a file
	config := `{
		"serf" : {
			"Addr" : "127.0.0.1:7373"
		},
		"collins" : {
			"URL" : "http://localhost:9000",
			"User": "blake",
			"Password": "admin:first"
		},
		"ansible" : {
			"configure-playbook": "site.yml",
			"cleanup-playbook": "cleanup.yml",
			"upgrade-playbook": "rolling-upgrade.yml",
			"playbook-location": "/vagrant/src/demo/files/",
			"user": "vagrant",
			"priv_key_file": "/vagrant/src/demo/files/insecure_private_key"
		},
		"manager" : {
			"Addr" : "localhost:9999"
		}
	}`
	mgr, err := manager.NewManager([]byte(config))
	if err != nil {
		log.Fatalf("Failed to initialize the manager. Error: %s", err)
	}

	// start manager's processing loop
	errCh := make(chan error)
	go mgr.Run(errCh)
	select {
	case err := <-errCh:
		log.Fatalf("Error encountered. Error: %s", err)
	}
}
