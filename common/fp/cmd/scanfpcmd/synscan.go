package scanfpcmd

import (
	"context"
	"fmt"
	"github.com/urfave/cli"
	"github.com/yaklang/yaklang/common/fp"
	"github.com/yaklang/yaklang/common/fp/webfingerprint"
	"github.com/yaklang/yaklang/common/hybridscan"
	"github.com/yaklang/yaklang/common/log"
	"github.com/yaklang/yaklang/common/synscan"
	"github.com/yaklang/yaklang/common/utils"
	"net"
	"os"
	"sync"
	"time"
)

var SynScanCmd = cli.Command{
	Name:      "synscan",
	ShortName: "syn",
	Usage:     "SYN port scan",
	Before:    nil,
	After:     nil,

	OnUsageError: nil,
	Subcommands:  nil,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name: "target,host,t",
		},
		cli.StringFlag{
			Name:  "port,p",
			Value: "22,80,443,3389,3306,8080-8082,9000-9002,7000-7002",
		},
		cli.IntFlag{
			Name:  "wait,waiting",
			Usage: "How long to wait for finishing after the SYN packet is sent (Seconds)",
			Value: 5,
		},

		// fingerprint recognition related configuration
		cli.BoolFlag{
			Name:  "fingerprint,fp,x",
			Usage: "Enable fingerprint scanning",
		},
		cli.IntFlag{
			Name:  "request-timeout",
			Usage: "Timeout for a single request (Seconds)",
			Value: 10,
		},
		cli.StringFlag{
			Name:  "rule-path,rule,r",
			Usage: "Manually load rule file/Folder",
		},
		cli.BoolFlag{
			Name:  "only-rule",
			Usage: "Only load web fingerprints in this folder",
		},
		cli.StringFlag{
			Name:  "fp-json,fpo",
			Usage: "Detailed results output json to file",
		},

		// output real-time Open port information
		cli.StringFlag{
			Name:  "output",
			Usage: "output port open information to file",
		},

		cli.StringFlag{
			Name:  "output-line-prefix",
			Value: "",
			Usage: "Output the prefix of each line of OUTPUT, for example: https:// http://",
		},

		cli.IntFlag{
			Name:  "fingerprint-concurrent,fc",
			Value: 20,
			Usage: "Set the concurrency of fingerprint scanning (how many fingerprint scanning modules are performed at the same time)",
		},
	},

	Action: func(c *cli.Context) {
		target := c.String("target")
		targetList := utils.ParseStringToHosts(target)
		if len(targetList) <= 0 {
			log.Errorf("empty target: %s", c.String("target"))
			return
		}

		var sampleTarget string
		if len(targetList) == 1 {
			sampleTarget = targetList[0]
		} else {
			for _, target := range targetList {
				if !utils.IsLoopback(target) {
					sampleTarget = target
					break
				}
			}
			if sampleTarget == "" {
				sampleTarget = targetList[1]
			}
		}

		options, err := synscan.CreateConfigOptionsByTargetNetworkOrDomain(sampleTarget, 10*time.Second)
		if err != nil {
			log.Errorf("init syn scanner failed: %s", err)
			return
		}
		synScanConfig, err := synscan.NewConfig(options...)
		if err != nil {
			log.Errorf("create synscan config failed: %s", err)
			return
		}

		log.Infof("default config: \n    iface:%v src:%v gateway:%v", synScanConfig.Iface.Name, synScanConfig.SourceIP, synScanConfig.GatewayIP)

		// Analyze fingerprint configuration
		// web rule
		webRules, _ := fp.GetDefaultWebFingerprintRules()
		userRule := webfingerprint.FileOrDirToWebRules(c.String("rule-path"))

		if c.Bool("only-rule") {
			webRules = userRule
		} else {
			webRules = append(webRules, userRule...)
		}

		fingerprintMatchConfigOptions := []fp.ConfigOption{
			// active detection mode - proactively sending qualified packets
			fp.WithActiveMode(true),

			// Timeout for each fingerprint detection request
			fp.WithProbeTimeout(time.Second * time.Duration(c.Int("request-timeout"))),

			// web fingerprint full firepower
			fp.WithWebFingerprintUseAllRules(true),

			// web fingerprints
			fp.WithWebFingerprintRule(webRules),

			// Turn on web fingerprinting
			fp.WithForceEnableAllFingerprint(true),

			// enable TCP scan
			fp.WithTransportProtos(fp.TCP),
		}
		fpConfig := fp.NewConfig(fingerprintMatchConfigOptions...)

		scanCenterConfig, err := hybridscan.NewDefaultConfigWithSynScanConfig(synScanConfig)
		if err != nil {
			log.Error("default config failed: %s", err)
			return
		}

		// fingerprint scan switch
		// Scan fingerprint scanning separately
		scanCenterConfig.DisableFingerprintMatch = true

		log.Info("start create hyper scan center...")
		scanCenter, err := hybridscan.NewHyperScanCenter(context.Background(), scanCenterConfig)
		if err != nil {
			log.Error(err)
			return
		}

		log.Info("preparing for result collectors")
		var fpLock = new(sync.Mutex)
		var openPortLock = new(sync.Mutex)

		var fpResults []*fp.MatchResult
		var openPortCount int
		var openResult []string

		//// distribution task and callback function
		//err = scanCenter.RegisterMatcherResultHandler("cmd", func(matcherResult *fp.MatchResult, err error) {
		//	fpLock.Lock()
		//	defer fpLock.Unlock()
		//
		//	fpCount++
		//
		//	if matcherResult != nil {
		//		fpResults = append(fpResults, matcherResult)
		//		log.Infof("found open port fp -> %v", utils.HostPort(matcherResult.Target, matcherResult.Port))
		//	}
		//})
		//if err != nil {
		//	log.Error(err)
		//	return
		//}

		// outputfile
		var outputFile *os.File
		if c.String("output") != "" {
			outputFile, err = os.OpenFile(c.String("output"), os.O_RDWR|os.O_CREATE, os.ModePerm)
			if err != nil {
				log.Error("open file %v failed; %s", c.String("output"), err)
			}
			if outputFile != nil {
				defer outputFile.Close()
			}
		}

		log.Infof("start submit task and scan...")
		err = scanCenter.Scan(
			context.Background(),
			c.String("target"), c.String("port"), true, false,
			func(ip net.IP, port int) {
				openPortLock.Lock()
				defer openPortLock.Unlock()

				openPortCount++
				r := utils.HostPort(ip.String(), port)
				log.Debugf("found open port -> tcp://%v", r)
				openResult = append(openResult, r)

				if outputFile != nil {
					//outputFile.Write([]byte(fmt.Sprintf("%v\n", r)))
					outputFile.Write(
						[]byte(fmt.Sprintf(
							"%s%v\n",
							c.String("output-line-prefix"),
							r,
						)),
					)
				}
			},
		)
		if err != nil {
			log.Error(err)
			return
		}
		log.Infof("finished submitting.")

		if c.Bool("fingerprint") {
			fpTargetChan := make(chan *fp.PoolTask)
			go func() {
				defer close(fpTargetChan)
				for _, i := range openResult {
					host, port, err := utils.ParseStringToHostPort(i)
					if err != nil {
						continue
					}

					fpTargetChan <- &fp.PoolTask{
						Host:    host,
						Port:    port,
						Options: fingerprintMatchConfigOptions,
					}
				}
			}()
			pool, err := fp.NewExecutingPool(context.Background(), c.Int("fingerprint-concurrent"), fpTargetChan, fpConfig)
			if err != nil {
				log.Error("create fingerprint execute pool failed: %s", err)
				return
			}
			pool.AddCallback(func(matcherResult *fp.MatchResult, err error) {
				fpLock.Lock()
				defer fpLock.Unlock()

				if matcherResult != nil {
					fpResults = append(fpResults, matcherResult)
					log.Infof("scan fingerprint finished: -> %v", utils.HostPort(matcherResult.Target, matcherResult.Port))
				}
			})
			err = pool.Run()
			if err != nil {
				log.Error("fingerprint execute pool run failed: %v", err)
				return
			}
		}

		analysis := fp.MatcherResultsToAnalysis(fpResults)

		log.Infof("waiting last packet (SYN) for %v seconds", c.Int("waiting"))
		select {
		case <-time.After(time.Second * time.Duration(c.Int("waiting"))):
		}

		hosts := utils.ParseStringToHosts(c.String("target"))
		ports := utils.ParseStringToPorts(c.String("port"))
		analysis.TotalScannedPort = len(hosts) * len(ports)

		if c.Bool("fp") || len(analysis.OpenPortCPEMap) > 0 {
			analysis.Show()
			analysis.ToJson(c.String("fp-json"))
		} else {
			log.Infof("open ports ...\n===================================")
			for _, port := range openResult {
				println(port)
			}
		}
	},
}
