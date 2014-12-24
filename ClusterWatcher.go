package main

import (
	. "github.com/arkenio/goarken"
	"github.com/coreos/go-etcd/etcd"
	"github.com/golang/glog"
	metrics "github.com/rcrowley/go-metrics"
	datadog "github.com/vistarmedia/go-datadog"
	"os"
	"time"
)

type ClusterWatcher struct {
	Watcher       *Watcher
	Client        *etcd.Client
	SingleRun     bool
	DataDogAPIKey string

	inError map[string]*ServiceCluster

	errorsGauge     metrics.Gauge
	startedGauge    metrics.Gauge
	passivatedGauge metrics.Gauge
}

func (cw *ClusterWatcher) Watch(stop chan interface{}) error {

	if cw.DataDogAPIKey != "" {
		cw.errorsGauge = metrics.NewGauge()
		cw.startedGauge = metrics.NewGauge()
		cw.passivatedGauge = metrics.NewGauge()

		metrics.Register("arken.environments.stats.errors", cw.errorsGauge)
		metrics.Register("arken.environments.stats.started", cw.startedGauge)
		metrics.Register("arken.environments.stats.passivated", cw.passivatedGauge)

		host, _ := os.Hostname()
		dog := datadog.New(host, cw.DataDogAPIKey)
		go dog.DefaultReporter().Start(60 * time.Second)

	}

	go cw.updateMetrics()

	return cw.watchServiceKeys(stop)

}

func (cw *ClusterWatcher) updateMetrics() {
	interval := time.Minute
	ticker := time.NewTicker(interval)

	for {
		select {
		case <-ticker.C:
			errors := int64(0)
			passivated := int64(0)
			started := int64(0)
			for _, cluster := range cw.Watcher.Services {
				_, err := cw.Watcher.Services[cluster.Name].Next()
				if err != nil {
					if stError, ok := err.(StatusError); ok {
						switch stError.ComputedStatus {
						case PASSIVATED_STATUS:
							passivated++
						case STARTING_STATUS, STOPPED_STATUS, STOPPING_STATUS:
							break
						default:
							// If status is nil, then we can't say it's an error... it's in an unknown status
							if stError.Status != nil {
								errors++
							}
						}
					} else {
						errors++
					}
				}
				started++
			}

			cw.errorsGauge.Update(errors)
			cw.passivatedGauge.Update(passivated)
			cw.startedGauge.Update(started)

			ticker = time.NewTicker(interval)
		}
	}
}

func (cw *ClusterWatcher) watchServiceKeys(stop chan interface{}) error {
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
