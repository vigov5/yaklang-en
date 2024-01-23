package core

import (
	"encoding/binary"
	"github.com/yaklang/yaklang/common/cuckoo"
	"github.com/yaklang/yaklang/common/filter"
	"github.com/yaklang/yaklang/common/utils"
	"sync"

	"github.com/valyala/bytebufferpool"
)

var bufferPool = bytebufferpool.Pool{}

type StringFilterwithCount struct {
	sync.Mutex
	container   *cuckoo.Filter
	conf        *filter.Config
	lastUpdated int64
	count       int64
}

func (s *StringFilterwithCount) build(str string) []byte {
	buf := bufferPool.Get()
	defer func() {
		bufferPool.Put(buf)
	}()

	if s.conf.TTL > 0 {
		// . If the last element is expired, the previous container is released directly.
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

func (s *StringFilterwithCount) Exist(str string) bool {
	s.Lock()
	defer s.Unlock()
	return s.container.Lookup(s.build(str))
}

// The return value is true, the insertion is successful, false when the container is full and cannot continue to add
func (s *StringFilterwithCount) Insert(str string) bool {
	s.Lock()
	defer s.Unlock()
	status := s.container.Insert(s.build(str))
	if status {
		s.count++
	}
	return status
}

func (s *StringFilterwithCount) Count() int64 {
	return s.count
}

func NewStringFilterwithCount(config *filter.Config, container *cuckoo.Filter) *StringFilterwithCount {
	return &StringFilterwithCount{
		conf:      config,
		container: container,
	}
}

func NewFilterwithCount() *StringFilterwithCount {
	filterConfig := filter.NewDefaultConfig()
	filterConfig.CaseSensitive = true
	f := NewStringFilterwithCount(filterConfig, filter.NewGenericCuckoo())
	return f
}
