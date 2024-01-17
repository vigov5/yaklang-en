package subdomain

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
)

type SubdomainResult struct {
	FromTarget    string
	FromDNSServer string
	FromModeRaw   int

	IP     string
	Domain string

	// Tag is used to store some other information
	// such as data source and so on
	Tags []string
}

func (s *SubdomainResult) Hash() string {
	bs := md5.Sum([]byte(fmt.Sprintf("%v:%v:%v", s.IP, s.Domain, s.FromModeRaw)))
	return hex.EncodeToString(bs[:])
}

func (s *SubdomainResult) ToString() string {
	return fmt.Sprintf("%48s IP:[%15s] From:%v", s.Domain, s.IP, s.Tags)
}

func (s *SubdomainResult) Show() {
	println(s.ToString())
}
