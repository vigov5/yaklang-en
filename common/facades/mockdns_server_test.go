package facades

// This test duplicates the test content of raw_net_dns.go to prevent looping of package imports, so it is commented out.
//func TestMockDNSServerDefault(t *testing.T) {
//	for i := 0; i < 10; i++ {
//
//	}
//	randomStr := utils.RandStringBytes(10)
//	var check = false
//	var a = MockDNSServerDefault("", func(record string, domain string) string {
//		spew.Dump(domain)
//		if strings.Contains(domain, randomStr) {
//			check = true
//		}
//		return "1.1.1.1"
//	})
//	var result = yakdns.LookupFirst(randomStr+".baidu.com", yakdns.WithTimeout(5*time.Second), yakdns.WithDNSServers(a))
//
//	spew.Dump(result)
//	if !check {
//		panic("GetFirstIPByDnsWithCache failed")
//	}
//}
