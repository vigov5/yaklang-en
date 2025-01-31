package antlr4nasl

import (
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/yaklang/yaklang/common/consts"
	"github.com/yaklang/yaklang/common/log"
	"github.com/yaklang/yaklang/common/utils"
	"github.com/yaklang/yaklang/common/yak"
	_ "github.com/yaklang/yaklang/common/yak"
	utils2 "github.com/yaklang/yaklang/common/yak/antlr4nasl/lib"
	"github.com/yaklang/yaklang/common/yak/antlr4yak/yakvm"
	"github.com/yaklang/yaklang/common/yakgrpc/yakit"
	"os"
	"strings"
	"testing"
	"time"
)

func init() {
	yak.SetNaslExports(Exports)
}
func PatchEngine(engine *Engine) {
	engine.AddNaslLibPatch("http_func", func(s string) string {
		s += `

function http_get_port( default_list, host, ignore_broken, ignore_unscanned, ignore_cgi_disabled, dont_use_vhosts ) {
 local_var final_port_list;

  final_port_list = http_get_ports(default_list:default_list,host:host,ignore_broken:ignore_broken,ignore_unscanned:ignore_unscanned,ignore_cgi_disabled:ignore_cgi_disabled,dont_use_vhosts:dont_use_vhosts);
  foreach port( final_port_list ) {
	return port;
  }
  return -1;
}
`
		return s
	})
	engine.AddNaslLibPatch("smtp_func", func(s string) string {
		s += `
function smtp_get_port( default_list, ignore_broken, ignore_unscanned ) {

  local_var final_port_list;

  final_port_list = smtp_get_ports(default_list:default_list,ignore_broken:ignore_broken,ignore_unscanned:ignore_unscanned);
	foreach port( final_port_list ) {
		return port;
	}
	return -1;
}
`
		return s
	})
}

//func BuildInMethodCheck(engine *ScriptEngine) {
//	includeLibCodes := []string{}
//	missMethod := map[string]struct{}{}
//	naslLibPath := engine.naslLibsPath
//	files, err := utils.GetAllFiles(naslLibPath)
//	if err != nil {
//		panic(err)
//	}
//	for _, file := range files {
//		fileName := filepath.Base(file)
//		if !strings.HasSuffix(fileName, ".inc") {
//			continue
//		}
//		includeLibCodes = append(includeLibCodes, fmt.Sprintf(`include("%s");`, fileName))
//	}
//	err = engine.SafeEval(strings.Join(includeLibCodes, "\n"))
//	if err != nil {
//		panic(err)
//	}
//	for script, _ := range engine.scripts {
//		engine.compiler.RegisterVisitHook("a", func(c *visitors.Compiler, ctx antlr.ParserRuleContext) {
//			if v, ok := ctx.(*nasl.IdentifierExpressionContext); ok {
//				id := v.GetText()
//				hasMethod := false
//				if _, ok := NaslLib[id]; ok {
//					hasMethod = true
//				}
//				if _, ok := lib.NaslBuildInNativeMethod[id]; ok {
//					hasMethod = true
//				}
//				if _, ok := engine.GetVirtualMachine().GetVar(id); ok {
//					hasMethod = true
//				}
//				if !hasMethod {
//					missMethod[id] = struct{}{}
//				}
//			}
//		})
//		err := engine.Compile(script)
//		if err != nil {
//			panic(err)
//		}
//		spew.Dump(missMethod)
//	}
//}

