package coreplugin

import (
	"testing"

	"github.com/yaklang/yaklang/common/yakgrpc"
)

func TestGRPCMUSTPASS_Shiro(t *testing.T) {
	client, err := yakgrpc.NewLocalClient()
	if err != nil {
		panic(err)
	}
	pluginName := "Shiro fingerprinting + weak password detection"
	vul1 := VulInfo{
		Path: []string{"/shiro/cbc"},
		ExpectedResult: map[string]int{
			"detected Shiro (Cookie) framework usage.": 1,
			"(Shiro default KEY)":         1,
			"(Shiro Header echo)":      1,
		},
		StrictMode: false,
	}
	vul2 := VulInfo{
		Path: []string{"/shiro/gcm"},
		ExpectedResult: map[string]int{
			"detected Shiro (Cookie) framework usage.": 1,
			"(Shiro default KEY)":         1,
			"(Shiro Header echo)":      1,
		},
		StrictMode: false,
	}

	Must(CoreMitmPlugTest(pluginName, server, vul1, client, t), "The Shiro plug-in’s detection results for lower versions of Shiro do not meet expectations.")
	Must(CoreMitmPlugTest(pluginName, server, vul2, client, t), "The Shiro plug-in’s detection results for high versions of Shiro do not meet expectations.")
}
