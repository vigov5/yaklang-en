package bruteutils

import (
	"bufio"
	"bytes"
	"container/list"
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/yaklang/yaklang/common/log"
	"github.com/yaklang/yaklang/common/mutate"
	"github.com/yaklang/yaklang/common/utils"
	"github.com/yaklang/yaklang/common/utils/mixer"
	"io/ioutil"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type BruteItem struct {
	Type     string
	Target   string
	Username string
	Password string
}

func (b *BruteItem) Result() *BruteItemResult {
	return &BruteItemResult{
		Type:             b.Type,
		Ok:               false,
		Finished:         false,
		UserEliminated:   false,
		OnlyNeedPassword: false,
		Target:           b.Target,
		Username:         b.Username,
		Password:         b.Password,
	}
}

func (b *BruteItem) String() string {
	return fmt.Sprintf("%s:%s@%v", b.Username, b.Password, b.Target)
}

type targetProcessing struct {
	Target   string
	Swg      *utils.SizedWaitGroup
	count    int32
	Items    []*BruteItem
	Finished bool
}

func (t *targetProcessing) GetCurrentCount() int {
	return int(atomic.LoadInt32(&t.count))
}

func (t *targetProcessing) Finish() {
	t.Finished = true
}

type BruteUtil struct {
	processes            sync.Map
	TargetTaskConcurrent int

	targetsSwg *utils.SizedWaitGroup

	targetList     *list.List
	targetListLock sync.Mutex

	delayer *utils.DelayWaiter

	// Callbacks of the blasting task
	callback BruteCallback

	// Each execution ends. The result is returned
	resultCallback BruteItemResultCallback

	// This option marks that if OK is encountered, stop blasting the current target
	OkToStop bool

	// to complete the threshold, which is an integer.
	// During the blasting process, the number of Finished tasks will be counted.
	// Once the number of Finished results given by the task execution reaches the value set by this parameter
	// Immediately end the blasting of the current target.
	FinishingThreshold int

	// OnlyNeedPassword marks that this blast only requires Password blasting
	OnlyNeedPassword bool

	//
	beforeBruteCallback func(string) bool
}

func (b *BruteUtil) SetResultCallback(cb BruteItemResultCallback) {
	b.resultCallback = cb
}

type BruteItemResult struct {
	// Explosion type
	Type string

	// Marks blasting success
	Ok bool

	// marks the completion of blasting/Because of the protocol Wrong, or there is a network verification error, wait.
	Finished bool

	// indicates that there is a problem with the user name and should not be used again.
	UserEliminated bool

	// This blast only requires a password, not a user name.
	OnlyNeedPassword bool

	// The target of the blasting.
	Target string

	// Username and password for blasting
	Username string
	Password string

	// Banner basis for blasting results, additional information
	ExtraInfo []byte
}

func (r *BruteItemResult) String() string {
	var result = "FAIL"
	if r.Ok {
		result = "OK"
	} else {
		result = "FAIL"
	}
	return fmt.Sprintf("[%v]: %v:\\\\%v:%v@%v", result, r.Type, r.Username, r.Password, r.Target)
}

func (r *BruteItemResult) Show() {
	println(r.String())
}

type BruteCallback func(item *BruteItem) *BruteItemResult
type BruteItemResultCallback func(b *BruteItemResult)

func NewMultiTargetBruteUtil(targetsConcurrent, minDelay, maxDelay int, callback BruteCallback) (*BruteUtil, error) {
	delayer, err := utils.NewDelayWaiter(int32(minDelay), int32(maxDelay))
	if err != nil {
		return nil, errors.Errorf("create delayer failed: %s", err)
	}
	// first is 0 delay
	delayer.Wait()
	return &BruteUtil{
		TargetTaskConcurrent: 1,
		targetList:           list.New(),
		targetsSwg:           utils.NewSizedWaitGroup(targetsConcurrent),
		callback:             callback,
		delayer:              delayer,
	}, nil
}

func (b *BruteUtil) Feed(item *BruteItem) {
	process, err := b.GetProcessingByTarget(item.Target)
	if err != nil {
		// new target
		swg := utils.NewSizedWaitGroup(b.TargetTaskConcurrent)
		process = &targetProcessing{
			Target: item.Target,
			Swg:    swg,
		}
		b.targetList.PushBack(item.Target)
		b.processes.Store(item.Target, process)
	}

	process.Items = append(process.Items, item)
}

func (b *BruteUtil) GetProcessingByTarget(target string) (*targetProcessing, error) {
	if raw, ok := b.processes.Load(target); ok {
		return raw.(*targetProcessing), nil
	} else {
		return nil, errors.New("no such target")
	}
}

func (b *BruteUtil) GetAllTargetsProcessing() []*targetProcessing {
	var ct []*targetProcessing
	b.processes.Range(func(key, value interface{}) bool {
		p := value.(*targetProcessing)
		ct = append(ct, p)
		return true
	})
	return ct
}

func (b *BruteUtil) RemoteProcessingByTarget(target string) {
	b.processes.Delete(target)
}

func (b *BruteUtil) Run() error {
	return b.run(context.Background())
}

func (b *BruteUtil) RunWithContext(ctx context.Context) error {
	return b.run(ctx)
}

func (b *BruteUtil) run(ctx context.Context) error {
	defer b.targetsSwg.Wait()

	for {
		target, err := b.popFirstTarget()
		if err != nil {
			log.Trace("finished poping target from target list")
			break
		}

		if target == "" {
			continue
		}

		// context cancel
		if err := ctx.Err(); err != nil {
			log.Info("context canceled")
			return errors.New("user canceled")
		}

		err = b.targetsSwg.AddWithContext(ctx)
		if err != nil {
			break
		}

		go func(t string) {
			defer b.targetsSwg.Done()

			log.Tracef("start processing for target: %s", t)
			err := b.startProcessingTarget(t, ctx)
			if err != nil {
				log.Errorf("start processing brute target failed: %s", err)
			}
		}(target)
	}
	return nil
}

func (b *BruteUtil) startProcessingTarget(target string, parentCtx context.Context) error {
	currCtx, cancel := context.WithCancel(parentCtx)
	defer cancel()
	defer func() {
		go func() {
			select {
			case <-time.NewTimer(5 * time.Second).C:
				b.RemoteProcessingByTarget(target)
			}
		}()
	}()

	process, err := b.GetProcessingByTarget(target)
	if err != nil {
		return errors.Errorf("start processing target failed: %s", err)
	}
	defer func() {
		process.Swg.Wait()
		process.Finish()
	}()

	var (
		finishedCount int32 = 0
		//finished               = utils.NewBool(false)
		onlyNeedPassword = utils.NewBool(b.OnlyNeedPassword)
		eliminatedUsers  = sync.Map{}
		usedPassword     = sync.Map{}
	)

	// Do a pre-blasting check to check the rationality of the target. If it is unreasonable, end it immediately
	// Usually contains the following parts:
	//    1. Check target plausibility
	//    2. Check the target fingerprint
	if b.beforeBruteCallback != nil {
		if !b.beforeBruteCallback(target) {
			return errors.Errorf("pre-checking target[%s] failed", target)
		}
	}

	for _, i := range process.Items {
		if err := currCtx.Err(); err != nil {
			return errors.New("context canceled")
		}

		//// Exit blasting
		//if finished.IsSet() {
		//	break
		//}

		// Calculate the number of times the subtask requires exiting blasting
		if atomic.LoadInt32(&finishedCount) >= int32(b.FinishingThreshold) && b.FinishingThreshold != 0 {
			break
		}

		// If the blasting only requires a password and does not require a user name
		if onlyNeedPassword.IsSet() {
			if _, ok := usedPassword.Load(i.Password); ok {
				// If this password has been used, immediately enter the next set of
				continue
			} else {
				// If this password has not been used, record the password and the next
				usedPassword.Store(i.Password, 1)
			}
		}

		err := process.Swg.AddWithContext(currCtx)
		if err != nil {
			return nil
		}

		i := i
		go func(item *BruteItem) {
			defer func() {
				process.Swg.Done()
				atomic.AddInt32(&process.count, 1)
			}()

			// Check whether the context has been cancelled.
			if err := currCtx.Err(); err != nil {
				return
			}

			// Abandoned user name
			if _, ok := eliminatedUsers.Load(item.Username); ok {
				// If the user name is discarded, the blasting of the task should not be started directly
				return
			}

			// Execute the blast function
			result := b.callback(item)
			if result == nil {
				return
			}

			if b.resultCallback != nil {
				b.resultCallback(result)
			}

			// Have you encountered a situation where the blasting was successful?
			if result.Ok && b.OkToStop {
				//finished.Set()
				cancel()
			}

			// Is the current result complete?
			if result.Finished {
				atomic.AddInt32(&finishedCount, 1)
			}

			// Is there any result and it is found that this target only requires a password?
			if result.OnlyNeedPassword {
				onlyNeedPassword.Set()
			}

			// Make sure that the current user name is an obsolete user name. The current user name will no longer be used for the current target.
			if result.UserEliminated {
				eliminatedUsers.Store(item.Username, 1)
			}
			b.delayer.Wait()
		}(i)
	}

	log.Tracef("finished handling target: %s", target)
	return nil
}

func (b *BruteUtil) popFirstTarget() (string, error) {
	b.targetListLock.Lock()
	defer b.targetListLock.Unlock()

	e := b.targetList.Front()
	if e == nil {
		return "", errors.New("emtpy targets")
	}

	defer func() {
		_ = b.targetList.Remove(e)
	}()

	return e.Value.(string), nil
}

// Use a more reasonable interface to build BruteUtil

type OptionsAction func(util *BruteUtil)

// This option controls the overall target concurrency. The default value is 200
func WithTargetsConcurrent(targetsConcurrent int) OptionsAction {
	return func(util *BruteUtil) {
		util.targetsSwg = utils.NewSizedWaitGroup(targetsConcurrent)
	}
}

// This option is used Controls how many blasting tasks each target can perform at the same time. The default is 1.
func WithTargetTasksConcurrent(targetTasksConcurrent int) OptionsAction {
	return func(util *BruteUtil) {
		util.TargetTaskConcurrent = targetTasksConcurrent
	}
}

// This option controls the settings Delayer
func WithDelayerWaiter(minDelay, maxDelay int) (OptionsAction, error) {
	dlr, err := utils.NewDelayWaiter(int32(minDelay), int32(maxDelay))
	if err != nil {
		return nil, errors.Errorf("delay waiter build failed: %s", err)
	}
	return func(util *BruteUtil) {
		util.delayer = dlr
	}, nil
}

// Set up the blasting task
func WithBruteCallback(callback BruteCallback) OptionsAction {
	return func(util *BruteUtil) {
		util.callback = callback
	}
}

// setting results. Callback
func WithResultCallback(callback BruteItemResultCallback) OptionsAction {
	return func(util *BruteUtil) {
		util.resultCallback = callback
	}
}

// Set OkToStop option
func WithOkToStop(t bool) OptionsAction {
	return func(util *BruteUtil) {
		util.OkToStop = t
	}
}

// Set the threshold
func WithFinishingThreshold(t int) OptionsAction {
	return func(util *BruteUtil) {
		util.FinishingThreshold = t
	}
}

// Set up only password blasting
func WithOnlyNeedPassword(t bool) OptionsAction {
	return func(util *BruteUtil) {
		util.OnlyNeedPassword = t
	}
}

// Set blasting pre-check function
func WithBeforeBruteCallback(c func(string) bool) OptionsAction {
	return func(util *BruteUtil) {
		util.beforeBruteCallback = c
	}
}

func NewMultiTargetBruteUtilEx(options ...OptionsAction) (*BruteUtil, error) {
	delayer, err := utils.NewDelayWaiter(0, 0)
	if err != nil {
		return nil, errors.Errorf("init delay waiter failed: %s", err)
	}

	bu := &BruteUtil{
		TargetTaskConcurrent: 1,
		targetsSwg:           utils.NewSizedWaitGroup(200),
		OkToStop:             false,
		FinishingThreshold:   0,
		OnlyNeedPassword:     false,
		delayer:              delayer,
		targetList:           list.New(),
	}

	for _, option := range options {
		option(bu)
	}

	if bu.callback == nil {
		return nil, errors.New("callback is not set")
	}
	return bu, nil
}

func BruteItemStreamWithContext(ctx context.Context, typeStr string, target []string, users []string, pass []string) (chan *BruteItem, error) {
	mixerIns, err := mixer.NewMixer(target, pass, users)
	if err != nil {
		return nil, utils.Errorf("create target/user/password mixer failed: %s", err)
	}

	ch := make(chan *BruteItem)
	go func() {
		defer close(ch)

		for {
			result := mixerIns.Value()
			select {
			case ch <- &BruteItem{
				Type:     typeStr,
				Target:   result[0],
				Password: result[1],
				Username: result[2],
			}:
			case <-ctx.Done():
				return
			}
			if err := mixerIns.Next(); err != nil {
				return
			}
		}
	}()
	return ch, nil
}

func FileOrMutateTemplateForStrings(divider string, t ...string) []string {
	var r []string
	for _, item := range t {
		r = append(r, FileOrMutateTemplate(item, divider)...)
	}
	return r
}

func FileOrMutateTemplate(t string, divider string) []string {
	targetList := FileToDictList(t)

	if targetList == nil {
		for _, user := range utils.PrettifyListFromStringSplited(t, divider) {
			_l, err := mutate.QuickMutate(user, nil)
			if err != nil {
				continue
			}
			targetList = append(targetList, _l...)
		}
	}

	if targetList == nil {
		targetList = append(targetList, t)
	}
	return targetList
}

func FileToDictList(fileName string) []string {
	fd, err := ioutil.ReadFile(fileName)
	if err != nil {
		log.Error(err)
		return nil
	}
	scanner := bufio.NewScanner(bytes.NewBuffer(fd))
	scanner.Split(bufio.ScanLines)

	var lines []string
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		lines = append(lines, line)
	}
	return lines
}

