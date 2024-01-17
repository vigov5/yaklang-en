package yaktest

import (
	"github.com/yaklang/yaklang/common/utils"
	"testing"
)

func TestMisc_JS(t *testing.T) {
	randomPort := utils.GetRandomAvailableTCPPort()
	utils.NewWebHookServer(randomPort, func(data interface{}) {

	})

	jsCode1 := `// Sample xyzzy example
    (function(){
        if (3.14159 > 0) {
            console.log("Hello, World.");
            return;
        }

        var xyzzy = NaN;
        console.log("Nothing happens.");
        return xyzzy;
    })();`

	jsCode2 := `// Sample xyzzy example
    function test(){
        if (3.14159 > 0) {
            console.log("Call By Function!!!!!!!!!!!!.");
            return "CallByFunc";
        }

        var xyzzy = NaN;
        console.log("Nothing happens.");
        return xyzzy;
    }`
	cases := []YakTestCase{
		{
			Name: "test js",
			Src:  `die(js.Run("1+1")[2])`,
		},
		{
			Name: "Test closure function js",
			Src:  "vm, value, err = js.Run(`" + jsCode1 + "`); die(err); dump(value)",
		},
		{
			Name: "Test function definition execution js",
			Src:  "value, err = js.CallFunctionFromCode(`" + jsCode2 + "`, `test`); die(err); dump(value)",
		},
	}

	Run("JS OTTO usability test", t, cases...)
}
