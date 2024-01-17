package wsm

import (
	"bytes"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
)

// SecretKey The result after converting the string into md5[0:16] that meets the encryption requirements of Ice Scorpion and Godzilla
func secretKey(pwd string) []byte {
	return []byte(pass2MD5(pwd))
}

// Obtain the first sixteen digits of md5 value
func pass2MD5(input string) string {
	md5hash := md5.New()
	md5hash.Write([]byte(input))
	return hex.EncodeToString(md5hash.Sum(nil))[0:16]
}

func decodeBase64Values(i interface{}) (interface{}, error) {
	switch v := i.(type) {
	case string:
		decoded, err := base64.StdEncoding.DecodeString(v)
		if err != nil {
			return nil, fmt.Errorf("failed to decode base64 value: %w", err)
		}
		// Fix the error json returned by Ice Scorpion
		decoded = bytes.Replace(decoded, []byte(",]"), []byte("]"), 1)
		var j interface{}
		err = json.Unmarshal(decoded, &j)
		if err != nil {
			return string(decoded), nil
		}
		return decodeBase64Values(j)
	case []interface{}:
		for i, e := range v {
			decoded, err := decodeBase64Values(e)
			if err != nil {
				return nil, err
			}
			v[i] = decoded
		}
	case map[string]interface{}:
		for k, e := range v {
			decoded, err := decodeBase64Values(e)
			if err != nil {
				return nil, err
			}
			v[k] = decoded
		}
	}
	return i, nil
}
