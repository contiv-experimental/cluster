package ansible

import (
	"io/ioutil"
	"os"
	"text/template"
)

// InventoryHost contains information about a host in ansible inventory
type InventoryHost struct {
	Alias string
	Addr  string
	group HostGroup
}

// NewInventoryHost instantiates and returns ansible inventory host info structure
func NewInventoryHost(alias, addr, group string) InventoryHost {
	return InventoryHost{
		Alias: alias,
		Addr:  addr,
		group: HostGroup(group),
	}
}

// HostGroup is type for the group name
type HostGroup string

// Inventory contains ansible's inventory
type Inventory struct {
	Hosts map[HostGroup][]InventoryHost
}

// NewInventory returns inventory with specified hosts, grouped by respective groups
func NewInventory(hosts []InventoryHost) Inventory {
	i := Inventory{
		Hosts: make(map[HostGroup][]InventoryHost),
	}
	for _, h := range hosts {
		if _, ok := i.Hosts[h.group]; !ok {
			i.Hosts[h.group] = make([]InventoryHost, 0)
		}
		i.Hosts[h.group] = append(i.Hosts[h.group], h)
	}
	return i
}

// NewInventoryFile creates a hosts file from inventory information. The caller shall
// delete the file after use
func NewInventoryFile(inventory Inventory) (*os.File, error) {
	f, err := ioutil.TempFile("", "hosts")
	if err != nil {
		return nil, err
	}
	// Close the file to ensure that contents are flushed.
	// XXX: We expect the caller to just use the file name to be passed to
	// `ansible` command. This will need to be done different this assumption changes.
	defer f.Close()

	templateText := `
{{/* walk over the groups and print the group name*/}}{{ range $group, $hosts := .Hosts }}[{{ $group }}]
{{/* walk over the hosts in the group and print the host name and address*/}}{{ range $i, $host := $hosts }}{{ $host.Alias }} ansible_ssh_host={{ $host.Addr }}
{{ end }}
{{ end }}
	`
	if err := template.Must(template.New("entry").Parse(templateText)).Execute(f, inventory); err != nil {
		os.Remove(f.Name())
		return nil, err
	}
	return f, nil
}
