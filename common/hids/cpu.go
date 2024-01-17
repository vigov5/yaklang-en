package hids

import (
	"github.com/yaklang/yaklang/common/consts"
	"github.com/yaklang/yaklang/common/log"
	"github.com/yaklang/yaklang/common/yak/yaklib/codec"
	"github.com/yaklang/yaklang/common/yakgrpc/yakit"
)

const LASTCPUPERCENT_KEY = "LastCPUPercent"

// CPUPercentCallback When the CPU usage changes, call the callback function
// Example:
// ```
// hids.Init()
// hids.CPUPercentCallback(func(i) {
// if (i > 50) { println("cpu precent is over 50%") } // When the CPU usage exceeds 50%, output information
// })
// ```
func CPUPercentCallback(callback func(i float64)) {
	GetGlobalHealthManager().RegisterCPUPercentCallback(callback)
}

// CPUPercentCallback When the average CPU usage changes, call the callback function
// Example:
// ```
// hids.Init()
// hids.CPUAverageCallback(func(i) {
// if (i > 50) { println("cpu average precent is over 50%") } // When the average CPU usage exceeds 50%, output information
// })
// ```
func CPUAverageCallback(callback func(i float64)) {
	GetGlobalHealthManager().RegisterCPUAverageCallback(callback)
}

// CPUPercent to obtain the CPU usage of the current system.
// Example:
// ```
// printf("%f%%\n", hids.CPUPercent())
// ```
func CPUPercent() float64 {
	if info, err := SystemHealthStats(); err != nil {
		log.Errorf("cannot get system-health-stats, reason: %s", err)
		return 0
	} else {
		return info.CPUPercent
	}
}

// CPUAverage Get the average CPU usage of the current system
// Example:
// ```
// printf("%f%%\n", hids.CPUAverage())
// ```
func CPUAverage() float64 {
	if ret := codec.Atof(yakit.GetKey(consts.GetGormProfileDatabase(), LASTCPUPERCENT_KEY)); ret > 0 {
		return (CPUPercent() + ret) / 2.0
	}
	return CPUPercent()
}
