package hids

import (
	"github.com/yaklang/yaklang/common/log"
	"github.com/yaklang/yaklang/common/spec/health"
	"github.com/yaklang/yaklang/common/utils"
	"github.com/yaklang/yaklang/common/utils/healthinfo"
)

func SystemHealthStats() (*health.HealthInfo, error) {
	return healthinfo.NewHealthInfo(utils.TimeoutContextSeconds(3))
}

// MemoryPercent to obtain the memory usage of the current system
// Example:
// ```
// printf("%f%%\n", hids.MemoryPercent())
// ```
func MemoryPercent() float64 {
	if info, err := SystemHealthStats(); err != nil {
		log.Errorf("cannot get system-health-stats, reason: %s", err)
		return 0
	} else {
		return info.MemoryPercent
	}
}

// MemoryPercentCallback When memory usage changes, call callback
// Example:
// ```
// hids.Init()
// hids.MemoryPercentCallback(func(i) {
// if (i > 50) { println("memory precent is over 50%") } // Output information when memory usage exceeds 50%
// })
// ```
func MemoryPercentCallback(callback func(i float64)) {
	GetGlobalHealthManager().RegisterMemPercentCallback(callback)
}
