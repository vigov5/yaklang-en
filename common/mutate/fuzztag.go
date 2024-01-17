package mutate

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/samber/lo"
	"github.com/yaklang/yaklang/common/consts"
	"github.com/yaklang/yaklang/common/filter"
	"github.com/yaklang/yaklang/common/fuzztagx/parser"
	"github.com/yaklang/yaklang/common/utils/regen"
	"github.com/yaklang/yaklang/common/yak/yaklib/codec"
	"github.com/yaklang/yaklang/common/yso"

	"github.com/yaklang/yaklang/common/fuzztag"
	"github.com/yaklang/yaklang/common/log"
	"github.com/yaklang/yaklang/common/utils"
)

// empty content
var fuzztagfallback = []string{""}

type FuzzTagDescription struct {
	TagName          string
	Handler          func(string) []string
	HandlerEx        func(string) []*fuzztag.FuzzExecResult
	ErrorInfoHandler func(string) ([]string, error)
	IsDyn            bool
	IsDynFun         func(name, params string) bool
	Alias            []string
	Description      string
}

func AddFuzzTagDescriptionToMap(methodMap map[string]*parser.TagMethod, f *FuzzTagDescription) {
	if f == nil {
		return
	}
	name := f.TagName
	alias := f.Alias
	var expand map[string]any
	if f.IsDynFun != nil {
		expand = map[string]any{
			"IsDynFun": f.IsDynFun,
		}
	}
	methodMap[name] = &parser.TagMethod{
		Name:   name,
		IsDyn:  f.IsDyn,
		Expand: expand,
		Fun: func(s string) (result []*parser.FuzzResult, err error) {
			defer func() {
				if r := recover(); r != nil {
					if v, ok := r.(error); ok {
						err = v
					} else {
						err = errors.New(utils.InterfaceToString(r))
					}
				}
			}()
			if f.Handler != nil {
				for _, d := range f.Handler(s) {
					result = append(result, parser.NewFuzzResultWithData(d))
				}
			} else if f.HandlerEx != nil {
				for _, data := range f.HandlerEx(s) {
					var verbose string
					showInfo := data.ShowInfo()
					if len(showInfo) != 0 {
						verbose = utils.InterfaceToString(showInfo)
					}
					result = append(result, parser.NewFuzzResultWithDataVerbose(data.Data(), verbose))
				}
			} else if f.ErrorInfoHandler != nil {
				res, err := f.ErrorInfoHandler(s)
				fuzzRes := []*parser.FuzzResult{}
				for _, r := range res {
					fuzzRes = append(fuzzRes, parser.NewFuzzResultWithData(r))
				}
				return fuzzRes, err
			} else {
				return nil, errors.New("no handler")
			}
			return
		},
	}
	for _, a := range alias {
		methodMap[a] = methodMap[name]
	}
}

var (
	existedFuzztag []*FuzzTagDescription
	tagMethodMap   = map[string]*parser.TagMethod{}
)

func AddFuzzTagToGlobal(f *FuzzTagDescription) {
	AddFuzzTagDescriptionToMap(tagMethodMap, f)
}

