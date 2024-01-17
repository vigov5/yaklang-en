package yakgrpc

import (
	"encoding/json"
	"mime"
	"regexp"
	"strings"

	"github.com/gobwas/glob"
	"github.com/jinzhu/gorm"
	"github.com/yaklang/yaklang/common/log"
	"github.com/yaklang/yaklang/common/utils"
	"github.com/yaklang/yaklang/common/yakgrpc/yakit"
)

type MITMFilterManager struct {
	db               *gorm.DB `json:"-"`
	IncludeHostnames []string `json:"includeHostnames"`
	ExcludeHostnames []string `json:"excludeHostnames"`
	IncludeSuffix    []string `json:"includeSuffix"`
	ExcludeSuffix    []string `json:"excludeSuffix"`
	ExcludeMethods   []string `json:"excludeMethods"`
	ExcludeMIME      []string `json:"excludeMIME"`
	ExcludeUri       []string `json:"excludeUri"`
	IncludeUri       []string `json:"includeUri"`
}

var (
	defaultExcludeHostnames = []string{"google.com", "*gstatic.com", "*bdstatic.com", "*google*.com"}
	defaultExcludeSuffix    = []string{
		".css",
		".jpg", ".jpeg", ".png",
		".mp3", ".mp4", ".ico", ".bmp",
		".flv", ".aac", ".ogg", "avi",
		".svg", ".gif", ".woff", ".woff2",
		".doc", ".docx", ".pptx",
		".ppt", ".pdf",
	}
)

var (
	defaultExcludeMethods = []string{"OPTIONS", "CONNECT"}
	defaultExcludeMIME    = []string{
		"image/*",
		"audio/*", "video/*", "*octet-stream*",
		"application/ogg", "application/pdf", "application/msword",
		"application/x-ppt", "video/avi", "application/x-ico",
		"*zip",
	}
)

func _exactChecker(includes, excludes []string, target string) bool {
	excludes = utils.StringArrayFilterEmpty(excludes)
	includes = utils.StringArrayFilterEmpty(includes)

	for _, exclude := range excludes {
		if match, err := regexp.MatchString(exclude, target); err == nil && match {
			return false
		} else if exclude == target {
			return false
		}
	}

	if includes == nil {
		return true
	}

	for _, include := range includes {
		if match, err := regexp.MatchString(include, target); err == nil && match {
			return true
		} else if include == target {
			return true
		}
	}

	return false
}

func _checker(includes, excludes []string, target string) bool {
	excludes = utils.StringArrayFilterEmpty(excludes)
	includes = utils.StringArrayFilterEmpty(includes)

	for _, exclude := range excludes {
		if match, err := regexp.MatchString(exclude, target); err == nil && match {
			return false
		} else if utils.StringGlobContains(exclude, target) {
			return false
		} else if exclude == target {
			return false
		}
	}

	if includes == nil {
		return true
	}

	for _, include := range includes {
		if match, err := regexp.MatchString(include, target); err == nil && match {
			return true
		} else if utils.StringGlobContains(include, target) {
			return true
		} else if include == target {
			return true
		}
	}

	return false
}

func fixSuffix(suf string) string {
	if strings.HasPrefix(suf, ".") {
		return suf
	} else {
		return "." + suf
	}
}

func _suffixChecker(includes, excludes []string, target string) bool {
	excludes = utils.StringArrayFilterEmpty(excludes)
	includes = utils.StringArrayFilterEmpty(includes)

	for _, exclude := range excludes {
		if strings.HasSuffix(target, fixSuffix(exclude)) {
			return false
		} else if exclude == target {
			return false
		}
	}

	if includes == nil {
		return true
	}

	for _, include := range includes {
		if strings.HasSuffix(target, fixSuffix(include)) {
			return true
		} else if include == target {
			return true
		}
	}

	return false
}

func mimeCheckGlobRule(rule string, target string) bool {
	if strings.Contains(rule, "/") && strings.Contains(target, "/") { // If both contain/, then split and match
		ruleType := strings.SplitN(rule, "/", 2)
		targetType := strings.SplitN(target, "/", 2)
		for i := 0; i < 2; i++ {
			if strings.Contains(ruleType[i], "*") {
				rule, err := glob.Compile(ruleType[i])
				if err != nil || !rule.Match(targetType[i]) {
					return false // , any part of the match will fail. False, including glob compilation failure
				}
			} else {
				if ruleType[i] != targetType[i] {
					return false // , if any part of the match fails, false
				}
			}
		}
		return true // , all pass true
	}

	if !strings.Contains(target, "/") && !strings.Contains(rule, "/") { // If neither contains /
		if strings.Contains(rule, "*") { // Try glob to match
			rule, err := glob.Compile(rule)
			if err == nil && rule.Match(target) {
				return true
			}
		} else { // , directly contains
			if utils.IContains(target, rule) {
				return true
			}
		}
		return false
	}

	if strings.Contains(target, "/") && !strings.Contains(rule, "/") { // will be released, only rule does not include /
		targetType := strings.SplitN(target, "/", 2)
		for i := 0; i < 2; i++ {
			if strings.Contains(rule, "*") {
				rule, err := glob.Compile(rule)
				if err != nil {
					continue
				}
				if rule.Match(targetType[i]) {
					return true // , any part of the match is successful, then true
				}
			} else {
				if rule == targetType[i] {
					return true // , any part of the match is successful, then true
				}
			}
		}
		return false // False if all fails
	}

	return false // Only rule has / , then it will directly return false,
}

