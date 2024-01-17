package subdomain

import "time"

const (
	_ = iota
	BRUTE
	SEARCH
	ZONE_TRANSFER
)

type SubdomainScannerConfig struct {
	// Set
	Modes []int

	// Allows recursive scanning
	AllowToRecursive bool

	// DNS Worker Count
	// How many DNS query Goroutines are allowed simultaneously?
	WorkerCount int

	// DNS Servers Default DNS server
	DNSServers []string

	// Subdomain name traversal The maximum depth of // Default is 5
	MaxDepth int

	// Number of parallel execution tasks Default: 10
	ParallelismTasksCount int

	// Timeout for each target
	TimeoutForEachTarget time.Duration

	// Subdomain name blasting primary dictionary
	MainDictionary []byte

	// Subdomain name blasting secondary dictionary
	SubDictionary []byte

	// turning on automatic Web discovery?
	//	allow_auto_web_discover: bool = True
	// Automatic discovery of the Web should not be left to this module to implement
	// The subdomain name scan results should be handed over to the crawler module externally for processing

	// The timeout for each query, the default value is 3s
	TimeoutForEachQuery time.Duration

	// Encounter In the case of pan-parsing, stop parsing immediately
	//   There are two situations here. The first is the pan-parsing set by oneself, and the second is the pan-parsing set by the operator. Is
	//   If pan-parsing is encountered and do not stop immediately, subdomain blasting will automatically add the IP address of pan-parsed to the blacklist
	//   can be cached. The default value is false.
	WildCardToStop bool

	// When searching for various data sources, the HTTP timeout that needs to be set
	// Default is 10s
	TimeoutForEachHTTPSearch time.Duration
}

func (s *SubdomainScannerConfig) init() {
	s.Modes = []int{BRUTE, SEARCH, ZONE_TRANSFER}
	s.AllowToRecursive = true
	s.WorkerCount = 50
	s.DNSServers = []string{"114.114.114.114", "8.8.8.8"}
	s.MaxDepth = 5
	s.ParallelismTasksCount = 10
	s.TimeoutForEachTarget = 300 * time.Second
	s.MainDictionary = DefaultMainDictionary
	s.SubDictionary = DefaultSubDictionary
	s.TimeoutForEachQuery = 3 * time.Second
	s.WildCardToStop = false
	s.TimeoutForEachHTTPSearch = 10 * time.Second
}

type ConfigOption func(s *SubdomainScannerConfig)

// Configure subdomain discovery mode
func WithModes(modes ...int) ConfigOption {
	return func(s *SubdomainScannerConfig) {
		s.Modes = modes
	}
}

func WithAllowToRecursive(b bool) ConfigOption {
	return func(s *SubdomainScannerConfig) {
		s.AllowToRecursive = b
	}
}

func WithWorkerCount(c int) ConfigOption {
	return func(s *SubdomainScannerConfig) {
		s.WorkerCount = c
	}
}

func WithDNSServers(servers []string) ConfigOption {
	return func(s *SubdomainScannerConfig) {
		s.DNSServers = servers
	}
}

func WithMaxDepth(d int) ConfigOption {
	return func(s *SubdomainScannerConfig) {
		s.MaxDepth = d
	}
}

func WithParallelismTasksCount(c int) ConfigOption {
	return func(s *SubdomainScannerConfig) {
		s.ParallelismTasksCount = c
	}
}

func WithTimeoutForEachTarget(t time.Duration) ConfigOption {
	return func(s *SubdomainScannerConfig) {
		s.TimeoutForEachTarget = t
	}
}

func WithMainDictionary(raw []byte) ConfigOption {
	return func(s *SubdomainScannerConfig) {
		s.MainDictionary = raw
	}
}

func WithSubDictionary(raw []byte) ConfigOption {
	return func(s *SubdomainScannerConfig) {
		s.SubDictionary = raw
	}
}

func WithTimeoutForEachQuery(timeout time.Duration) ConfigOption {
	return func(s *SubdomainScannerConfig) {
		s.TimeoutForEachQuery = timeout
	}
}

func WithWildCardToStop(t bool) ConfigOption {
	return func(s *SubdomainScannerConfig) {
		s.WildCardToStop = t
	}
}

func WithTimeoutForEachHTTPSearch(timeout time.Duration) ConfigOption {
	return func(s *SubdomainScannerConfig) {
		s.TimeoutForEachHTTPSearch = timeout
	}
}

func NewSubdomainScannerConfig(options ...ConfigOption) *SubdomainScannerConfig {
	config := &SubdomainScannerConfig{}
	config.init()

	for _, option := range options {
		option(config)
	}
	return config
}
