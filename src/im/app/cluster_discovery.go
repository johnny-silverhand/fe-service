package app

import (
	"fmt"
	"time"

	"im/mlog"
	"im/model"
)

const (
	DISCOVERY_SERVICE_WRITE_PING = 60 * time.Second
)

type ClusterDiscoveryService struct {
	model.ClusterDiscovery
	app  *App
	stop chan bool
}

func (a *App) NewClusterDiscoveryService() *ClusterDiscoveryService {
	ds := &ClusterDiscoveryService{
		ClusterDiscovery: model.ClusterDiscovery{},
		app:              a,
		stop:             make(chan bool),
	}

	return ds
}

func (me *ClusterDiscoveryService) Start() {

	<-me.app.Srv.Store.ClusterDiscovery().Cleanup()

	if cresult := <-me.app.Srv.Store.ClusterDiscovery().Exists(&me.ClusterDiscovery); cresult.Err != nil {
		mlog.Error(fmt.Sprintf("ClusterDiscoveryService failed to check if row exists for %v with err=%v", me.ClusterDiscovery.ToJson(), cresult.Err))
	} else {
		if cresult.Data.(bool) {
			if u := <-me.app.Srv.Store.ClusterDiscovery().Delete(&me.ClusterDiscovery); u.Err != nil {
				mlog.Error(fmt.Sprintf("ClusterDiscoveryService failed to start clean for %v with err=%v", me.ClusterDiscovery.ToJson(), u.Err))
			}
		}
	}

	if result := <-me.app.Srv.Store.ClusterDiscovery().Save(&me.ClusterDiscovery); result.Err != nil {
		mlog.Error(fmt.Sprintf("ClusterDiscoveryService failed to save for %v with err=%v", me.ClusterDiscovery.ToJson(), result.Err))
		return
	}

	go func() {
		mlog.Debug(fmt.Sprintf("ClusterDiscoveryService ping writer started for %v", me.ClusterDiscovery.ToJson()))
		ticker := time.NewTicker(DISCOVERY_SERVICE_WRITE_PING)
		defer func() {
			ticker.Stop()
			if u := <-me.app.Srv.Store.ClusterDiscovery().Delete(&me.ClusterDiscovery); u.Err != nil {
				mlog.Error(fmt.Sprintf("ClusterDiscoveryService failed to cleanup for %v with err=%v", me.ClusterDiscovery.ToJson(), u.Err))
			}
			mlog.Debug(fmt.Sprintf("ClusterDiscoveryService ping writer stopped for %v", me.ClusterDiscovery.ToJson()))
		}()

		for {
			select {
			case <-ticker.C:
				if u := <-me.app.Srv.Store.ClusterDiscovery().SetLastPingAt(&me.ClusterDiscovery); u.Err != nil {
					mlog.Error(fmt.Sprintf("ClusterDiscoveryService failed to write ping for %v with err=%v", me.ClusterDiscovery.ToJson(), u.Err))
				}
			case <-me.stop:
				return
			}
		}
	}()
}

func (me *ClusterDiscoveryService) Stop() {
	me.stop <- true
}

func (a *App) IsLeader() bool {

	//return a.Cluster.IsLeader()

	return true
}

func (a *App) GetClusterId() string {
	if a.Cluster == nil {
		return ""
	}

	return a.Cluster.GetClusterId()
}
