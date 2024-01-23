package tools

import (
	"context"
	"fmt"
	filter2 "github.com/yaklang/yaklang/common/filter"
	"github.com/yaklang/yaklang/common/fp"
	"github.com/yaklang/yaklang/common/log"
	"github.com/yaklang/yaklang/common/synscan"
	"github.com/yaklang/yaklang/common/utils"
	"github.com/yaklang/yaklang/common/utils/pingutil"
	"github.com/yaklang/yaklang/common/utils/spacengine"
	"reflect"
	"strings"
)

func scanFingerprint(target string, port string, opts ...fp.ConfigOption) (chan *fp.MatchResult, error) {
	config := fp.NewConfig(opts...)
	return _scanFingerprint(context.Background(), config, 50, target, port)
}

func scanOneFingerprint(target string, port int, opts ...fp.ConfigOption) (*fp.MatchResult, error) {
	config := fp.NewConfig(opts...)
	matcher, err := fp.NewFingerprintMatcher(nil, config)
	if err != nil {
		return nil, err
	}
	return matcher.Match(target, port)
}

func _scanFingerprint(ctx context.Context, config *fp.Config, concurrent int, host, port string) (chan *fp.MatchResult, error) {
	matcher, err := fp.NewDefaultFingerprintMatcher(config)
	if err != nil {
		return nil, err
	}

	log.Infof("start to scan [%s] 's port: %s", host, port)

	if matcher.Config.PoolSize > 0 {
		concurrent = matcher.Config.PoolSize
	}

	filter := filter2.NewFilter()

	outC := make(chan *fp.MatchResult)
	go func() {
		swg := utils.NewSizedWaitGroup(concurrent)
		var portsInt = utils.ParseStringToPorts(port)
		for _, p := range portsInt {
			for _, hRaw := range utils.ParseStringToHosts(host) {
				h := utils.ExtractHost(hRaw)
				if h != hRaw {
					buildinHost, buildinPort, _ := utils.ParseStringToHostPort(hRaw)
					if buildinPort > 0 {
						swg.Add()
						go func() {
							defer swg.Done()
							proto, portWithoutProto := utils.ParsePortToProtoPort(buildinPort) // The protocol and port are separated here to facilitate later printing of logs
							addr := utils.HostPort(buildinHost, buildinPort)
							if filter.Exist(addr) {
								return
							}
							filter.Insert(addr)
							log.Infof("start task to scan: [%s://%s]", proto, utils.HostPort(buildinHost, portWithoutProto))
							result, err := matcher.MatchWithContext(ctx, buildinHost, buildinPort)
							if err != nil {
								if len(portsInt) <= 0 {
									if strings.Contains(fmt.Sprint(err), "excludeHosts/Ports") {
										return
									}
								} else {
									if strings.Contains(fmt.Sprint(err), "filtered by servicescan") {
										return
									}
								}
								log.Errorf("failed to scan [%s://%s]: %v", proto, utils.HostPort(buildinHost, portWithoutProto), err)
								return
							}

							outC <- result
						}()
					}
				}

				swg.Add()
				rawPort := p
				rawHost := h
				proto, portWithoutProto := utils.ParsePortToProtoPort(p) // The protocol and port are separated here to facilitate later printing of logs
				go func() {
					defer swg.Done()

					addr := utils.HostPort(rawHost, rawPort)
					if filter.Exist(addr) {
						return
					}
					filter.Insert(addr)

					log.Infof("start task to scan: [%s://%s]", proto, utils.HostPort(rawHost, portWithoutProto))
					result, err := matcher.MatchWithContext(ctx, rawHost, rawPort)
					if err != nil {
						log.Errorf("failed to scan [%s://%s]: %v", proto, utils.HostPort(rawHost, portWithoutProto), err)
						return
					}

					outC <- result
				}()
			}
		}
		go func() {
			swg.Wait()
			close(outC)
		}()
	}()

	return outC, nil
}

func _scanFromPingUtils(res chan *pingutil.PingResult, ports string, opts ...fp.ConfigOption) (chan *fp.MatchResult, error) {
	var synResults = make(chan *synscan.SynScanResult, 1000)
	portsInt := utils.ParseStringToPorts(ports)
	go func() {
		defer close(synResults)
		for result := range res {
			if !result.Ok {
				log.Errorf("%v is may not alive.", result.IP)
				continue
			}

			var hostRaw, portRaw, _ = utils.ParseStringToHostPort(result.IP)
			if portRaw > 0 {
				synResults <- &synscan.SynScanResult{Host: hostRaw, Port: portRaw}
			}

			//log.Errorf("%v is alive", result.IP)
			for _, port := range portsInt {
				synResults <- &synscan.SynScanResult{
					Host: result.IP,
					Port: port,
				}
			}
		}
	}()
	return _scanFromTargetStream(synResults, opts...)
}

