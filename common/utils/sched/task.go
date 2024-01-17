package sched

import (
	"context"
	"github.com/pkg/errors"
	"github.com/yaklang/yaklang/common/log"
	"github.com/yaklang/yaklang/common/utils"
	"time"
)

type Task struct {
	// Task interval
	interval time.Duration

	// Task ID
	ID string

	// Task startup time
	Start time.Time

	// The stop time of the task
	End time.Time

	// The task execution function
	f func()

	isDisabled *utils.AtomicBool

	// Has it been executed?
	isExecuted *utils.AtomicBool

	// Is it running?
	isWorking *utils.AtomicBool

	// Is the schedule effective?
	isScheduling *utils.AtomicBool

	// will be executed immediately. Has it ended?
	isFinished *utils.AtomicBool

	// context
	ctx context.Context

	// will only be executed once. Cancel the task
	cancel context.CancelFunc

	// Is it executed for the first time?
	first *utils.AtomicBool

	// Last execution time and next execution time
	last, next time.Time

	// Hook function
	onFinished        map[string]TaskCallback
	onBeforeExecuting map[string]TaskCallback
	onEveryExecuted   map[string]TaskCallback
	onScheduleStart   map[string]TaskCallback
	onCanceled        map[string]TaskCallback
}

func NewTask(interval time.Duration, id string, start, end time.Time, f func(), first bool) *Task {
	return &Task{
		interval: interval, ID: id, Start: start, End: end, f: f,
		isExecuted:        utils.NewAtomicBool(),
		isWorking:         utils.NewAtomicBool(),
		isScheduling:      utils.NewAtomicBool(),
		isFinished:        utils.NewAtomicBool(),
		isDisabled:        utils.NewAtomicBool(),
		ctx:               context.Background(),
		first:             utils.NewBool(first),
		onFinished:        make(map[string]TaskCallback),
		onBeforeExecuting: make(map[string]TaskCallback),
		onCanceled:        make(map[string]TaskCallback),
		onScheduleStart:   make(map[string]TaskCallback),
		onEveryExecuted:   make(map[string]TaskCallback),
	}
}

func (t *Task) JustExecuteNotRecording() {
	t.f()
}

func (t *Task) SetDisabled(b bool) {
	t.isDisabled.SetTo(b)
}

func (t *Task) runWithContext(ctx context.Context) {
	t.isScheduling.Set()
	callbackLock.Lock()
	for _, f := range t.onScheduleStart {
		f(t)
	}
	callbackLock.Unlock()

	defer func() {
		t.isScheduling.UnSet()
		t.isFinished.Set()

		callbackLock.Lock()
		for _, f := range t.onFinished {
			f(t)
		}
		callbackLock.Unlock()
	}()

	// Set hook to record the last execution time
	taskFunc := func() {
		t.isWorking.Set()
		t.last = time.Now()

		// Set the callback function before executing the task
		callbackLock.Lock()
		for _, f := range t.onBeforeExecuting {
			f(t)
		}
		callbackLock.Unlock()

		// . If it has been disabled,
		if !t.isDisabled.IsSet() {
			t.f()
		}

		callbackLock.Lock()
		for _, f := range t.onEveryExecuted {
			f(t)
		}
		t.next = time.Now().Add(t.interval)
		callbackLock.Unlock()

		t.isWorking.UnSet()
	}

	// Set time execution time
	var taskCtx = ctx
	if t.End.After(time.Now()) {
		taskCtx, _ = context.WithDeadline(ctx, t.End)
	}

	if t.Start.After(time.Now()) {
		startCtx, _ := context.WithDeadline(ctx, t.Start)
		select {
		case <-startCtx.Done():
			break
		case <-ctx.Done():
			callbackLock.Lock()
			for _, f := range t.onCanceled {
				f(t)
			}
			callbackLock.Unlock()
			return
		}
	}

	// cannot be executed. If the first execution is set,
	if t.first.IsSet() {
		taskFunc()
	}

	// Enter loop mode
	ticker := time.Tick(t.interval)
	for {
		if t.isFinished.IsSet() || !t.isScheduling.IsSet() {
			if t.cancel != nil {
				t.cancel()
			}
			return
		}

		select {
		case <-taskCtx.Done():
			callbackLock.Lock()
			for _, f := range t.onCanceled {
				f(t)
			}
			callbackLock.Unlock()
			return
		case <-ticker:
			taskFunc()
		}
	}
}

func (t *Task) ExecuteWithContext(ctx context.Context) error {
	if t.isExecuted.IsSet() {
		return errors.Errorf("execute failed: %s is executed", t.ID)
	}
	t.isExecuted.Set()

	var c context.Context
	c, t.cancel = context.WithCancel(ctx)
	go t.runWithContext(c)
	return nil
}

func (t *Task) Execute() error {
	return t.ExecuteWithContext(t.ctx)
}

func (t *Task) Cancel() {
	log.Infof("schedule task in memory: %v is canceled", t.ID)
	if t.cancel != nil {
		t.cancel()
	}
	t.isScheduling.UnSet()
	t.isFinished.Set()
}

// Status function
func (t *Task) IsFinished() bool {
	return t.isFinished.IsSet()
}

func (t *Task) IsDisabled() bool {
	return t.isDisabled.IsSet()
}

func (t *Task) GetIntervalSeconds() int64 {
	return int64(t.interval.Seconds())
}

func (t *Task) IsExecuted() bool {
	return t.isExecuted.IsSet()
}

func (t *Task) IsWorking() bool {
	return t.isWorking.IsSet()
}

func (t *Task) IsInScheduling() bool {
	return t.isScheduling.IsSet()
}

// Other parameters
func (t *Task) LastExecutedDate() (time.Time, error) {
	if t.last.IsZero() {
		return time.Time{}, errors.New("not executed yet")
	}
	return t.last, nil
}

func (t *Task) NextExecuteDate() (time.Time, error) {
	if !t.IsInScheduling() || t.IsFinished() {
		return t.next, nil
	}
	return time.Time{}, errors.New("sched is finished")
}
