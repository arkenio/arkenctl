package main

import (
	"bytes"
	. "github.com/arkenio/goarken"
	"github.com/coreos/go-etcd/etcd"
	"github.com/golang/glog"
	metrics "github.com/rcrowley/go-metrics"
	datadog "github.com/vistarmedia/go-datadog"
	"os"
	"time"
	"fmt"
)

type ClusterWatcher struct {
	Watcher       *Watcher
	Client        *etcd.Client
	SingleRun     bool
	DataDogAPIKey string

	inError map[string]*ServiceCluster


	dog *datadog.Client

	errorsGauge     metrics.Gauge
	warningsGauge     metrics.Gauge
	startedGauge    metrics.Gauge
	passivatedGauge metrics.Gauge
}

func (cw *ClusterWatcher) Watch(stop chan interface{}) error {

	if cw.DataDogAPIKey != "" {
		cw.errorsGauge = metrics.NewGauge()
		cw.startedGauge = metrics.NewGauge()
		cw.passivatedGauge = metrics.NewGauge()
		cw.warningsGauge = metrics.NewGauge()

		metrics.Register("arken.environments.stats.errors", cw.errorsGauge)
		metrics.Register("arken.environments.stats.started", cw.startedGauge)
		metrics.Register("arken.environments.stats.passivated", cw.passivatedGauge)
		metrics.Register("arken.environments.stats.warning", cw.warningsGauge)

		host, _ := os.Hostname()
		cw.dog = datadog.New(host, cw.DataDogAPIKey)
		go cw.dog.DefaultReporter().Start(30 * time.Second)

	}

	go cw.updateMetrics()

	return cw.watchServiceKeys(stop)

}

func (cw *ClusterWatcher) updateMetrics() {
	interval := 30 * time.Second
	ticker := time.NewTicker(interval)

	for {
		select {
		case <-ticker.C:
			errors := int64(0)
			passivated := int64(0)
			started := int64(0)
			warning := int64(0)
			glog.Infof("Updating metrics...")
			for _, cluster := range cw.Watcher.Services {
				_, err := cw.Watcher.Services[cluster.Name].Next()
				if err != nil {
					if stError, ok := err.(StatusError); ok {
						switch stError.ComputedStatus {
						case PASSIVATED_STATUS:
							passivated++
							break
						case WARNING_STATUS:
							warning++
							break;
						case STARTING_STATUS, STOPPED_STATUS, STOPPING_STATUS:
							break
						default:
							// If status is nil, then we can't say it's an error... it's in an unknown status
							if stError.Status != nil {
								glog.Infof("Cluster in error : %s", cluster.Name)
								errors++
							}
						}
					} else {
						errors++
						glog.Infof("Cluster in error : %s", cluster.Name)
					}
				} else {
					started++
				}
			}
			glog.Infof("End metrics update...")

			cw.errorsGauge.Update(errors)
			cw.passivatedGauge.Update(passivated)
			cw.startedGauge.Update(started)
			cw.warningsGauge.Update(warning)

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
			case STARTING_STATUS, PASSIVATED_STATUS, STOPPED_STATUS, STOPPING_STATUS, WARNING_STATUS:
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

		cw.dog.PostEvent(&datadog.Event {
			Title: fmt.Sprintf("IO instance %s entered error state",cluster.Name),
			Text:      cw.getClusterDescriptionInMarkdown(cluster),
			Priority: "normal",
			Tags      : []string{fmt.Sprintf("ioinstance:%s", cluster.Name),"arkenwatch"},
			AlertType : "error",
		})

		cw.inError[cluster.Name] = cluster
	}
	var doc bytes.Buffer
	renderService(cluster, "", &doc)
	glog.Errorf(doc.String())
}


func (cw *ClusterWatcher) getClusterDescriptionInMarkdown(cluster *ServiceCluster) string {
	tpl := `%%%
{{range $index, $service := .GetInstances }}
# Instance : {{.Name}}

    * Name : {{.Name}}
    * UnitName : {{.UnitName }}
    * Etcd key : {{.NodeKey }}
    * Domain name : [https://{{.Domain}}/]()
    * Location : {{.Location.Host}}:{{.Location.Port}}
    * LastAccess: {{.LastAccess}}
    * Status: {{.Status.Compute}}
      * expected : {{.Status.Expected}}
      * current : {{.Status.Current}}
      * alive : {{.Status.Alive}}
{{end}}
%%%`

	var doc bytes.Buffer
	renderService(cluster, tpl, &doc)
	return doc.String()
}







func (cw *ClusterWatcher) removeInError(cluster *ServiceCluster) {
	if _, ok := cw.inError[cluster.Name]; ok {
		glog.Infof("Cluster %s is back to a stable state", cluster.Name)

		cw.dog.PostEvent(&datadog.Event {
			Title: fmt.Sprintf("IO instance %s recovered from error state",cluster.Name),
			Text:      cw.getClusterDescriptionInMarkdown(cluster),
			Priority: "normal",
			Tags      : []string{fmt.Sprintf("ioinstance:%s", cluster.Name),"arkenwatch"},
			AlertType : "info",
		})



		var doc bytes.Buffer
		renderService(cluster, "", &doc)
		glog.Errorf(doc.String())
		delete(cw.inError, cluster.Name)
	}
}