func _scanFromTargetStream(res interface{}, opts ...fp.ConfigOption) (chan *fp.MatchResult, error) {
	var synResults = make(chan *synscan.SynScanResult, 1000)

	// generates scanning results.
	go func() {
		defer close(synResults)

		switch ret := res.(type) {
		case chan *synscan.SynScanResult:
			for r := range ret {
				synResults <- r
			}
		case []interface{}:
			for _, r := range ret {
				switch subResult := r.(type) {
				case *synscan.SynScanResult:
					synResults <- subResult
				case synscan.SynScanResult:
					synResults <- &subResult
				case string:
					host, port, err := utils.ParseStringToHostPort(subResult)
					if err != nil {
						continue
					}
					synResults <- &synscan.SynScanResult{
						Host: host,
						Port: port,
					}
				case *spacengine.NetSpaceEngineResult:
					host, port, err := utils.ParseStringToHostPort(subResult.Addr)
					if err != nil {
						continue
					}
					synResults <- &synscan.SynScanResult{
						Host: host,
						Port: port,
					}
				}

			}
		case []*synscan.SynScanResult:
			for _, r := range ret {
				synResults <- r
			}
		case chan *spacengine.NetSpaceEngineResult:
			for r := range ret {
				host, port, err := utils.ParseStringToHostPort(r.Addr)
				if err != nil {
					continue
				}
				synResults <- &synscan.SynScanResult{
					Host: host,
					Port: port,
				}
			}
		case []*spacengine.NetSpaceEngineResult:
			for _, r := range ret {
				host, port, err := utils.ParseStringToHostPort(r.Addr)
				if err != nil {
					continue
				}
				synResults <- &synscan.SynScanResult{
					Host: host,
					Port: port,
				}
			}
		case []string:
			for _, r := range ret {
				host, port, err := utils.ParseStringToHostPort(r)
				if err != nil {
					continue
				}
				synResults <- &synscan.SynScanResult{
					Host: host,
					Port: port,
				}
			}
		default:
			log.Errorf("not a valid param: %v", reflect.TypeOf(res))
		}
	}()

	// Scan
	config := fp.NewConfig(opts...)
	concurrent := config.PoolSize
	ctx := context.Background()

	matcher, err := fp.NewDefaultFingerprintMatcher(config)
	if err != nil {
		return nil, err
	}

	if matcher.Config.PoolSize > 0 {
		concurrent = matcher.Config.PoolSize
	}

	outC := make(chan *fp.MatchResult)
	go func() {
		swg := utils.NewSizedWaitGroup(concurrent)
		for synRes := range synResults {
			swg.Add()
			rawPort := synRes.Port
			rawHost := synRes.Host
			proto, portWithoutProto := utils.ParsePortToProtoPort(rawPort) // The protocol and port are separated here to facilitate later printing of logs
			go func() {
				defer swg.Done()

				log.Infof("start task to scan: [%s://%s]", proto, utils.HostPort(rawHost, portWithoutProto))
				result, err := matcher.MatchWithContext(ctx, rawHost, rawPort)
				if err != nil {
					log.Errorf("failed to scan [%s://%s]: %v", proto, utils.HostPort(rawHost, portWithoutProto), err)
					return
				}

				outC <- result
			}()
		}

		go func() {
			swg.Wait()
			close(outC)
		}()
	}()

	return outC, nil
}

var FingerprintScanExports = map[string]interface{}{
	"Scan":                scanFingerprint,
	"ScanOne":             scanOneFingerprint,
	"ScanFromSynResult":   _scanFromTargetStream,
	"ScanFromSpaceEngine": _scanFromTargetStream,
	"ScanFromPing":        _scanFromPingUtils,

	"proto": func(proto ...interface{}) fp.ConfigOption {
		return fp.WithTransportProtos(fp.ParseStringToProto(proto...)...)
	},

	// Overall scan concurrency
	"concurrent": fp.WithPoolSize,

	"excludePorts": fp.WithExcludePorts,
	"excludeHosts": fp.WithExcludeHosts,

	// Single request timeout
	"probeTimeout": fp.WithProbeTimeoutHumanRead,

	// proxies
	"proxy": fp.WithProxy,

	// enables caching.
	"cache":         fp.WithCache,
	"databaseCache": fp.WithDatabaseCache,

	// Use web fingerprint identification rules Scan
	"webRule": fp.WithWebFingerprintRule,

	// can use nmap rules to scan, or write nmap rules to scan.
	"nmapRule": fp.WithNmapRule,

	// nmap rule filtering, by rarity
	"nmapRarityMax": fp.WithRarityMax,

	// turns on active packet sending mode.
	"active": fp.WithActiveMode,

	// . How many packets can each service send actively?
	"maxProbes": fp.WithProbesMax,

	// In active packet sending mode, what is the amount of concurrency?
	"maxProbesConcurrent": fp.WithProbesConcurrentMax,

	// Specify and select the scan target protocol: refers to turning on web service scanning
	"web": func() fp.ConfigOption {
		return func(config *fp.Config) {
			config.OnlyEnableWebFingerprint = true
		}
	},

	// turns on nmap rule base.
	"service": func() fp.ConfigOption {
		return func(config *fp.Config) {
			config.DisableWebFingerprint = true
		}
	},

	// Scan all services
	"all": func() fp.ConfigOption {
		return func(config *fp.Config) {
			config.ForceEnableAllFingerprint = true
		}
	},
}
