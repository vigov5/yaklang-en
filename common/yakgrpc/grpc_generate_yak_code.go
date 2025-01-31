package yakgrpc

import (
	"bytes"
	"context"
	"fmt"
	"regexp"
	"strings"
	"text/template"

	"github.com/yaklang/yaklang/common/utils"
	"github.com/yaklang/yaklang/common/utils/lowhttp"
	"github.com/yaklang/yaklang/common/yak/yaklib"
	"github.com/yaklang/yaklang/common/yakgrpc/ypb"
)

var extractHostRegexp = regexp.MustCompile(`[Hh]ost:\s+([^\r\n]+)`)

func extractPacketToGenerateParams(isHttps bool, req []byte) map[string]interface{} {
	res := make(map[string]interface{})
	res["https"] = fmt.Sprint(isHttps)
	var target = ""
	var packetRaw = req
	results := extractHostRegexp.FindSubmatchIndex(req)
	if len(results) > 3 {
		start, end := results[2], results[3]
		target = string(req[start:end])
		isMultipart := false
		header, body := lowhttp.SplitHTTPHeadersAndBodyFromPacket(req, func(line string) {
			if !isMultipart {
				isMultipart = strings.Contains(strings.ToLower(line), "multipart/form-data")
			}
		})
		header = strings.ReplaceAll(header, target, "{{params(target)}}")
		header = strings.ReplaceAll(header, "`", "` + \"`\" + `")
		if !isMultipart {
			// If it is not uploading the data packet, just process the escape.
			body = bytes.ReplaceAll(body, []byte("`"), []byte("` + \"`\" + `"))
		} else {
			// If you are uploading a data packet, you need to be able to identify the uploaded content and re-encode it
		}
		packetRaw = lowhttp.ReplaceHTTPPacketBody([]byte(header), body, false)
	}
	res["target"] = target
	res["packetTemplate"] = string(packetRaw)
	return res
}

var BatchPoCTemplate, _ = template.New("BatchPoCTemplate").Parse(`

isHttps = cli.Have("https", cli.setDefault({{ .https }}))
target = cli.String("target", cli.setDefault("{{ .target }}"))
concurrent = cli.Int("concurrent", cli.setDefault(10))

packet = ` + "`" + `{{ .packetTemplate }}` + "`" + `

debug = cli.Have("debug", cli.setDefault(true))

if debug {
    loglevel("debug")
}

batchPacket = func(target) {
    return httpool.Pool(
        packet, 
        # httpool.proxy("http://127.0.0.1:8083"),
        # httpool.proxy("http://127.0.0.1:7890"),
        httpool.rawMode(true),
        httpool.size(concurrent), 
        httpool.redirectTimes(5),
        httpool.perRequestTimeout(10),
        httpool.fuzz(true),
        httpool.fuzzParams({
            "target": target,
        }),
    )
}

if YAK_MAIN {
    res, err = batchPacket(target)
    if err != nil {
        log.error("send packet error: %s", err)
        return
    }

    for result = range res {

        if result.Error != nil {
            yakit.Error("Request[%v] Payload: %v Failed: %v", result.Url, result.Payloads, result.Error)
            continue
        }

        yakit.Info("[%v] Request Result Received! payloads: %v", result.Url, result.Payloads)

        reqBytes := result.RequestRaw
        rspBytes := result.ResponseRaw

        if debug {
            println(string(reqBytes))
            println("---------------------------------")
            println(string(rspBytes))
        }

        // processing result
        riskTarget = target
        if str.MatchAllOfRegexp(rspBytes, ` + "`" + `(?i)foundtextinRsp!` + "`" + `) || str.MatchAllOfSubString(rspBytes, "FoundTextInResponse") {
            urlIns, _ = str.ExtractURLFromHTTPRequestRaw(reqBytes, isHttps)
            if urlIns == nil {
                riskTarget = urlIns.String()
            }
            yakit.Info("Matched for %v", target)
            # Save to RiskTable
            risk.NewRisk(
                riskTarget, risk.severity("high"), risk.type("poc"),
                risk.title("English Title"),            ## English Title for Risk
                risk.titleVerbose("Chinese title"), ## Chinese Title
                risk.details({
                    "target": riskTarget,
                    "request": reqBytes,
                    "response": rspBytes,
                }),
            )
        }
    }
}

/*
type palm/common/mutate.(_httpResult) struct {
  Fields (available fields): 
      Url: string  
      Request: *http.Request  
      Error: error  
      RequestRaw: []uint8  
      ResponseRaw: []uint8  
      Response: *http.Response  
      DurationMs: int64  
      Timestamp: int64  
      Payloads: []string  
  StructMethods (structural method/function): 
  PtrStructMethods (pointer structure method/function): 
}
*/


`)

