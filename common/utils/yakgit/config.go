package yakgit

import (
	"context"
	"github.com/yaklang/yaklang/common/utils/lowhttp/poc"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	gitHttp "github.com/go-git/go-git/v5/plumbing/transport/http"
)

type config struct {
	Context context.Context
	Cancel  context.CancelFunc

	VerifyTLS          bool
	Username           string
	Password           string
	Depth              int
	RecursiveSubmodule bool

	// remote operation
	Remote string

	// Force
	Force        bool
	NoFetchTags  bool
	FetchAllTags bool

	CheckoutCreate bool
	CheckoutForce  bool
	CheckoutKeep   bool

	// GitHack
	Threads           int
	UseLocalGitBinary bool
	HTTPOptions       []poc.PocConfig

	// handler
	HandleGitReference func(r *plumbing.Reference) error
	FilterGitReference func(r *plumbing.Reference) bool
	HandleGitCommit    func(r *object.Commit) error
	FilterGitCommit    func(r *object.Commit) bool
}

func NewConfig() *config {
	ctx, cancel := context.WithCancel(context.Background())
	return &config{
		Context:            ctx,
		Cancel:             cancel,
		VerifyTLS:          true,
		Depth:              1,
		RecursiveSubmodule: true,
		Remote:             "origin",
	}
}

func (c *config) ToRecursiveSubmodule() git.SubmoduleRescursivity {
	var recursiveSubmodule git.SubmoduleRescursivity
	if c.RecursiveSubmodule {
		recursiveSubmodule = git.SubmoduleRescursivity(10)
	} else {
		recursiveSubmodule = git.NoRecurseSubmodules
	}
	return recursiveSubmodule
}

func (c *config) ToAuth() gitHttp.AuthMethod {
	var auth gitHttp.AuthMethod
	if c.Username != "" && c.Password != "" {
		auth = &gitHttp.BasicAuth{
			Username: c.Username,
			Password: c.Password,
		}
	}
	return auth
}

type Option func(*config) error

// handleReference is an option function, which receives a callback function. This function has a Parameter, which is the reference record structure (reference). Every time the filtered reference is traversed, this callback function will be called.
// Example:
// ```
// // traverses the submission records, filters the reference records whose names contain ci, and prints each remaining reference record
// git.IterateCommit("D:/coding/golang/src/yaklang",
// git.filterReference((ref) => {return !ref.Name().Contains("ci")}),
// git.handleReference((ref) => { println(ref.String()) }))
// ```
func WithHandleGitReference(f func(r *plumbing.Reference) error) Option {
	return func(c *config) error {
		c.HandleGitReference = f
		return nil
	}
}

// filterReference is an option function that receives a callback function. This function has one parameter, which is the reference record structure (reference). This callback function will be called every time the reference is traversed. This function also has a return value. Use this return value to decide whether to filter out this reference.
// Example:
// ```
// // traverses the submission records, filters the reference records whose names contain ci, and prints each remaining reference record
// git.IterateCommit("D:/coding/golang/src/yaklang",
// git.filterReference((ref) => {return !ref.Name().Contains("ci")}),
// git.handleReference((ref) => { println(ref.String()) }))
// ```
func WithFilterGitReference(f func(r *plumbing.Reference) bool) Option {
	return func(c *config) error {
		c.FilterGitReference = f
		return nil
	}
}

// handleCommit is an option function that receives a callback function. This function has one parameter, which is the commit record structure (commit). Each time a filtered commit record is traversed This callback function will be called when
// Example:
// ```
// // during a checkout operation. Traverse the commit records and print each commit record
// git.IterateCommit("D:/coding/golang/src/yaklang", git.handleCommit((c) => { println(c.String()) }))
// ```
func WithHandleGitCommit(f func(r *object.Commit) error) Option {
	return func(c *config) error {
		c.HandleGitCommit = f
		return nil
	}
}

// filterCommit is an option function. It receives a callback function. This function has one parameter, which is commit. Record structure (commit), every time a commit record is traversed, this callback function will be called. This function also has a return value, which is used to decide whether to filter out the commit record.
// Example:
// ```
// // traverses the submission records, filters the submission records whose author name is xxx, and prints each remaining submission record
// git.IterateCommit("D:/coding/golang/src/yaklang",
// git.filterCommit((c) => { return c.Author.Name != "xxx" }),
// git.handleCommit((c) => { println(c.String()) }))
// ```
func WithFilterGitCommit(f func(r *object.Commit) bool) Option {
	return func(c *config) error {
		c.FilterGitCommit = f
		return nil
	}
}

// verify is an option function used to specify whether to verify the TLS certificate when other Git operations (such as Clone)
// Example:
// ```
// git.Clone("https://github.com/yaklang/yaklang", "C:/Users/xxx/Desktop/yaklang", git.recursive(true), git.verify(false))
// ```
func WithVerifyTLS(b bool) Option {
	return func(c *config) error {
		c.VerifyTLS = b
		return nil
	}
}

// noFetchTags is an option function used to specify whether not to pull during the fetch operation The label
// Example:
// ```
// git.Fetch("C:/Users/xxx/Desktop/yaklang", git.noFetchTags(true)) // Do not pull tags
// ```
func WithNoFetchTags(b bool) Option {
	return func(c *config) error {
		c.NoFetchTags = b
		return nil
	}
}

