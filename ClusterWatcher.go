package main

import (
	. "github.com/arkenio/goarken"
	"github.com/coreos/go-etcd/etcd"
	"github.com/golang/glog"
)

type ClusterWatcher struct {
	Watcher   *Watcher
	Client    *etcd.Client
	SingleRun bool
	inError   map[string]*ServiceCluster
}

func (cw *ClusterWatcher) Watch(stop chan interface{}) error {
	cw.inError = make(map[string]*ServiceCluster)

	// First check that no instance has to be passivated
	var err error = nil

	for _, cluster := range cw.Watcher.Services {
		err = cw.check(cluster)
	}

	if !cw.SingleRun {
		// Then watch for changes
		updateChannel := cw.Watcher.Listen()
		for {
			select {
			case <-stop:
				return nil
			case serviceOrDomain := <-updateChannel:
				if cluster, ok := serviceOrDomain.(*ServiceCluster); ok {
					cw.check(cluster)
				}
			}
		}
		return nil
	} else {
		return err
	}

}

func (cw *ClusterWatcher) check(cluster *ServiceCluster) error {
	_, err := cw.Watcher.Services[cluster.Name].Next()
	if err != nil {
		if stError, ok := err.(StatusError); ok {
			switch stError.ComputedStatus {
			case STARTING_STATUS, PASSIVATED_STATUS, STOPPED_STATUS, STOPPING_STATUS:
				break
			default:
				// If status is nil, then we can't say it's an error... it's in an unknown status
				if stError.Status != nil {
					cw.addInError(cluster, stError)
				}
			}
		} else {
			cw.addInError(cluster, err)
		}
		return err
	}
	cw.removeInError(cluster)
	return nil

}

func (cw *ClusterWatcher) addInError(cluster *ServiceCluster, err error) {
	if _, ok := cw.inError[cluster.Name]; ok {
		glog.Errorf("Cluster %s is still in error, computedStatus : %v", cluster.Name, err)
	} else {
		glog.Errorf("Cluster %s is in error : %v ", cluster.Name, err)
		cw.inError[cluster.Name] = cluster
	}
}

func (cw *ClusterWatcher) removeInError(cluster *ServiceCluster) {
	if _, ok := cw.inError[cluster.Name]; ok {
		glog.Infof("Cluster %s is back to a stable state", cluster.Name)
		delete(cw.inError, cluster.Name)
	}
}
