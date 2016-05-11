package main

import (
	"bytes"
	"encoding/json"
	"os"
	"reflect"
	"text/template"

	"github.com/codegangsta/cli"
	"github.com/contiv/cluster/management/src/clusterm/manager"
)

type nodeInfo struct {
	Mon map[string]interface{} `json:"monitoring_state"`
	Inv map[string]interface{} `json:"inventory_state"`
	Cfg map[string]interface{} `json:"configuration_state"`
}

type nodesInfo map[string]nodeInfo

type jobInfo map[string]interface{}

type globalInfo map[string]interface{}

// printHelper stores indent related metadat along with the value being printed
type printHelper struct {
	Indent string
	Val    interface{}
}

func newPrintHelper(indent string, val interface{}) *printHelper {
	return &printHelper{
		Indent: indent,
		Val:    val,
	}
}

var (
	typeFuncs = template.FuncMap{
		"valueOf":        reflect.ValueOf,
		"newPrintHelper": newPrintHelper,
	}
	typePrint = `
{{- define "typePrint" }}
	{{- $indent := .Indent }}
	{{- $inVal := .Val }}
	{{- with valueOf $inVal }}
		{{- $type := .Kind.String }}
		{{- if eq $type "map" }}
			{{- template "mapPrint" newPrintHelper $indent $inVal }}
		{{- else if eq $type "slice" }}
			{{- template "slicePrint" newPrintHelper $indent $inVal }}
		{{- else if eq $type "ptr" }}
			{{- $val := .Elem.Interface }}
			{{- template "typePrint" newPrintHelper $indent $val }}
		{{- else }}
			{{- $indent }}{{ $inVal }}{{ "\n" }}
		{{- end }}
	{{- end }}
{{- end }}
{{- define "slicePrint" }}
	{{- $indent := .Indent }}
	{{- $inVal := .Val }}
	{{- range $idx, $elem := $inVal }}
		{{- $indent }}{{ $elem }}{{ "\n" }}
	{{- end }}
{{- end }}
{{- define "mapPrint" }}
	{{- $indent := .Indent }}
	{{- $inVal := .Val }}
	{{- range $key, $val := $inVal }}
		{{- with valueOf $val }}
			{{- $type := .Kind.String }}
			{{- if eq $type "map" }}
				{{- $indent }}{{ $key }}:{{ "\n" }}
				{{- $newIndent := printf "%s    " $indent }}{{ template "typePrint" newPrintHelper $newIndent $val }}
			{{- else if eq $type "slice" }}
				{{- $indent }}{{ $key }}:{{ "\n" }}
				{{- $newIndent := printf "%s    " $indent }}{{ template "typePrint" newPrintHelper $newIndent $val }}
			{{- else }}
				{{- $indent }}{{ $key }}: {{ $val }}{{ "\n" }}
			{{- end }}
		{{- end }}
	{{- end }}
{{- end }}
	`
	typeTemplate = template.Must(template.New("").Funcs(typeFuncs).Parse(typePrint))

	globalPrint    = `{{ template "typePrint" newPrintHelper "" .}}`
	globalTemplate = template.Must(template.Must(typeTemplate.Clone()).Parse(globalPrint))

	nodePrint = `
{{- define "nodePrint" }}
	{{- $invName := .Inv.name }}
	{{- $indent := printf "%s:    " $invName }}
	{{- $invName }}: Inventory State{{ "\n" }}
	{{- template "typePrint" newPrintHelper $indent .Inv }}
	{{- $invName }}: Monitoring State{{ "\n" }}
	{{- template "typePrint" newPrintHelper $indent .Mon }}
	{{- $invName }}: Configuration State{{ "\n" }}
	{{- template "typePrint" newPrintHelper $indent .Cfg }}
{{ end }}
`
	nodeTemplate = template.Must(template.Must(typeTemplate.Clone()).Parse(nodePrint))

	oneNodePrint    = `{{- template "nodePrint" . }}`
	oneNodeTemplate = template.Must(template.Must(nodeTemplate.Clone()).Parse(oneNodePrint))

	multiNodePrint    = `{{- range $key, $val := . }}{{ template "nodePrint" $val }}{{ end }}`
	multiNodeTemplate = template.Must(template.Must(nodeTemplate.Clone()).Parse(multiNodePrint))

	jobPrint = `
Description: {{ .desc }}
Status: {{ .status }}
Error: {{ .error }}
Logs:
{{ template "typePrint" newPrintHelper "    " .logs }}
`
	jobTemplate = template.Must(template.Must(typeTemplate.Clone()).Parse(jobPrint))
)

type getCallback func(c *manager.Client, arg string, flags parsedFlags) error

type getActioner struct {
	arg   string
	flags parsedFlags
	getCb getCallback
}

func newGetActioner(getCb getCallback) *getActioner {
	return &getActioner{getCb: getCb}
}

func (nga *getActioner) procFlags(c *cli.Context) {
	nga.flags.jsonOutput = c.Bool("json")
	return
}

func (nga *getActioner) procArgs(c *cli.Context) {
	nga.arg = c.Args().First()
}

func (nga *getActioner) action(c *manager.Client) error {
	return nga.getCb(c, nga.arg, nga.flags)
}

func ppJSON(out []byte) {
	var outBuf bytes.Buffer
	json.Indent(&outBuf, out, "", "    ")
	outBuf.WriteTo(os.Stdout)
}

func printTemplate(out []byte, t *template.Template, i interface{}) error {
	if err := json.Unmarshal(out, i); err != nil {
		return err
	}
	return t.Execute(os.Stdout, i)
}

func nodeGet(c *manager.Client, nodeName string, flags parsedFlags) error {
	if nodeName == "" {
		return errUnexpectedArgCount("1", 0)
	}

	out, err := c.GetNode(nodeName)
	if err != nil {
		return err
	}

	if !flags.jsonOutput {
		return printTemplate(out, oneNodeTemplate, &nodeInfo{})
	}

	ppJSON(out)
	return nil
}

func nodesGet(c *manager.Client, noop string, flags parsedFlags) error {
	out, err := c.GetAllNodes()
	if err != nil {
		return err
	}

	if !flags.jsonOutput {
		return printTemplate(out, multiNodeTemplate, &nodesInfo{})
	}

	ppJSON(out)
	return nil
}

func globalsGet(c *manager.Client, noop string, flags parsedFlags) error {
	out, err := c.GetGlobals()
	if err != nil {
		return err
	}

	if !flags.jsonOutput {
		return printTemplate(out, globalTemplate, &globalInfo{})
	}

	ppJSON(out)
	return nil
}

func jobGet(c *manager.Client, job string, flags parsedFlags) error {
	if job == "" {
		return errUnexpectedArgCount("1", 0)
	}

	out, err := c.GetJob(job)
	if err != nil {
		return err
	}

	if !flags.jsonOutput {
		return printTemplate(out, jobTemplate, &jobInfo{})
	}

	ppJSON(out)
	return nil
}
