package yaktest

import (
	"fmt"
	"testing"
)

func TestMisc_PingScan(t *testing.T) {

	cases := []YakTestCase{
		{
			Name: "test ping scan",
			Src:  fmt.Sprintf(`loglevel("info");for result = range ping.Scan("47.52.100.0", ping.concurrent(20)) {if(result.Ok){println(result.Reason);};}`),
		},
	}

	Run("pingscan usability test", t, cases...)
}

func TestMisc_SynPingScan(t *testing.T) {

	cases := []YakTestCase{
		{
			Name: "test ping scan",
			Src:  fmt.Sprintf(`loglevel("info");for result = range ping.Scan("47.52.100.0", ping.concurrent(20)) {if(result.Ok){println(result.Reason);};}`),
		},
	}

	Run("pingscan usability test", t, cases...)
}
