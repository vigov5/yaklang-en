package hids

import (
	"fmt"
	"sync"
	"time"

	"github.com/yaklang/yaklang/common/log"
	"github.com/yaklang/yaklang/common/utils"
	"github.com/yaklang/yaklang/common/utils/healthinfo"
)

var (
	monitorDuration             = 5 * time.Second
	setGlobalHealthManagerMutex sync.Mutex
	_globalHealthManager        *healthinfo.Manager
)

func resetGlobalHealthManager() {
	setGlobalHealthManagerMutex.Lock()
	defer setGlobalHealthManagerMutex.Unlock()

	if _globalHealthManager != nil {
		_globalHealthManager.Cancel()
	}
	_globalHealthManager = nil
}

func setGlobalHealthManager(i *healthinfo.Manager) {
	resetGlobalHealthManager()

	setGlobalHealthManagerMutex.Lock()
	_globalHealthManager = i
	setGlobalHealthManagerMutex.Unlock()
}

func GetGlobalHealthManager() *healthinfo.Manager {
	if _globalHealthManager == nil {
		m, err := healthinfo.NewHealthInfoManager(monitorDuration, 30*time.Minute)
		if err != nil {
			log.Warnf("cannot create health-info-manager, reason: %s", err)
			return nil
		}
		setGlobalHealthManager(m)
		return m
	}
	return _globalHealthManager
}

// SetMonitorInterval Sets the monitoring interval (unit: seconds) of the global health manager. If called when the global health manager is running, it will reset the global health manager
// Example:
// ```
// hids.SetMonitorInterval(1)
// ```
func SetMonitorIntervalFloat(i float64) {
	if i < 1 {
		log.Warnf("invalid monitor-interval: %fs, at least 1s", i)
		return
	}
	monitorDuration = utils.FloatSecondDuration(i)

	if _globalHealthManager != nil {
		log.Info("monitor duration(interval) has been modified, reset health manager...")
		resetGlobalHealthManager()
		GetGlobalHealthManager()
	}
}

// Init Initializes the global health manager
// Example:
// ```
// hids.Init()
// ```
func InitHealthManager() {
	GetGlobalHealthManager()
}

// ShowMonitorInterval Outputs the global health manager in the standard output monitoring interval (unit: seconds)
// Example:
// ```
// hids.ShowMonitorInterval()
// ```
func ShowMonitorInterval() {
	fmt.Println(monitorDuration.String())
}
