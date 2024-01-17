package yakit

import (
	"context"
	"net/url"
	"sort"
	"strings"

	"github.com/jinzhu/gorm"
	"github.com/yaklang/yaklang/common/log"
)

func FilterHTTPFlowBySchema(db *gorm.DB, schema string) *gorm.DB {
	if schema != "" {
		db = db.Where("SUBSTR(url, 1, ?) = ?", len(schema+"://"), schema+"://") //.Debug()
	}
	return db
}

type WebsiteTree struct {
	Path         string
	NextParts    []*WebsiteNextPart
	HaveChildren bool
}

type WebsiteNextPart struct {
	Schema       string
	NextPart     string
	HaveChildren bool
	Count        int
	IsQuery      bool
	RawQueryKey  string
	RawNextPart  string
	IsFile       bool
}

func trimPathWithOneSlash(path string) string {
	path, _, _ = strings.Cut(path, "?")
	if strings.HasPrefix(path, "/") {
		path = path[1:]
	}
	if strings.HasSuffix(path, "/") {
		path = path[:len(path)-1]
	}
	return path
}

func GetHTTPFlowDomainsByDomainSuffix(db *gorm.DB, domainSuffix string) []*WebsiteNextPart {
	db = FilterHTTPFlowByDomain(db, domainSuffix)
	db = db.Select(
		"DISTINCT SUBSTR(url, INSTR(url, '://') + 3, INSTR(SUBSTR(url, INSTR(url, '://') + 3), '/') - 1) as next_part,\n" +
			"SUBSTR(url, 0, INSTR(url, '://'))",
	).Table("http_flows").Limit(1000) //.Debug()
	if rows, err := db.Rows(); err != nil {
		log.Error("query nextPart for website tree failed: %s", err)
		return nil
	} else {
		resultMap := make(map[string]*WebsiteNextPart)
		for rows.Next() {
			var nextPart string
			var schema string
			rows.Scan(&nextPart, &schema)
			if nextPart == "" {
				continue
			}
			haveChildren := false
			nextPartItem, after, splited := strings.Cut(nextPart, "/")
			if splited && after != "" {
				haveChildren = true
			}
			// Create a schema and nextPart Unique key
			uniqueKey := schema + "://" + nextPartItem
			if result, ok := resultMap[uniqueKey]; ok {
				result.Count++
			} else {
				resultMap[uniqueKey] = &WebsiteNextPart{
					NextPart: uniqueKey, HaveChildren: haveChildren, Count: 1,
					Schema: schema,
				}
			}
		}
		var data []*WebsiteNextPart
		for _, r := range resultMap {
			data = append(data, r)
		}
		sort.SliceStable(data, func(i, j int) bool {
			return data[i].NextPart > data[j].NextPart
		})
		return data
	}
}

func matchURL(u string, searchPath string) bool {
	var err error
	if strings.Contains(searchPath, "%") {
		searchPath, err = url.PathUnescape(searchPath)
		if err != nil {
			return false
		}
	}

	// Parse the URL
	parsedURL, _ := url.Parse(u)

	normalizedPath := strings.Join(strings.FieldsFunc(parsedURL.Path, func(r rune) bool {
		return r == '/'
	}), "/")

	normalizedSearchPath := strings.Join(strings.FieldsFunc(searchPath, func(r rune) bool {
		return r == '/'
	}), "/")

	// Make sure the search path ends with "/" at the beginning
	searchPath = "/" + strings.TrimLeft(normalizedSearchPath, "/")
	searchPath = strings.TrimRight(searchPath, "/")
	// Get the path and make sure it ends with "/" at the beginning
	path := "/" + strings.TrimLeft(normalizedPath, "/")

	// Split the search path and URL path
	searchSegments := strings.Split(searchPath, "/")
	pathSegments := strings.Split(path, "/")

	// Check if the path starts with a segment of the search path
	match := true
	for i := 1; i < len(searchSegments); i++ {
		if i >= len(pathSegments) || searchSegments[i] != pathSegments[i] {
			match = false
			break
		}
	}

	return match
}

// findNextPathSegment Return the path segment immediately after the target in the url, containing the correct number of slashes
func findNextPathSegment(url, target string) string {
	// Split url and target
	urlSegments := strings.Split(url, "/")
	targetSegments := strings.Split(target, "/")

	targetIndex, slashCount := 0, 0

	// Traverse urlSegments to find targetSegments
	for _, segment := range urlSegments {
		targetIndex++

		if segment == "" {
			slashCount++
			continue
		}

		if targetIndex-1 < len(targetSegments) && segment == targetSegments[targetIndex-1] {
			slashCount = 1 // Reset the slash count
			continue
		}

		// Find the target, return the next non-empty segment, along with the previously calculated slash
		if segment != "" {
			return strings.Repeat("/", slashCount) + segment
		}

	}

	return ""
}

