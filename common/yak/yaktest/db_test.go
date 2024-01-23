package yaktest

import (
	"fmt"
	"testing"
)

func TestMisc_DatabaseTest(t *testing.T) {

	cases := []YakTestCase{
		{
			Name: "Test Database",
			Src: fmt.Sprintf(`
for result = range db.QueryHTTPFlowsByID(1443) {
	dump(result)
}
`),
		},
	}

	Run("Test Database Link", t, cases...)
}
