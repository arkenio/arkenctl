package main

import (
	"errors"
	"fmt"
	. "github.com/arkenio/goarken"
	"github.com/codegangsta/cli"
	"github.com/coreos/go-etcd/etcd"
	"os"
	"strings"
	"text/tabwriter"
	"text/template"
)

type ServiceCommand struct {
	Watcher *Watcher
	Client  *etcd.Client
	Cli     *cli.Context
}

func (sc *ServiceCommand) getServiceCluster() (*ServiceCluster, error) {
	if len(sc.Cli.Args()) > 0 {
		serviceName := sc.Cli.Args()[0]
		path := sc.Cli.GlobalString("serviceDir") + "/" + serviceName
		return GetServiceClusterFromPath(path, sc.Client)
	} else {
		return nil, errors.New("You must pass the service name as an argument")
	}

}

func (sc *ServiceCommand) List(stop chan interface{}) error {

	statusFilter := sc.Cli.String("status")

	tpl := sc.Cli.String("template")

	if tpl == "" {
		w := new(tabwriter.Writer)
		w.Init(os.Stdout, 0, 8, 2, '\t', 0)
		fmt.Fprintln(w, "Name\tIndex\tDomain\tStatus\tLastAccess")
		fmt.Fprintln(w, "----\t-----\t------\t------\t----------")
		for _, cluster := range sc.Watcher.Services {
			for _, service := range cluster.GetInstances() {

				if statusFilter == "" || statusFilter == service.Status.Compute() {
					fmt.Fprintln(w, strings.Join([]string{
						service.Name,
						service.Index,
						service.Domain,
						service.Status.Compute(),
						fmt.Sprintf("%s", service.LastAccess),
					}, "\t"))
				}
			}
		}
		fmt.Fprintln(w)
		w.Flush()
	} else {
		t := template.Must(template.New("service").Parse(tpl))
		for _, cluster := range sc.Watcher.Services {
			for _, service := range cluster.GetInstances() {
				if statusFilter == "" || statusFilter == service.Status.Compute() {
					t.Execute(os.Stdout, service)
					fmt.Fprintln(os.Stdout, "")
				}
			}
		}
	}

	return nil
}

func (sc *ServiceCommand) Cat(stop chan interface{}) error {
	cluster, err := sc.getServiceCluster()
	if err != nil {
		return err
	} else {
		tpl := sc.Cli.String("template")
		renderService(cluster, tpl)
	}
	return nil
}

func (sc *ServiceCommand) Start(stop chan interface{}) error {
	cluster, err := sc.getServiceCluster()
	if err != nil {
		return err
	}

	for _, service := range cluster.GetInstances() {
		err = service.Start(sc.Client)
		if err != nil {
			break
		}
	}

	return err
}

func (sc *ServiceCommand) Stop(stop chan interface{}) error {
	cluster, err := sc.getServiceCluster()
	if err != nil {
		return err
	}

	for _, service := range cluster.GetInstances() {
		err = service.Stop(sc.Client)
		if err != nil {
			break
		}
	}

	return err
}

func (sc *ServiceCommand) Passivate(stop chan interface{}) error {
	cluster, err := sc.getServiceCluster()
	if err != nil {
		return err
	}

	for _, service := range cluster.GetInstances() {
		err = service.Passivate(sc.Client)
		if err != nil {
			break
		}
	}

	return err
}

func renderService(cluster *ServiceCluster, tpl string) {
	if tpl == "" {

		tpl = `{{range $index, $service := .GetInstances }}===========================================
    Node index : {{.Index}}
    Name : {{.Name}}
    UnitName : {{.UnitName }}
    Etcd key : {{.NodeKey }}
    Domain name : {{.Domain}}
    Location : {{.Location.Host}}:{{.Location.Port}}
    LastAccess: {{.LastAccess}}
    Status: {{.Status.Compute}}
      * expected : {{.Status.Expected}}
      * current : {{.Status.Current}}
      * alive : {{.Status.Alive}}
    {{end}}
`
	}

	t := template.Must(template.New("service").Parse(tpl))
	t.Execute(os.Stdout, cluster)
}