var OrdinaryPoCTemplate, _ = template.New("OrdinaryPoCTemplate").Parse(`
isHttps = cli.Have("https", cli.setDefault({{ .https }}))
target = cli.String("target", cli.setDefault("{{ .target }}"))

packet = ` + "`" + `
{{ .packetTemplate }}` + "`" + `

debug = cli.Have("debug", cli.setDefault(true))

if debug {
    loglevel("debug")
}

sendPacket = func(target) {
    return poc.HTTP(
        packet, 
        poc.timeout(10),
        # poc.proxy("http://127.0.0.1:8083"),
        # poc.proxy("http://127.0.0.1:7890"),
        poc.redirectTimes(3), # Number of redirects
        poc.https(isHttps),
        poc.params({
            "target": target,
        },
    ))
}

if YAK_MAIN {
    rspBytes, reqBytes, err = sendPacket(target)
    if err != nil {
        log.error("send packet error: %s", err)
        return
    }

    if debug {
        println(string(reqBytes))
        println("---------------------------------")
        println(string(rspBytes))
    }

    riskTarget = target
    if str.MatchAllOfRegexp(rspBytes, ` + "`" + `(?i)foundtextinRsp!` + "`" + `) || str.MatchAllOfSubString(rspBytes, "FoundTextInResponse") {
        urlIns, _ = str.ExtractURLFromHTTPRequestRaw(reqBytes, isHttps)
        if urlIns == nil {
            riskTarget = urlIns.String()
        }
        yakit.Info("Matched for %v", target)
        # Save to RiskTable
        risk.NewRisk(
            riskTarget, risk.severity("high"), risk.type("poc"),
            risk.title("English Title"),            ## English Title for Risk
            risk.titleVerbose("Chinese title"), ## Chinese Title
            risk.details({
                "target": riskTarget,
                "request": reqBytes,
                "response": rspBytes,
            }),
        )
    }
}
















`)

func (s *Server) GenerateCSRFPocByPacket(ctx context.Context, req *ypb.GenerateCSRFPocByPacketRequest) (*ypb.GenerateCSRFPocByPacketResponse, error) {
	poc, err := yaklib.GenerateCSRFPoc(req.GetRequest(), yaklib.CsrfOptWithHTTPS(req.IsHttps))
	if err != nil {
		return nil, err
	}
	return &ypb.GenerateCSRFPocByPacketResponse{Code: []byte(poc)}, nil
}

func (s *Server) GenerateYakCodeByPacket(ctx context.Context, req *ypb.GenerateYakCodeByPacketRequest) (*ypb.GenerateYakCodeByPacketResponse, error) {
	multipartReq := lowhttp.IsMultipartFormDataRequest(req.GetRequest())
	if multipartReq {
		// processes the uploaded data packet
		return nil, utils.Errorf("multipart/form-data; need generate specially!")
	}

	switch req.GetCodeTemplate() {
	case ypb.GenerateYakCodeByPacketRequest_Ordinary:
		var buf bytes.Buffer
		err := OrdinaryPoCTemplate.Execute(&buf, extractPacketToGenerateParams(req.GetIsHttps(), req.GetRequest()))
		if err != nil {
			return nil, utils.Errorf("generate yak code[ordinary] failed: %s", err)
		}
		return &ypb.GenerateYakCodeByPacketResponse{Code: buf.Bytes()}, nil
	case ypb.GenerateYakCodeByPacketRequest_Batch:
		var buf bytes.Buffer
		err := BatchPoCTemplate.Execute(&buf, extractPacketToGenerateParams(req.GetIsHttps(), req.GetRequest()))
		if err != nil {
			return nil, utils.Errorf("generate yak code[ordinary] failed: %s", err)
		}
		return &ypb.GenerateYakCodeByPacketResponse{Code: buf.Bytes()}, nil
	default:
		var buf bytes.Buffer
		err := OrdinaryPoCTemplate.Execute(&buf, extractPacketToGenerateParams(req.GetIsHttps(), req.GetRequest()))
		if err != nil {
			return nil, utils.Errorf("generate yak code[ordinary] failed: %s", err)
		}
		return &ypb.GenerateYakCodeByPacketResponse{Code: buf.Bytes()}, nil
	}
}
