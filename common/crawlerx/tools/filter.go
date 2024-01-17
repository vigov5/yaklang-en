// Package tools
// @Author bcy2007  2023/7/13 15:49
package tools

import (
	"encoding/binary"
	"github.com/valyala/bytebufferpool"
	"github.com/yaklang/yaklang/common/cuckoo"
	"github.com/yaklang/yaklang/common/filter"
	"github.com/yaklang/yaklang/common/utils"
	"sync"
)

var bufferPool = bytebufferpool.Pool{}

type StringCountFilter struct {
	sync.Mutex
	container   *cuckoo.Filter
	conf        *filter.Config
	lastUpdated int64
	count       int64
}

func (s *StringCountFilter) build(str string) []byte {
	buf := bufferPool.Get()
	defer func() {
		bufferPool.Put(buf)
	}()

	if s.conf.TTL > 0 {
		// If the last element is expired, directly release the previous container
		now := utils.TimestampMs() / 1000
		if s.lastUpdated != 0 && (now-s.lastUpdated >= s.conf.TTL) {
			s.container = filter.NewDirCuckoo()
		}
		_ = binary.Write(buf, binary.LittleEndian, now/s.conf.TTL)
	}
	_, _ = buf.WriteString(str)
	b := buf.Bytes()
	newB := append(b[:0:0], b...)
	return newB
}

func (s *StringCountFilter) Insert(str string) bool {
	s.Lock()
	defer s.Unlock()
	status := s.container.Insert(s.build(str))
	if status {
		s.count++
	}
	return status
}

func (s *StringCountFilter) Exist(str string) bool {
	s.Lock()
	defer s.Unlock()
	return s.container.Lookup(s.build(str))
}

func (s *StringCountFilter) Count() int64 {
	return s.count
}

func NewStringCountFilter(config *filter.Config, container *cuckoo.Filter) *StringCountFilter {
	return &StringCountFilter{
		conf:      config,
		container: container,
	}
}

func NewCountFilter() *StringCountFilter {
	filterConfig := filter.NewDefaultConfig()
	filterConfig.CaseSensitive = true
	f := NewStringCountFilter(filterConfig, filter.NewGenericCuckoo())
	return f
}
