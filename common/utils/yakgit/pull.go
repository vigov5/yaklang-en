package yakgit

import (
	"os"

	"github.com/go-git/go-git/v5"
	"github.com/yaklang/yaklang/common/log"
	"github.com/yaklang/yaklang/common/utils"
)

// Pull is used to specify a local repository, obtain code from its remote repository and merge it into the local repository. This behavior is called pull. It can also receive zero to multiple option functions for influencing pull behavior
// Example:
// ```
// git.Pull("C:/Users/xxx/Desktop/yaklang", git.verify(false), git.remote("origin"))
// ```
func pull(localPath string, opts ...Option) error {
	c := &config{Remote: "origin"}
	for _, o := range opts {
		if err := o(c); err != nil {
			return err
		}
	}

	repos, err := git.PlainOpen(localPath)
	if err != nil {
		return utils.Errorf("git.PlainOpen failed: %s", err)
	}

	tree, err := repos.Worktree()
	if err != nil {
		return utils.Errorf("git.Worktree failed: %s", err)
	}

	err = tree.PullContext(c.Context, &git.PullOptions{
		RemoteName:        c.Remote,
		Depth:             c.Depth,
		Auth:              c.ToAuth(),
		RecurseSubmodules: c.ToRecursiveSubmodule(),
		Progress:          os.Stdout,
		Force:             c.Force,
		InsecureSkipTLS:   !c.VerifyTLS,
	})
	if err != nil {
		return utils.Errorf("git fetch failed: %s", err)
	}
	ref, err := repos.Head()
	if err != nil {
		log.Errorf("git fetch head failed: %s", err)
	}
	log.Infof("git pull success: %s", ref.String())
	return nil
}
