package main

import (
	//	"errors"
	"errors"
	"fmt"
	. "github.com/arkenio/goarken"
	"github.com/codegangsta/cli"
	"github.com/coreos/go-etcd/etcd"
	"os"
	"strings"
	"text/tabwriter"
)

type DomainCommand struct {
	Watcher *Watcher
	Client  *etcd.Client
	Cli     *cli.Context
}

func (dc *DomainCommand) getDomain() (*Domain, error) {
	if len(dc.Cli.Args()) > 0 {
		domainName := dc.Cli.Args()[0]
		path := dc.Cli.GlobalString("domainDir") + "/" + domainName
		return GetDomainFromPath(path, dc.Client)
	} else {
		return nil, errors.New("You must pass the domain name as an argument")
	}

}

func (dc *DomainCommand) List(stop chan interface{}) error {

	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 0, 8, 2, '\t', 0)
	fmt.Fprintln(w, "Host\tTyp\tValue")
	fmt.Fprintln(w, "----\t-----\t-------")
	for host, domain := range dc.Watcher.Domains {
		fmt.Fprintln(w, strings.Join([]string{
			host,
			domain.Typ,
			domain.Value,
		}, "\t"))
	}

	fmt.Fprintln(w)
	w.Flush()

	return nil

}

func (dc *DomainCommand) Cat(stop chan interface{}) error {
	domain, err := dc.getDomain()
	if err != nil {
		return err
	}

	if domain.Typ == "service" {
		path := dc.Cli.GlobalString("serviceDir") + "/" + domain.Value
		service, _ := GetServiceClusterFromPath(path, dc.Client)
		renderService(service, "")
	} else {
		fmt.Printf("Redirecting to : %s", domain.Value)
	}
	return nil
}

func (dc *DomainCommand) getAssociatedCluster() (*ServiceCluster, error) {
	domain, err := dc.getDomain()
	if err != nil {
		return nil, err
	}

	if domain.Typ == "service" {
		path := dc.Cli.GlobalString("serviceDir") + "/" + domain.Value
		return GetServiceClusterFromPath(path, dc.Client)
	} else {
		return nil, errors.New("This domain is not of type service")
	}
}

func (dc *DomainCommand) Start(stop chan interface{}) error {

	cluster, err := dc.getAssociatedCluster()
	if err != nil {
		return err
	}

	for _, service := range cluster.GetInstances() {
		err = service.Start(dc.Client)
		if err != nil {
			break
		}
	}
	return nil

}

func (dc *DomainCommand) Stop(stop chan interface{}) error {
	cluster, err := dc.getAssociatedCluster()
	if err != nil {
		return err
	}

	for _, service := range cluster.GetInstances() {
		err = service.Stop(dc.Client)
		if err != nil {
			break
		}
	}
	return nil
}

func (dc *DomainCommand) Passivate(stop chan interface{}) error {
	cluster, err := dc.getAssociatedCluster()
	if err != nil {
		return err
	}

	for _, service := range cluster.GetInstances() {
		err = service.Passivate(dc.Client)
		if err != nil {
			break
		}
	}
	return nil
}
