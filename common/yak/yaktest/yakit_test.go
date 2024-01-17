package yaktest

import (
	"fmt"
	"github.com/yaklang/yaklang/common/utils"
	"os"
	"testing"
)

func TestMisc_YAKIT(t *testing.T) {
	randomPort := utils.GetRandomAvailableTCPPort()
	utils.NewWebHookServer(randomPort, func(data interface{}) {

	})

	cases := []YakTestCase{
		{
			Name: "test yakit.File",
			Src:  fmt.Sprintf(`yakit.File("/etc/hosts", "HOSTS", "this is hosts")`),
		},
	}

	Run("yakit.File Usability Test", t, cases...)
}

func TestMisc_YAKIT2(t *testing.T) {
	os.Setenv("YAKMODE", "vm")
	randomPort := utils.GetRandomAvailableTCPPort()
	utils.NewWebHookServer(randomPort, func(data interface{}) {

	})

	cases := []YakTestCase{
		{
			Name: "Test ",
			Src: `println()))
a = 123;
a()
`,
		},
	}

	Run(")))))) Test", t, cases...)
}
