package fp

import (
	"context"
	"net"
	"strconv"
	"strings"
	"sync"

	"github.com/yaklang/yaklang/common/fp/iotdevfp"
	"github.com/yaklang/yaklang/common/fp/webfingerprint"
	"github.com/yaklang/yaklang/common/log"
	utils2 "github.com/yaklang/yaklang/common/utils"
	"github.com/yaklang/yaklang/common/utils/lowhttp"
)

func (f *Matcher) webDetector(result *MatchResult, ctx context.Context, config *Config, host string, ip net.IP, port int) (*MatchResult, error) {
	//////////////////////////////////////////////////////////////////////////
	////////////////////////Start Web fingerprinting from here//////////////////////////
	//////////////////////////////////////////////////////////////////////////
	// First perform configuration detection of various IoT devices (highest priority)
	iotDetectCtx, cancel := context.WithTimeout(ctx, config.ProbeTimeout)
	defer cancel()

	//////////////////////////////////////////////////////////////////////////
	////////////////////////////// IoT device optimization ///////////////////////////////
	//////////////////////////////////////////////////////////////////////////
	//log.Infof("start to check iotdevfp: %v", utils2.HostPort(ip.String(), port))
	h := host
	if h == "" {
		h = ip.String()
	}
	isOpen, httpBanners, err := FetchBannerFromHostPortEx(iotDetectCtx, nil, h, port, int64(config.FingerprintDataSize), config.Proxies...)
	if err != nil {
		if !isOpen {
			return &MatchResult{
				Target: host,
				Port:   port,
				State:  CLOSED,
				Reason: err.Error(),
				Fingerprint: &FingerprintInfo{
					IP:   ip.String(),
					Port: port,
				},
			}, nil
		}
		return &MatchResult{
			Target: host,
			Port:   port,
			State:  OPEN,
			Reason: "",
			Fingerprint: &FingerprintInfo{
				IP:   ip.String(),
				Port: port,
			},
		}, nil
	}

	if httpBanners == nil {
		// Set initialization matching result
		return &MatchResult{
			Target: host,
			Port:   port,
			State:  OPEN,
			Fingerprint: &FingerprintInfo{
				IP:   ip.String(),
				Port: port,
			},
		}, nil
	}

	// If Web fingerprint detection is forcibly enabled, Bypass fingerprint detection conditions are required
	// Force Fingerprint to a value that can perform Web fingerprinting
	if result.Fingerprint == nil {
		result.Fingerprint = &FingerprintInfo{
			IP:   ip.String(),
			Port: port,
		}
	}

	var (
		//wg                      = new(sync.WaitGroup)
		results     = new(sync.Map)
		cpeAnalyzer = webfingerprint.NewCPEAnalyzer()
		httpflows   []*HTTPFlow
	)
	if httpBanners != nil {
		log.Debugf("finished to check iotdevfp: %v fetch response[%v]", utils2.HostPort(ip.String(), port), len(httpBanners))
		result.State = OPEN
		result.Fingerprint.ServiceName = "http"
		for _, i := range httpBanners {
			var iotdevResults = iotdevfp.MatchAll(i.Bytes())
			if result.Fingerprint == nil {
				result.Fingerprint = &FingerprintInfo{
					IP:          ip.String(),
					Port:        port,
					Proto:       TCP,
					ServiceName: "http",
					Banner:      strconv.Quote(string(i.Bytes())),
				}
			}

			if i.IsHttps {
				name := strings.ToLower(result.Fingerprint.ServiceName)
				if !strings.Contains(name, "https") &&
					strings.Contains(name, "http") {
					// does not contain https but contains http
					result.Fingerprint.ServiceName = strings.ReplaceAll(name, "http", "https")
				}

				if name == "" {
					result.Fingerprint.ServiceName = "https"
				}

			}

			var currentCPE []*webfingerprint.CPE
			for _, iotdevResult := range iotdevResults {
				result.Fingerprint.CPEs = append(result.Fingerprint.CPEs, iotdevResult.GetCPE())
				cpeIns, _ := webfingerprint.ParseToCPE(iotdevResult.GetCPE())
				if cpeIns != nil {
					currentCPE = append(currentCPE, cpeIns)
				}
			}

			info := i
			cpes, err := f.wfMatcher.MatchWithOptions(info, config.GenerateWebFingerprintConfigOptions()...)
			if err != nil {
				if !strings.Contains(err.Error(), "no rules matched") {
					continue
				}
			}

			// If fingerprint information is detected
			if len(cpes) > 0 {
				currentCPE = append(currentCPE, cpes...)
				urlStr := info.URL.String()
				cpeAnalyzer.Feed(urlStr, cpes...)
				results.Store(urlStr, cpes)
			}

			requestHeader, requestBody := lowhttp.SplitHTTPHeadersAndBodyFromPacket(info.RequestRaw)
			flow := &HTTPFlow{
				StatusCode:     info.StatusCode,
				IsHTTPS:        info.IsHttps,
				RequestHeader:  []byte(requestHeader),
				RequestBody:    requestBody,
				ResponseHeader: info.ResponseHeaderBytes(),
				ResponseBody:   info.Body,
				CPEs:           currentCPE,
			}
			httpflows = append(httpflows, flow)
		}
	}
	urlCpe := map[string][]*webfingerprint.CPE{}
	results.Range(func(key, value interface{}) bool {
		log.Debugf("url: %s cpes: %#v", key, value)
		_url := key.(string)
		cpes := value.([]*webfingerprint.CPE)
		urlCpe[_url] = cpes
		return true
	})

	// Complete URL for FingerprintResult Fingerprint information
	result.Fingerprint.CPEFromUrls = urlCpe

	// If possible, fingerprint recognition needs to be improved HTTP related request
	result.Fingerprint.HttpFlows = httpflows

	// Update the new cpes to the original cpe list
	cpes := result.Fingerprint.CPEs
	var cpesStrRaw []string
	for _, c := range cpeAnalyzer.AvailableCPE() {
		cpesStrRaw = append(cpesStrRaw, c.String())
	}
	result.Fingerprint.CPEs = append(cpes, cpesStrRaw...)

	// You need to check the necessary fields of Fingerprint before returning the result
	if result.Fingerprint.ServiceName == "" {
		result.Fingerprint.ServiceName = strings.ToLower(string(result.Fingerprint.Proto))
	}

	if httpBanners != nil {
		result.State = OPEN
		result.Fingerprint.Proto = TCP
		if result.Fingerprint.ServiceName == "" {
			if httpBanners[0].IsHttps {
				result.Fingerprint.ServiceName = "https"
			} else {
				result.Fingerprint.ServiceName = "http"
			}
		}
	}

	switch result.Fingerprint.Proto {
	case TCP, UDP:
	default:
		result.Fingerprint.Proto = TCP
	}
	return result, nil
}
