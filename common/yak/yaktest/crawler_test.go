package yaktest

import "testing"

func TestCrawler(t *testing.T) {
	cases := []YakTestCase{
		{
			Name: "For crawler testing, add additional configuration: interfaces to be tested under each path",
			Src: `
res, err = crawler.Start(
	"159.65.125.15:80", 
	crawler.maxRequest(500),
	crawler.autoLogin("admin", "password"),
	// crawler.urlRegexpExclude("(?i).*?\/?(logout|reset|delete|setup).*"),
)
die(err)
dump(res)
for crawlerReq = range res {
	if crawlerReq == nil {
		continue
	}
	println(crawlerReq.Url())
}
`,
		},
	}

	Run("Yak Crawler test", t, cases...)
}
