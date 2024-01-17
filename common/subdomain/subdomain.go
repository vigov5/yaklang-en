package subdomain

import (
	"context"
	"fmt"
	"github.com/miekg/dns"
	"github.com/pkg/errors"
	"github.com/yaklang/yaklang/common/log"
	"github.com/yaklang/yaklang/common/utils"
	"strings"
	"sync"
)

func qualifyDomain(domain string) string {
	return fmt.Sprintf("%s.", formatDomain(domain))
}

func formatDomain(target string) string {
	for strings.HasPrefix(target, ".") {
		target = target[1:]
	}
	return target
}

func removeRepeatedMode(modes ...int) []int {
	var ret sync.Map
	for _, m := range modes {
		switch m {
		case BRUTE:
		case SEARCH:
		case ZONE_TRANSFER:
		default:
			continue
		}
		ret.Store(m, 1)
	}

	modes = []int{}
	ret.Range(func(key, value interface{}) bool {
		modes = append(modes, key.(int))
		return true
	})
	return modes
}

func removeRepeatedTargets(targets ...string) []string {
	var ret sync.Map
	for _, m := range targets {
		if strings.TrimSpace(m) != "" {
			ret.Store(m, 1)
		}
	}

	targets = []string{}
	ret.Range(func(key, value interface{}) bool {
		targets = append(targets, key.(string))
		return true
	})
	return targets
}

type SubdomainScanner struct {
	logger *log.Logger

	targets []string
	config  *SubdomainScannerConfig

	dnsQuerierSwg *utils.SizedWaitGroup
	dnsClient     *dns.Client

	// Result callback function
	resultCallbacks []ResultCallback

	// Parse failure callback
	// Parse Failure is not called by blast, this call chain only involves search and domain transfer
	resultFailedCallbacks []ResultCallback

	resultCacher *sync.Map
}

func (s *SubdomainScanner) GetConfig() *SubdomainScannerConfig {
	return s.config
}

func NewSubdomainScannerWithLogger(config *SubdomainScannerConfig, logger *log.Logger, targets ...string) (*SubdomainScanner, error) {
	if config.WorkerCount <= 0 {
		config.WorkerCount = 50
	}

	client := &dns.Client{}
	client.Timeout = config.TimeoutForEachQuery

	if logger == nil {
		logger = log.DefaultLogger
	}

	return &SubdomainScanner{
		logger: logger,

		targets: targets,
		config:  config,

		dnsClient:     client,
		dnsQuerierSwg: utils.NewSizedWaitGroup(config.WorkerCount),

		resultCacher: new(sync.Map),
	}, nil
}

func NewSubdomainScanner(config *SubdomainScannerConfig, targets ...string) (*SubdomainScanner, error) {
	return NewSubdomainScannerWithLogger(config, nil, targets...)
}

// Append target
func (s *SubdomainScanner) Feed(targets ...string) {
	s.targets = append(s.targets, targets...)
}

func (s *SubdomainScanner) RunWithContext(ctx context.Context) error {
	if len(s.targets) <= 0 {
		return errors.New("empty targets list")
	}

	// Normalized modes
	modes := removeRepeatedMode(s.config.Modes...)
	if len(modes) <= 0 {
		return errors.New("subdomain scan modes is empty.")
	}

	// Limit target concurrency
	swg := utils.NewSizedWaitGroup(s.config.ParallelismTasksCount)
	defer swg.Wait()

	// Normalized targets
	for _, t := range removeRepeatedTargets(s.targets...) {
		// This swg is used to limit the concurrency of the entire target
		err := swg.AddWithContext(ctx)
		if err != nil || ctx.Err() != nil {
			return err
		}

		go func(target string) {
			defer swg.Done()

			// When detecting subdomain names for specific targets, you should use the one generated separately for the target. Context with TimeoutSeconds
			ctx, _ := context.WithTimeout(ctx, s.config.TimeoutForEachTarget)

			// Start goroutine for different modes Concurrency
			wg := utils.NewSizedWaitGroup(3)
			defer wg.Wait()
			for _, mode := range modes {

				// Use AddWithContext to safely cancel tasks in the queue
				err := wg.AddWithContext(ctx)
				if err != nil {
					return
				}

				log.Debugf("start target: %v mode: %v", target, mode)
				switch mode {
				case BRUTE:
					go func() {
						defer wg.Done()
						s.Brute(ctx, target)
					}()
					continue
				case SEARCH:
					go func() {
						defer wg.Done()

						s.Search(ctx, target)
					}()
				case ZONE_TRANSFER:
					go func() {
						defer wg.Done()

						s.ZoneTransfer(ctx, target)
					}()
				default:
					wg.Done()
				}
			}
		}(t)
	}

	return nil
}

func (s *SubdomainScanner) Run() error {
	return s.RunWithContext(context.Background())
}

// Set the callback for discovering subdomain names
type ResultCallback func(*SubdomainResult)

func (s *SubdomainScanner) OnResult(cb ResultCallback) {
	s.resultCallbacks = append(s.resultCallbacks, cb)
}

func (s *SubdomainScanner) onResult(result *SubdomainResult) {
	// If it is a cached domain name result, do not report it
	if _, ok := s.resultCacher.LoadOrStore(result.Hash(), result); ok {
		return
	}

	for _, cb := range s.resultCallbacks {
		cb(result)
	}
}

// Set the callback function for subdomain names that failed to resolve
func (s *SubdomainScanner) OnResolveFailedResult(cb ResultCallback) {
	s.resultFailedCallbacks = append(s.resultFailedCallbacks, cb)
}

func (s *SubdomainScanner) onResolveFailedResult(result *SubdomainResult) {
	for _, cb := range s.resultFailedCallbacks {
		cb(result)
	}
}