//	func TestScriptLib(t *testing.T) {
//		engine := New()
//		engine.Debug()                                       // Turn on the debugging mode, and the debugging information will be printed when the script exits
//		engine.Init()                                        // Import the built-in native library
//		InitPluginGroup(engine)                              // Initialize the plug-in group
//		PatchEngine(engine)                                  // Some libraries are missing functions.
//		engine.SetIncludePath("/Users/z3/nasl/nasl-plugins") // Set the location of the nasl dependent library
//		engine.LoadGroup(PluginGroupApache)
//
//		//Get the dependencies of all scripts
//		libs := map[string]struct{}{}
//		engine.compiler.RegisterVisitHook("includeHook", func(c *visitors.Compiler, ctx antlr.ParserRuleContext) {
//			if v, ok := ctx.(*nasl.CallExpressionContext); ok {
//				if v.SingleExpression().GetText() != "include" {
//					return
//				}
//				if v.ArgumentList() == nil {
//					return
//				}
//				argumentsCtx, ok := v.ArgumentList().(*nasl.ArgumentListContext)
//				if !ok {
//					return
//				}
//				arguments := argumentsCtx.AllArgument()
//				if arguments == nil || len(arguments) == 0 {
//					return
//				}
//				libs[strings.Trim(arguments[0].GetText(), `"`)] = struct{}{}
//			}
//		})
//		for path, _ := range engine.scripts {
//			code, err := os.ReadFile(path)
//			if err != nil {
//				panic(err)
//			}
//			err = engine.Compile(string(code))
//			if err != nil {
//				panic(err)
//			}
//		}
//		engine.compiler.UnregisterVisitHook("includeHook")
//		//for lib, _ := range libs {
//		//	fmt.Println(lib)
//		//}
//		//Check whether the built-in functions used by the dependent libraries exist
//		missMethod := map[string]struct{}{}
//		userDefinedMethod := map[string]struct{}{}
//		engine.compiler.RegisterVisitHook("buildInMethodCheck", func(c *visitors.Compiler, ctx antlr.ParserRuleContext) {
//
//			if v, ok := ctx.(*nasl.FunctionDeclarationStatementContext); ok {
//				id := v.Identifier()
//				if v1, ok := id.(*nasl.IdentifierContext); ok {
//					userDefinedMethod[v1.GetText()] = struct{}{}
//				}
//			}
//			if v, ok := ctx.(*nasl.CallExpressionContext); ok {
//				id := v.SingleExpression().GetText()
//				hasMethod := false
//				if _, ok := NaslLib[id]; ok {
//					hasMethod = true
//				}
//				if _, ok := lib.NaslBuildInNativeMethod[id]; ok {
//					hasMethod = true
//				}
//				if _, ok := engine.GetVirtualMachine().GetVar(id); ok {
//					hasMethod = true
//				}
//				if !hasMethod {
//					missMethod[id] = struct{}{}
//				}
//			}
//		})
//		libsPath := []string{}
//		for lib, _ := range libs {
//			libsPath = append(libsPath, path.Join(engine.naslLibsPath, lib))
//		}
//		for _, path := range libsPath {
//			code, err := os.ReadFile(path)
//			if err != nil {
//				panic(err)
//			}
//			err = engine.Compile(string(code))
//			if err != nil {
//				panic(err)
//			}
//		}
//		engine.compiler.UnregisterVisitHook("buildInMethodCheck")
//		for s, _ := range missMethod {
//			if _, ok := userDefinedMethod[s]; !ok {
//				fmt.Println(s)
//			}
//		}
//		//BuildInMethodCheck(engine) // Detect the currently loaded scripts Does the built-in function exist?
//	}
func TestPocScanner(t *testing.T) {
	consts.GetGormProjectDatabase()
	engine := NewScriptEngine()
	//engine.vm.GetConfig().SetStopRecover(true)
	engine.Debug() // Turn on the debugging mode, and the debugging information will be printed when the script exits
	engine.LoadFamilys("Product detection")
	//engine.LoadScriptFromDb("gb_cisco_asa_detect.nasl")
	//engine.LoadScriptFromDb("gb_apache_hadoop_detect.nasl")
	engine.SetGoroutineNum(10)
	engine.AddEngineHooks(func(engine *Engine) {
		inFun := false
		engine.vm.AddBreakPoint(func(v *yakvm.VirtualMachine) bool {
			defer func() {
				if err := recover(); err != nil {
					fmt.Println(err)
				}
			}()
			fm := v.CurrentFM()
			if fm == nil {
				return false
			}
			if fm.GetVerbose() == "function: recv_mysql_server_handshake" {
				inFun = true
			}
			if inFun {
				if fm.CurrentCode().StartLineNumber == 96 {
					v, ok := fm.CurrentScope().GetValueByName("buf")
					if ok {
						println(v.Value)
					}
				}
				inFun = false
			}
			return false
		})
		engine.RegisterBuildInMethodHook("build_detection_report", func(origin NaslBuildInMethod, engine *Engine, params *NaslBuildInMethodParam) (interface{}, error) {
			scriptObj := engine.scriptObj
			app := params.getParamByName("app", "").String()
			version := params.getParamByName("version", "").String()
			install := params.getParamByName("install", "").String()
			cpe := params.getParamByName("cpe", "").String()
			concluded := params.getParamByName("concluded", "").String()
			if strings.TrimSpace(concluded) == "" || concluded == "Concluded from:" || concluded == "unknown" {
				return origin(engine, params)
			}
			riskType := ""
			if v, ok := utils2.ActToChinese[scriptObj.Category]; ok {
				riskType = v
			} else {
				riskType = scriptObj.Category
			}
			source := "[NaslScript] " + engine.scriptObj.ScriptName
			concludedUrl := params.getParamByName("concludedUrl", "").String()
			solution := utils.MapGetString(engine.scriptObj.Tags, "solution")
			summary := utils.MapGetString(engine.scriptObj.Tags, "summary")
			cve := strings.Join(scriptObj.CVE, ", ")
			//xrefStr := ""
			//for k, v := range engine.scriptObj.Xrefs {
			//	xrefStr += fmt.Sprintf("\n Reference: %s(%s)", v, k)
			//}
			title := fmt.Sprintf("The detection target exists [%s] application, the version number is [%s]", app, version)
			if cve != "" {
				title += fmt.Sprintf(", CVE: %s", summary)
			}
			yakit.NewRisk(concludedUrl,
				yakit.WithRiskParam_Title(title),
				yakit.WithRiskParam_RiskType(riskType),
				yakit.WithRiskParam_Severity("low"),
				yakit.WithRiskParam_YakitPluginName(source),
				yakit.WithRiskParam_Description(summary),
				yakit.WithRiskParam_Solution(solution),
				yakit.WithRiskParam_Details(map[string]interface{}{
					"app":       app,
					"version":   version,
					"install":   install,
					"cpe":       cpe,
					"concluded": concluded,
					"source":    source,
					"cve":       cve,
				}),
			)
			return origin(engine, params)
		})
		// You need to patch all the ACT_SCAN scripts
		engine.AddNaslLibPatch("ping_host.nasl", func(code string) string {
			codeBytes, err := os.ReadFile("/Users/z3/Downloads/ping_host_patch.nasl")
			if err != nil {
				return code
			}
			return string(codeBytes)
		})
		engine.AddNaslLibPatch("http_keepalive.inc", func(code string) string {
			codeLines := strings.Split(code, "\n")
			if len(codeLines) > 341 {
				codeLines[341] = "if( \" HTTP/1.1\" >< data && ! egrep( pattern:\"User-Agent:.+\", string:data, icase:TRUE ) ) {"
				code = strings.Join(codeLines, "\n")
			}
			return code
		})
		engine.AddNaslLibPatch("gb_altn_mdaemon_http_detect.nasl", func(code string) string {
			codeLines := strings.Split(code, "\n")
			if len(codeLines) > 55 {
				codeLines[55] = "if ((res =~ \"MDaemon[- ]Webmail\" || res =~ \"Server\\s*:\\s*WDaemon\") && \"WorldClient.dll\" >< res) {"
				code = strings.Join(codeLines, "\n")
			}
			return code
		})
		engine.AddNaslLibPatch("gb_apache_tomcat_open_redirect_vuln_lin.nasl", func(code string) string {
			codeBytes, err := os.ReadFile("/Users/z3/nasl/nasl-plugins/2018/apache/gb_apache_tomcat_open_redirect_vuln_lin.nasl")
			if err != nil {
				return code
			}
			return string(codeBytes)
		})
	})
	start := time.Now()
	err := engine.ScanTarget("https://uat.sdeweb.hkcsl.com")
	if err != nil {
		log.Error(err)
	}
	log.Info("scan time: ", time.Since(start))
	data := engine.GetKBData()
	data["Host/port_infos"] = nil
	spew.Dump(data)
}
