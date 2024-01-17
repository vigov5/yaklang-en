package rpa

import (
	"github.com/yaklang/yaklang/common/rpa/core"
	"github.com/yaklang/yaklang/common/rpa/implement/bruteforce"
)

var Exports = map[string]interface{}{
	"Start":       Start,
	"depth":       core.WithSpiderDepth,     //Scanning depth.
	"proxy":       core.WithBrowserProxy,    //Proxy address. Optional username and password.
	"headers":     core.WithHeader,          //Specify headers. You can choose headers file address or json string.
	"strictUrl":   core.WithStrictUrlDetect, //Check whether the URL is risky.
	"maxUrl":      core.WithUrlCount,        //Limit on the total number of URLs obtained. No more scanning after exceeding
	"whiteDomain": core.WithWhiteDomain,     //Whitelist
	"blackDomain": core.WithBlackDomain,     //Blacklist
	"timeout":     core.WithTimeout,         //Single link timeout

	// BruteForce
	"Bruteforce":          bruteforce.BruteForceStart,
	"bruteUserPassPath":   bruteforce.WithUserPassPath,
	"bruteUsername":       bruteforce.WithUsername,
	"brutePassword":       bruteforce.WithPassword,
	"bruteUserElement":    bruteforce.WithUsernameElement,
	"brutePassElement":    bruteforce.WithPasswordElement,
	"bruteCaptchaElement": bruteforce.WithCaptchaElement,
	"bruteButtonElement":  bruteforce.WithButtonElement,

	// BruteForce do_before_brute
	"click":  bruteforce.ClickMethod,
	"select": bruteforce.SelectMethod,
	"input":  bruteforce.InputMethod,
}
