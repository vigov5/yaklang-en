package yakit

import (
	"github.com/jinzhu/gorm"
	"github.com/yaklang/yaklang/common/utils"
	"regexp"
	"sort"
	"strings"
	"time"
)

type RssFeed struct {
	gorm.Model

	SourceXmlUrl    string
	Hash            string     `gorm:"columns:hash;unique_index"`
	Title           string     `json:"title,omitempty"`
	Description     string     `json:"description,omitempty"`
	Link            string     `json:"link,omitempty"`
	FeedLink        string     `json:"feedLink,omitempty"`
	Updated         string     `json:"updated,omitempty"`
	UpdatedParsed   *time.Time `json:"updatedParsed,omitempty"`
	Published       string     `json:"published,omitempty"`
	PublishedParsed *time.Time `json:"publishedParsed,omitempty"`
	Author          string     `json:"author,omitempty"`
	AuthorEmail     string     `json:"author_email,omitempty"`
	Language        string     `json:"language,omitempty"`
	ImageUrl        string     `json:"image_url,omitempty"`
	ImageName       string     `json:"image_name,omitempty"`
	Copyright       string     `json:"copyright,omitempty"`
	Generator       string     `json:"generator,omitempty"`
	Categories      string     `json:"categories,omitempty"`
	FeedType        string     `json:"feedType"`
	FeedVersion     string     `json:"feedVersion"`
}

func (b *RssFeed) CalcHash() string {
	return utils.CalcSha1(
		b.Title, b.Description,
		b.Link, b.FeedLink, b.Author, b.AuthorEmail, b.ImageUrl,
	)
}

func (b *RssFeed) BeforeSave() error {
	b.Hash = b.CalcHash()
	return nil
}

type Briefing struct {
	gorm.Model

	SourceXmlUrl    string
	RssFeedHash     string
	Hash            string     `gorm:"columns:hash;unique_index"`
	Title           string     `json:"title,omitempty"`
	Description     string     `json:"description,omitempty"`
	Content         string     `json:"content,omitempty"`
	Link            string     `json:"link,omitempty"`
	Updated         string     `json:"updated,omitempty"`
	UpdatedParsed   *time.Time `json:"updatedParsed,omitempty"`
	Published       string     `json:"published,omitempty"`
	PublishedParsed *time.Time `json:"publishedParsed,omitempty"`
	Author          string     `json:"author,omitempty"`
	AuthorEmail     string     `json:"author_email,omitempty"`
	GUID            string     `json:"guid,omitempty"`
	ImageUrl        string     `json:"image_url,omitempty"`
	ImageName       string     `json:"image_name,omitempty"`
	Categories      string     `json:"categories,omitempty"`
	Tags            string     `json:"tags"`
	IsRead          bool       `json:"is_read"`
}

var (
	tagConditions = map[string][]string{
		"industrial control":      {"industrial control", "IoT", "Nuclear power"},
		"Kernel security":    {"Kernel", "Kernel Vul", "Kernel vulnerability", "Kernel module", "Privilege escalation", "BIOS"},
		"RCE":     {"Remote Code Execution", "Remote code execution", "Arbitrary code execution", "rbitrary code"},
		"Remote code execution":  {"Remote Code Execution", "Remote code execution", "Arbitrary code execution", "rbitrary code", "code execution"},
		"Black intelligence":    {"APT", "Hijacking", "Poisoning", "Puddle", "Harpoon", "Worm", "Phishing", "Selling", "Profit"},
		"SDLC":    {"security life cycle", "sdl", "Devsecops", "devops"},
		"CTF":     {"awd", "ctf"},
		"0day":    {"0 day", "0day"},
		"Privilege escalation":      {"privilege escalation", "Privilege escalation"},
		"Brute force":    {"brute-force", "Brute force", "Exploding"},
		"DoS":     {"dos", "Denial of service attack"},
		"SQL injection":   {"SQL injection", "SQL injection", "SQL injection", "sqlinjection", "sqli"},
		"XSS":     {"cross-site scripting", "XSS", "cross-site scripting"},
		"CSRF":    {"cross-site request forgery", "CSRF", "cross-site request forgery"},
		"Exploit": {"exploit", "Exploit", "Using"},
	}

	cveTitleRegexp = regexp.MustCompile(`CVE-\d+-\d+[ \t]*(\(.*?\))`)
)

func (b *Briefing) BeforeSave() error {
	b.Hash = b.CalcHash()

	rawMaterial := strings.Join([]string{b.Title, b.Description, b.Content, b.Link}, " ")
	var tags []string
	for tag, conds := range tagConditions {
		if utils.IStringContainsAnyOfSubString(rawMaterial, conds) {
			tags = append(tags, tag)
			//b.Tags = append(b.Tags, tag)
		}
	}

	if strings.HasPrefix(b.Title, "CVE") {
		for _, subs := range cveTitleRegexp.FindAllStringSubmatch(b.Title, -1) {
			if len(subs) > 1 {
				tag := subs[1]
				tag = strings.ReplaceAll(tag, ",", "|")
				if len(tag) > 20 {
					tag = tag[:17] + "..."
				}
				tags = append(tags, tag)
				//b.Tags = append(b.Tags, tag)
			}
		}
	}

	sort.Strings(tags)
	b.Tags = strings.Join(tags, ",")

	return nil
}

func (b *Briefing) CalcHash() string {
	return utils.CalcSha1(b.Title, b.Description, b.Content, b.Updated, b.Published, b.Tags)
}
