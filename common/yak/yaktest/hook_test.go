package yaktest

import (
	"fmt"
	"github.com/yaklang/yaklang/common/consts"
	"github.com/yaklang/yaklang/common/utils"
	"github.com/yaklang/yaklang/common/yakgrpc/yakit"
	"testing"
)

func TestMisc_Hook(t *testing.T) {
	randomPort := utils.GetRandomAvailableTCPPort()
	utils.NewWebHookServer(randomPort, func(data interface{}) {

	})

	s := &yakit.YakScript{
		ScriptName: "yakit-plugin-test-abcccc",
		Type:       "testtype",
		Content:    `clear = func() {println("Hello World")}`,
	}
	yakit.CreateOrUpdateYakScriptByName(consts.GetGormProfileDatabase(), s.ScriptName, s)
	defer yakit.DeleteYakScriptByName(consts.GetGormProfileDatabase(), s.ScriptName)

	cases := []YakTestCase{
		{
			Name: "test hook",
			Src: fmt.Sprintf(`
a = hook.NewManager()
err = hook.LoadYakitPlugin(a, "asdfhuiasdhfhasdf", "clear")
if err == nil {
	die("load failed")
}
`),
		},
		{Name: "Test hooks, known plug-ins", Src: `
a = hook.NewManager()
err = hook.LoadYakitPlugin(a, "testtype", "clear")
if err != nil {
    die(err)
}

a.CallByName("clear")
`},
	}

	Run("hooks test", t, cases...)
}
