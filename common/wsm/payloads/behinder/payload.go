package behinder

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"github.com/yaklang/yaklang/common/javaclassparser"
	"github.com/yaklang/yaklang/common/utils"
	"github.com/yaklang/yaklang/common/wsm/payloads"
	"regexp"
	"strconv"
	"strings"
)

func GetRawClass(binPayload string, params map[string]string) ([]byte, error) {
	b, err := hex.DecodeString(binPayload)
	if err != nil {
		return nil, err
	}
	clsObj, err := javaclassparser.Parse(b)
	if err != nil {
		return nil, err
	}
	for k, v := range params {
		fields := clsObj.FindConstStringFromPool(k)
		fields.Value = v
	}
	// Randomly replace the class name. The original class name is like this. net/behinder/payload/java/xxx
	err = clsObj.SetClassName(payloads.RandomClassName())
	if err != nil {
		return nil, err
	}
	// Randomly replace the file name.
	err = clsObj.SetSourceFileName(utils.RandNumberStringBytes(6))
	if err != nil {
		return nil, err
	}
	// Modified to Jdk 1.5. The original version of Ice Scorpion is 50 (1.6). After a few tests, it was found that 49 (1.5) is also OK.
	clsObj.MajorVersion = 49
	return clsObj.Bytes(), nil
}

func GetRawPHP(binPayload string, params map[string]string) ([]byte, error) {
	payloadBytes, err := hex.DecodeString(binPayload)
	if err != nil {
		return nil, err
	}
	code := strings.Replace(string(payloadBytes), "<?", "", 1)
	if v, ok := params["customEncoderFromText"]; ok {
		code += v + "\r\n"
	}
	paramsList := getPhpParams(payloadBytes)
	for i, paraName := range paramsList {
		paraValue := ""
		if v, ok := params[paraName]; ok {
			paraValue = base64.StdEncoding.EncodeToString([]byte(v))
			code += fmt.Sprintf("$%s=\"%s\";$%s=base64_decode($%s);", paraName, paraValue, paraName, paraName)
		} else {
			code += fmt.Sprintf("$%s=\"%s\";", paraName, "")
		}
		paramsList[i] = "$" + paraName
	}
	code += "\r\nmain(" + strings.Trim(strings.Join(paramsList, ","), ",") + ");"
	return []byte(code), nil
}

// Get the params that need to be changed in the php code.
func getPhpParams(phpPayload []byte) []string {
	paramList := make([]string, 0, 2)
	mainRegex := regexp.MustCompile(`main\s*\([^)]*\)`)
	mainMatch := mainRegex.Match(phpPayload)
	mainStr := mainRegex.FindStringSubmatch(string(phpPayload))

	if mainMatch && len(mainStr) > 0 {
		paramRegex := regexp.MustCompile(`\$([a-zA-Z]*)`)
		//paramMatch := paramRegex.FindStringSubmatch(mainStr[0])
		paramMatch := paramRegex.FindAllStringSubmatch(mainStr[0], -1)
		if len(paramMatch) > 0 {
			for _, v := range paramMatch {
				paramList = append(paramList, v[1])
			}
		}
	}

	return paramList
}

func GetRawAssembly(binPayload string, params map[string]string) ([]byte, error) {
	payloadBytes, err := hex.DecodeString(binPayload)
	if err != nil {
		return nil, err
	}

	if len(params) == 0 {
		return payloadBytes, nil
	}

	var paramTokens []string
	for key, value := range params {
		value = base64.StdEncoding.EncodeToString([]byte(value))
		paramTokens = append(paramTokens, key+":"+value)
	}

	paramsStr := strings.Join(paramTokens, ",")
	token := "~~~~~~" + paramsStr

	return append(payloadBytes, []byte(token)...), nil
}

func GetRawASP(binPayload string, params map[string]string) ([]byte, error) {
	payloadBytes, err := hex.DecodeString(binPayload)
	if err != nil {
		return nil, err
	}

	if v, ok := params["customEncoderFromText"]; ok {
		payloadBytes = bytes.Replace(payloadBytes, []byte("__Encrypt__"), []byte(v), 1)
		delete(params, "customEncoderFromText")
	}
	var code strings.Builder
	code.WriteString(string(payloadBytes))
	paraList := ""
	if len(params) > 0 {
		paraList = paraList + "Array("
		for _, paramValue := range params {
			var paraValueEncoded string
			for _, v := range paramValue {
				paraValueEncoded = paraValueEncoded + "chrw(" + strconv.Itoa(int(v)) + ")&"
			}
			paraValueEncoded = strings.TrimRight(paraValueEncoded, "&")
			paraList = paraList + "," + paraValueEncoded
		}
		paraList = paraList + ")"
	}
	paraList = strings.Replace(paraList, ",", "", 1)
	code.WriteString("\r\nmain " + paraList + "")
	return []byte(code.String()), nil
}
