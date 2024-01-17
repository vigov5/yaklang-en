package yaktest

import (
	"github.com/davecgh/go-spew/spew"
	"github.com/yaklang/yaklang/common/yak"
	"os"
	"testing"
)

/*
conn, err = tcp.Connect("127.0.0.1", 22, tcp.clientTimeout(3))
if err != nil {
	die(err)
}

dump(conn.RecvString())
*/

func TestMisc_Syntax_FixLine(t *testing.T) {
	cases := []YakTestCase{
		{
			Name: "Test null pointer reference 1",
			Src: `
defer fn{
	err = recover()
	if sprint(err) != "line 14: invalid value[<nil>](a) cannot fetch member[abc]" {
		die(err)
	}
} 



a=nil;
;
;
a.abc
;
;
a.abc
asdfa
asdf
a
sdfa
sd
`,
		},
	}

	Run("yaklang syntax fix", t, cases...)
}

func TestMisc_Syntax_FixLine_TCP(t *testing.T) {
	cases := []YakTestCase{
		{
			Name: "Test null pointer reference 1",
			Src: `
conn, err = tcp.Connect("127.0.0.1", 9999, tcp.clientTimeout(3))
if err != nil {
	die(err)
}

dump(conn)
println("123123")
dump(conn.Send("asdfasd"))
`,
		},
	}

	Run("yaklang syntax fix", t, cases...)
}

func TestMisc_Syntax_FixUndefinedCall(t *testing.T) {
	cases := []YakTestCase{
		{
			Name: "Test undefined reference",
			Src: `defer fn{
	err = recover()
	if !str.MatchAllOfSubString(err, "sd.Abc", "sd.abc") {
		die(err)
	}
} 
sd.abc
`,
		},
	}

	Run("yaklang syntax fix", t, cases...)
}

func TestMisc_Syntax_FixValueCall(t *testing.T) {
	cases := []YakTestCase{
		{
			Name: "Test undefined reference",
			Src: `
loglevel("info")
defer fn{
	err = recover()
	log.info("checking syntax error result")
	dump(err)
	if !str.MatchAllOfSubString(err, "sd.Abc", "sd.abc") {
		die(err)
	}
} 
a = 1;
a.abc
`,
		},
	}

	Run("yaklang syntax fix", t, cases...)
}

func TestMisc_Syntax_FixValueCall1(t *testing.T) {
	cases := []YakTestCase{
		{
			Name: "Test undefined reference",
			Src: `
loglevel("info")
defer fn{
	err = recover()
	log.info("checking syntax error result")
	dump(err)
	if !str.MatchAllOfSubString(err, "a[0].abc", "a[0].Abc") {
		die(err)
	}
} 
a = [ab,1];
a[0].abc
`,
		},
	}

	Run("yaklang syntax fix", t, cases...)
}

func TestMisc_Syntax_FixValueCall2(t *testing.T) {
	cases := []YakTestCase{
		{
			Name: "Test undefined reference",
			Src: `
loglevel("info")
defer fn{
	err = recover()
	log.info("checking syntax error result")
	dump(err)
	if !str.MatchAllOfSubString(err, "a[1].abc", "a[1].Abc") {
		die(err)
	}
} 
a = [ab,1];
a[1].abc
`,
		},
	}

	Run("yaklang syntax fix", t, cases...)
}

func TestMisc_Syntax_FixValueCall3(t *testing.T) {
	cases := []YakTestCase{
		{
			Name: "Test undefined reference",
			Src: `
loglevel("info")
defer fn{
	err = recover()
	log.info("checking syntax error result")
	dump(err)
	if !str.MatchAllOfSubString(err, "a[31]") {
		die(err)
	}
} 
a = [ab,1];
a[31].abc
`,
		},
	}

	Run("yaklang syntax fix", t, cases...)
}

