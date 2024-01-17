package coreplugin

import (
	"fmt"
	"testing"

	"github.com/yaklang/yaklang/common/yakgrpc"
)

func TestGRPCMUSTPASS_SSTI(t *testing.T) {
	client, err := yakgrpc.NewLocalClient()
	if err != nil {
		panic(err)
	}

	pluginName := "SSTI Expr Server Template Expression Injection"
	vul := VulInfo{
		Path: []string{"/expr/injection?a=1", "/expr/injection?b={%22a%22:%201}", "/expr/injection?c=abc"},
		ExpectedResult: map[string]int{
			fmt.Sprintf("SSTI Expr Injection (Param:a): %s/expr/injection?a=", vulAddr): 3,
			fmt.Sprintf("SSTI Expr Injection (Param:b): %s/expr/injection?b=", vulAddr): 3,
			fmt.Sprintf("SSTI Expr Injection (Param:c): %s/expr/injection?c=", vulAddr): 3,
		},
		StrictMode: false,
	}

	Must(CoreMitmPlugTest(pluginName, server, vul, client, t), "The injection detection results of the SSTI plug-in do not meet expectations")
}
