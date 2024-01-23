package fp

import (
	"context"
	"fmt"
	"github.com/yaklang/yaklang/common/filter"
	"github.com/yaklang/yaklang/common/fp/webfingerprint"
	"github.com/yaklang/yaklang/common/utils"
	"github.com/yaklang/yaklang/common/utils/hostsparser"
	"io/ioutil"
	"time"
)

type ConfigOption func(config *Config)

type Config struct {
	// For that Transport layer protocol for fingerprinting?
	TransportProtos []TransportProto

	// Probe control based on active mode packet sending
	// The greater the rarity, the less likely this service is to exist in reality.
	// The value range is 1-9
	// The default value is 5
	RarityMax int

	/*
		Probe is to actively send some Data packets are used to detect fingerprint information. The following options can control the use of Probe.
	*/
	// Active mode. In this mode, packets will be actively sent to detect fingerprints (Probe is enabled).
	// can be cached. The default value is false.
	ActiveMode bool

	// defaults to the timeout of each Probe.
	ProbeTimeout time.Duration

	// The limit on the number of probes sent, the default value is 5
	ProbesMax int

	// Concurrency of sending Probe, the default value is 5
	ProbesConcurrentMax int

	// Specify rules
	FingerprintRules map[*NmapProbe][]*NmapMatch

	// Data taken during fingerprint detection Size, means how much data will be involved in fingerprinting.
	// 2048 is the default value
	// The time of host fingerprinting is proportional to this value.
	FingerprintDataSize int

	/* Web Fingerprint */
	// Active Mode. The ActiveMode here should be consistent with the current configuration, so there is no need to set

	//
	// ForceEnableAllFingerprint indicates forced detection of Web fingerprints.
	ForceEnableAllFingerprint bool

	// OnlyEnableWebFingerprint indicates the value of Web fingerprint recognition.
	//    When this option is True, the behavior will override ForceEnableAllFingerprint.
	OnlyEnableWebFingerprint bool

	// Disable special Web fingerprint scanning
	DisableWebFingerprint bool

	// This option indicates that if some fingerprints have been detected during the Web fingerprint detection, other fingerprints should also continue to be detected
	WebFingerprintUseAllRules bool

	// The maximum number of URLs discovered by the crawler, the default is 5
	CrawlerMaxUrlCount int

	// Use the specified WebRule to test the Web Fingerprint, the default is to use the default fingerprint
	WebFingerprintRules []*webfingerprint.WebRule

	// Concurrency pool size configuration (single does not take effect)
	PoolSize int

	// is the port. Scan and set proxy
	Proxies []string

	// In the same engine process,
	EnableCache bool

	// for the time being. Set up the database cache, which can be used across processes.
	EnableDatabaseCache bool

	// Exclude
	ExcludeHostsFilter *hostsparser.HostsParser
	ExcludePortsFilter *filter.StringFilter
}

func (c *Config) IsFiltered(host string, port int) bool {
	if c == nil {
		return false
	}
	if c.ExcludeHostsFilter != nil {
		if c.ExcludeHostsFilter.Contains(host) {
			return true
		}
	}

	if c.ExcludePortsFilter != nil {
		if c.ExcludePortsFilter.Exist(fmt.Sprint(port)) {
			return true
		}
	}
	return false
}

func WithCache(b bool) ConfigOption {
	return func(config *Config) {
		config.EnableCache = b
	}
}

func WithDatabaseCache(b bool) ConfigOption {
	return func(config *Config) {
		config.EnableDatabaseCache = b
	}
}

func WithProxy(proxies ...string) ConfigOption {
	return func(config *Config) {
		config.Proxies = utils.StringArrayFilterEmpty(proxies)
	}
}

func WithPoolSize(size int) ConfigOption {
	return func(config *Config) {
		config.PoolSize = size
		if config.PoolSize <= 0 {
			config.PoolSize = 50
		}
	}
}

func WithExcludeHosts(hosts string) ConfigOption {
	return func(config *Config) {
		config.ExcludeHostsFilter = hostsparser.NewHostsParser(context.Background(), hosts)
	}
}

func WithExcludePorts(ports string) ConfigOption {
	return func(config *Config) {
		config.ExcludePortsFilter = filter.NewFilter()
		for _, port := range utils.ParseStringToPorts(ports) {
			config.ExcludePortsFilter.Insert(fmt.Sprint(port))
		}
	}
}

func (f *Config) GenerateWebFingerprintConfigOptions() []webfingerprint.ConfigOption {
	return []webfingerprint.ConfigOption{
		webfingerprint.WithActiveMode(f.ActiveMode),
		webfingerprint.WithForceAllRuleMatching(f.WebFingerprintUseAllRules),
		webfingerprint.WithProbeTimeout(f.ProbeTimeout),
		webfingerprint.WithWebFingerprintRules(f.WebFingerprintRules),
		webfingerprint.WithWebFingerprintDataSize(f.FingerprintDataSize),
		webfingerprint.WithWebProxy(f.Proxies...),
	}
}

func NewConfig(options ...ConfigOption) *Config {
	config := &Config{}
	config.init()

	for _, p := range options {
		p(config)
	}

	config.lazyInit()
	return config
}

func (c *Config) Configure(ops ...ConfigOption) {
	for _, op := range ops {
		op(c)
	}
}

