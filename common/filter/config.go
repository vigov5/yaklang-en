package filter

type Config struct {
	TTL           int64 `json:"ttl" yaml:"ttl"`                       // deduplication time window, in seconds
	CaseSensitive bool  `json:"case_sensitive" yaml:"case_sensitive"` // Whether to be case sensitive in URLs
	Hash          bool  `json:"hash" yaml:"hash"`                     // Whether to use URL hash as a deduplication factor
	Query         bool  `json:"query" yaml:"query"`                   // Whether to use query as a deduplication factor, if true, then ?a=b and ?a=c will be regarded as two links
	Credential    bool  `json:"credential" yaml:"credential"`         // Whether to use authentication information as a deduplication factor, if true, the same page before and after login May be scanned twice
	Rewrite       bool  `json:"rewrite" yaml:"rewrite"`               // Whether to enable rewrite to identify
}

func (c *Config) Clone() *Config {
	newC := *c
	return &newC
}

func NewDefaultConfig() *Config {
	return &Config{Credential: true}
}