func _mimeChecker(includes, excludes []string, target string) bool {
	excludes = utils.StringArrayFilterEmpty(excludes)
	includes = utils.StringArrayFilterEmpty(includes)

	if includes == nil {
		for _, rule := range excludes {
			if mimeCheckGlobRule(rule, target) {
				return false // , if excludes is hit, it will be false, that is, filter
			}
		}
		return true
	}

	for _, rule := range excludes {
		if mimeCheckGlobRule(rule, target) {
			return false // , if excludes is hit, it will be false, that is, filter
		}
	}

	for _, rule := range includes {
		if mimeCheckGlobRule(rule, target) {
			return true // , if it hits includes, it will be true, that is,
		}
	}
	return false
}

func (m *MITMFilterManager) Recover() {
	m.ExcludeMethods = defaultExcludeMethods
	m.ExcludeSuffix = defaultExcludeSuffix
	m.ExcludeHostnames = defaultExcludeHostnames
	m.ExcludeMIME = defaultExcludeMIME
	m.ExcludeUri = nil
	m.IncludeUri = nil
	m.IncludeHostnames = nil
	m.IncludeSuffix = nil
	m.Save()
}

var defaultMITMFilterManager = &MITMFilterManager{
	ExcludeHostnames: defaultExcludeHostnames,
	ExcludeSuffix:    defaultExcludeSuffix,
	ExcludeMethods:   defaultExcludeMethods,
	ExcludeMIME:      defaultExcludeMIME,
}

func getInitFilterManager(db *gorm.DB) func() (*MITMFilterManager, error) {
	if db == nil {
		return nil
	}
	return func() (*MITMFilterManager, error) {
		if db == nil {
			return nil, utils.Error("no database set")
		}
		results := yakit.GetKey(db, MITMFilterKeyRecords)
		var manager MITMFilterManager
		err := json.Unmarshal([]byte(results), &manager)
		if err != nil {
			return nil, err
		}
		managerP := &manager
		//managerP.saveHandler = func(filter *MITMFilterManager) {
		//
		//}
		return managerP, nil
	}
}

func NewMITMFilterManager(db *gorm.DB) *MITMFilterManager {
	initFilter := getInitFilterManager(db)
	if initFilter == nil {
		defaultMITMFilterManager.db = db
		return defaultMITMFilterManager
	}
	result, err := initFilter()
	if err != nil || result == nil {
		defaultMITMFilterManager.db = db
		return defaultMITMFilterManager
	}
	result.db = db
	return result
}

func (m *MITMFilterManager) IsEmpty() bool {
	return len(m.ExcludeMIME) <= 0 && len(m.ExcludeMethods) <= 0 &&
		len(m.ExcludeSuffix) <= 0 && len(m.ExcludeHostnames) <= 0 &&
		len(m.IncludeHostnames) <= 0 && len(m.IncludeSuffix) <= 0 &&
		len(m.ExcludeUri) <= 0 && len(m.IncludeUri) <= 0
}

func (m *MITMFilterManager) Save() {
	db := m.db
	if db == nil {
		return
	}

	if m.IsEmpty() {
		m.Recover()
		return
	}

	result, err := json.Marshal(m)
	if err != nil {
		log.Errorf("marshal mitm filter failed: %s", err)
		return
	}
	err = yakit.SetKey(db, MITMFilterKeyRecords, string(result))
	if err != nil {
		log.Errorf("set filter db key failed: %s", err)
	}
}

func (m *MITMFilterManager) IsMIMEPassed(ct string) bool {
	parsed, _, _ := mime.ParseMediaType(ct)
	if parsed != "" {
		ct = parsed
	}
	return _mimeChecker(nil, m.ExcludeMIME, ct)
}

// IsPassed return true if passed, false if filtered out
func (m *MITMFilterManager) IsPassed(method string, hostport, urlStr string, ext string, isHttps bool) bool {
	var passed bool

	passed = _exactChecker(nil, m.ExcludeMethods, method)
	if !passed {
		log.Debugf("[%v] url: %s is filtered via method", method, truncate(urlStr))
		return false
	}

	passed = _suffixChecker(m.IncludeSuffix, m.ExcludeSuffix, strings.ToLower(ext))
	if !passed {
		log.Debugf("url: %v is filtered via suffix(%v)", truncate(urlStr), ext)
		return false
	}

	passed = _checker(m.IncludeHostnames, m.ExcludeHostnames, hostport)
	if !passed {
		log.Debugf("url: %s is filtered via hostnames(%v)", truncate(urlStr), hostport)
		return false
	}

	passed = _checker(m.IncludeUri, m.ExcludeUri, urlStr)
	if !passed {
		log.Debugf("url: %s is filtered via uri(url)", truncate(urlStr))
		return false
	}
	return true
}
