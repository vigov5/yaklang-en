package coreplugin

import (
	"fmt"
	"testing"

	"github.com/yaklang/yaklang/common/yakgrpc"
)

func TestGRPCMUSTPASS_CSRF(t *testing.T) {
	client, err := yakgrpc.NewLocalClient()
	if err != nil {
		panic(err)
	}

	pluginName := "CSRF form protection and CORS misconfiguration detection"
	vul := VulInfo{
		Path: []string{
			"/csrf/unsafe",
			"/csrf/safe",
		},
		ExpectedResult: map[string]int{
			fmt.Sprintf("csrf for: %v/csrf/unsafe", vulAddr): 1,
			fmt.Sprintf("csrf for: %v/csrf/safe", vulAddr):   0,
		},
		StrictMode: true,
	}

	Must(CoreMitmPlugTest(pluginName, server, vul, client, t), " ")

}