func TestMisc_Syntax_FixValueCall4(t *testing.T) {
	cases := []YakTestCase{
		{
			Name: "Test undefined reference",
			Src: `
loglevel("info")
defer fn{
	err = recover()
	log.info("checking syntax error result")
	dump(err)
	if !str.MatchAllOfSubString(err, "a[31]") {
		die(err)
	}
} 
a = [ab,1];
b = 22222
a[b].abc
`,
		},
	}

	Run("yaklang syntax fix", t, cases...)
}

func TestMisc_Syntax_FixValueCall4_MAP(t *testing.T) {
	cases := []YakTestCase{
		{
			Name: "Test undefined reference: map",
			Src: `
loglevel("info")
defer fn{
	err = recover()
	log.info("checking syntax error result")
	dump(err)
	if !str.MatchAllOfSubString(err, "a[b]") {
		die(err)
	}
} 
a = {"abc": asdfasd, "2222": 1111, "ab": 1};
b = 22222
a[b].abc
`,
		},
	}

	Run("yaklang syntax fix", t, cases...)
}

func TestMisc_Syntax_FixValueCall4_MAP11(t *testing.T) {
	cases := []YakTestCase{
		{
			Name: "Test undefined reference: map",
			Src: `
loglevel("info")
defer fn{
	err = recover()
	log.info("checking syntax error result")
	dump(err)
	if !str.MatchAllOfSubString(err, "a[b]") {
		die(err)
	}
} 
a = {"abc": asdfasd, "2222": 1111, "ab": 1};
b = 22222
a[111].abc
`,
		},
	}

	Run("yaklang syntax fix", t, cases...)
}

func TestMisc_Syntax_FixValueCall4_MAP1(t *testing.T) {
	cases := []YakTestCase{
		{
			Name: "Test undefined reference: map",
			Src: `
loglevel("info")
defer fn{
	err = recover()
	log.info("checking syntax error result")
	dump(err)
	if !str.MatchAllOfSubString(err, "(int) 12") {
		die(err)
	}
} 
a = {"abc": asdfasd, "2222": 1111, "ab": 1, 12: 12};
`,
		},
	}

	Run("yaklang syntax fix", t, cases...)
}

func TestMisc_Syntax_FixValueCall4_UndefinedCALL(t *testing.T) {
	os.Setenv("YAKLANGDEBUG", "1")
	cases := []YakTestCase{
		{
			Name: "Test undefined reference: map",
			Src: `
loglevel("info")
defer fn{
	err = recover()
	log.info("checking syntax error result")
	dump(err)
	if !str.MatchAllOfSubString(err, "ccc") {
		die(err)
	}
}

ccc()
`,
		},
	}

	Run("yaklang syntax fix", t, cases...)
}

func TestMisc_Syntax_FixValueCall4_UndefinedCALL1(t *testing.T) {
	os.Setenv("YAKLANGDEBUG", "1")
	cases := []YakTestCase{
		{
			Name: "Test undefined reference: map",
			Src: `
loglevel("info")
defer fn{
	err = recover()
	log.info("checking syntax error result")
	dump(err)
	if !str.MatchAllOfSubString(err, "ccc(a,b,c,d,e)") {
		die(err)
	}
}
ccc = func(abcccc) {
}
ccc(a,b,c,d,e)
`,
		},
	}

	Run("yaklang syntax fix", t, cases...)
}

func TestMisc_Syntax_StringCast(t *testing.T) {
	os.Setenv("YAKLANGDEBUG", "1")
	cases := []YakTestCase{
		{
			Name: "Test undefined reference: map",
			Src: `
loglevel("info")
defer fn{
	err = recover()
	log.info("checking syntax error result")
	if !str.MatchAllOfSubString(err, "ccc") {
		die(err)
	}
}
println(string(asdfasdsd))
`,
		},
	}

	Run("yaklang syntax fix", t, cases...)
}

func TestMisc_Syntax_DumpZero(t *testing.T) {
	cases := []YakTestCase{
		{
			Name: "Test undefined reference: map",
			Src: `
defer fn{
	err := recover()
	println(err)
	if !str.MatchAllOfSubString(string(err), "str.ParseStringToHosts(ccccc)") {
		die(err)
	}
}

str.MatchAllOfSubString(abc, "aa")
str["MatchAllOfSubString"](ccc, "bb")
str.ParseStringToHosts(ccccc)
`,
		},
	}

	Run("yaklang syntax fix", t, cases...)
}

