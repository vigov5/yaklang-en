package yakgit

import (
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/yaklang/yaklang/common/log"
	"github.com/yaklang/yaklang/common/utils"
)

// . Checkout is used to specify a local warehouse, switch its branches or restore the files of the working tree. This behavior is called checkout. It can also receive zero to more option functions. Use to affect detection behavior
// Example:
// ```
// git.Checkout("C:/Users/xxx/Desktop/yaklang", "feat/new-branch", git.checkoutCreate(true)) // Create a new branch
// git.Checkout("C:/Users/xxx/Desktop/yaklang", "old-branch", git.checkoutForce(true)) // and force switching
// ```
func checkout(localPath string, ref string, opts ...Option) error {
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

	checkoutOpt := &git.CheckoutOptions{
		Create: c.CheckoutCreate,
		Force:  c.CheckoutForce || c.Force,
		Keep:   c.CheckoutKeep,
	}

	if ref != "" {
		branch, err := repos.Branch(ref)
		if err != nil {
			log.Infof("git branch %s not found, try tag", ref)
			tag, err := repos.Tag(ref)
			if err != nil {
				log.Infof("git tag %s not found, try commit", ref)
				commit, err := repos.CommitObject(plumbing.NewHash(ref))
				if err != nil {
					return utils.Errorf("git commit %s not found", ref)
				}
				checkoutOpt.Hash = commit.Hash
			} else {
				checkoutOpt.Hash = tag.Hash()
			}
		} else {
			checkoutOpt.Branch = branch.Merge
		}
	}

	err = tree.Checkout(checkoutOpt)
	if err != nil {
		return utils.Errorf("git fetch failed: %s", err)
	}
	if ref == "" {
		log.Info("git checkout success")
	} else {
		log.Infof("git checkout success: %s", ref)
	}
	return nil
}
