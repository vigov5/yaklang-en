package yaktest

import (
	"github.com/yaklang/yaklang/common/consts"
	"os"
	"testing"

	"github.com/yaklang/yaklang/common/yakgrpc/yakit"
)

/*
manager = hook.NewMixPluginCaller()[0]
manager.LoadPlugin("Tomcat login blasting")
manager.SetConcurrent(20)

loglevel("info")
yakit.Info("loading port")
for result = range servicescan.Scan("47.52.100.104", "443,22")[0] {
    manager.GetNativeCaller().CallByName("handle", result)
}

manager.Wait()
*/

func TestMisc_RuntimeDB(t *testing.T) {
	//os.Setenv("YAKLANGDEBUG", "123")
	consts.GetGormProjectDatabase()
	consts.GetGormProjectDatabase()
	err := yakit.CallPostInitDatabase()
	if err != nil {
		panic(err)
	}
	_ = yakit.ExecResult{}
	os.Setenv(consts.CONST_YAK_SAVE_HTTPFLOW, "true")
	cases := []YakTestCase{
		{
			Name: "Function return value fails",
			Src: `
yakit.AutoInitYakit()

manager = hook.NewMixPluginCaller()[0]
//err = manager.LoadPlugin("Tomcat login blasting")
//dump(err)

err = manager.LoadPlugin("test111")
dump(err)

manager.SetConcurrent(20)
manager.SetDividedContext(true) // Set independent context

loglevel("info")
for result = range servicescan.Scan("47.52.100.104", "443,22")[0] {
    manager.GetNativeCaller().CallByName("handle", result)
}

manager.Wait()
`,
		},
	}

	Run("yaklang syntax fix", t, cases...)
}
