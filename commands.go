package main

import (
	"errors"
	"github.com/arkenio/goarken"
	"github.com/codegangsta/cli"
	"github.com/coreos/go-etcd/etcd"
)

type Runnable func(stop chan interface{}) error

func GetGlobalFlags() []cli.Flag {
	flags := []cli.Flag{

		cli.StringFlag{
			Name:  "etcdAddress",
			Value: "http://127.0.0.1:4001/",
			Usage: "etcd http endpoint",
		},
		cli.StringFlag{
			Name:  "domainDir",
			Value: "/domains",
			Usage: "etcd prefix to get domains",
		},
		cli.StringFlag{
			Name:  "serviceDir",
			Value: "/services",
			Usage: "etcd prefix to get services",
		},
		cli.BoolFlag{
			Name:  "logtostderr",
			Usage: "log to stderr instead of files",
		},
	}
	return flags
}

func GetCommands(stop chan interface{}) []cli.Command {

	commands := []cli.Command{
		{
			Name:  "watch",
			Usage: "Watch the cluster for inconsistency and log errors",
			Flags: GetGlobalFlags(),
			Action: func(c *cli.Context) {
				NewClusterWatcher(c)(stop)
			},
		},
		{
			Name:  "service",
			Usage: "Show informations about services",
			Subcommands: []cli.Command{
				{
					Name:  "list",
					Usage: "List all services in the cluster",
					Action: func(c *cli.Context) {
						NewServiceListCommand(c)(stop)
					},
					Flags: []cli.Flag{

						cli.StringFlag{
							Name:  "status",
							Value: "",
							Usage: "Show only services in the given status",
						},
						cli.StringFlag{
							Name:  "template",
							Value: "",
							Usage: "template to use to render the output",
						},
					},
				},
				{
					Name:  "cat",
					Usage: "Get the infos for a service",
					Action: func(c *cli.Context) {
						NewServiceInfoCommand(c)(stop)
					},
					Flags: []cli.Flag{

						cli.StringFlag{
							Name:  "template",
							Value: "",
							Usage: "template to use to render the output",
						},
					},
				},
				{
					Name:  "start",
					Usage: "Starts the given service",
					Action: func(c *cli.Context) {
						NewServiceStartCommand(c)(stop)
					},
				},
				{
					Name:  "stop",
					Usage: "Watch the cluster for inconsistency and log errors",
					Action: func(c *cli.Context) {
						NewServiceStopCommand(c)(stop)
					},
				},
				{
					Name:  "passivate",
					Usage: "Watch the cluster for inconsistency and log errors",
					Action: func(c *cli.Context) {
						NewServicePassivateCommand(c)(stop)
					},
				},
			},
		},
		{
			Name:  "domain",
			Usage: "Show informations about domains",
			Subcommands: []cli.Command{
				{

					Name:  "list",
					Usage: "list the domains",
					Action: func(c *cli.Context) {
						NewDomainListCommand(c)(stop)
					},
				},
				{
					Name:  "cat",
					Usage: "Gets the info of a domain",
					Action: func(c *cli.Context) {
						NewDomainInfoCommand(c)(stop)
					},
				},
			},
		},
	}

	return commands
}

func CreateEtcdClientFromCli(c *cli.Context) *etcd.Client {
	return etcd.NewClient([]string{c.GlobalString("etcdAddress")})
}

func CreateWatcherFromCli(c *cli.Context, client *etcd.Client) *goarken.Watcher {
	w := &goarken.Watcher{
		Client:        client,
		DomainPrefix:  c.GlobalString("domainDir"),
		ServicePrefix: c.GlobalString("serviceDir"),
		Domains:       make(map[string]*goarken.Domain),
		Services:      make(map[string]*goarken.ServiceCluster),
	}
	w.Init()
	return w
}

func NewClusterWatcher(c *cli.Context) Runnable {
	client := CreateEtcdClientFromCli(c)
	w := CreateWatcherFromCli(c, client)

	cw := &ClusterWatcher{
		Watcher:   w,
		Client:    client,
		SingleRun: c.Bool("single"),
	}
	return cw.Watch
}

type NotImplementedCommand struct{}

func (ni *NotImplementedCommand) Run(stop chan interface{}) error {
	return errors.New("Command is not implemented")
}

func CreateServiceCommand(c *cli.Context) *ServiceCommand {
	client := CreateEtcdClientFromCli(c)
	w := CreateWatcherFromCli(c, client)

	return &ServiceCommand{
		Watcher: w,
		Client:  CreateEtcdClientFromCli(c),
		Cli:     c,
	}

}

func NewServiceListCommand(c *cli.Context) Runnable {
	return CreateServiceCommand(c).List
}

func NewServiceInfoCommand(c *cli.Context) Runnable {
	return CreateServiceCommand(c).Cat
}

func NewServiceStartCommand(c *cli.Context) Runnable {
	return CreateServiceCommand(c).Start
}

func NewServiceStopCommand(c *cli.Context) Runnable {
	return CreateServiceCommand(c).Stop
}

func NewServicePassivateCommand(c *cli.Context) Runnable {
	return CreateServiceCommand(c).Passivate
}

func NewDomainListCommand(c *cli.Context) Runnable {
	return (&NotImplementedCommand{}).Run
}

func NewDomainInfoCommand(c *cli.Context) Runnable {
	return (&NotImplementedCommand{}).Run
}
