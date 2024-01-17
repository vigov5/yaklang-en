package hids

var Exports = map[string]interface{}{
	// Basic settings
	"Init":                InitHealthManager,
	"SetMonitorInterval":  SetMonitorIntervalFloat,
	"ShowMonitorInterval": ShowMonitorInterval,

	// CPU metrics
	"CPUPercent":            CPUPercent,
	"MemoryPercent":         MemoryPercent,
	"CPUAverage":            CPUAverage,
	"CPUPercentCallback":    CPUPercentCallback,
	"CPUAverageCallback":    CPUAverageCallback,
	"MemoryPercentCallback": MemoryPercentCallback,
}
