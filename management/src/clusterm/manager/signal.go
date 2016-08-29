package manager

import (
	"bufio"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/Sirupsen/logrus"
	"github.com/contiv/errored"
)

func (m *Manager) reparseConfig() (*Config, error) {
	f, err := os.Open(m.configFile)
	if err != nil {
		return nil, errored.Errorf("failed to open config file. Error: %v", err)
	}
	defer func() { f.Close() }()
	logrus.Debugf("re-reading configuration from file: %q", m.configFile)
	reader := bufio.NewReader(f)
	config, err := DefaultConfig().MergeFromReader(reader)
	if err != nil {
		return nil, errored.Errorf("failed to merge configuration. Error: %v", err)
	}
	return config, nil
}

func (m *Manager) signalLoop() {
	if strings.TrimSpace(m.configFile) == "" {
		logrus.Infof("clusterm started without a config file, not registering signal handler")
		return
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP)

	for {
		select {
		case <-c:
			config, err := m.reparseConfig()
			if err != nil {
				logrus.Errorf("failed to reparse config. Error: %v", err)
				continue
			}
			if err := NewClient(m.addr).PostConfig(config); err != nil {
				logrus.Errorf("error posting config. Error: %v", err)
			}
		}
	}
}