// fetchAllTags is an option function, used to specify the fetch operation Whether to pull all tags when
// Example:
// ```
// git.Fetch("C:/Users/xxx/Desktop/yaklang", git.fetchAllTags(true)) // pulls all tags
// ```
func WithFetchAllTags(b bool) Option {
	return func(c *config) error {
		c.FetchAllTags = b
		return nil
	}
}

// during a checkout operation. fetchAllTags is an option function used to specify whether to create a new branch
// Example:
// ```
// git.Checkout("C:/Users/xxx/Desktop/yaklang", "feat/new-branch", git.checkoutCreate(true))
// ```
func WithCheckoutCreate(b bool) Option {
	return func(c *config) error {
		c.CheckoutCreate = b
		return nil
	}
}

// fetchAllTags is an option function used to specify whether to force
// Example:
// ```
// git.Checkout("C:/Users/xxx/Desktop/yaklang", "old-branch", git.checkoutForce(true))
// ```
func WithCheckoutForce(b bool) Option {
	return func(c *config) error {
		c.CheckoutForce = b
		return nil
	}
}

// checkoutKeep is an option function used to specify whether local changes (index or working tree changes) are retained during checkout (checkout) operations. If If retained, they can be committed to the target branch. The default is false.
// Example:
// ```
// git.Checkout("C:/Users/xxx/Desktop/yaklang", "old-branch", git.checkoutKeep(true))
// ```
func WithCheckoutKeep(b bool) Option {
	return func(c *config) error {
		c.CheckoutKeep = b
		return nil
	}
}

// depth is an option function used to specify the maximum depth when other Git operations (such as Clone), the default is 1
// Example:
// ```
// git.Clone("https://github.com/yaklang/yaklang", "C:/Users/xxx/Desktop/yaklang", git.Depth(1))
// ```
func WithDepth(depth int) Option {
	return func(c *config) error {
		c.Depth = depth
		return nil
	}
}

// force is an options function used to specify other Git operations (such as Pull) Whether to enforce execution, the default is false
// Example:
// ```
// git.Pull("C:/Users/xxx/Desktop/yaklang", git.verify(false), git.force(true))
// ```
func WithForce(b bool) Option {
	return func(c *config) error {
		c.Force = b
		return nil
	}
}

// remote is an option function, used When specifying the remote warehouse name for other Git operations (such as Pull), the default is origin
// Example:
// ```
// git.Pull("C:/Users/xxx/Desktop/yaklang", git.verify(false), git.remote("origin"))
// ```
func WithRemote(remote string) Option {
	return func(c *config) error {
		c.Remote = remote
		return nil
	}
}

// recursive is an option function used to specify whether to recursively clone submodules during other Git operations (such as Clone), the default is false
// Example:
// ```
// git.Clone("https://github.com/yaklang/yaklang", "C:/Users/xxx/Desktop/yaklang", git.recursive(true))
// ```
func WithRecuriveSubmodule(b bool) Option {
	return func(c *config) error {
		c.RecursiveSubmodule = b
		return nil
	}
}

// context is an option function, used to specify the context for other Git operations (such as Clone)
// Example:
// ```
// git.Clone("https://github.com/yaklang/yaklang", "C:/Users/xxx/Desktop/yaklang", git.context(context.New()))
// ```
func WithContext(ctx context.Context) Option {
	return func(c *config) error {
		c.Context, c.Cancel = context.WithCancel(ctx)
		return nil
	}
}

// auth is an option function used to specify other Git operations ( For example, the authentication username and password when Clone)
// Example:
// ```
// git.Clone("https://github.com/yaklang/yaklang", "C:/Users/xxx/Desktop/yaklang", git.auth("admin", "admin"))
// ```
func WithUsernamePassword(username, password string) Option {
	return func(c *config) error {
		c.Username = username
		c.Password = password
		return nil
	}
}

// threads Is a GitHack option function, used to specify the number of concurrency, the default is 8
// Example:
// ```
// git.GitHack("http://127.0.0.1:8787/git/website", "C:/Users/xxx/Desktop/githack-test", git.threads(8))
// ```
func WithThreads(threads int) Option {
	return func(c *config) error {
		c.Threads = threads
		return nil
	}
}

// useLocalGitBinary is a GitHack option function, used to specify whether to use the git binary file of the local environment variable to execute the `git fsck` command. This command is used to restore the complete git repository as much as possible. The default is true
// Example:
// ```
// git.GitHack("http://127.0.0.1:8787/git/website", "C:/Users/xxx/Desktop/githack-test", git.useLocalGitBinary(true))
// ```
func WithUseLocalGitBinary(b bool) Option {
	return func(c *config) error {
		c.UseLocalGitBinary = b
		return nil
	}
}

// httpOpts is a GitHack options function used to specify GitHacks HTTP options, which receives zero to multiple POC request options function
// Example:
// ```
// git.GitHack("http://127.0.0.1:8787/git/website", "C:/Users/xxx/Desktop/githack-test", git.httpOpts(poc.timeout(10), poc.https(true)))
// ```
func WithHTTPOptions(opts ...poc.PocConfig) Option {
	return func(c *config) error {
		c.HTTPOptions = opts
		return nil
	}
}