func init() {
	AddFuzzTagToGlobal(&FuzzTagDescription{
		TagName: "trim",
		Handler: func(s string) []string {
			return []string{strings.TrimSpace(s)}
		},
		Description: "removes excess before and after. spaces, for example: `{{trim( abc )}}`, the result is: `abc`",
	})
	AddFuzzTagToGlobal(&FuzzTagDescription{
		TagName: "substr",
		Handler: func(s string) []string {
			index := strings.LastIndexByte(s, '|')
			if index == -1 {
				return []string{s}
			}
			before, after := s[:index], s[index+1:]
			if strings.Contains(after, ",") {
				start, length := sepToEnd(after, ",")
				startInt := codec.Atoi(start)
				lengthInt := codec.Atoi(length)
				if lengthInt <= 0 {
					lengthInt = len(before)
				}
				if startInt >= len(before) {
					return []string{""}
				}
				if startInt+lengthInt >= len(before) {
					return []string{before[startInt:]}
				}
				return []string{before[startInt : startInt+lengthInt]}
			} else {
				start := codec.Atoi(after)
				if len(before) > start {
					return []string{before[start:]}
				}
				return []string{""}
			}
		},
		Description: "outputs a substring of a string, defined as {{substr(abc|start,length)}}, for example:{{substr(abc|1)}}according to the step size, the result is: bc,{{substr(abcddd|1,2)}}, the result is: bc",
	})

	AddFuzzTagToGlobal(&FuzzTagDescription{
		TagName: "fuzz:password",
		Handler: func(s string) []string {
			origin, level := sepToEnd(s, "|")
			levelInt := atoi(level)
			return fuzzpass(origin, levelInt)
		},
		Alias:       []string{"fuzz:pass"},
		Description: "randomly generates possible passwords based on the input operation (default is root/admin generates)",
	})

	AddFuzzTagToGlobal(&FuzzTagDescription{
		TagName: "zlib:encode",
		Handler: func(s string) []string {
			res, _ := utils.ZlibCompress(s)
			return []string{
				string(res),
			}
		},
		Alias:       []string{"zlib:enc", "zlibc", "zlib"},
		Description: "Zlib encoding, zlib compresses the tag content,",
	})

	AddFuzzTagToGlobal(&FuzzTagDescription{
		TagName: "zlib:decode",
		Handler: func(s string) []string {
			res, _ := utils.ZlibDeCompress([]byte(s))
			return []string{
				string(res),
			}
		},
		Alias:       []string{"zlib:dec", "zlibdec", "zlibd"},
		Description: "Zlib decoding, zlib the content in the tag Decode",
	})

	AddFuzzTagToGlobal(&FuzzTagDescription{
		TagName: "gzip:encode",
		Handler: func(s string) []string {
			res, _ := utils.GzipCompress(s)
			return []string{
				string(res),
			}
		},
		Alias:       []string{"gzip:enc", "gzipc", "gzip"},
		Description: "Gzip encoding, gzip compresses the tag content",
	})

	AddFuzzTagToGlobal(&FuzzTagDescription{
		TagName: "gzip:decode",
		Handler: func(s string) []string {
			res, _ := utils.GzipDeCompress([]byte(s))
			return []string{
				string(res),
			}
		},
		Alias:       []string{"gzip:dec", "gzipdec", "gzipd"},
		Description: "Gzip decoding, gzip the content in the tag",
	})

	AddFuzzTagToGlobal(&FuzzTagDescription{
		TagName: "fuzz:username",
		Handler: func(s string) []string {
			origin, level := sepToEnd(s, "|")
			levelInt := atoi(level)
			return fuzzuser(origin, levelInt)
		},
		Alias:       []string{"fuzz:user"},
		Description: "randomly generates possible user names based on the input operation (the default is root/admin generates)",
	})

	AddFuzzTagToGlobal(&FuzzTagDescription{
		TagName: "date",
		Handler: func(s string) []string {
			if s == "" {
				return []string{utils.JavaTimeFormatter(time.Now(), "YYYY-MM-dd")}
			}
			return []string{utils.JavaTimeFormatter(time.Now(), s)}
		},
		Description: "generates a time in the format YYYY-MM-dd. If the format is specified, the time will be generated in the specified format.",
	})

	AddFuzzTagToGlobal(&FuzzTagDescription{
		TagName: "datetime",
		Handler: func(s string) []string {
			if s == "" {
				return []string{utils.JavaTimeFormatter(time.Now(), "YYYY-MM-dd")}
			}
			return []string{utils.JavaTimeFormatter(time.Now(), s)}
		},
		Alias:       []string{"time"},
		Description: "to generate a time in the format YYYY-MM-dd HH:mm:ss. If the format is specified, the time",
	})

	AddFuzzTagToGlobal(&FuzzTagDescription{
		TagName: "timestamp",
		IsDyn:   true,
		Handler: func(s string) []string {
			switch strings.ToLower(s) {
			case "s", "sec", "seconds", "second":
				return []string{fmt.Sprint(time.Now().Unix())}
			case "ms", "milli", "millis", "millisecond", "milliseconds":
				return []string{fmt.Sprint(time.Now().UnixMilli())}
			case "ns", "nano", "nanos", "nanosecond", "nanoseconds":
				return []string{fmt.Sprint(time.Now().UnixNano())}
			}
			return []string{fmt.Sprint(time.Now().Unix())}
		},
		Description: "generates a timestamp, the default unit is seconds, you can specify the unit: s, ms, ns: {{timestamp(s)}}",
	})

	AddFuzzTagToGlobal(&FuzzTagDescription{
		TagName: "uuid",
		Handler: func(s string) []string {
			result := []string{}
			for i := 0; i < atoi(s); i++ {
				result = append(result, uuid.New().String())
			}
			if len(result) == 0 {
				return []string{uuid.New().String()}
			}
			return result
		},
		Description: "Generate a random uuid, if the number is specified, the specified number will be generated The uuid",
	})

	AddFuzzTagToGlobal(&FuzzTagDescription{
		TagName: "trim",
		Handler: func(s string) []string {
			return []string{
				strings.TrimSpace(s),
			}
		},
		Description: "Remove the spaces on both sides of the string, generally used with other tags, such as:{{trim({{x(dict)}})}}",
	})

	AddFuzzTagToGlobal(&FuzzTagDescription{
		TagName: "null",
		Handler: func(s string) []string {
			if ret := atoi(s); ret > 0 {
				return []string{
					strings.Repeat("\x00", ret),
				}
			}
			return []string{
				"\x00",
			}
		},
		Alias:       []string{"nullbyte"},
		Description: "generates a null byte, if the number is specified, the specified number of null bytes {{null(5)}} to generate 5 Null bytes",
	})

	AddFuzzTagToGlobal(&FuzzTagDescription{
		TagName: "crlf",
		Handler: func(s string) []string {
			if ret := codec.Atoi(s); ret > 0 {
				return []string{
					strings.Repeat("\r\n", ret),
				}
			}
			return []string{
				"\r\n",
			}
		},
		Description: "generates a CRLF, if the number is specified, the specified number of CRLF will be generated {{crlf(5)}} means generating 5 CRLFs.",
	})

	AddFuzzTagToGlobal(&FuzzTagDescription{
		TagName: "gb18030",
		Handler: func(s string) []string {
			g, err := codec.Utf8ToGB18030([]byte(s))
			if err != nil {
				return []string{s}
			}
			return []string{string(g)}
		},
		Description: `Convert the string to GB18030 encoding, for example:{{gb18030 (Hello)}}`,
	})
	AddFuzzTagToGlobal(&FuzzTagDescription{
		TagName: "gb18030toUTF8",
		Handler: func(s string) []string {
			g, err := codec.GB18030ToUtf8([]byte(s))
			if err != nil {
				return []string{s}
			}
			return []string{string(g)}
		},
		Description: `Convert a string to UTF8 encoding, for example:{{gb18030toUTF8({{hexd(c4e3bac3)}})}}`,
	})

	AddFuzzTagToGlobal(&FuzzTagDescription{
		TagName: "padding:zero",
		Handler: func(s string) []string {
			origin, paddingTotal := sepToEnd(s, "|")
			right := strings.HasPrefix(paddingTotal, "-")
			paddingTotal = strings.TrimLeft(paddingTotal, "-+")
			if ret := atoi(paddingTotal); ret > 0 {
				if right {
					return []string{
						strings.Repeat("0", ret-len(origin)) + origin,
					}
				}
				return []string{
					strings.Repeat("0", ret-len(origin)) + origin,
				}
			}
			return []string{origin}
		},
		Alias:       []string{"zeropadding", "zp"},
		Description: "uses 0 padding to compensate for insufficient string length,{{zeropadding(abc|5)}} means filling abc to a string of length 5 (00abc),{{zeropadding(abc|-5)}} means padding abc to a string of length 5, and padding (abc00) on the right",
	})

	AddFuzzTagToGlobal(&FuzzTagDescription{
		TagName: "padding:null",
		Handler: func(s string) []string {
			origin, paddingTotal := sepToEnd(s, "|")
			right := strings.HasPrefix(paddingTotal, "-")
			paddingTotal = strings.TrimLeft(paddingTotal, "-+")
			if ret := atoi(paddingTotal); ret > 0 {
				if right {
					return []string{
						strings.Repeat("\x00", ret-len(origin)) + origin,
					}
				}
				return []string{
					strings.Repeat("\x00", ret-len(origin)) + origin,
				}
			}
			return []string{origin}
		},
		Alias:       []string{"nullpadding", "np"},
		Description: "use \\x00 to fill the problem of insufficient string length,{{nullpadding(abc|5)}} means filling abc to a string of length 5 (\\x00\\x00abc),{{nullpadding(abc|-5)}} means padding abc to a string of length 5, and padding (abc\\ x00\\x00)",
	})

	AddFuzzTagToGlobal(&FuzzTagDescription{
		TagName: "char",
		Handler: func(s string) []string {
			if s == "" {
				return []string{""}
			}

			if !strings.Contains(s, "-") {
				log.Errorf("bad char params: %v, eg.: %v", s, "a-z")
				return []string{""}
			}

			ret := strings.Split(s, "-")
			switch len(ret) {
			case 2:
				p1, p2 := ret[0], ret[1]
				if !(len(p1) == 1 && len(p2) == 1) {
					log.Errorf("start char or end char is not char(1 byte): %v eg.: %v", s, "a-z")
					return []string{""}
				}

				p1Byte := []byte(p1)[0]
				p2Byte := []byte(p2)[0]
				var rets []string
				min, max := utils.MinByte(p1Byte, p2Byte), utils.MaxByte(p1Byte, p2Byte)
				for i := min; i <= max; i++ {
					rets = append(rets, string(i))
				}
				return rets
			default:
				log.Errorf("bad params[%s], eg.: %v", s, "a-z")
			}
			return []string{""}
		},
		Alias:       []string{"c", "ch"},
		Description: "generates a character, for example: `{{char(a-z)}}`, the result is [a b c ... x y z]",
	})

	AddFuzzTagToGlobal(&FuzzTagDescription{
		TagName: "repeat",
		Handler: func(s string) []string {
			i := atoi(s)
			if i == 0 {
				chr, right := sepToEnd(s, "|")
				if repeatTimes := atoi(right); repeatTimes > 0 {
					return lo.Times(repeatTimes, func(index int) string {
						return chr
					})
				}
			} else {
				return make([]string, i)
			}
			return []string{""}
		},
		Description: "Repeat a string or a number of times, for example:`{{repeat(abc|3)}}`, the result is: abcabcabc, or `{{repeat(3)}}`, the result is repeated three times",
	})

	AddFuzzTagToGlobal(&FuzzTagDescription{
		TagName: "repeat:range",
		Handler: func(s string) []string {
			origin, times := sepToEnd(s, "|")
			if ret := atoi(times); ret > 0 {
				results := make([]string, ret+1)
				for i := 0; i < ret+1; i++ {
					if i == 0 {
						results[i] = ""
						continue
					}
					results[i] = strings.Repeat(origin, i)
				}
				return results
			}
			return []string{s}
		},
		Description: "repeats a string and outputs all repeated steps, for example: `{{repeat(abc|3)}}`, the result is: ['' abc abcabc abcabcabc]",
	})

	AddFuzzTagToGlobal(&FuzzTagDescription{
		TagName: "payload",
		Handler: func(s string) []string {
			db := consts.GetGormProfileDatabase()
			if db == nil {
				return []string{s}
			}
			for _, s := range utils.PrettifyListFromStringSplited(s, ",") {
				group, folder := "", ""
				ss := strings.Split(s, "/")
				if len(ss) == 2 {
					folder = ss[0]
					group = ss[1]
					if group == "*" {
						group = ""
					}
				} else {
					group = ss[0]
				}

				if group != "" && folder != "" {
					db = db.Or("`group` = ? AND `folder` = ?", group, folder)
				} else if group != "" {
					db = db.Or("`group` = ?", group)
				} else if folder != "" {
					db = db.Or("`folder` = ?", folder)
				}
			}

			var payloads []string
			rows, err := db.Table("payloads").Select("content, is_file").Order("hit_count desc").Rows()
			if err != nil {
				return []string{s}
			}
			var (
				payloadRaw string
				isFile     bool
			)
			f := filter.NewFilter()
			for rows.Next() {
				err := rows.Scan(&payloadRaw, &isFile)
				if err != nil {
					log.Errorf("sql scan error: %v", err)
					return payloads
				}
				unquoted, err := strconv.Unquote(payloadRaw)
				if err == nil {
					payloadRaw = unquoted
				}

				if isFile {
					ch, err := utils.FileLineReader(payloadRaw)
					if err != nil {
						log.Errorf("read payload err: %v", err)
						continue
					}
					for line := range ch {
						lineStr := utils.UnsafeBytesToString(line)
						raw, err := strconv.Unquote(lineStr)
						if err == nil {
							lineStr = raw
						}
						if f.Exist(lineStr) {
							continue
						}
						f.Insert(lineStr)
						payloads = append(payloads, lineStr)
					}
				} else {
					if f.Exist(payloadRaw) {
						continue
					}
					f.Insert(payloadRaw)
					payloads = append(payloads, payloadRaw)
				}
			}
			return payloads
		},
		Alias:       []string{"x"},
		Description: "Load Payload from the database, you can specify the payload group or folder, `{{payload(groupName)}}`, `{{payload(folder/*)}}`",
	})

	AddFuzzTagToGlobal(&FuzzTagDescription{
		TagName:     "array",
		Alias:       []string{"list"},
		Description: "Set an array, use `|` split, for example: `{{array(1|2|3)}}`, the result is: [1,2,3],",
		Handler: func(s string) []string {
			return strings.Split(s, "|")
		},
	})

	AddFuzzTagToGlobal(&FuzzTagDescription{
		TagName:     "ico",
		Handler:     func(s string) []string { return []string{"\x00\x00\x01\x00\x01\x00\x20\x20"} },
		Description: "generates an ico file header, for example `{{ico}}`",
	})

	AddFuzzTagToGlobal(&FuzzTagDescription{
		TagName: "tiff",
		Handler: func(s string) []string {
			return []string{"\x4d\x4d", "\x49\x49"}
		},
		Description: "Generate a tiff file header, for example`{{tiff}}`",
	})

	AddFuzzTagToGlobal(&FuzzTagDescription{TagName: "bmp", Handler: func(s string) []string { return []string{"\x42\x4d"} }, Description: "generates a PNG file header. {{bmp}}"})

	AddFuzzTagToGlobal(&FuzzTagDescription{
		TagName:     "gif",
		Handler:     func(s string) []string { return []string{"GIF89a"} },
		Description: "Generate gif file header",
	})

	AddFuzzTagToGlobal(&FuzzTagDescription{
		TagName: "png",
		Handler: func(s string) []string {
			return []string{
				"\x89PNG" +
					"\x0d\x0a\x1a\x0a" +
					"\x00\x00\x00\x0D" +
					"IHDR\x00\x00\x00\xce\x00\x00\x00\xce\x08\x02\x00\x00\x00" +
					"\xf9\x7d\xaa\x93",
			}
		},
		Description: "generates PNG file header",
	})

	AddFuzzTagToGlobal(&FuzzTagDescription{
		TagName: "jpg",
		Handler: func(s string) []string {
			return []string{
				"\xff\xd8\xff\xe0\x00\x10JFIF" + s + "\xff\xd9",
				"\xff\xd8\xff\xe1\x00\x1cExif" + s + "\xff\xd9",
			}
		},
		Alias:       []string{"jpeg"},
		Description: "Generate jpeg / jpg File header",
	})

	const puncSet = `<>?,./:";'{}[]|\_+-=)(*&^%$#@!'"` + "`"
	var puncArr []string
	for _, s := range puncSet {
		puncArr = append(puncArr, string(s))
	}
	AddFuzzTagToGlobal(&FuzzTagDescription{
		TagName: "punctuation",
		Handler: func(s string) []string {
			return puncArr
		},
		Alias:       []string{"punc"},
		Description: "generates all punctuation marks",
	})
	AddFuzzTagToGlobal(&FuzzTagDescription{
		TagName: "regen",
		Handler: func(s string) []string {
			return regen.MustGenerate(s)
		},
		Alias:       []string{"re"},
		Description: "uses regular expressions to generate all possible characters.",
	})
	AddFuzzTagToGlobal(&FuzzTagDescription{
		TagName: "regen:one",
		Handler: func(s string) []string {
			return []string{regen.MustGenerateOne(s)}
		},
		Alias:       []string{"re:one"},
		Description: "using regular Generate a random",
		IsDyn:       true,
	})
	AddFuzzTagToGlobal(&FuzzTagDescription{
		TagName: "rangechar",
		Handler: func(s string) []string {
			var min byte = 0
			var max byte = 0xff

			ret := utils.PrettifyListFromStringSplited(s, ",")
			switch len(ret) {
			case 2:
				p1, p2 := ret[0], ret[1]
				p1Uint, _ := strconv.ParseUint(p1, 16, 8)
				min = uint8(p1Uint)
				p2Uint, _ := strconv.ParseUint(p2, 16, 8)
				max = uint8(p2Uint)
				if max <= 0 {
					max = 0xff
				}
			case 1:
				p2Uint, _ := strconv.ParseUint(ret[0], 16, 8)
				max = uint8(p2Uint)
				if max <= 0 {
					max = 0xff
				}
			}

			var res []string
			if min > max {
				min = 0
			}

			for i := min; true; i++ {
				res = append(res, string(i))
				if i >= max {
					break
				}
			}
			return res
		},
		Alias:       []string{"range:char", "range"},
		Description: "generates a range character set in order, such as `{{rangechar(20,7e)}}` Generate a character set of 0x20 - 0x7e",
	})
	AddFuzzTagToGlobal(&FuzzTagDescription{
		TagName:     "network",
		Handler:     utils.ParseStringToHosts,
		Alias:       []string{"host", "hosts", "cidr", "ip", "net"},
		Description: "generates a network address, such as `{{network(192.168.1.1/24)}}` corresponds to cidr 192.168.1.1/24 All addresses can be separated by commas, for example `{{network(8.8.8.8,192.168.1.1/25,example.com)}}`",
	})
	AddFuzzTagToGlobal(&FuzzTagDescription{
		TagName: "int",
		Handler: func(s string) []string {
			if s == "" {
				return []string{fmt.Sprint(rand.Intn(10))}
			}

			enablePadding := false
			paddingLength := 4
			paddingRight := false
			step := 1

			// uses pipe symbols to split parameters
			parts := strings.Split(s, "|")

			if len(parts) > 1 {
				s = parts[0]
				paddingSuffix := strings.TrimSpace(parts[1])
				enablePadding = true
				paddingRight = strings.HasPrefix(paddingSuffix, "-")
				rawLen := strings.TrimLeft(paddingSuffix, "-")
				paddingLength, _ = strconv.Atoi(rawLen)
			}

			if strings.Contains(s, "-") {
				splited := strings.Split(s, "-")
				left := splited[0]
				if len(left) > 1 && strings.HasPrefix(left, "0") {
					enablePadding = true
					paddingLength = len(left)
				}
			}

			if len(parts) > 2 {
				step, _ = strconv.Atoi(parts[2])
			}

			ints := utils.ParseStringToPorts(s)
			if len(ints) <= 0 {
				return []string{""}
			}

			if step > 1 {
				// generates the result
				var filteredResults []int
				for i := 0; i < len(ints); i += step {
					filteredResults = append(filteredResults, ints[i])
				}
				ints = filteredResults
			}

			var results []string
			for _, i := range ints {
				r := fmt.Sprint(i)
				if enablePadding && paddingLength > len(r) {
					repeatedPaddingCount := paddingLength - len(r)
					if paddingRight {
						r = r + strings.Repeat("0", repeatedPaddingCount)
					} else {
						r = strings.Repeat("0", repeatedPaddingCount) + r
					}
				}
				results = append(results, r)
			}
			return results
		},
		Alias:       []string{"port", "ports", "integer", "i"},
		Description: "generates an integer and range , for example, {{int(1,2,3,4,5)}} Generate an integer among 1, 2, 3, 4, 5, you can also use {{int(1-5)}} to generate an integer of 1-5, you can also use `{{int(1-5|4)}}` generates integers from 1-5, but each integer is 4 digits, such as 0001, 0002, 0003, 0004, 0005, etc. You can use `{{int(1-10|2|3)}}` to generate a list of integers with a stride.",
	})
	AddFuzzTagToGlobal(&FuzzTagDescription{
		TagName: "randint",
		IsDynFun: func(name, params string) bool {
			if len(utils.PrettifyListFromStringSplited(params, ",")) == 3 {
				return false
			}
			return true
		},
		Handler: func(s string) []string {
			var (
				min, max, count uint
				err             error

				enablePadding = false
				paddingRight  bool
				paddingLength int
			)

			splitted := strings.SplitN(s, "|", 2)
			if len(splitted) > 1 {
				s = splitted[0]
				paddingSuffix := strings.TrimSpace(splitted[1])
				enablePadding = true
				paddingRight = strings.HasPrefix(paddingSuffix, "-")
				rawLen := strings.TrimLeft(paddingSuffix, "-")
				paddingLength, _ = strconv.Atoi(rawLen)
			}

			count = 1
			fuzztagfallback := []string{fmt.Sprint(rand.Intn(10))}
			raw := utils.PrettifyListFromStringSplited(s, ",")
			switch len(raw) {
			case 3:
				count, err = parseUint(raw[2])
				if err != nil {
					return fuzztagfallback
				}

				if count <= 0 {
					count = 1
				}
				fallthrough
			case 2:
				min, err = parseUint(raw[0])
				if err != nil {
					return fuzztagfallback
				}
				max, err = parseUint(raw[1])
				if err != nil {
					return fuzztagfallback
				}

				min = uint(utils.Min(int(min), int(max)))
				max = uint(utils.Max(int(min), int(max)))
				break
			case 1:
				min = 0
				max, err = parseUint(raw[0])
				if err != nil {
					return fuzztagfallback
				}
				if max <= 0 {
					max = 10
				}
				break
			default:
				return fuzztagfallback
			}

			var results []string
			RepeatFunc(count, func() bool {
				res := int(max - min)
				if res <= 0 {
					res = 10
				}
				i := min + uint(rand.Intn(res))
				c := fmt.Sprint(i)
				if enablePadding && paddingLength > len(c) {
					repeatedPaddingCount := paddingLength - len(c)
					if paddingRight {
						c = c + strings.Repeat("0", repeatedPaddingCount)
					} else {
						c = strings.Repeat("0", repeatedPaddingCount) + c
					}
				}

				results = append(results, fmt.Sprint(c))
				return true
			})
			return results
		},
		Alias:       []string{"ri", "rand:int", "randi"},
		Description: "randomly generates integers, defined as {{randint(10)}} generates any random number from 0 to 10,{{randint(1,50)}} generates 1-50 Any random number,{{randint(1,50,10)}} generates any random number from 1 to 50, repeat 10 times",
	})
	AddFuzzTagToGlobal(&FuzzTagDescription{
		TagName: "randstr",
		IsDynFun: func(name, params string) bool {
			if len(utils.PrettifyListFromStringSplited(params, ",")) == 3 {
				return false
			}
			return true
		},
		ErrorInfoHandler: func(s string) ([]string, error) {
			var (
				min, max, count uint
				err             error
			)
			fuzztagfallback := []string{utils.RandStringBytes(8)}
			count = 1
			min = 1
			raw := utils.PrettifyListFromStringSplited(s, ",")
			switch len(raw) {
			case 3:
				count, err = parseUint(raw[2])
				if err != nil {
					return fuzztagfallback, err
				}

				if count <= 0 {
					count = 1
				}
				fallthrough
			case 2:
				min, err = parseUint(raw[0])
				if err != nil {
					return fuzztagfallback, err
				}
				max, err = parseUint(raw[1])
				if err != nil {
					return fuzztagfallback, err
				}
				min = uint(utils.Min(int(min), int(max)))
				max = uint(utils.Max(int(min), int(max)))
				if max >= 1e8 {
					max = 1e8
					err = fmt.Errorf("max length is 100000000")
				}
				break
			case 1:
				max, err = parseUint(raw[0])
				if err != nil {
					return fuzztagfallback, err
				}
				min = max
				if max <= 0 {
					max = 8
				}
				break
			default:
				return fuzztagfallback, err
			}

			var r []string
			RepeatFunc(count, func() bool {
				result := int(max - min)
				if result < 0 {
					result = 8
				}

				var offset uint = 0
				if result > 0 {
					offset = uint(rand.Intn(result))
				}
				c := min + offset
				r = append(r, utils.RandStringBytes(int(c)))
				return true
			})
			return r, err
		},
		Alias:       []string{"rand:str", "rs", "rands"},
		Description: "randomly generates a string, defined as {{randstr(10)}} generates a random string of length 10,{{randstr(1,30)}} Generate a random string with a length of 1-30.{{randstr(1,30,10)}} generates 10 random strings, length 1-30",
	})

	AddFuzzTagToGlobal(&FuzzTagDescription{
		TagName: "codec",
		Handler: func(s string) []string {
			s = strings.Trim(s, " ()")

			if codecCaller == nil {
				return []string{s}
			}

			lastDividerIndex := strings.LastIndexByte(s, '|')
			if lastDividerIndex < 0 {
				script, err := codecCaller(s, "")
				if err != nil {
					log.Errorf("codec caller error: %s", err)
					return []string{s}
				}
				// log.Errorf("fuzz.codec no plugin / param specific")
				return []string{script}
			}
			name, params := s[:lastDividerIndex], s[lastDividerIndex+1:]
			script, err := codecCaller(name, params)
			if err != nil {
				log.Errorf("codec caller error: %s", err)
				return []string{s}
			}
			return []string{script}
		},
		Description: "Call the Yakit Codec plug-in",
	})

	AddFuzzTagToGlobal(&FuzzTagDescription{
		TagName: "codec:line",
		Handler: func(s string) []string {
			if codecCaller == nil {
				return fuzztagfallback
			}

			s = strings.Trim(s, " ()")
			lastDividerIndex := strings.LastIndexByte(s, '|')
			if lastDividerIndex < 0 {
				log.Errorf("fuzz.codec no plugin / param specific")
				return fuzztagfallback
			}
			name, params := s[:lastDividerIndex], s[lastDividerIndex+1:]
			script, err := codecCaller(name, params)
			if err != nil {
				log.Errorf("codec caller error: %s", err)
				return fuzztagfallback
			}
			var results []string
			for line := range utils.ParseLines(script) {
				results = append(results, line)
			}
			return results
		},
		Description: "Call Yakit Codec plug-in, parse the result into lines",
	})
	AddFuzzTagToGlobal(&FuzzTagDescription{
		TagName: "unquote",
		Handler: func(s string) []string {
			raw, err := strconv.Unquote(s)
			if err != nil {
				raw, err := strconv.Unquote(`"` + s + `"`)
				if err != nil {
					log.Errorf("unquoted failed: %s", err)
					return []string{s}
				}
				return []string{raw}
			}

			return []string{
				raw,
			}
		},
		Description: "convert the content to strconv.Unquote",
	})

	AddFuzzTagToGlobal(&FuzzTagDescription{
		TagName: "quote",
		Handler: func(s string) []string {
			return []string{
				strconv.Quote(s),
			}
		},
		Description: "strconv.Quote Convert",
	})

	AddFuzzTagToGlobal(&FuzzTagDescription{
		TagName: "lower",
		Handler: func(s string) []string {
			return []string{
				strings.ToLower(s),
			}
		},
		Description: "sets the incoming content to lowercase {{lower(Abc)}} => abc",
	})

	AddFuzzTagToGlobal(&FuzzTagDescription{
		TagName: "upper",
		Handler: func(s string) []string {
			return []string{
				strings.ToUpper(s),
			}
		},
		Description: "Change the incoming content to uppercase {{upper(abc)}} => ABC",
	})

	AddFuzzTagToGlobal(&FuzzTagDescription{
		TagName: "base64enc",
		Handler: func(s string) []string {
			return []string{
				base64.StdEncoding.EncodeToString([]byte(s)),
			}
		},
		Alias:       []string{"base64encode", "base64e", "base64", "b64"},
		Description: "performs base64 encoding,{{base64enc(abc)}} => YWJj",
	})

	AddFuzzTagToGlobal(&FuzzTagDescription{
		TagName: "base64dec",
		Handler: func(s string) []string {
			r, err := codec.DecodeBase64(s)
			if err != nil {
				return []string{s}
			}
			return []string{string(r)}
		},
		Alias:       []string{"base64decode", "base64d", "b64d"},
		Description: "performs base64 decoding,{{base64dec(YWJj)}} => abc",
	})

	AddFuzzTagToGlobal(&FuzzTagDescription{
		TagName: "md5",
		Handler: func(s string) []string {
			return []string{codec.Md5(s)}
		},
		Description: "for md5 encoding,{{md5(abc)}} => 900150983cd24fb0d6963f7d28e17f72",
	})

	AddFuzzTagToGlobal(&FuzzTagDescription{
		TagName: "hexenc",
		Handler: func(s string) []string {
			return []string{codec.EncodeToHex(s)}
		},
		Alias:       []string{"hex", "hexencode"},
		Description: "HEX encoding,{{hexenc(abc)}} => 616263",
	})

	AddFuzzTagToGlobal(&FuzzTagDescription{
		TagName: "hexdec",
		Handler: func(s string) []string {
			raw, err := codec.DecodeHex(s)
			if err != nil {
				return []string{s}
			}
			return []string{string(raw)}
		},
		Alias:       []string{"hexd", "hexdec", "hexdecode"},
		Description: "HEX decoding,{{hexdec(616263)}} => abc",
	})

	AddFuzzTagToGlobal(&FuzzTagDescription{
		TagName:     "sha1",
		Handler:     func(s string) []string { return []string{codec.Sha1(s)} },
		Description: "performs sha1 encoding,{{sha1(abc)}} => a9993e364706816aba3e25717850c26c9cd0d89d",
	})

	AddFuzzTagToGlobal(&FuzzTagDescription{
		TagName:     "sha256",
		Handler:     func(s string) []string { return []string{codec.Sha256(s)} },
		Description: "for sha256 encoding,{{sha256(abc)}} => ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad",
	})

	AddFuzzTagToGlobal(&FuzzTagDescription{
		TagName:     "sha224",
		Handler:     func(s string) []string { return []string{codec.Sha224(s)} },
		Description: "",
	})

	AddFuzzTagToGlobal(&FuzzTagDescription{
		TagName:     "sha512",
		Handler:     func(s string) []string { return []string{codec.Sha512(s)} },
		Description: "for sha512 encoding,{{sha512(abc)}} => ddaf35a193617abacc417349ae20413112e6fa4e89a97ea20a9eeee64b55d39a2192992a274fc1a836ba3c23a3feebbd454d4423643ce80e2a9ac94fa54ca49f",
	})

	AddFuzzTagToGlobal(&FuzzTagDescription{
		TagName: "sha384",
		Handler: func(s string) []string { return []string{codec.Sha384(s)} },
	})

	AddFuzzTagToGlobal(&FuzzTagDescription{
		TagName:     "sm3",
		Handler:     func(s string) []string { return []string{codec.EncodeToHex(codec.SM3(s))} },
		Description: "calculates sm3 hash value,{{sm3(abc)}} => 66c7f0f462eeedd9d1f2d46bdc10e4e24167c4875cf2f7a3f0b8ddb27d8a7eb3",
	})

	AddFuzzTagToGlobal(&FuzzTagDescription{
		TagName: "hextobase64",
		Handler: func(s string) []string {
			raw, err := codec.DecodeHex(s)
			if err != nil {
				return []string{s}
			}
			return []string{base64.StdEncoding.EncodeToString(raw)}
		},
		Alias:       []string{"h2b64", "hex2base64"},
		Description: "converts the HEX string into base64 encoding.{{hextobase64(616263)}} => YWJj",
	})

	AddFuzzTagToGlobal(&FuzzTagDescription{
		TagName: "base64tohex",
		Handler: func(s string) []string {
			raw, err := codec.DecodeBase64(s)
			if err != nil {
				return []string{s}
			}
			return []string{codec.EncodeToHex(string(raw))}
		},
		Alias:       []string{"b642h", "base642hex"},
		Description: "Put the Base64 string Convert to HEX encoding,{{base64tohex(YWJj)}} => 616263",
	})

	AddFuzzTagToGlobal(&FuzzTagDescription{
		TagName: "urlescape",
		Handler: func(s string) []string {
			return []string{codec.QueryEscape(s)}
		},
		Alias:       []string{"urlesc"},
		Description: "url encoding (only encodes special characters),{{urlescape(abc=)}} => abc%3d",
	})

	AddFuzzTagToGlobal(&FuzzTagDescription{
		TagName: "urlenc",
		Handler: func(s string) []string {
			return []string{codec.EncodeUrlCode(s)}
		},
		Alias:       []string{"urlencode", "url"},
		Description: "URL Forced encoding,{{urlenc(abc)}} => %61%62%63",
	})

	AddFuzzTagToGlobal(&FuzzTagDescription{
		TagName: "urldec",
		Handler: func(s string) []string {
			r, err := codec.QueryUnescape(s)
			if err != nil {
				return []string{s}
			}
			return []string{string(r)}
		},
		Alias:       []string{"urldecode", "urld"},
		Description: "URL force decoding,{{urldec(%61%62%63)}} => abc",
	})

	AddFuzzTagToGlobal(&FuzzTagDescription{
		TagName: "doubleurlenc",
		Handler: func(s string) []string {
			return []string{codec.DoubleEncodeUrl(s)}
		},
		Alias:       []string{"doubleurlencode", "durlenc", "durl"},
		Description: "Double URL encoding,{{doubleurlenc(abc)}} => %2561%2562%2563",
	})

	AddFuzzTagToGlobal(&FuzzTagDescription{
		TagName: "doubleurldec",
		Handler: func(s string) []string {
			r, err := codec.DoubleDecodeUrl(s)
			if err != nil {
				return []string{s}
			}
			return []string{r}
		},
		Alias:       []string{"doubleurldecode", "durldec", "durldecode"},
		Description: "Double URL decoding,{{doubleurldec(%2561%2562%2563)}} => abc",
	})

	AddFuzzTagToGlobal(&FuzzTagDescription{
		TagName: "htmlenc",
		Handler: func(s string) []string {
			return []string{codec.EncodeHtmlEntity(s)}
		},
		Alias:       []string{"htmlencode", "html", "htmle", "htmlescape"},
		Description: "HTML entity encoding,{{htmlenc(abc)}} => &#97;&#98;&#99;",
	})

	AddFuzzTagToGlobal(&FuzzTagDescription{
		TagName: "htmlhexenc",
		Handler: func(s string) []string {
			return []string{codec.EncodeHtmlEntityHex(s)}
		},
		Alias:       []string{"htmlhex", "htmlhexencode", "htmlhexescape"},
		Description: "HTML hexadecimal entity encoding,{{htmlhexenc(abc)}} => &#x61;&#x62;&#x63;",
	})

	AddFuzzTagToGlobal(&FuzzTagDescription{
		TagName: "htmldec",
		Handler: func(s string) []string {
			return []string{codec.UnescapeHtmlString(s)}
		},
		Alias:       []string{"htmldecode", "htmlunescape"},
		Description: "HTML decode,{{htmldec(&#97;&#98;&#99;)}} => abc",
	})

	AddFuzzTagToGlobal(&FuzzTagDescription{
		TagName: "repeatstr",
		Handler: func(s string) []string {
			if !strings.Contains(s, "|") {
				return []string{s + s}
			}
			index := strings.LastIndex(s, "|")
			if index <= 0 {
				return []string{s}
			}
			n, _ := strconv.Atoi(s[index+1:])
			s = s[:index]
			return []string{strings.Repeat(s, n)}
		},
		Alias:       []string{"repeat:str"},
		Description: "Repeat string, `{{repeatstr(abc|3)}}` => abcabcabc",
	})

	AddFuzzTagToGlobal(&FuzzTagDescription{
		TagName: "randomupper",
		Handler: func(s string) []string {
			return []string{codec.RandomUpperAndLower(s)}
		},
		Alias:       []string{"random:upper", "random:lower"},
		Description: "random case,{{randomupper(abc)}} => aBc",
	})

	AddFuzzTagToGlobal(&FuzzTagDescription{
		TagName: "yso:exec",
		HandlerEx: func(s string) []*fuzztag.FuzzExecResult {
			var result []*fuzztag.FuzzExecResult
			pushNewResult := func(d []byte, verbose []string) {
				result = append(result, fuzztag.NewFuzzExecResult(d, verbose))
			}
			for _, gadget := range yso.GetAllRuntimeExecGadget() {
				javaObj, err := gadget(s)
				if javaObj == nil || err != nil {
					continue
				}
				objBytes, err := yso.ToBytes(javaObj)
				if err != nil {
					continue
				}
				pushNewResult(objBytes, []string{javaObj.Verbose().GetNameVerbose(), "runtime exec evil class", s})
			}
			for _, gadget := range yso.GetAllTemplatesGadget() {
				javaObj, err := gadget(yso.SetProcessImplExecEvilClass(s))
				if javaObj == nil || err != nil {
					continue
				}
				objBytes, err := yso.ToBytes(javaObj)
				if err != nil {
					continue
				}
				pushNewResult(objBytes, []string{javaObj.Verbose().GetNameVerbose(), "processImpl exec evil class", s})
			}
			if len(result) > 0 {
				return result
			}

			return []*fuzztag.FuzzExecResult{fuzztag.NewFuzzExecResult([]byte(s), []string{s})}
		},
	})
	AddFuzzTagToGlobal(&FuzzTagDescription{
		TagName: "yso:dnslog",
		HandlerEx: func(s string) []*fuzztag.FuzzExecResult {
			var getDomain func() string
			sSplit := strings.Split(s, "|")
			if len(sSplit) == 2 {
				n := 0
				getDomain = func() string {
					n++
					return fmt.Sprintf("%s%d.%s", sSplit[1], n, sSplit[0])
				}
			} else {
				getDomain = func() string {
					return s
				}
			}
			var result []*fuzztag.FuzzExecResult
			pushNewResult := func(d []byte, verbose []string) {
				result = append(result, fuzztag.NewFuzzExecResult(d, verbose))
			}
			for _, gadget := range yso.GetAllTemplatesGadget() {
				domain := getDomain()
				javaObj, err := gadget(yso.SetDnslogEvilClass(domain))
				if javaObj == nil || err != nil {
					continue
				}
				objBytes, err := yso.ToBytes(javaObj)
				if err != nil {
					continue
				}
				pushNewResult(objBytes, []string{javaObj.Verbose().GetNameVerbose(), "dnslog evil class", domain})
			}
			if len(result) > 0 {
				return result
			}

			return []*fuzztag.FuzzExecResult{fuzztag.NewFuzzExecResult([]byte(s), []string{s})}
		},
	})
	AddFuzzTagToGlobal(&FuzzTagDescription{
		TagName: "yso:urldns",
		HandlerEx: func(s string) []*fuzztag.FuzzExecResult {
			var result []*fuzztag.FuzzExecResult
			javaObj, err := yso.GetURLDNSJavaObject(s)
			if err == nil {
				objBytes, err := yso.ToBytes(javaObj)
				if err == nil {
					return append(result, fuzztag.NewFuzzExecResult(objBytes, []string{javaObj.Verbose().GetNameVerbose()}))
				}
			}
			return []*fuzztag.FuzzExecResult{fuzztag.NewFuzzExecResult([]byte(s), []string{s})}
		},
	})
	// The tag is too long, so a simple fuzztag is added to the front end.
	AddFuzzTagToGlobal(&FuzzTagDescription{
		TagName: "yso:find_gadget_by_dns",
		HandlerEx: func(s string) []*fuzztag.FuzzExecResult {
			var result []*fuzztag.FuzzExecResult
			javaObj, err := yso.GetFindGadgetByDNSJavaObject(s)
			if err == nil {
				objBytes, err := yso.ToBytes(javaObj)
				if err == nil {
					return append(result, fuzztag.NewFuzzExecResult(objBytes, []string{javaObj.Verbose().GetNameVerbose()}))
				}
			}
			return []*fuzztag.FuzzExecResult{fuzztag.NewFuzzExecResult([]byte(s), []string{s})}
		},
	})
	AddFuzzTagToGlobal(&FuzzTagDescription{
		TagName: "yso:find_gadget_by_bomb",
		HandlerEx: func(s string) []*fuzztag.FuzzExecResult {
			var result []*fuzztag.FuzzExecResult
			pushNewResult := func(d []byte, verbose []string) {
				result = append(result, fuzztag.NewFuzzExecResult(d, verbose))
			}

			if s == "all" {
				for gadget, className := range yso.GetGadgetChecklist() {
					javaObj, err := yso.GetFindClassByBombJavaObject(className)
					if javaObj == nil || err != nil {
						continue
					}
					objBytes, err := yso.ToBytes(javaObj)
					if err != nil {
						continue
					}
					pushNewResult(objBytes, []string{javaObj.Verbose().GetNameVerbose(), "Gadget", gadget})
				}
			} else {
				javaObj, err := yso.GetFindClassByBombJavaObject(s)
				if javaObj == nil || err != nil {
					return result
				}
				objBytes, err := yso.ToBytes(javaObj)
				if err != nil {
					return result
				}
				pushNewResult(objBytes, []string{javaObj.Verbose().GetNameVerbose(), "Class name", s})
			}

			if len(result) > 0 {
				return result
			}

			return []*fuzztag.FuzzExecResult{fuzztag.NewFuzzExecResult([]byte(s), []string{s})}
		},
	})
	// . These tags are unstable.
	AddFuzzTagToGlobal(&FuzzTagDescription{
		TagName:     "yso:headerecho",
		Description: "Try your best to use header echo to generate multiple chains",
		HandlerEx: func(s string) []*fuzztag.FuzzExecResult {
			headers := strings.Split(s, "|")
			if len(headers) != 2 {
				return []*fuzztag.FuzzExecResult{fuzztag.NewFuzzExecResult([]byte(s), []string{s})}
			}
			var result []*fuzztag.FuzzExecResult
			pushNewResult := func(d []byte, verbose []string) {
				result = append(result, fuzztag.NewFuzzExecResult(d, verbose))
			}
			for _, gadget := range yso.GetAllTemplatesGadget() {
				javaObj, err := gadget(yso.SetMultiEchoEvilClass(), yso.SetHeader(headers[0], headers[1]))
				if javaObj == nil || err != nil {
					continue
				}
				objBytes, err := yso.ToBytes(javaObj)
				if err != nil {
					continue
				}
				pushNewResult(objBytes, []string{javaObj.Verbose().GetNameVerbose(), "tomcat header echo evil class", s})
			}
			if len(result) > 0 {
				return result
			}
			return []*fuzztag.FuzzExecResult{fuzztag.NewFuzzExecResult([]byte(s), []string{s})}
		},
	})
	AddFuzzTagToGlobal(&FuzzTagDescription{
		TagName:     "yso:bodyexec",
		Description: "to try your best to use class body exec to generate multiple chains.",
		HandlerEx: func(s string) []*fuzztag.FuzzExecResult {
			var result []*fuzztag.FuzzExecResult
			pushNewResult := func(d []byte, verbose []string) {
				result = append(result, fuzztag.NewFuzzExecResult(d, verbose))
			}
			for _, gadget := range yso.GetAllTemplatesGadget() {
				javaObj, err := gadget(yso.SetMultiEchoEvilClass(), yso.SetExecAction(), yso.SetParam(s), yso.SetEchoBody())
				if javaObj == nil || err != nil {
					continue
				}
				objBytes, err := yso.ToBytes(javaObj)
				if err != nil {
					continue
				}
				pushNewResult(objBytes, []string{javaObj.Verbose().GetNameVerbose(), "tomcat body exec echo evil class", s})
			}
			if len(result) > 0 {
				return result
			}

			return []*fuzztag.FuzzExecResult{fuzztag.NewFuzzExecResult([]byte(s), []string{s})}
		},
	})
	AddFuzzTagToGlobal(&FuzzTagDescription{
		TagName: "headerauth",
		Handler: func(s string) []string {
			return []string{"Accept-Language: zh-CN,zh;q=1.9"}
		},
	})
}

