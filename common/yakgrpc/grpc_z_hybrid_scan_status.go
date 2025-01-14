package yakgrpc

import (
	"fmt"
	"github.com/yaklang/yaklang/common/utils"
	"github.com/yaklang/yaklang/common/yak/yaklib/codec"
	"github.com/yaklang/yaklang/common/yakgrpc/yakit"
	"github.com/yaklang/yaklang/common/yakgrpc/ypb"
	"sync"
	"sync/atomic"
)

type HybridScanStatusManager struct {
	// tasks when resuming tasks
	TargetTotal    int64
	PluginTotal    int64
	TotalTaskCount int64

	// Completed tasks
	TargetFinished int64
	TaskFinished   int64

	// dynamic
	ActiveTask    int64
	ActiveTarget  int64
	ActiveTaskMap *sync.Map

	TaskId string

	// Task counter as an index
	TaskCount int64

	// uses the total number of
	minTaskCount int64
}

func (h *HybridScanStatusManager) SetCurrentTaskIndex(i int64) {
	h.minTaskCount = i
}

func newHybridScanStatusManager(id string, targets int, plugins int) *HybridScanStatusManager {
	return &HybridScanStatusManager{
		TargetTotal:   int64(targets),
		PluginTotal:   int64(plugins),
		TaskId:        id,
		ActiveTaskMap: new(sync.Map),
	}
}

func fitStatusToHybridScanTaskRecord(status *ypb.HybridScanResponse, task *yakit.HybridScanTask) {
	task.TotalTargets = status.TotalTargets
	task.TotalPlugins = status.TotalPlugins
	task.TotalTasks = status.TotalTasks
	task.FinishedTargets = status.FinishedTargets
	task.FinishedTasks = status.FinishedTasks
}

func (h *HybridScanStatusManager) GetStatus(r ...*yakit.HybridScanTask) *ypb.HybridScanResponse {
	status := &ypb.HybridScanResponse{
		TotalTargets:     h.TargetTotal,
		TotalPlugins:     h.PluginTotal,
		TotalTasks:       int64(h.TargetTotal) * int64(h.PluginTotal),
		FinishedTasks:    h.TaskFinished,
		FinishedTargets:  h.TargetFinished,
		ActiveTasks:      h.ActiveTask,
		ActiveTargets:    h.ActiveTarget,
		HybridScanTaskId: h.TaskId,
	}
	if h.minTaskCount > 0 {
		if status.FinishedTasks < h.minTaskCount {
			status.FinishedTasks = h.minTaskCount
		}
	}
	for _, data := range r {
		fitStatusToHybridScanTaskRecord(status, data)
	}
	return status
}

func (h *HybridScanStatusManager) DoActiveTarget() int64 {
	return atomic.AddInt64(&h.ActiveTarget, 1)
}

// DoActiveTask returns index of task
func (h *HybridScanStatusManager) DoActiveTask() int64 {
	atomic.AddInt64(&h.ActiveTask, 1)
	index := atomic.AddInt64(&h.TotalTaskCount, 1)
	h.ActiveTaskMap.Store(index, struct{}{})
	return index
}

func (h *HybridScanStatusManager) PushActiveTask(index int64, t *HybridScanTarget, pluginName string, stream HybridScanRequestStream) {
	rsp := h.GetStatus()
	rsp.UpdateActiveTask = &ypb.HybridScanUpdateActiveTaskTable{
		Operator:    "create",
		Index:       fmt.Sprint(index),
		IsHttps:     t.IsHttps,
		HTTPRequest: t.Request,
		Url:         utils.EscapeInvalidUTF8Byte([]byte(t.Url)),
		PluginName:  pluginName,
	}
	stream.Send(rsp)
}

func (h *HybridScanStatusManager) RemoveActiveTask(index int64, t *HybridScanTarget, pluginName string, stream HybridScanRequestStream) {
	rsp := h.GetStatus()
	rsp.UpdateActiveTask = &ypb.HybridScanUpdateActiveTaskTable{
		Operator:    "remove",
		Index:       fmt.Sprint(index),
		IsHttps:     t.IsHttps,
		HTTPRequest: t.Request,
		Url:         utils.EscapeInvalidUTF8Byte([]byte(t.Url)),
		PluginName:  pluginName,
	}
	stream.Send(rsp)
}

func (h *HybridScanStatusManager) DoneTask(index int64) {
	atomic.AddInt64(&h.TaskFinished, 1)
	atomic.AddInt64(&h.ActiveTask, -1)
	h.ActiveTaskMap.Delete(index)
}

func (h *HybridScanStatusManager) DoneTarget() {
	atomic.AddInt64(&h.TargetFinished, 1)
	atomic.AddInt64(&h.ActiveTarget, -1)
}

func (h *HybridScanStatusManager) Feedback(stream HybridScanRequestStream) error {
	return stream.Send(h.GetStatus())
}

func (h *HybridScanStatusManager) GetCurrentActiveTaskIndexes() []int {
	var vals []int
	h.ActiveTaskMap.Range(func(key, value any) bool {
		if ret := codec.Atoi(fmt.Sprint(key)); ret > 0 {
			vals = append(vals, ret)
		}
		return true
	})
	return vals
}
