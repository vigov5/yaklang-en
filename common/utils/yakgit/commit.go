package yakgit

import (
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/yaklang/yaklang/common/log"
	"github.com/yaklang/yaklang/common/utils"
)

// IterateCommit is used to specify a local warehouse, traverse all its commit records (commits), and perform specified operations on each filtered commit record. It can also receive zero to multiple option functions for configuring the callback function
// Example:
// ```
// // Traverse the submission records, filter the reference records whose names contain ci, filter the submission records whose author name is xxx, and print each remaining submission record
// git.IterateCommit("D:/coding/golang/src/yaklang",
// git.filterReference((ref) => {return !ref.Name().Contains("ci")}),
// git.filterCommit((c) => { return c.Author.Name != "xxx" }),
// git.handleCommit((c) => { println(c.String()) }))
// ```
func EveryCommit(localRepos string, opt ...Option) error {
	c := NewConfig()
	for _, i := range opt {
		err := i(c)
		if err != nil {
			return err
		}
	}
	r, err := git.PlainOpen(localRepos)
	if err != nil {
		return utils.Errorf("open repository failed: %s", err)
	}
	refs, err := r.References()
	if err != nil {
		return utils.Errorf("get references failed: %s", err)
	}
	return refs.ForEach(func(ref *plumbing.Reference) error {
		if c.FilterGitReference != nil && !c.FilterGitReference(ref) {
			return nil
		}
		if c.HandleGitReference != nil {
			err := c.HandleGitReference(ref)
			if err != nil {
				return err
			}
		}
		commitIter, err := r.Log(&git.LogOptions{
			From: ref.Hash(),
		})
		if err != nil {
			log.Errorf("fetch %v's logs failed: %s", ref.Hash(), err)
			return nil
		}
		commitIter.ForEach(func(commit *object.Commit) error {
			if c.FilterGitCommit != nil && !c.FilterGitCommit(commit) {
				return nil
			}

			if c.HandleGitCommit != nil {
				err := c.HandleGitCommit(commit)
				if err != nil {
					return err
				}
			}

			return nil
		})
		return nil
	})
}
