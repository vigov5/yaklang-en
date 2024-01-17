package yaktest

import "testing"

func TestHookCaller(t *testing.T) {
	Run("Test loading plug-in", t, YakTestCase{
		Name: "Test loading plug-in crashes",
		Src: `
m = hook.NewMixPluginCaller()[0]
m.SetDividedContext(true)

err =  m.LoadPlugin("sleep3")
die(err)

start = time.Now().Unix()
m.SetConcurrent(2)
m.HandleServiceScanResult(result)
m.HandleServiceScanResult(result)
m.HandleServiceScanResult(result)
m.HandleServiceScanResult(result)
m.Wait()
du = time.Now().Unix() - start
if du >= 7 {
    panic("concurrent panic")
}




`,
	})
}

// Test mixcaller calling nasl plug-in (not completed)
func TestMixHookCaller(t *testing.T) {
	Run("Test loading plug-in", t, YakTestCase{
		Name: "Test loading plug-in crashes",
		Src: `
m = hook.NewMixPluginCaller()[0]
m.SetDividedContext(true)

err =  m.LoadPlugin("__NaslScript__gb_apache_tomcat_consolidation.nasl")
die(err)
m.SetConcurrent(2)
res,err = servicescan.Scan("183.234.44.226", "8099")
for result in res{
	m.HandleServiceScanResult(result)    
}
m.Wait()
`,
	})
}
