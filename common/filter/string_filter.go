package filter

import (
	"encoding/binary"
	"sort"
	"strconv"
	"sync"

	"github.com/valyala/bytebufferpool"
	"github.com/yaklang/yaklang/common/cuckoo"
	"github.com/yaklang/yaklang/common/utils"
)

var bufferPool = bytebufferpool.Pool{}

type StringFilter struct {
	sync.Mutex
	container   *cuckoo.Filter
	conf        *Config
	lastUpdated int64
}

func (s *StringFilter) build(str string) []byte {
	buf := bufferPool.Get()
	defer func() {
		bufferPool.Put(buf)
	}()

	if s.conf.TTL > 0 {
		// If the last element is expired, directly release the previous container
		now := utils.TimestampMs() / 1000
		if s.lastUpdated != 0 && (now-s.lastUpdated >= s.conf.TTL) {
			s.container = NewDirCuckoo()
		}
		_ = binary.Write(buf, binary.LittleEndian, now/s.conf.TTL)
	}
	_, _ = buf.WriteString(str)
	b := buf.Bytes()
	newB := append(b[:0:0], b...)
	return newB
}

func (s *StringFilter) Exist(str string) bool {
	s.Lock()
	defer s.Unlock()
	return s.container.Lookup(s.build(str))
}

// is true, the insertion is successful, false the container is full, and no more additions can be made.
func (s *StringFilter) Insert(str string) bool {
	s.Lock()
	defer s.Unlock()
	return s.container.Insert(s.build(str))
}

func NewStringFilter(config *Config, container *cuckoo.Filter) *StringFilter {
	return &StringFilter{
		conf:      config,
		container: container,
	}
}

// NewFilter creates a default string cuckoo filter. The cuckoo filter is used to determine whether an element is in a set. It has extremely low false positives (that is, the element that is said to exist does not actually exist). Usually in this set The cuckoo filter is only used if the number of elements is very large.
// Example:
// ```
// f = str.NewFilter()
// f.Insert("hello")
// f.Exist("hello") // true
// ```
func NewFilter() *StringFilter {
	filterConfig := NewDefaultConfig()
	filterConfig.CaseSensitive = true
	f := NewStringFilter(filterConfig, NewGenericCuckoo())
	return f
}

func NewFilterWithSize(entries, total uint) *StringFilter {
	filterConfig := NewDefaultConfig()
	filterConfig.CaseSensitive = true
	f := NewStringFilter(filterConfig, cuckoo.New(
		cuckoo.BucketEntries(entries),
		cuckoo.BucketTotal(total),
		cuckoo.Kicks(300),
	))
	return f
}

// RemoveDuplicatePorts Parses two lists of ports as strings and uses a cuckoo filter for deduplication.
// function first creates a cuckoo filter and then parses the two input strings. is the port list.
// Next, it iterates through these two lists and adds each port to the cuckoo filter. If the port has not been added before, the
// then it will also be added to the results list. Finally, the function returns a resulting list containing all unique ports from the two input strings.
// Example:
// ```
// RemoveDuplicatePorts("10086-10088,23333", "10086,10089,23333") // [10086, 10087, 10088, 23333, 10089]
// ```
func RemoveDuplicatePorts(ports1, ports2 string) []int {
	filter := NewFilter()

	parsedPorts1 := utils.ParseStringToPorts(ports1)
	parsedPorts2 := utils.ParseStringToPorts(ports2)

	// merge and sort ports1 and ports2
	allPorts := append(parsedPorts1, parsedPorts2...)
	sort.Ints(allPorts)

	var uniquePorts []int

	// and adds the unique element
	for _, port := range allPorts {
		if !filter.Exist(strconv.Itoa(port)) {
			filter.Insert(strconv.Itoa(port))
			uniquePorts = append(uniquePorts, port)
		}
	}

	return uniquePorts
}

// FilterPorts accepts two port lists as strings as arguments, returns a new port list,
// to the result containing all ports that are in `ports1` but not in `ports2`.
// This function first parses the two input strings into port lists, and then creates a map (or collection) to store all ports in `ports2`.
// Then, it iterates through each port in `ports1`, and if the port is not in `ports2`, then it is added to the results list. The return value of
// Finally, the function returns a list of all ports that only appear in `ports1`.
// Example:
// ```
// FilterPorts("1-10", "2-10") // [1]
// ```
func FilterPorts(sourcePorts, excludePorts string) []int {
	p1 := utils.ParseStringToPorts(sourcePorts)
	p2 := utils.ParseStringToPorts(excludePorts)

	// Create a cuckoo filter for quick lookup of ports in ports2
	f := NewFilter()
	for _, v := range p2 {
		f.Insert(strconv.Itoa(v)) // Convert int to string before inserting
	}

	// Filter ports in ports1
	result := make([]int, 0)
	for _, v := range p1 {
		if !f.Exist(strconv.Itoa(v)) { // Convert int to string before checking
			result = append(result, v)
		}
	}

	return result
}
