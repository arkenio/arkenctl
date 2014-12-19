package main

import (
	"errors"
	"flag"
	"github.com/coreos/go-etcd/etcd"
	"github.com/golang/glog"
)

type Config struct {
	port          int
	domainPrefix  string
	servicePrefix string
	etcdAddress   string
	client        *etcd.Client
	cpuProfile    string
}

func (c *Config) getEtcdClient() (*etcd.Client, error) {
	if c.client == nil {
		c.client = etcd.NewClient([]string{c.etcdAddress})
		if !c.client.SyncCluster() {
			return nil, errors.New("Unable to sync with etcd cluster, check your configuration or etcd status")
		}
	}
	return c.client, nil
}

func parseConfig() *Config {
	config := &Config{}
	flag.StringVar(&config.domainPrefix, "domainDir", "/domains", "etcd prefix to get domains")
	flag.StringVar(&config.servicePrefix, "serviceDir", "/services", "etcd prefix to get services")
	flag.StringVar(&config.etcdAddress, "etcdAddress", "http://127.0.0.1:4001/", "etcd client host")
	flag.StringVar(&config.cpuProfile, "cpuProfile", "/tmp/arken-wd.prof", "File to dump cpuProfile")
	flag.Parse()

	glog.Infof("Dumping Configuration")
	glog.Infof("  domainPrefix : %s", config.domainPrefix)
	glog.Infof("  servicesPrefix : %s", config.servicePrefix)
	glog.Infof("  etcdAddress : %s", config.etcdAddress)
	glog.Infof("  cpuProfile: %s", config.cpuProfile)

	return config
}
