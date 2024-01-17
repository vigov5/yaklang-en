package coreplugin

import (
	"strings"

	uuid "github.com/satori/go.uuid"
	"github.com/yaklang/yaklang/common/consts"
	"github.com/yaklang/yaklang/common/log"
	"github.com/yaklang/yaklang/common/utils"
	"github.com/yaklang/yaklang/common/yakgrpc/yakit"
)

var (
	buildInPlugin = make(map[string]*yakit.YakScript)
)

type pluginConfig struct {
	Help   string
	Author []string
}

type pluginOption func(*pluginConfig)

func withPluginHelp(pluginHelp string) pluginOption {
	return func(config *pluginConfig) {
		config.Help = pluginHelp
	}
}

func withPluginAuthors(authors ...string) pluginOption {
	return func(config *pluginConfig) {
		config.Author = authors
	}
}

func registerBuildInPlugin(pluginType string, name string, opt ...pluginOption) {
	var codes = string(GetCorePluginData(name))
	if len(codes) <= 0 {
		return
	}

	config := &pluginConfig{}
	for _, o := range opt {
		o(config)
	}

	var plugin = &yakit.YakScript{
		ScriptName:         name,
		Type:               pluginType,
		Content:            codes,
		Help:               config.Help,
		Author:             "yaklang.io",
		OnlineContributors: strings.Join(config.Author, ","),
		Uuid:               uuid.NewV4().String(),
		OnlineOfficial:     true,
		IsCorePlugin:       true,
		HeadImg:            `https://yaklang.oss-cn-beijing.aliyuncs.com/yaklang-avator-logo.png`,
	}
	buildInPlugin[name] = plugin
	OverWriteYakPlugin(plugin.ScriptName, plugin)
}

func init() {
	yakit.RegisterPostInitDatabaseFunction(func() error {
		log.Debug("start to load core plugin")
		registerBuildInPlugin(
			"mitm",
			"HTTP request smuggling",
			withPluginAuthors("V1ll4n"),
			withPluginHelp("HTTP request smuggling vulnerability detection, by setting malformed Content-Length (CL) and Transfer-Encoding (TE) to detect whether the server responds to malformed The packet generated an unsafe response."),
		)
		registerBuildInPlugin(
			"mitm", "CSRF form protection and CORS improper configuration detection",
			withPluginHelp("Detect whether the application has CSRF form protection and improper CORS configuration"),
			withPluginAuthors("Rookie"),
		)
		registerBuildInPlugin(
			"mitm", "Fastjson Comprehensive detection",
			withPluginHelp("Comprehensive FastJSON deserialization vulnerability detection"),
			withPluginAuthors("z3"),
		)
		registerBuildInPlugin(
			"mitm", "Shiro fingerprinting + weak password detection",
			withPluginHelp("Identify whether the application is a Shiro application and try to detect the default KEY (CBC/GCM mode is supported), when the default KEY is found, an exploit chain detection is performed"),
			withPluginAuthors("z3", "go0p"),
		)
		registerBuildInPlugin(
			"mitm", "SSRF HTTP Public",
			withPluginHelp("Detect SSRF vulnerability in parameters"),
		)
		registerBuildInPlugin(
			"mitm", "SQL injection-UNION injection-MD5 function",
			withPluginHelp("Union injection, use md5 function to detect feature output (mysql/postgres）"),
			withPluginAuthors("V1ll4n"),
		)
		registerBuildInPlugin(
			"mitm", "SQL injection-MySQL-ErrorBased",
			withPluginHelp("MySQL error injection (using MySQL 16 Hexadecimal string feature detection)"),
			withPluginAuthors("V1ll4n"),
		)
		registerBuildInPlugin(
			"mitm",
			"SSTI Expr Server Template Expression Injection",
			withPluginHelp("SSTI server template expression injection vulnerability (General vulnerability detection)"),
			withPluginAuthors("V1ll4n"),
		)
		registerBuildInPlugin(
			"mitm", "Swagger JSON leak",
			withPluginHelp("Check whether the website is open API information of Swagger JSON"),
			withPluginAuthors("V1ll4n"),
		)
		registerBuildInPlugin(
			"mitm", "Heuristic SQL injection detection",
			withPluginHelp("Detect sql injection for various situation parameters in the request package"),
			withPluginAuthors("After the rain, the sky clears & the umbrella falls away"),
		)
		registerBuildInPlugin(
			"mitm", "Basic XSS detection",
			withPluginHelp("XSS in one detection parameter Algorithm, supports various XSS encoded or in JSON detection"),
			withPluginAuthors("WaY"),
		)
		registerBuildInPlugin(
			"mitm", "file contains",
			withPluginHelp(`Using PHP pseudo-protocol features and base64 convergence features to test file contains`),
			withPluginAuthors("V1ll4n"),
		)
		registerBuildInPlugin(
			"mitm", "Open URL redirect vulnerability",
			withPluginHelp("detects open URL redirection vulnerabilities, you can check meta / js / Content in location"),
			withPluginAuthors("Rookie"),
		)
		registerBuildInPlugin(
			"mitm", "Echo command injection",
			withPluginHelp("Detection of echo command injection vulnerabilities ( Do not detect command injection in cookies)"),
			withPluginAuthors("V1ll4n"),
		)
		return nil
	})
}

func OverWriteCorePluginToLocal() {
	for pluginName, instance := range buildInPlugin {
		OverWriteYakPlugin(pluginName, instance)
	}
}

func OverWriteYakPlugin(name string, scriptData *yakit.YakScript) {
	codeBytes := GetCorePluginData(name)
	if codeBytes == nil {
		log.Errorf("fetch buildin-plugin: %v failed", name)
		return
	}
	backendSha1 := utils.CalcSha1(string(codeBytes), scriptData.HeadImg)
	databasePlugins := yakit.QueryYakScriptByNames(consts.GetGormProfileDatabase(), name)
	if len(databasePlugins) == 0 {
		log.Infof("add core plugin %v to plugin database", name)
		// Add core plug-in field
		scriptData.IsCorePlugin = true
		err := yakit.CreateOrUpdateYakScriptByName(consts.GetGormProfileDatabase(), name, scriptData)
		if err != nil {
			log.Errorf("create/update yak script[%v] failed: %s", name, err)
		}
		return
	}
	databasePlugin := databasePlugins[0]
	if databasePlugin.Content != "" && utils.CalcSha1(databasePlugin.Content, databasePlugin.HeadImg) == backendSha1 && databasePlugin.IsCorePlugin {
		log.Debugf("existed plugin's code is not changed, skip: %v", name)
		return
	} else {
		log.Infof("start to override existed plugin: %v", name)
		databasePlugin.Content = string(codeBytes)
		databasePlugin.IsCorePlugin = true
		databasePlugin.HeadImg = scriptData.HeadImg
		err := yakit.CreateOrUpdateYakScriptByName(consts.GetGormProfileDatabase(), name, databasePlugin)
		if err != nil {
			log.Errorf("override %v failed: %s", name, err)
			return
		}
		log.Debugf("override buildin-plugin %v success", name)
	}
}