func TestMisc_Syntax_DumpZero_UndefinedCall(t *testing.T) {
	cases := []YakTestCase{
		{
			Name: "Test undefined reference: map",
			Src: `
defer fn{
	err := recover()
	println(err)
	if !str.MatchAllOfSubString(string(err), "str.MatchAllOfSubS") {
		die(err)
	}
}

str.MatchAllOfSubS(abc, "aa")
`,
		},
	}

	Run("yaklang syntax fix", t, cases...)
}

func TestMisc_Syntax_DStructCall(t *testing.T) {
	cases := []YakTestCase{
		{
			Name: "Test undefined reference: map",
			Src: `
defer fn{
	err := recover()
	println(err)
	if !str.MatchAllOfSubString(string(err), ")[0].CAAAA", "fuzz.HTTPRequest(") {
		die(err)
	}
}

fuzz.HTTPRequest("GET / HTTP/1.1\r\nHost: www.baidu.com\r\n\r\n")[0].CAAAA()
`,
		},
	}

	Run("yaklang syntax fix", t, cases...)
}

func TestMisc_Syntax_NilCall(t *testing.T) {

	//os.Setenv("YAKLANGDEBUG", "123")
	cases := []YakTestCase{
		{
			Name: "Test null pointer reference 1",
			Src: `a=nil;


// a.abc();


abc = 1
abc = func(a,b,c...) {
	println("1231")
	println(a)
	println(b)
	dump(c)
}
abc("123", "aaa", "123123123", "asdfasdf", "aaaccc")

servicescan.proto(["tcp"]...)
`,
		},
	}

	Run("yaklang syntax fix", t, cases...)
}

func TestMisc_Syntax_NilCall111(t *testing.T) {

	//os.Setenv("YAKLANGDEBUG", "123")
	cases := []YakTestCase{
		{
			Name: "Test null pointer reference 1",
			Src: `a=nil;


// a.abc();


abc = 1
abc = func(a,b) {
	println("1231")
	println(a)
	println(b)
}
abc1 = abc("123", "")
`,
		},
	}

	Run("yaklang syntax fix", t, cases...)
}

func TestMisc_FuncAsParamCall(t *testing.T) {
	//os.Setenv("YAKLANGDEBUG", "123")
	cases := []YakTestCase{
		{
			Name: "Function parameter call fails",
			Src: `a=nil;

defer fn{
err = recover()
if err != nil && !str.MatchAllOfSubString(sprint(err), "hook.NewMixPluginCaller()[0].LoadPlugin") {
	die(1)
}
}

hook.NewMixPluginCaller()[0].LoadPlugin(i)
`,
		},
	}

	Run("yaklang syntax fix", t, cases...)
}

func TestMisc_FuncAsParamCall_RetError(t *testing.T) {
	//os.Setenv("YAKLANGDEBUG", "123")
	cases := []YakTestCase{
		{
			Name: "Function return value fails",
			Src: `a=nil;

defer fn{
err = recover()
if err != nil && !str.MatchAllOfSubString(sprint(err), "a,bc,c = hook.NewMixPluginCaller()") {
	die(err)
}
}

a,bc,c = hook.NewMixPluginCaller()
`,
		},
	}

	Run("yaklang syntax fix", t, cases...)
}

/*
av = []byte("abc*****")
a = str.MatchAllOfGlob(av, "abc*")
dump(a)
*/
func TestMisc_FuncAsParamCall_RetError1111(t *testing.T) {
	//os.Setenv("YAKLANGDEBUG", "123")
	cases := []YakTestCase{
		{
			Name: "Function return value fails",
			Src: `a=nil;

av = []byte("abc*****")
a = str.MatchAllOfGlob(av, "abc*")
dump(a)
`,
		},
	}

	Run("yaklang syntax fix", t, cases...)
}

