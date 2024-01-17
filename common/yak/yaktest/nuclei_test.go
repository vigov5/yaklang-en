package yaktest

import (
	"fmt"
	"github.com/yaklang/yaklang/common/utils"
	"testing"
)

func TestMisc_NUCLEI(t *testing.T) {
	randomPort := utils.GetRandomAvailableTCPPort()
	utils.NewWebHookServer(randomPort, func(data interface{}) {

	})

	cases := []YakTestCase{
		{
			Name: "test nuclei",
			Src: fmt.Sprintf(`
defer func{
	err := recover()
	if err != nil {
		panic("FAILED")
	}
}

rs, err := nuclei.Scan("http://127.0.0.1:8004", nuclei.noInteractsh(true),
	nuclei.tags("thinkphp"),
)
if err != nil {
    println(err)
	return
}

for r = range rs {
    dump(r)
}
`),
		},
	}

	Run("nuclei usability test (external network)", t, cases...)
}

func TestMisc_NUCLEI_1(t *testing.T) {
	code := `
// CURRENT_NUCLEI_PLUGIN_NAME = "[thinkphp-5023-rce]: ThinkPHP 5.0.23 RCE"
// This script needs to operate, set CURRENT_NUCLEI_PLUGIN_NAME as the variable name
nucleiPoCName = "[thinkphp-5023-rce]: ThinkPHP 5.0.23 RCE" // MITM_PARAMS.CURRENT_NUCLEI_PLUGIN_NAME
yakit.Info("loading yakit nuclei plugin: %s", nucleiPoCName)
script, err := db.GetYakitPluginByName(nucleiPoCName)
if err != nil {
	yakit.Error("load yakit-plugin(nuclei) failed: %s", err)
	return
}
f, err := file.TempFile()
if err != nil {
	yakit.Error("load tempfile to save nuclei poc failed: %s", err)
	return
}
pocName := f.Name()
f.WriteString(script.Content)
f.Close()

execNuclei := func(target) {
	res, err = nuclei.Scan(
        target, nuclei.templates(pocName),
        nuclei.retry(0), nuclei.stopAtFirstMatch(true), nuclei.timeout(10), 
        //nuclei.debug(cli.Have("debug")), 
        //nuclei.verbose(cli.Have("debug")),
    )
	if err != nil {
		yakit.Error("nuclei scan %v failed: %s", target, err)
		return
	}
	for pocVul = range res {
        if pocVul == nil {
			continue
        }
		yakit.Output(nuclei.PocVulToRisk(pocVul))		
	}
}

execNuclei("127.0.0.1:8080")
`
	cases := []YakTestCase{
		{
			Name: "test nuclei package function",
			Src:  code,
		},
	}

	Run("nuclei package function usability smoke test", t, cases...)

}
