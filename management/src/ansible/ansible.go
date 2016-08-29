package ansible

import (
	"io"
	"os"
	"os/exec"

	"golang.org/x/net/context"

	"github.com/Sirupsen/logrus"
	"github.com/contiv/executor"
)

// Runner facilitates running a playbook on specified inventory
type Runner struct {
	inventory   Inventory
	playbook    string
	user        string
	privKeyFile string
	extraVars   string
	ctxt        context.Context
}

// NewRunner returns an instance of Runner for specified playbook and inventory.
// The caller passes a ctxt that can be used to control runner's state using a
// cancellable context or a timeout based context or a dummy context if no control is desired.
func NewRunner(inventory Inventory, playbook, user, privKeyFile, extraVars string, ctxt context.Context) *Runner {
	return &Runner{
		inventory:   inventory,
		playbook:    playbook,
		user:        user,
		privKeyFile: privKeyFile,
		extraVars:   extraVars,
		ctxt:        ctxt,
	}
}

// Run runs a playbook and return's it's status as well the stdout and
// stderr outputs respectively.
func (r *Runner) Run(stdout, stderr io.Writer) error {
	hostsFile, err := NewInventoryFile(r.inventory)
	if err != nil {
		return err
	}
	defer os.Remove(hostsFile.Name())

	logrus.Debugf("going to run playbook: %q with hosts file: %q and vars: %s", r.playbook, hostsFile.Name(), r.extraVars)
	cmd := exec.Command("ansible-playbook", "-i", hostsFile.Name(), "--user", r.user,
		"--private-key", r.privKeyFile, "--extra-vars", r.extraVars, r.playbook)
	// turn off host key checking as we are in non-interactive mode
	cmd.Env = append(cmd.Env, "ANSIBLE_HOST_KEY_CHECKING=false")
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	e := executor.New(cmd)
	res, err := e.Run(r.ctxt)
	if err != nil {
		return err
	}
	logrus.Debugf("executor result: %s", res)
	return nil
}
