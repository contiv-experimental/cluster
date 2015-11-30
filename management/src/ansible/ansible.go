package ansible

import (
	"io"
	"os"
	"os/exec"

	log "github.com/Sirupsen/logrus"
)

// Runner facillitaes running a playbook on specifed inventory
type Runner struct {
	inventory   Inventory
	playbook    string
	user        string
	privKeyFile string
	extraVars   string
}

// NewRunner returns an instance of Runner for specified playbook and inventory
func NewRunner(inventory Inventory, playbook, user, privKeyFile, extraVars string) *Runner {
	return &Runner{
		inventory:   inventory,
		playbook:    playbook,
		user:        user,
		privKeyFile: privKeyFile,
		extraVars:   extraVars,
	}
}

// Run runs a playbook and return's it's status as well the stdout and
// stderr outputs respectively.
func (r *Runner) Run(stdout, stderr io.Writer) error {
	var (
		hostsFile *os.File
		cmd       *exec.Cmd
		err       error
	)
	if hostsFile, err = NewInventoryFile(r.inventory); err != nil {
		return err
	}
	defer os.Remove(hostsFile.Name())

	log.Debugf("going to run playbook: %q with hosts file: %q and vars: %s", r.playbook, hostsFile.Name(), r.extraVars)
	cmd = exec.Command("ansible-playbook", "-i", hostsFile.Name(), "--user", r.user,
		"--private-key", r.privKeyFile, "--extra-vars", r.extraVars, r.playbook)
	// turn off host key checking as we are in non-interactive mode
	cmd.Env = append(cmd.Env, "ANSIBLE_HOST_KEY_CHECKING=false")
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	if err = cmd.Run(); err != nil {
		return err
	}
	return nil
}
