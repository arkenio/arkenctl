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
		if cluster, ok := sc.Watcher.Services[serviceName]; ok {
			return cluster, nil
		} else {
			return nil, errors.New("Service not foudn")
		}

	} else {
		return nil, errors.New("You must pass the service name as an argument")
	}

}

func (sc *ServiceCommand) List(stop chan interface{}) error {

	statusFilter := sc.Cli.String("status")

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
	return nil
}

func (sc *ServiceCommand) Cat(stop chan interface{}) error {
	cluster, err := sc.getServiceCluster()
	if err != nil {
		return err
	} else {
		sc.renderService(cluster)
	}
	return nil
}

func (sc *ServiceCommand) Start(stop chan interface{}) error {
	cluster, err := sc.getServiceCluster()
	if err != nil {
		return err
	}

	for _, service := range cluster.GetInstances() {
		err = service.Start()
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
		err = service.Stop()
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
		err = service.Passivate()
		if err != nil {
			break
		}
	}

	return err
}

func (sc *ServiceCommand) renderService(cluster *ServiceCluster) {
	tpl := `{{range $index, $service := .GetInstances }}===========================================
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

	t := template.Must(template.New("service").Parse(tpl))
	t.Execute(os.Stdout, cluster)
}