func (c *Config) init() {
	c.TransportProtos = []TransportProto{TCP}
	c.ActiveMode = true
	c.RarityMax = 5
	c.ProbesMax = 5
	c.ProbeTimeout = 5 * time.Second
	c.ProbesConcurrentMax = 5
	c.OnlyEnableWebFingerprint = false
	c.DisableWebFingerprint = false
	c.ForceEnableAllFingerprint = false
	c.WebFingerprintUseAllRules = true
	c.CrawlerMaxUrlCount = 5
	c.PoolSize = 20

	c.FingerprintDataSize = 20480
}

func (c *Config) lazyInit() {
	if len(c.WebFingerprintRules) <= 0 {
		c.WebFingerprintRules, _ = GetDefaultWebFingerprintRules()
	}

	if len(c.FingerprintRules) <= 0 {
		c.FingerprintRules, _ = GetDefaultNmapServiceProbeRules()
	}
}

func WithProbesMax(m int) ConfigOption {
	if m <= 0 {
		m = 5
	}

	return func(config *Config) {
		config.ProbesMax = m
	}
}

func ParseStringToProto(protos ...interface{}) []TransportProto {
	var ret []TransportProto

	var raw []string
	for _, proto := range protos {
		raw = append(raw, utils.ToLowerAndStrip(fmt.Sprint(proto)))
	}

	if utils.StringSliceContain(raw, "tcp") {
		ret = append(ret, TCP)
	}

	if utils.StringSliceContain(raw, "udp") {
		ret = append(ret, UDP)
	}

	return ret
}

func WithProbesConcurrentMax(m int) ConfigOption {
	if m <= 0 {
		m = 5
	}

	return func(config *Config) {
		config.ProbesConcurrentMax = m
	}
}

func WithTransportProtos(protos ...TransportProto) ConfigOption {
	r := map[TransportProto]int{}
	for _, p := range protos {
		r[p] = 1
	}

	var results []TransportProto
	for key := range r {
		results = append(results, key)
	}

	if results == nil {
		results = []TransportProto{TCP}
	}

	return func(config *Config) {
		config.TransportProtos = results
	}
}

func WithRarityMax(rarity int) ConfigOption {
	return func(config *Config) {
		config.RarityMax = rarity
	}
}

func WithActiveMode(raw bool) ConfigOption {
	return func(config *Config) {
		config.ActiveMode = raw
	}
}

func WithProbeTimeout(timeout time.Duration) ConfigOption {
	return func(config *Config) {
		config.ProbeTimeout = timeout
	}
}

func WithProbeTimeoutHumanRead(f float64) ConfigOption {
	return func(config *Config) {
		config.ProbeTimeout = utils.FloatSecondDuration(f)
		if config.ProbeTimeout <= 0 {
			config.ProbeTimeout = 10 * time.Second
		}
	}
}

func WithForceEnableAllFingerprint(b bool) ConfigOption {
	return func(config *Config) {
		config.ForceEnableAllFingerprint = b
	}
}

func WithOnlyEnableWebFingerprint(b bool) ConfigOption {
	return func(config *Config) {
		config.OnlyEnableWebFingerprint = b
	}
}

func WithFingerprintRule(rules map[*NmapProbe][]*NmapMatch) ConfigOption {
	return func(config *Config) {
		config.FingerprintRules = rules
	}
}

func WithDisableWebFingerprint(t bool) ConfigOption {
	return func(config *Config) {
		config.DisableWebFingerprint = t
	}
}

func WithFingerprintDataSize(size int) ConfigOption {
	return func(config *Config) {
		config.FingerprintDataSize = size
	}
}

func WithWebFingerprintUseAllRules(b bool) ConfigOption {
	return func(config *Config) {
		config.WebFingerprintUseAllRules = b
	}
}

func WithWebFingerprintRule(i interface{}) ConfigOption {
	var rules []*webfingerprint.WebRule
	switch ret := i.(type) {
	case []byte:
		rules, _ = webfingerprint.ParseWebFingerprintRules(ret)
	case string:
		e := utils.GetFirstExistedPath(ret)
		if e != "" {
			raw, _ := ioutil.ReadFile(e)
			rules, _ = webfingerprint.ParseWebFingerprintRules(raw)
		} else {
			rules, _ = webfingerprint.ParseWebFingerprintRules([]byte(ret))
		}
	case []*webfingerprint.WebRule:
		rules = ret
	}

	return func(config *Config) {
		if rules == nil {
			return
		}

		config.WebFingerprintRules = rules
	}
}

func WithNmapRule(i interface{}) ConfigOption {
	var nmapRules map[*NmapProbe][]*NmapMatch
	switch ret := i.(type) {
	case []byte:
		nmapRules, _ = ParseNmapServiceProbeToRuleMap(ret)
	case string:
		e := utils.GetFirstExistedPath(ret)
		if e != "" {
			raw, _ := ioutil.ReadFile(e)
			nmapRules, _ = ParseNmapServiceProbeToRuleMap(raw)
		} else {
			nmapRules, _ = ParseNmapServiceProbeToRuleMap([]byte(ret))
		}
	case map[*NmapProbe][]*NmapMatch:
		nmapRules = ret
	}
	return func(config *Config) {
		if nmapRules == nil {
			return
		}

		if len(nmapRules) <= 0 {
			return
		}
		config.FingerprintRules = nmapRules
	}
}