func GetHTTPFlowNextPartPathByPathPrefix(db *gorm.DB, originPathPrefix string) []*WebsiteNextPart {
	//pathPrefix := strings.Join(strings.FieldsFunc(originPathPrefix, func(r rune) bool {
	//	return r == '/'
	//}), "/")
	pathPrefix := strings.TrimLeft(originPathPrefix, "/")
	db = db.Select("url").Table("http_flows").Where("url LIKE ?", `%`+pathPrefix+`%`).Limit(1000) //.Debug()
	urlsMap := make(map[string]bool)
	var urls []string
	for u := range YieldHTTPUrl(db, context.Background()) {
		if _, exists := urlsMap[u.Url]; !exists {
			urlsMap[u.Url] = true
			if matchURL(u.Url, originPathPrefix) {
				urls = append(urls, u.Url)
			}
		}
	}

	// Initialize a mapping to store website structure
	resultMap := make(map[string]*WebsiteNextPart)

	// Assume urls is your list of URLs
	for _, us := range urls {
		usC := strings.SplitN(us, "?", 2)[0] + "%2f"
		uc, _ := url.Parse(usC)
		u, _ := url.Parse(us)
		if u.Path == "" || uc.RawPath == "" {
			continue
		}
		// Find the target string, in order to solve multiple / question whether the path segment is already in the resultMap
		rawNextPart := findNextPathSegment(strings.TrimSuffix(uc.RawPath, "%2f"), originPathPrefix)
		// Remove the extra slashes in the URL path
		normalizedPath := strings.Join(strings.FieldsFunc(u.Path, func(r rune) bool {
			return r == '/'
		}), "/")

		normalizedOriginPathPrefix := strings.Join(strings.FieldsFunc(pathPrefix, func(r rune) bool {
			return r == '/'
		}), "/")

		path := strings.Trim(normalizedPath, "/")

		pathPrefix, err := url.PathUnescape(normalizedOriginPathPrefix)
		if err != nil {
			continue
		}

		suffix := strings.TrimPrefix(path, pathPrefix)

		suffix = strings.Trim(suffix, "/")

		nextSegment, after, splited := strings.Cut(suffix, "/")

		// According to whether the path Split, decide if there is a sub-path
		haveChildren := splited && after != ""

		if nextSegment == "" && u.RawQuery == "" {
			continue
		}

		if nextSegment != "" {
			node, ok := resultMap[nextSegment]
			// Check the root
			if !ok {
				if u.RawQuery != "" {
					haveChildren = true
				}
				node = &WebsiteNextPart{
					NextPart:     nextSegment,
					HaveChildren: haveChildren, // If there are multiple path segments, there are child nodes
					Count:        1,            // Initialize the count to 1
					Schema:       u.Scheme,
					RawNextPart:  rawNextPart,
				}
				if strings.Contains(nextSegment, ".") {
					node.IsFile = true
				}
				resultMap[nextSegment] = node

			} else {
				// If it already exists and is not a file, then Increment count
				if !strings.Contains(nextSegment, ".") && haveChildren {
					node.Count++
				}

				if u.RawQuery != "" {
					haveChildren = true
				}
				node.HaveChildren = haveChildren
			}
		}

		// If the query parameter exists, add it to the root path segment
		if u.RawQuery != "" && path == pathPrefix {
			for key := range u.Query() {
				if len(key) == 0 {
					continue
				}
				queryNode, ok := resultMap[key]
				if !ok {
					queryNode = &WebsiteNextPart{
						NextPart:     key,
						HaveChildren: false, // If there are multiple path segments, there are child nodes
						RawQueryKey:  key,
						IsQuery:      true,
						Schema:       u.Scheme,
					}
					//if strings.Contains(key, ".") {
					//	queryNode.HaveChildren = true
					//}
					resultMap[key] = queryNode
				} else {
					queryNode.Count++
				}
			}
		}

	}

	// resultMap now contains all paths and query parameters, as well as their hierarchies and counts
	var data []*WebsiteNextPart
	for _, r := range resultMap {
		data = append(data, r)
	}
	sort.SliceStable(data, func(i, j int) bool {
		return data[i].NextPart > data[j].NextPart
	})
	return data
}

func FilterHTTPFlowPathPrefix(db *gorm.DB, pathPrefix string) *gorm.DB {
	if pathPrefix != "" {
		pathPrefix := trimPathWithOneSlash(pathPrefix)
		template := `SUBSTR(url,INSTR(url, '://') + 3 + INSTR(SUBSTR(url, INSTR(url, '://') + 3), '/'),CASE WHEN INSTR(SUBSTR(url, INSTR(url, '://') + 3), ?) > 0 THEN INSTR(SUBSTR(url, INSTR(url, '://') + 3), ?) - INSTR(SUBSTR(url, INSTR(url, '://') + 3), '/') - 1 ELSE LENGTH(url) END) LIKE ?`
		db = db.Where(template, "?", "?", pathPrefix+"%")
	}
	return db
}

func FilterHTTPFlowByDomain(db *gorm.DB, domain string) *gorm.DB {
	// query url
	// schema://domain
	// no '/' in domain and schema
	if strings.Contains(domain, "%") {
		domain = strings.ReplaceAll(domain, "%", "%%")
		domain = strings.Trim(domain, "%")
	}

	if domain != "" {
		db = db.Where(`SUBSTR(url, INSTR(url, '://') + 3, INSTR(SUBSTR(url, INSTR(url, '://') + 3), '/') - 1) LIKE ?`, "%"+domain)
		// db = db.Where(`SUBSTR(url, INSTR(url, '://') + 3, INSTR(SUBSTR(url, INSTR(url, '://') + 3), '/') - 1) = ?`, domain)

	}
	return db
}

func FilterHTTPFlowByRuntimeID(db *gorm.DB, runtimeID string) *gorm.DB {
	return db.Where("runtime_id = ?", runtimeID)
}
