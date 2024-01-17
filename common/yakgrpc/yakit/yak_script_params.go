package yakit

import "github.com/yaklang/yaklang/common/yakgrpc/ypb"

var mitmPluginDefaultPlugins = []*ypb.YakScriptParam{
	{
		Field:        "target",
		TypeVerbose:  "string",
		FieldVerbose: "target (URL/IP:Port)",
		Help:         "Enter the test target of the plug-in for basic crawling (up to 10 requests)",
		Required:     true,
	},
}
