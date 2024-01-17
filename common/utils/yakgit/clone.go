package yakgit

import (
	"context"
	"os"

	"github.com/go-git/go-git/v5"
	gitClient "github.com/go-git/go-git/v5/plumbing/transport/client"
	gitHttp "github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/yaklang/yaklang/common/log"
	"github.com/yaklang/yaklang/common/netx"
	"github.com/yaklang/yaklang/common/utils"
)

func init() {
	tr := gitHttp.NewClient(netx.NewDefaultHTTPClient())
	gitClient.InstallProtocol("https", tr)
	gitClient.InstallProtocol("http", tr)
}

// SetProxy is an auxiliary function used to specify the proxy for other Git operations (such as Clone).
// Example:
// ```
// git.SetProxy("http://127.0.0.1:1080")
// ```
func SetProxy(proxies ...string) {
	tr := gitHttp.NewClient(netx.NewDefaultHTTPClient(proxies...))
	gitClient.InstallProtocol("https", tr)
	gitClient.InstallProtocol("http", tr)
}

// Clone is used to clone the remote repository and store it in the local path. It can also receive zero to more option functions to affect the cloning behavior.
// Example:
// ```
// git.Clone("https://github.com/yaklang/yaklang", "C:/Users/xxx/Desktop/yaklang", git.recursive(true), git.verify(false))
// ```
func clone(u string, localPath string, opt ...Option) error {
	c := &config{}
	for _, o := range opt {
		if err := o(c); err != nil {
			return err
		}
	}
	if c.Context == nil {
		c.Context, c.Cancel = context.WithCancel(context.Background())
	}

	respos, err := git.PlainCloneContext(c.Context, localPath, false, &git.CloneOptions{
		URL:               u,
		Auth:              c.ToAuth(),
		Depth:             c.Depth,
		RecurseSubmodules: c.ToRecursiveSubmodule(),
		InsecureSkipTLS:   !c.VerifyTLS,
		Progress:          os.Stdout,
	})
	if err != nil {
		return utils.Wrapf(err, "git clone: %v to %v failed", u, localPath)
	}
	_ = respos
	h, err := respos.Head()
	if h != nil {
		log.Infof("git clone: %v to %v success: %v", u, localPath, h.String())
	} else {
		log.Infof("git clone: %v to %v success", u, localPath)
	}
	return nil
}