func (b *BruteUtil) StreamBruteContext(
	ctx context.Context, typeStr string, target, users, pass []string,
	resultCallback BruteItemResultCallback,
) error {
	ch, err := BruteItemStreamWithContext(ctx, typeStr, target, users, pass)
	if err != nil {
		return err
	}
	b.SetResultCallback(resultCallback)
	log.Infof("brute task with target[%v] user[%v] password[%v]", len(target), len(users), len(pass))
	for item := range ch {
		b.Feed(item)
	}
	err = b.RunWithContext(ctx)
	if err != nil {
		return err
	}
	return nil
}

func autoSetFinishedByConnectionError(err error, result *BruteItemResult) *BruteItemResult {
	switch true {
	case utils.IContains(err.Error(), "connect: connection refused"):
		fallthrough
	case utils.IContains(err.Error(), "no pg_hba.conf entry for host"):
		fallthrough
	case utils.IContains(err.Error(), "network unreachable"):
		fallthrough
	case utils.IContains(err.Error(), "network is unreachable"):
		fallthrough
	//case utils.IContains(err.Error(), "remote error: tls: access denied"):
	//	fallthrough
	case utils.IContains(err.Error(), "no reachable servers"):
		fallthrough
	case utils.IContains(err.Error(), "i/o timeout"):
		result.Finished = true
		return result
	default:
		log.Error(err.Error())
		return result
	}
}
