package core

import (
	"github.com/yaklang/yaklang/common/log"
	"github.com/yaklang/yaklang/common/utils"
	"github.com/yaklang/yaklang/common/yak/yaklib/codec"
	"strings"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
)

func (m *Manager) WaitLoad(page *rod.Page) {
	page.MustWaitLoad()
	//defer page.Close()
	// defer page.MustClose()
	// defer m.PutPage(page)
	//defer m.PutPage(page)
	var lastHtmlHash = ""
	var exitThroshold = 0
	var totalFetcher = 0
	var htmlStr string
	for {
		totalFetcher++
		if totalFetcher > 10 {
			break
		}

		if exitThroshold >= 2 {
			break
		}
		time.Sleep(500 * time.Millisecond)
		htmlStr, err := page.HTML()
		if err != nil {
			log.Errorf("fetch html failed: %s", err)
		}

		if lastHtmlHash == "" && htmlStr == "" {
			exitThroshold++
			continue
		}

		// . Cache effectively.
		if lastHtmlHash == "" && htmlStr != "" {
			lastHtmlHash = codec.Sha256(htmlStr)
			exitThroshold = 0
			continue
		}

		// . Repeat three times.
		if lastHtmlHash == codec.Sha256(htmlStr) {
			exitThroshold++
			continue
		} else {
			exitThroshold = 0
			lastHtmlHash = codec.Sha256(htmlStr)
		}
	}

	if htmlStr == "" {
		log.Error("page empty")
	}

}

func (m *Manager) page(url string, depth int) error {

	defer m.pageSizedWaitGroup.Done()

	page_block, err := m.GetPage(proto.TargetCreateTarget{
		URL: "about:blank",
	}, depth)
	page := page_block.page.Timeout(time.Duration(m.config.timeout) * time.Second)
	// page := page_block.page
	if err != nil {
		return utils.Errorf("get blank page[%v] failed: %s", url, err)
	}
	if page == nil {
		return utils.Errorf("get blank page[%v] nil: %s", url, err)
	}
	// . The page exits normally.
	// defer page.Close()
	defer m.PutPage(page)
	// page.Navigate(url)
	err = page.Navigate(url)
	if err != nil {
		return utils.Errorf("navigate page[%v] failed: %s", url, err)
	}
	page = page.Context(m.rootContext)

	// m.visitedUrl.Insert(url)

	// page.MustWaitLoad()
	// time.Sleep(1 * time.Second)
	wait_count := 0
	for {
		state, err := m.GetReadyState(page)
		if err != nil {
			return utils.Errorf("get ready state error: %s", err)
		}
		if state == "complete" {
			break
		}
		wait_count++
		if wait_count > 20 {
			log.Infof("wait ready state too long error: %s.", state)
			break
		}
		time.Sleep(1 * time.Second)
	}
	var lastHtmlHash = ""
	var exitThroshold = 0
	var totalFetcher = 0
	var htmlStr string
	for {
		totalFetcher++
		if totalFetcher > 10 {
			break
		}

		if exitThroshold >= 2 {
			break
		}
		time.Sleep(500 * time.Millisecond)
		htmlStr, err = page.HTML()
		if err != nil && !strings.Contains(err.Error(), "context canceled") {
			log.Errorf("fetch [%v]'s html failed: %s", url, err)
		}

		if lastHtmlHash == "" && htmlStr == "" {
			exitThroshold++
			continue
		}

		// . Cache effectively.
		if lastHtmlHash == "" && htmlStr != "" {
			lastHtmlHash = codec.Sha256(htmlStr)
			exitThroshold = 0
			continue
		}

		// . Repeat three times.
		if lastHtmlHash == codec.Sha256(htmlStr) {
			exitThroshold++
			continue
		} else {
			exitThroshold = 0
			lastHtmlHash = codec.Sha256(htmlStr)
		}
	}

	if htmlStr == "" {
		// log.Error("page empty")
		return nil
	}

	//Store the response to chan
	// . At this point, I think the page is almost loaded. Got it
	// . Next, you need to do other things, such as:
	//   . 1. Get the clickable links of the page.
	//   . 2. Get the table that can be processed by the page.
	// page.MustScreenshot("test.png")

	err = m.extractInput(page)
	if err != nil && !strings.Contains(err.Error(), "context canceled") {
		log.Errorf("extract blanks and do input failed: %s", err)
	}

	err = m.extractUrls(page_block)
	if err != nil && !strings.Contains(err.Error(), "context canceled") {
		log.Errorf("extract urls failed: %s", err)
	}

	err = m.extractCommit(page_block)
	if err != nil && !strings.Contains(err.Error(), "context canceled") {
		log.Errorf("extract commit failed: %s", err)
	}

	return nil
}
