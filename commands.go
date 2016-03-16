package main

import (
	"errors"
	"github.com/arkenio/goarken"
	"github.com/arkenio/goarken/drivers"
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
		cli.StringFlag{
			Name: "driver",
			Value: "fleet",
			Usage: "Service driver to use (fleet, rancher)",
		},

		cli.StringFlag{
			Name: "rancherEndpoint",
			EnvVar: "RANCHER_ENDPOINT",
			Value: "",
			Usage: "Rancher endpoint on which to send commands",
		},

		cli.StringFlag{
			Name: "rancherAccessKey",
			EnvVar: "RANCHER_ACCESSKEY",
			Value: "",
			Usage: "The access key to use to connect to Rancher",
		},

		cli.StringFlag{
			Name: "rancherSecretKey",
			EnvVar: "RANCHER_SECRETKEY",
			Value: "",
			Usage: "The secret key to use to connect to Rancher",
		},
	}

	return flags
}

func GetCommands(stop chan interface{}) []cli.Command {

	commands := []cli.Command{
		{
			Name:  "watch",
			Usage: "Watch the cluster for inconsistency and log errors",
			Flags: []cli.Flag{

				cli.StringFlag{
					Name:   "datadogApiKey",
					Value:  "",
					Usage:  "The datadog API key, if set, metrics are sent to datadog",
					EnvVar: "DD_API_KEY",
				},
				cli.IntFlag{
					Name:   "checkCount",
					Value:  1,
					Usage:  "Number of failed check before marking a service to failed",
				},
				cli.IntFlag{
					Name:   "checkGracePeriod",
					Value:  5,
					Usage:  "Number of seconds before rechecking a service status",
				},
			},
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
				{
					Name:  "start",
					Usage: "Starts the service associated to the domain",
					Action: func(c *cli.Context) {
						NewDomainStartCommand(c)(stop)
					},
				},
				{
					Name:  "stop",
					Usage: "Stop the service associated to the domain",
					Action: func(c *cli.Context) {
						NewDomainStopCommand(c)(stop)
					},
				},
				{
					Name:  "passivate",
					Usage: "Passivate the service associated to the domain",
					Action: func(c *cli.Context) {
						NewDomainPassivateCommand(c)(stop)
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

func CreateServiceDriverFromCli(c *cli.Context, etcdClient *etcd.Client ) drivers.ServiceDriver {
	switch c.GlobalString("driver") {
	case "rancher":
		return drivers.NewRancherServiceDriver(etcdClient,
			c.GlobalString("rancherHost"),c.GlobalString("rancherAccessKey"),c.GlobalString("rancherSecretKey"))

	default:
		return drivers.NewFleetServiceDriver(etcdClient)
	}
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
		Watcher:       w,
		Client:        client,
		SingleRun:     c.Bool("single"),
		DataDogAPIKey: c.String("datadogApiKey"),
		CheckCount:    c.Int("checkCount"),
		GracePeriod:   c.Int("checkGracePeriod"),
	}
	return cw.Watch
}

type NotImplementedCommand struct{}

func (ni *NotImplementedCommand) Run(stop chan interface{}) error {
	return errors.New("Command is not implemented")
}

func CreateServiceCommand(c *cli.Context) *ServiceCommand {

	goarken.SetDomainPrefix(c.GlobalString("domainDir"))
	goarken.SetServicePrefix(c.GlobalString("serviceDir"))

	etcdClient := CreateEtcdClientFromCli(c)

	return &ServiceCommand{
		Client: etcdClient,
		Driver: CreateServiceDriverFromCli(c, etcdClient),
		Cli:    c,
	}

}

func NewServiceListCommand(c *cli.Context) Runnable {
	client := CreateEtcdClientFromCli(c)
	w := CreateWatcherFromCli(c, client)
	sc := &ServiceCommand{
		Watcher: w,
		Client:  CreateEtcdClientFromCli(c),
		Cli:     c,
	}

	return sc.List

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
	client := CreateEtcdClientFromCli(c)
	w := CreateWatcherFromCli(c, client)

	dc := &DomainCommand{
		Client:  client,
		Cli:     c,
		Watcher: w,
	}

	return dc.List
}

func NewDomainCommand(c *cli.Context) *DomainCommand {
	goarken.SetDomainPrefix(c.GlobalString("domainDir"))
	goarken.SetServicePrefix(c.GlobalString("serviceDir"))

	dc := &DomainCommand{
		Client: CreateEtcdClientFromCli(c),
		Cli:    c,
	}

	return dc
}

func NewDomainInfoCommand(c *cli.Context) Runnable {
	return NewDomainCommand(c).Cat

}

func NewDomainStartCommand(c *cli.Context) Runnable {
	return NewDomainCommand(c).Start
}

func NewDomainStopCommand(c *cli.Context) Runnable {
	return NewDomainCommand(c).Stop
}

func NewDomainPassivateCommand(c *cli.Context) Runnable {
	return NewDomainCommand(c).Passivate
}
