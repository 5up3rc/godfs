package libclient

import (
	"app"
	"encoding/json"
	"libservicev2"
	"sync"
	"util/logger"
)

var (
	managedStatistic       = make(map[string][]app.StorageDO)
	managedTrackerInstance = make(map[string]*TrackerInstance)
	managedLock            = new(sync.Mutex)
)

// trackerUUID -> host:port
// TODO fileCount
func updateStatistic(trackerUUID string, fileCount int, statistic []app.StorageDO) {
	ret, _ := json.Marshal(statistic)
	logger.Info("update statistic info:( ", string(ret), ")")
	if statistic == nil || len(statistic) == 0 {
		return
	}
	managedStatistic[trackerUUID] = statistic
	libservicev2.SaveStorage(trackerUUID, statistic...)
}

// registerTrackerInstance register tracker instance when start track with tracker
func registerTrackerInstance(instance *TrackerInstance) {
	if instance == nil {
		return
	}
	managedLock.Lock()
	defer managedLock.Unlock()
	if managedTrackerInstance[instance.ConnStr] != nil {
		logger.Info("tracker instance already started, ignore registration")
	} else {
		logger.Debug("register tracker instance", instance.ConnStr)
		managedTrackerInstance[instance.ConnStr] = instance
	}
}

// unRegisterTrackerInstance when delete a tracker, agent must remove the tracker instance and disconnect from tracker
func unRegisterTrackerInstance(connStr string) {
	managedLock.Lock()
	defer managedLock.Unlock()
	if managedTrackerInstance[connStr] == nil {
		logger.Info("tracker instance not exists")
	} else {
		delete(managedTrackerInstance, connStr)
	}
}

// UpdateTrackerInstanceState update nextRun flag of tracker instance
// secret is optional
func UpdateTrackerInstanceState(connStr string, secret string, nextRun bool, trackerMaintainer *TrackerMaintainer) {
	managedLock.Lock()
	defer managedLock.Unlock()
	if managedTrackerInstance[connStr] == nil {
		logger.Info("tracker instance not exists")
		if nextRun {
			logger.Info("start new tracker instance:", connStr)
			temp := make(map[string]string)
			temp[connStr] = secret
			trackerMaintainer.Maintain(temp)
		}
	} else {
		// logger.Info("unload tracker instance:", connStr)
		ins := managedTrackerInstance[connStr]
		ins.nextRun = nextRun
	}
}

/*func SyncTrackerAliveStatus(trackerMaintainer *TrackerMaintainer) {
	timer := time.NewTicker(app.SyncStatisticInterval + 3)
	execTimes := 0
	for {
		common.Try(func() {
			trackers, e := libservice.GetAllWebTrackers()
			if e != nil {
				logger.Error(e)
			} else {
				if trackers != nil && trackers.Len() > 0 {
					for ele := trackers.Front(); ele != nil; ele = ele.Next() {
						tracker := ele.Value.(*bridge.WebTracker)
						UpdateTrackerInstanceState(tracker.Host+":"+strconv.Itoa(tracker.Port),
							tracker.Secret, tracker.Status == app.StatusEnabled, trackerMaintainer)
					}
				}
			}
		}, func(i interface{}) {
			logger.Error("error fetch web tracker status:", i)
		})
		execTimes++
		<-timer.C
	}
}
*/