func TestMisc_FuncAsParamCall_RetError1(t *testing.T) {
	//os.Setenv("YAKLANGDEBUG", "123")
	cases := []YakTestCase{
		{
			Name: "Function return value fails",
			Src: `a=nil;

defer fn{
err = recover()
if err != nil && !str.MatchAllOfSubString(sprint(err), "a,bc,c = hook.NewMixPluginCaller()") {
	die(err)
}
}

abc = func() {
	return hook.NewMixPluginCaller()[0]
}

a,bc,c = abc()
`,
		},
	}

	Run("yaklang syntax fix", t, cases...)
}

func TestMisc_FuncAsParamCall_PhpMyAdmin(t *testing.T) {
	//os.Setenv("YAKLANGDEBUG", "123")
	cases := []YakTestCase{
		{
			Name: "Function return value fails",
			Src: `a=nil;

res, _ = servicescan.Scan("188.165.236.5", "443")
for result = range res {
println(result.String())
}

`,
		},
	}

	Run("yaklang syntax fix", t, cases...)
}

func TestMisc_TypeConvert(t *testing.T) {
	//os.Setenv("YAKLANGDEBUG", "123")
	cases := []YakTestCase{
		{
			Name: "Function return value fails",
			Src: `a=nil;

a = ["abc"]
dump(a)

aa = append(a, "abc")
dump(aa)

result = str.ToStringSlice(aa)
dump(result)

`,
		},
	}

	Run("yaklang syntax fix", t, cases...)
}

func TestMisc_NeNil(t *testing.T) {
	os.Setenv("DEBUG", "123")
	cases := []YakTestCase{
		{
			Name: "Function domain problem",
			Src: `
a, err := http.Get("asdfasdfssss")
dump(a, err)

if a != nil {
	panic("Determination of inequality fails")
}
`,
		},
	}

	Run("yaklang syntax fix function scope", t, cases...)
}

func TestMisc_TestFunctionDefination(t *testing.T) {
	os.Setenv("DEBUG", "123")
	cases := []YakTestCase{
		{
			Name: "Function domain problem",
			Src: `a=nil;

def b() {
	println("Hello!")
}

def aaa() {
	b()
}

aaa()
`,
		},
	}

	Run("yaklang syntax fix function scope", t, cases...)
}

func TestMisc_HTTPRequestToHistory(t *testing.T) {
	os.Setenv("DEBUG", "123")
	cases := []YakTestCase{
		{
			Name: "Function domain problem",
			Src: `req, err := http.NewRequest("GET", "https://baidu.com")
die(err)

freq, err := fuzz.HTTPRequest(req)
die(err)

freq.Show()
`,
		},
	}

	Run("yaklang http to fuzz.HTTPRequest", t, cases...)
}

// Test the situation when the function is written incorrectly
func TestMisc_ParseError(t *testing.T) {
	var src = `
1
req, err := http.NewRequest("GET", "https://baidu.com", a
)
die(err)

freq, err := fuzz.HTTPRequest(req)
die(err)

freq.Show()`
	result := yak.AnalyzeStaticYaklang(src)
	if len(result) <= 0 {
		panic("parse params failed")
	}
	if result[0].StartLineNumber <= 0 {
		panic("error invalid")
	}

	spew.Dump(result[0])
}

// Test common errors
func TestMisc_ParseError1(t *testing.T) {
	var src = `
1,
req, err := http.NewRequest("GET", "https://baidu.com", a
)
die(err)

freq, err := fuzz.HTTPRequest(req)
die(err)

freq.Show()`
	result := yak.AnalyzeStaticYaklang(src)
	if len(result) <= 0 {
		panic("parse params failed")
	}
	if result[0].StartLineNumber <= 0 {
		panic("error invalid")
	}

	spew.Dump(result[0])
}

// Test statement problem
func TestMisc_ParseError2(t *testing.T) {
	var src = `
a = {"123": 1, 
"1111111":}
`
	result := yak.AnalyzeStaticYaklang(src)
	if len(result) <= 0 {
		panic("parse params failed")
	}
	if result[0].StartLineNumber <= 0 {
		panic("error invalid")
	}

	spew.Dump(result[0])
}

