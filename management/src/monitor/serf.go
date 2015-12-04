package monitor

import (
	"encoding/json"
	"fmt"
	"os/exec"

	log "github.com/Sirupsen/logrus"
	"github.com/mapuri/serf/client"
	"github.com/mapuri/serfer"
)

const (
	nodeLabel  = "NodeLabel"
	nodeSerial = "NodeSerial"
	nodeAddr   = "NodeAddr"
)

// SerfSubsys implements monitoring sub-system for a serf based cluster
type SerfSubsys struct {
	config        *client.Config
	router        *serfer.Router
	discoveredCb  EventCb
	disappearedCb EventCb
}

// NewSerfSubsys initializes and return a SerfSubsys instance
func NewSerfSubsys(config *client.Config) *SerfSubsys {
	sm := &SerfSubsys{
		config: config,
		router: serfer.NewRouter(),
	}
	return sm
}

func serferCb(cb EventCb) serfer.HandlerFunc {
	return func(name string, se client.EventRecord) {
		mer := se.(client.MemberEventRecord)
		events := []Event{}
	for_label:
		for _, mbr := range mer.Members {
			n := &Node{}
			n.label = mbr.Tags[nodeLabel]
			n.serial = mbr.Tags[nodeSerial]
			n.addr = mbr.Tags[nodeAddr]
			e := Event{Node: n}
			switch name {
			case "member-join":
				e.Type = Discovered
			case "member-failed":
				e.Type = Disappeared
			default:
				log.Infof("Unexpected serf event: %q", name)
				break for_label
			}
			log.Debugf("monitor event: %+v", e)
			events = append(events, e)
		}
		cb(events)
		return
	}
}

// RegisterCb implements the callback registration interface of monitoring sub-system
func (sm *SerfSubsys) RegisterCb(e EventType, cb EventCb) error {
	if e == Discovered {
		sm.router.AddMemberJoinHandler(serferCb(cb))
		sm.discoveredCb = cb
		return nil
	}
	if e == Disappeared {
		sm.router.AddMemberFailedHandler(serferCb(cb))
		sm.disappearedCb = cb
		return nil
	}
	return fmt.Errorf("Unsupported event type: %d", e)
}

// Start implements the start interface of monitoring sub-system
func (sm *SerfSubsys) Start() error {
	// read any members and call the Discovered callback.
	// XXX: should this happen elsewhere?
	// XXX: for now just screen scrape the 'serf members' command, need to
	// use the client rpc
	type serfMemberInfo struct {
		Members []struct {
			Name   string            `json:"name"`
			Status string            `json:"status"`
			Tags   map[string]string `json:"tags"`
		} `json:"members"`
	}
	output, err := exec.Command("serf", "members", "-format", "json").CombinedOutput()
	if err != nil {
		log.Errorf("serf members failed. Output: %s, Error: %s", output, err)
		return err
	}
	info := &serfMemberInfo{}
	if err := json.Unmarshal(output, info); err != nil {
		log.Errorf("failed to parse serf members. Output: %s, Error: %s", output, err)
		return err
	}
	events := []Event{}
	for _, mbr := range info.Members {
		log.Debugf("considering member: %+v", mbr)
		if mbr.Status != "alive" {
			continue
		}

		e := Event{
			Type: Discovered,
			Node: &Node{
				label:  mbr.Tags[nodeLabel],
				serial: mbr.Tags[nodeSerial],
				addr:   mbr.Tags[nodeAddr],
			},
		}
		log.Debugf("monitor event: %+v", e)
		events = append(events, e)
	}
	sm.discoveredCb(events)

	// start serf event loop for member join and leave events.
	if err := sm.router.InitSerfFromConfigAndServe(sm.config); err != nil {
		log.Errorf("error occured in event loop. Error: %s", err)
		return err
	}
	return nil
}
