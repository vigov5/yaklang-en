package coreplugin

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/yaklang/yaklang/common/cybertunnel/tpb"
	"github.com/yaklang/yaklang/common/log"
	"github.com/yaklang/yaklang/common/utils"
	"github.com/yaklang/yaklang/common/vulinbox"
	"github.com/yaklang/yaklang/common/yak/yaklang"
	"github.com/yaklang/yaklang/common/yak/yaklib"
	"github.com/yaklang/yaklang/common/yak/yaklib/codec"
	"github.com/yaklang/yaklang/common/yakgrpc"
)

func TestGRPCMUSTPASS_Fastjson(t *testing.T) {
	domainMap := map[string]string{}
	yaklib.RiskExports["NewDNSLogDomain"] = func() (string, string, error) {
		token := utils.RandStringBytes(10)
		domainMap[token] = token + ".dnslog.cn"
		return token + ".dnslog.cn", token, nil
	}
	yaklib.RiskExports["CheckDNSLogByToken"] = func(token string, timeout ...float64) ([]*tpb.DNSLogEvent, error) {
		timeout1 := 1.0
		if len(timeout) > 0 {
			timeout1 = timeout[0]
		}
		if v, ok := domainMap[token]; ok {
			res := []*tpb.DNSLogEvent{}
			for i := 0; i < 3; i++ {
				hasRecord := false
				vulinbox.DnsRecord.Range(func(key, value any) bool {
					domain := key.(string)
					if strings.HasSuffix(domain, v) {
						hasRecord = true
						res = append(res, &tpb.DNSLogEvent{
							Domain: domain,
						})
						return false
					}
					return true
				})
				if hasRecord {
					return res, nil
				} else {
					time.Sleep(utils.FloatSecondDuration(timeout1))
				}
			}
		}
		return nil, errors.New("not found record")
	}
	yaklang.Import("risk", yaklib.RiskExports)
	client, err := yakgrpc.NewLocalClient()
	if err != nil {
		panic(err)
	}

	log.Infof("vulAddr: %v", vulAddr)
	//time.Sleep(5 * time.Hour)
	pluginName := "Fastjson Comprehensive detection of"
	//wg := sync.WaitGroup{}
	addFastjsonTestCase := func(vulInfo VulInfo, msg ...string) {
		//wg.Add(1)
		//go func() {
		//	defer wg.Done()
		//	Must(TestCoreMitmPlug(pluginName, server, vulInfo, client, t), msg...)
		//}()
		Must(CoreMitmPlugTest(pluginName, server, vulInfo, client, t), msg...)
	}
	//defer wg.Wait()
	vulInGet := VulInfo{
		Path: []string{
			"/fastjson/json-in-query?auth=" + codec.EncodeUrlCode(`{"user":"admin","password":"password"}`) + "&action=login",
		},
		ExpectedResult: map[string]int{
			"target fastjson framework may have RCE vulnerabilities (DNSLog Check )": 1,
		},
		StrictMode: true,
		Id:         "json in query test",
	}

	//vulInForm := VulInfo{
	//	Method: "POST",
	//	Path: []string{
	//		"/fastjson/json-in-form",
	//	},
	//	Headers: []*ypb.KVPair{
	//		{
	//			Key:   "Content-Type",
	//			Value: "application/x-www-form-urlencoded",
	//		},
	//	},
	//	Body: []byte(`auth={"user":"admin","password":"password"}`),
	//	ExpectedResult: map[string]int{
	//		"target fastjson framework may have RCE vulnerabilities (DNSLog Check )": 1,
	//	},
	//	StrictMode: true,
	//	Id: "json in form",
	//}
	//vulInBodyJson := VulInfo{
	//	Method: "POST",
	//	Path: []string{
	//		"/fastjson/json-in-body",
	//	},
	//	Body: []byte(`{"user":"admin","password":"password"}`),
	//	Headers: []*ypb.KVPair{
	//		{
	//			Key:   "Content-Type",
	//			Value: "application/json",
	//		},
	//	},
	//	ExpectedResult: map[string]int{
	//		"target fastjson framework may have RCE vulnerabilities (DNSLog Check )": 1,
	//	},
	//	StrictMode: true,
	//	Id: "json in body",
	//}
	//vulInGetServeByJackson := VulInfo{ // No vulnerabilities should be detected here, and the number of packets sent should be 1
	//	Method: "GET",
	//	Path: []string{
	//		"/fastjson/jackson-in-query?auth=" + codec.EncodeUrlCode(`{"user":"admin","password":"password"}`) + "&action=login",
	//	},
	//	ExpectedResult: map[string]int{},
	//	StrictMode:     true,
	//	Id: "jackson in query",
	//}
	addFastjsonTestCase(vulInGet, "Fastjson comprehensive detection plug-in’s detection results for json in query are not as expected.")
	//addFastjsonTestCase(vulInForm, "Fastjson comprehensive detection plug-in’s detection results for json in form are not as expected.")
	//addFastjsonTestCase(vulInBodyJson, "Fastjson comprehensive detection plug-in’s detection results for json in body are not as expected.")
	//addFastjsonTestCase(vulInGetServeByJackson, "Fastjson comprehensive detection plug-in’s detection results for Jackson are not as expected.")
	// TODO: Need to fix the problem of not being able to obtain Duration after fuzz request error.
	//vulInGetIntranet := VulInfo{
	//	Method: "GET",
	//	Path: []string{
	//		"/fastjson/get-in-query-intranet?auth=" + codec.EncodeUrlCode(`{"user":"admin","password":"password"}`) + "&action=login",
	//	},
	//	ExpectedResult: map[string]int{
	//		"target fastjson framework may have RCE vulnerabilities (Delay Check)": 1,
	//	},
	//	StrictMode: true,
	//}
	//Must(TestCoreMitmPlug(pluginName, server, vulInGetIntranet, client, t), "Fastjson comprehensive detection plug-in’s detection results are not as expected.")
	// TODO: Cookie Fuzz needs to support automatic decoding of
	//vulInGet := VulInfo{
	//	Method: "GET",
	//	Path: []string{
	//		"/fastjson/json-in-cookie?action=login",
	//	},
	//	Headers: []*ypb.KVPair{
	//		{
	//			Key:   "Cookie",
	//			Value: `auth=` + codec.EncodeBase64Url(`{"id":"-1"}`),
	//		},
	//	},
	//	ExpectedResult: map[string]int{
	//		"target fastjson framework may have RCE vulnerabilities (DNSLog Check )": 1,
	//	},
	//	StrictMode: true,
	//}
	//Must(TestCoreMitmPlug(pluginName, server, vulInGet, client, t), "Fastjson comprehensive detection plug-in’s detection results are not as expected.")
	// TODO: Authorization Fuzz needs to support automatic decoding.
	//vulInAuthorization := VulInfo{
	//	Method: "GET",
	//	Path: []string{
	//		"/fastjson/json-in-authorization?action=login",
	//	},
	//	Headers: []*ypb.KVPair{
	//		{
	//			Key:   "Authorization",
	//			Value: `Basic ` + codec.EncodeBase64Url(`{"user":"admin","password":"password"}`),
	//		},
	//	},
	//	ExpectedResult: map[string]int{
	//		"target fastjson framework may have RCE vulnerabilities (DNSLog Check )": 1,
	//	},
	//	StrictMode: true,
	//}
	//addFastjsonTestCase(vulInAuthorization, "Fastjson comprehensive detection plug-in’s detection results for Jackson are not as expected.")
}
func TestFastjson(t *testing.T) {
	client, err := yakgrpc.NewLocalClient()
	if err != nil {
		panic(err)
	}

	log.Infof("vulAddr: %v", vulAddr)
	//time.Sleep(5 * time.Hour)
	pluginName := "Fastjson Comprehensive detection of"
	vulInGet := VulInfo{
		Method: "GET",
		Path: []string{
			"/fastjson/get-in-query-intranet?auth=" + codec.EncodeUrlCode(`{"user":"admin","password":"password"}`) + "&action=login",
		},
		ExpectedResult: map[string]int{
			"target fastjson framework may have RCE vulnerabilities (Delay Check)": 1,
		},
		StrictMode: true,
	}
	Must(CoreMitmPlugTest(pluginName, server, vulInGet, client, t), "Fastjson comprehensive detection plug-in’s detection results are not as expected.")
}