// Test statement problem
func TestMisc_ParseError3(t *testing.T) {
	var src = `
a = {"123": 1, 
"1111111": 1

}
`
	result := yak.AnalyzeStaticYaklang(src)
	if len(result) <= 0 {
		panic("parse params failed")
	}
	if result[0].StartLineNumber <= 0 {
		panic("error invalid")
	}

	spew.Dump(result[0])
}

// Test declaration problem: Chinese Comma
func TestMisc_ParseError3_ChineseComma(t *testing.T) {
	var src = `
a = {"123": 1, 
"1111111": 1，

}
`
	result := yak.AnalyzeStaticYaklang(src)
	if len(result) <= 0 {
		panic("Unable to detect mistake!")
	}
	if result[0].StartLineNumber <= 0 {
		panic("error invalid")
	}

	spew.Dump(result[0])
}

// Test declaration problem: Chinese Comma
func TestMisc_ParseError3_Chinese(t *testing.T) {
	var src = `
a = {"123": 1, 



"1111111": 1 Chinese

}
`
	result := yak.AnalyzeStaticYaklang(src)
	if len(result) <= 0 {
		panic("Unable to detect mistake!")
	}
	if result[0].StartLineNumber <= 0 {
		panic("error invalid")
	}

	spew.Dump(result[0])
}

// Test declaration variable mixed Chinese syntax error: Chinese Comma
func TestMisc_ParseError3_ChineseWithUndefined(t *testing.T) {
	var src = `

println(abc)

a = {"123": 1, 



"1111111": 1 Chinese

}
`
	result := yak.AnalyzeStaticYaklang(src)
	if len(result) <= 0 {
		panic("Unable to detect mistake!")
	}
	if result[0].StartLineNumber <= 0 {
		panic("error invalid")
	}

	spew.Dump(result[0])
}

func TestMisc_TailF(t *testing.T) {
	os.Setenv("DEBUG", "123")
	cases := []YakTestCase{
		{
			Name: "Function domain problem",
			Src: `a=nil;

go func{file.TailF("/tmp/test.tailf", func(i){
	println("LINE")
	println(i)
})}

go func{
	f, err = file.Open("/tmp/test.tailf")
	if err != nil { panic(err) }
    
    f.Write([]byte("123123\n"+randstr(20)));sleep(0.5)
    f.Write([]byte("123123\n"+randstr(20)));sleep(0.5)
    f.Write([]byte("123123\n"+randstr(20)));sleep(0.5)
    f.Write([]byte("123123\n"+randstr(20)));sleep(0.5)
}

sleep(3)
`,
		},
	}

	Run("yaklang file.TailF", t, cases...)
}

func TestMisc_ZeroValueCall(t *testing.T) {
	os.Setenv("DEBUG", "123")
	cases := []YakTestCase{
		{
			Name: "Function domain problem",
			Src: `f, _ = poc.ParseBytesToHTTPRequest([]byte("asfsadfasd"))
f.Close()
`,
		},
	}

	Run("yaklang Zero Value Calling", t, cases...)
}

func TestSyntaxError(t *testing.T) {
	/*

			for _, c := range cve.Query(cve.keywrod("tomcat"), cve.cpe("tomcat:9.4"), cve.startyear(2010), cve.endyear(2022)) {
				if c.Year > 2014 && c.Year < 2022 {
					yakit.Output(...)
				}
		 	}


			for _, c := range cve.Query(cve.keywrod("tomcat"), cve.cpe("tomcat:9.4"), cve.before()), cve.endyear(2022)) {
					if c.Year > 2014 && c.Year < 2022 {
						yakit.Output(...)
					}
			 	}

	*/

	os.Setenv("YAKMODE", "vm")
	StaticAnalyze("Syntax error analysis", t, YakTestCase{
		Name: "Test basic syntax error",
		Src:  "if {abc+=}",
	})
}
