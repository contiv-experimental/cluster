package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/mapuri/serf/client"
	"github.com/mapuri/serfer"
)

func main() {
	r := serfer.NewRouter()
	r.AddMemberJoinHandler(handleJoin)
	sr := r.NewSubRouter("events/")
	sr.AddHandler("foo", handleFoo)

	if err := r.InitSerfAndServe(""); err != nil {
		log.Fatalf("Failed to initialize serfer. Error: %s", err)
	}
}

func handleJoin(name string, e client.EventRecord) {
	log.Infof("Received event: %q: %v", name, e)
}

func handleFoo(name string, e client.EventRecord) {
	log.Infof("Received event: %q: %v", name, e)
}