func FuzzFileOptions() []FuzzConfigOpt {
	var opt []FuzzConfigOpt
	for _, t := range Filetag() {
		opt = append(opt, Fuzz_WithExtraFuzzTagHandler(t.TagName, t.Handler))
		for _, a := range t.Alias {
			opt = append(opt, Fuzz_WithExtraFuzzTagHandler(a, t.Handler))
		}
	}
	return opt
}

func Fuzz_WithEnableFiletag() FuzzConfigOpt {
	return func(config *FuzzTagConfig) {
		for _, opt := range FuzzFileOptions() {
			opt(config)
		}
	}
}

func Filetag() []*FuzzTagDescription {
	return []*FuzzTagDescription{
		{
			TagName: "file:line",
			Handler: func(s string) []string {
				s = strings.Trim(s, " ()")
				var result []string
				for _, lineFile := range utils.PrettifyListFromStringSplited(s, "|") {
					lineChan, err := utils.FileLineReader(lineFile)
					if err != nil {
						log.Errorf("fuzztag read file failed: %s", err)
						continue
					}
					for line := range lineChan {
						result = append(result, string(line))
					}
				}
				if len(result) <= 0 {
					return fuzztagfallback
				}
				return result
			},
			Alias:       []string{"fileline", "file:lines"},
			Description: "parses the file name (you can use `|` Split), return the contents of the file into an array line by line, defined as `{{file:line(/tmp/test.txt)}}` or `{{file:line(/tmp/test.txt|/tmp/1.txt)}}`",
		},
		{
			TagName: "file:dir",
			Handler: func(s string) []string {
				s = strings.Trim(s, " ()")
				var result []string
				for _, lineFile := range utils.PrettifyListFromStringSplited(s, "|") {
					fileRaw, err := ioutil.ReadDir(lineFile)
					if err != nil {
						log.Errorf("fuzz.filedir read dir failed: %s", err)
						continue
					}
					for _, info := range fileRaw {
						if info.IsDir() {
							continue
						}
						fileContent, err := ioutil.ReadFile(info.Name())
						if err != nil {
							continue
						}
						result = append(result, string(fileContent))
					}
				}
				if len(result) <= 0 {
					return fuzztagfallback
				}
				return result
			},
			Alias:       []string{"filedir"},
			Description: "Parse the folder, read the contents of the files in the folder, read it into an array and return it, defined as `{{file:dir(/tmp/test)}}` or `{{file:dir(/tmp/test|/tmp/1)}}`",
		},
		{
			TagName: "file",
			Handler: func(s string) []string {
				s = strings.Trim(s, " ()")
				var result []string
				for _, lineFile := range utils.PrettifyListFromStringSplited(s, "|") {
					fileRaw, err := ioutil.ReadFile(lineFile)
					if err != nil {
						log.Errorf("fuzz.files read file failed: %s", err)
						continue
					}
					result = append(result, string(fileRaw))
				}
				if len(result) <= 0 {
					return fuzztagfallback
				}
				return result
			},
			Description: "reads the file content, can support multiple files, separated by vertical lines, `{{file(/tmp/1.txt)}}` or `{{file(/tmp/1.txt|/tmp/test.txt)}}`",
		},
	}
}
