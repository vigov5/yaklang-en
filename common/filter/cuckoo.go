package filter

import "github.com/yaklang/yaklang/common/cuckoo"

// Filter container with default parameters. Using one globally is easy to cause collision, so it is split into four.
// NewGenericCuckoo has a limit capacity of about 4 million.
func NewGenericCuckoo() *cuckoo.Filter {
	return cuckoo.New(
		cuckoo.BucketEntries(18),
		cuckoo.BucketTotal(1<<18),
		cuckoo.Kicks(300),
	)
}

// NewPathCuckoo has a limit capacity of about 2 million.
func NewPathCuckoo() *cuckoo.Filter {
	return cuckoo.New(cuckoo.BucketEntries(16),
		cuckoo.BucketTotal(1<<17),
		cuckoo.Kicks(300))
}

// NewDirCuckoo has a limit capacity of about 1 million.
func NewDirCuckoo() *cuckoo.Filter {
	return cuckoo.New(cuckoo.BucketEntries(14),
		cuckoo.BucketTotal(1<<16),
		cuckoo.Kicks(300))
}

// NewWebsiteCuckoo has a limit capacity of about 400,000.
func NewWebsiteCuckoo() *cuckoo.Filter {
	return cuckoo.New(cuckoo.BucketEntries(12),
		cuckoo.BucketTotal(1<<15),
		cuckoo.Kicks(300))
}
