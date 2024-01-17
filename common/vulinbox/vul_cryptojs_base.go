package vulinbox

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/yaklang/yaklang/common/log"
	"github.com/yaklang/yaklang/common/mutate"
	"github.com/yaklang/yaklang/common/utils"
	"github.com/yaklang/yaklang/common/utils/lowhttp"
	"github.com/yaklang/yaklang/common/utils/tlsutils"
	"github.com/yaklang/yaklang/common/yak/yaklib/codec"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"unicode/utf8"
)

//go:embed html/vul_cryptojs_basic.html
var cryptoBasicHtml []byte

//go:embed html/vul_cryptojs_rsa.html
var cryptoRsaHtml []byte

//go:embed html/vul_cryptojs_rsa_keyfromserver.html
var cryptoRsaKeyFromServerHtml []byte

//go:embed html/vul_cryptojs_rsa_keyfromserver_withresponse.html
var cryptoRsaKeyFromServerHtmlWithResponse []byte

//go:embed html/vul_cryptojs_rsa_and_aes.html
var cryptoRsaKeyAndAesHtml []byte

//go:embed html/vul_cryptojslib_template.html
var cryptoJSlibTemplateHtml string

func (s *VulinServer) registerCryptoJS() {
	r := s.router

	var (
		backupPass = []string{"admin", "123456", "admin123", "88888888", "666666"}
		pri, pub   []byte
		username   = "admin"
		password   = backupPass[rand.Intn(len(backupPass))]
	)

	log.Infof("frontend end crypto js user:pass = %v:%v", username, password)
	var isLogined = func(loginUser, loginPass string) bool {
		return loginUser == username && loginPass == password
	}
	var isLoginedFromRaw = func(i any) bool {
		var params = make(map[string]any)
		switch i.(type) {
		case string:
			params = utils.ParseStringToGeneralMap(i)
		case []byte:
			params = utils.ParseStringToGeneralMap(i)
		default:
			params = utils.InterfaceToGeneralMap(i)
		}
		username := utils.MapGetString(params, "username")
		password := utils.MapGetString(params, "password")
		return isLogined(username, password)
	}
	var renderLoginSuccess = func(writer http.ResponseWriter, loginUsername, loginPassword string, fallback []byte, success ...[]byte) {
		if loginUsername != username || loginPassword != password {
			writer.WriteHeader(403)
			writer.Write(fallback)
			return
		}

		if len(success) > 0 {
			writer.Write(success[0])
			return
		}

		writer.Write([]byte(`<!doctype html>
<html>
<head>
    <title>Example DEMO</title>

    <meta charset="utf-8" />
    <meta http-equiv="Content-type" content="text/html; charset=utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    <style type="text/css">
    body {
        background-color: #f0f0f2;
        margin: 0;
        padding: 0;
        font-family: -apple-system, system-ui, BlinkMacSystemFont, "Segoe UI", "Open Sans", "Helvetica Neue", Helvetica, Arial, sans-serif;
        
    }
    div {
        width: 600px;
        margin: 5em auto;
        padding: 2em;
        background-color: #fdfdff;
        border-radius: 0.5em;
        box-shadow: 2px 3px 7px 2px rgba(0,0,0,0.02);
    }
    </style>    
</head>

<body>
<div>
	<p class="success-container">
        <h1>Congratulations! login successful!</h1>
        <p>Welcome, you have logged in successfully.</p>
    </p>
</div>
</body>
</html>`))
	}

	var initKey = func() {
		log.Infof("start to GeneratePrivateAndPublicKeyPEMWithPrivateFormatter")
		pri, pub, _ = tlsutils.GeneratePrivateAndPublicKeyPEMWithPrivateFormatter("pkcs8")
	}
	var onceGenerator = sync.Once{}

	r.Use(func(handler http.Handler) http.Handler {
		onceGenerator.Do(initKey)
		return handler
	})
	cryptoGroup := r.PathPrefix("/crypto").Name("Advanced front-end encryption and decryption and signature verification.").Subrouter()
	cryptoRoutes := []*VulInfo{
		{
			Path:  "/sign/hmac/sha256",
			Title: ": HMAC-SHA256",
			Handler: func(writer http.ResponseWriter, request *http.Request) {
				unsafeTemplateRender(writer, request, cryptoJSlibTemplateHtml, map[string]any{
					"url":         `/crypto/sign/hmac/sha256/verify`,
					`extrakv`:     "username: jsonData.username, password: jsonData.password,",
					"title":       "HMAC-sha256 Signature verification",
					"datafield":   "signature",
					"key":         `CryptoJS.enc.Utf8.parse("1234123412341234")`,
					"info":        "Signature verification (also called signature verification or signature) ) is a common security method to verify whether request parameters have been tampered with. There are two mainstream methods for verifying signatures. One is KEY+hash algorithm, such as HMAC-MD5 / HMAC-SHA256, etc., this case is a typical case of this method. The rules for generating signatures are: username=*&password=*. The submitted data needs to be processed separately during submission and verification before the signature can be used and verified.",
					"encrypt":     `CryptoJS.HmacSHA256(word, key.toString(CryptoJS.enc.Utf8)).toString();`,
					"decrypt":     `"";`,
					"jsonhandler": "`username=${jsonData.username}&password=${jsonData.password}`;",
				})
			},
		},
		{
			Path: "/sign/hmac/sha256/verify",
			Handler: func(writer http.ResponseWriter, request *http.Request) {
				raw, err := utils.HttpDumpWithBody(request, true)
				if err != nil {
					Failed(writer, request, "dump request failed: %v", err)
					return
				}
				params := utils.ParseStringToGeneralMap(lowhttp.GetHTTPPacketBody(raw))
				keyEncoded := utils.MapGetString(params, "key")
				keyPlain, err := codec.DecodeHex(keyEncoded)
				if err != nil {
					Failed(writer, request, "key decode hex failed: %v", err)
					return
				}
				_ = keyPlain
				originSign := utils.MapGetString(params, "signature")
				siginatureHmacHex := originSign
				//siginatureHmacHex, err := tlsutils.PemPkcs1v15Decrypt(pri, []byte(originSignDecoded))
				//if err != nil {
				//	Failed(writer, request, "signature decrypt failed: %v", err)
				//	return
				//}
				username := utils.MapGetString(params, "username")
				password := utils.MapGetString(params, "password")
				backendCalcOrigin := fmt.Sprintf("username=%v&password=%v", username, password)
				dataRaw := codec.HmacSha256(keyPlain, backendCalcOrigin)
				var blocks []string
				var signFinished = string(siginatureHmacHex) == codec.EncodeToHex(dataRaw)
				msg := "ORIGIN -> " + originSign + "\n DECODED HMAC: " + string(siginatureHmacHex) + "\n" +
					"\n Expect: " + codec.EncodeToHex(dataRaw) +
					"\n Key: " + string(keyPlain) +
					"\n OriginData: " + backendCalcOrigin
				if signFinished {
					blocks = append(blocks, block("Signature verification is successful.", msg))
				} else {
					blocks = append(blocks, block("Signature verification failed", msg))
				}
				if isLoginedFromRaw(params) && signFinished {
					blocks = append(blocks, block("Username and password verification successful", "Congratulations, your login is successful!"))
				} else {
					blocks = append(blocks, block("Username and password verification failed", "origin data: "+backendCalcOrigin))
				}

				DefaultRender(BlockContent(blocks...), writer, request)
			},
		},
		{
			Path:  "/sign/rsa/hmacsha256",
			Title: "Front-end verification signature (signature verification) form: HMAC-SHA256 first and then RSA",
			Handler: func(writer http.ResponseWriter, request *http.Request) {
				unsafeTemplateRender(writer, request, cryptoJSlibTemplateHtml, map[string]any{
					"url":     `/crypto/sign/rsa/hmacsha256/verify`,
					`extrakv`: "username: jsonData.username, password: jsonData.password,",
					"title":   "First HMAC-SHA256 and then RSA",
					"initcode": `
document.getElementById('submit').disabled = true;
document.getElementById('submit').innerText = 'Need to wait for the public key to be obtained.';
let pubkey;
setTimeout(function(){
	fetch('/crypto/js/rsa/public/key').then(async function(rsp) {
		pubkey = await rsp.text()
		document.getElementById('submit').disabled = false;
		document.getElementById('submit').innerText = 'Submit form data';
		console.info(pubkey)
	})
},300)`,
					"datafield":   "signature",
					"key":         `CryptoJS.enc.Utf8.parse("1234123412341234")`,
					"info":        "Signature verification (also called signature verification or signature) ) is a common security method to verify whether request parameters have been tampered with. There are two mainstream methods for verifying signatures. One is KEY+hash algorithm, such as HMAC-MD5 / HMAC-SHA256, etc. The other is to use asymmetric encryption to encrypt the HMAC signature information. This case is a typical case of this method. The rules for generating signatures are: username=*&password=*. During submission and verification, the submitted data needs to be processed separately before the signature can be used and verified. This situation is relatively complicated.",
					"encrypt":     `KEYUTIL.getKey(pubkey).encrypt(CryptoJS.HmacSHA256(word, key.toString(CryptoJS.enc.Utf8)).toString());`,
					"decrypt":     `"";`,
					"jsonhandler": "`username=${jsonData.username}&password=${jsonData.password}`;",
				})
			},
		},
		{
			Path: "/sign/rsa/hmacsha256/verify",
			Handler: func(writer http.ResponseWriter, request *http.Request) {
				raw, err := utils.HttpDumpWithBody(request, true)
				if err != nil {
					Failed(writer, request, "dump request failed: %v", err)
					return
				}
				params := utils.ParseStringToGeneralMap(lowhttp.GetHTTPPacketBody(raw))
				keyEncoded := utils.MapGetString(params, "key")
				keyPlain, err := codec.DecodeHex(keyEncoded)
				if err != nil {
					Failed(writer, request, "key decode hex failed: %v", err)
					return
				}
				_ = keyPlain
				originSign := utils.MapGetString(params, "signature")
				originSignDecoded, err := codec.DecodeHex(originSign)
				if err != nil {
					Failed(writer, request, "originSign decode hex failed: %v", err)
					return
				}
				siginatureHmacHex, err := tlsutils.PemPkcs1v15Decrypt(pri, []byte(originSignDecoded))
				if err != nil {
					Failed(writer, request, "signature decrypt failed: %v", err)
					return
				}
				username := utils.MapGetString(params, "username")
				password := utils.MapGetString(params, "password")
				backendCalcOrigin := fmt.Sprintf("username=%v&password=%v", username, password)
				dataRaw := codec.HmacSha256(keyPlain, backendCalcOrigin)
				var blocks []string
				var signFinished = string(siginatureHmacHex) == codec.EncodeToHex(dataRaw)
				msg := "RSA -> " + originSign + "\n DECODED HMAC: " + string(siginatureHmacHex) + "\n" +
					"\n Expect: " + codec.EncodeToHex(dataRaw) +
					"\n Key: " + string(keyPlain) +
					"\n OriginData: " + backendCalcOrigin
				if signFinished {
					blocks = append(blocks, block("Signature verification is successful.", msg))
				} else {
					blocks = append(blocks, block("Signature verification failed", msg))
				}
				if isLoginedFromRaw(params) && signFinished {
					blocks = append(blocks, block("Username and password verification successful", "Congratulations, your login is successful!"))
				} else {
					blocks = append(blocks, block("Username and password verification failed", "origin data: "+backendCalcOrigin))
				}

				DefaultRender(BlockContent(blocks...), writer, request)
			},
		},
		{
			Path:  "/js/lib/aes/cbc",
			Title: "CryptoJS.AES (CBC) Front-end encrypted login form",
			Handler: func(writer http.ResponseWriter, request *http.Request) {
				unsafeTemplateRender(writer, request, cryptoJSlibTemplateHtml, map[string]any{
					"url":      `/crypto/js/lib/aes/cbc/handler`,
					"initcode": "var iv = CryptoJS.lib.WordArray.random(128/8);",
					`extrakv`:  "iv: iv.toString(),",
					"title":    "AES-CBC (4.0.0 default) encryption",
					"key":      `CryptoJS.enc.Utf8.parse("1234123412341234")`,
					"info":     "CryptoJS.AES (CBC requires IV).encrypt is used by default/decrypt, default PKCS7Padding, key length If it is less than 16 bytes, add it with NULL. If it exceeds 16 bytes, it will be truncated.\n Note: This encryption method may have different ciphertext every time.",
					"encrypt":  `CryptoJS.AES.encrypt(word, key, {iv: iv}).toString();`,
					"decrypt":  `CryptoJS.AES.decrypt(word, key, {iv: iv}).toString();`,
				})
			},
		},
		{
			Path: "/js/lib/aes/cbc/handler",
			Handler: func(writer http.ResponseWriter, request *http.Request) {
				raw, err := utils.HttpDumpWithBody(request, true)
				if err != nil {
					Failed(writer, request, "dump request failed: %v", err)
					return
				}
				params := utils.ParseStringToGeneralMap(lowhttp.GetHTTPPacketBody(raw))
				keyEncoded := utils.MapGetString(params, "key")
				keyPlain, err := codec.DecodeHex(keyEncoded)
				if err != nil {
					Failed(writer, request, "key decode hex failed: %v", err)
					return
				}
				dataBase64D := utils.MapGetString(params, "data")
				ivHex := utils.MapGetString(params, "iv")
				dataRaw, err := codec.DecodeBase64(dataBase64D)
				if err != nil {
					Failed(writer, request, "decode base64 failed: %v", err)
					return
				}
				ivRaw, err := codec.DecodeHex(ivHex)
				if err != nil {
					Failed(writer, request, "iv decode hex failed: %v", err)
					return
				}
				dec, err := codec.AESCBCDecrypt([]byte(keyPlain), dataRaw, ivRaw)
				if err != nil {
					Failed(writer, request, "decrypt failed: %v", err)
					return
				}
				var blocks []string
				blocks = append(blocks, block("decryption front end Content successful", string(dec)))
				if isLoginedFromRaw(dec) {
					blocks = append(blocks, block("Username and password verification successful", "Congratulations, your login is successful!"))
				} else {
					blocks = append(blocks, block("Username and password verification failed", "origin data: "+string(dataBase64D)))
				}

				DefaultRender(BlockContent(blocks...), writer, request)
			},
		},
		{
			Path:  "/js/lib/aes/ecb",
			Title: "CryptoJS.AES(ECB) front-end encrypted login form",
			Handler: func(writer http.ResponseWriter, request *http.Request) {
				unsafeTemplateRender(writer, request, cryptoJSlibTemplateHtml, map[string]any{
					"url":      `/crypto/js/lib/aes/ecb/handler`,
					"initcode": "// ignore:  var iv = CryptoJS.lib.WordArray.random(128/8);",
					`extrakv`:  "// iv: iv.toString(),",
					"title":    "AES(ECB PKCS7) encryption",
					"key":      `CryptoJS.enc.Utf8.parse("1234123412341234")`,
					"info":     "CryptoJS.AES(ECB).encrypt/decrypt, default PKCS7Padding, the key length is less than 16 bytes, supplemented with NULL, if it exceeds 16 bytes, truncated",
					"encrypt":  `CryptoJS.AES.encrypt(word, key, {mode: CryptoJS.mode.ECB}).toString();`,
					"decrypt":  `CryptoJS.AES.decrypt(word, key, {mode: CryptoJS.mode.ECB}).toString();`,
				})
			},
		},
		{
			Path: "/js/lib/aes/ecb/handler",
			Handler: func(writer http.ResponseWriter, request *http.Request) {
				raw, err := utils.HttpDumpWithBody(request, true)
				if err != nil {
					Failed(writer, request, "dump request failed: %v", err)
					return
				}
				params := utils.ParseStringToGeneralMap(lowhttp.GetHTTPPacketBody(raw))
				keyEncoded := utils.MapGetString(params, "key")
				keyPlain, err := codec.DecodeHex(keyEncoded)
				if err != nil {
					Failed(writer, request, "key decode hex failed: %v", err)
					return
				}
				dataBase64D := utils.MapGetString(params, "data")
				dataRaw, err := codec.DecodeBase64(dataBase64D)
				if err != nil {
					Failed(writer, request, "decode base64 failed: %v", err)
					return
				}
				dec, err := codec.AESECBDecrypt([]byte(keyPlain), dataRaw, nil)
				if err != nil {
					Failed(writer, request, "decrypt failed: %v", err)
					return
				}
				var blocks []string
				blocks = append(blocks, block("decryption front end Content successful", string(dec)))
				if isLoginedFromRaw(dec) {
					blocks = append(blocks, block("Username and password verification successful", "Congratulations, your login is successful!"))
				} else {
					blocks = append(blocks, block("Username and password verification failed", "origin data: "+string(dataBase64D)))
				}

				DefaultRender(BlockContent(blocks...), writer, request)
			},
		},
		{
			Path:  "/js/basic",
			Title: "AES-ECB encryption form (with password)",
			Handler: func(writer http.ResponseWriter, request *http.Request) {
				var params = make(map[string]interface{})

				var data, _ = utils.HttpDumpWithBody(request, true)
				params["packet"] = string(data)

				if request.Method == "GET" {
					results, err := mutate.FuzzTagExec(cryptoBasicHtml, mutate.Fuzz_WithParams(params))
					if err != nil {
						writer.Write([]byte("<pre>" + string(data) + "<pre> <br/> <br/> <h2>error</h2> <br/>" + err.Error()))
						return
					}
					writer.Write([]byte(results[0]))
					return
				}

				if request.Method == "POST" {
					_, body := lowhttp.SplitHTTPHeadersAndBodyFromPacket(data)
					err := json.Unmarshal(body, &params)
					if err != nil {
						writer.Write([]byte("<pre>" + string(data) + "<pre> <br/> <br/> <h2>error</h2> <br/>" + err.Error()))
						return
					}

					key, _ := codec.DecodeHex(utils.MapGetString(params, "key"))
					iv, _ := codec.DecodeHex(utils.MapGetString(params, "iv"))
					encrypted := utils.MapGetString(params, "data")
					encryptedBase64Decoded, _ := codec.DecodeBase64(encrypted)

					var origin, decErr = codec.AESECBDecryptWithPKCS7Padding([]byte(key), []byte(encryptedBase64Decoded), []byte(iv))
					spew.Dump(origin, decErr)
					var handled string
					var raw, _ = json.MarshalIndent(map[string]any{
						"key":             utils.MapGetString(params, "key"),
						"key_hex_decoded": string(key),
						"iv":              utils.MapGetString(params, "iv"),
						"iv_hex_deocded":  string(iv),

						"encrypted":                encrypted,
						"encrypted_base64_decoded": strconv.Quote(string(encryptedBase64Decoded)),
						"decrypted":                string(origin),
						"decrypted_error":          fmt.Sprint(decErr),
					}, "", "    ")
					handled = string(raw)
					_ = handled

					if !utf8.Valid(origin) {
						origin = []byte(strconv.Quote(string(origin)))
					} else {
						if strings.HasPrefix(string(origin), `"`) && strings.HasSuffix(string(origin), `"`) {
							var after, _ = strconv.Unquote(string(origin))
							if after != "" {
								origin = []byte(after)
							}
						}
					}

					var i any
					json.Unmarshal(origin, &i)
					params := utils.InterfaceToGeneralMap(i)
					username := utils.MapGetString(params, "username")
					password := utils.MapGetString(params, "password")
					var blocks []string
					blocks = append(blocks, block("decryption front end Content successful", string(origin)))
					if isLogined(username, password) {
						blocks = append(blocks, block("Username and password verification successful", "Congratulations, your login is successful!"))
					} else {
						blocks = append(blocks, block("Username and password verification failed", "origin data: "+string(origin)))
					}
					DefaultRender(BlockContent(blocks...), writer, request)
					//renderLoginSuccess(writer, username, password, []byte(
					//	`<br>`+
					//		`<pre>`+string(data)+`</pre> <br><br><br>	`+
					//		`<pre>`+handled+`</pre> <br><br>	`+
					//		`<pre>`+string(origin)+`</pre> <br><br>	`+
					//		`<pre>`+fmt.Sprint(err)+`</pre> <br><br>	`,
					//))
					return
				}

				writer.WriteHeader(http.StatusMethodNotAllowed)
			},
			RiskDetected: true,
		},
		{
			Path:  "/js/rsa",
			Title: "RSA: Encrypted form, attached key",
			Handler: func(writer http.ResponseWriter, request *http.Request) {
				var params = make(map[string]interface{})

				var data, _ = utils.HttpDumpWithBody(request, true)
				params["packet"] = string(data)

				if request.Method == "GET" {
					results, err := mutate.FuzzTagExec(cryptoRsaHtml, mutate.Fuzz_WithParams(params))
					if err != nil {
						writer.Write([]byte("<pre>" + string(data) + "<pre> <br/> <br/> <h2>error</h2> <br/>" + err.Error()))
						return
					}
					writer.Write([]byte(results[0]))
					return
				}

				if request.Method == "POST" {
					_, body := lowhttp.SplitHTTPHeadersAndBodyFromPacket(data)
					err := json.Unmarshal(body, &params)
					if err != nil {
						writer.Write([]byte("<pre>" + string(data) + "<pre> <br/> <br/> <h2>error</h2> <br/>" + err.Error()))
						return
					}

					pubkey := utils.MapGetString(params, "publicKey")
					prikey := utils.MapGetString(params, "privateKey")
					_ = pubkey

					println(prikey)

					encrypted := utils.MapGetString(params, "data")
					encryptedBase64Decoded, _ := codec.DecodeBase64(encrypted)

					var origin, decErr = tlsutils.PemPkcsOAEPDecrypt([]byte(prikey), encryptedBase64Decoded)
					spew.Dump(origin, decErr)
					var handled string
					var raw, _ = json.MarshalIndent(map[string]any{
						"publicKey":                pubkey,
						"privateKey":               prikey,
						"encrypted":                encrypted,
						"encrypted_base64_decoded": strconv.Quote(string(encryptedBase64Decoded)),
						"decrypted":                string(origin),
						"decrypted_error":          fmt.Sprint(decErr),
					}, "", "    ")
					handled = string(raw)

					_ = handled
					if !utf8.Valid(origin) {
						origin = []byte(strconv.Quote(string(origin)))
					} else {
						if strings.HasPrefix(string(origin), `"`) && strings.HasSuffix(string(origin), `"`) {
							var after, _ = strconv.Unquote(string(origin))
							if after != "" {
								origin = []byte(after)
							}
						}
					}

					var i any
					json.Unmarshal(origin, &i)
					if i != nil {
						params = utils.InterfaceToGeneralMap(i)
					} else {
						params = utils.InterfaceToGeneralMap(origin)
					}

					var blocks []string
					blocks = append(blocks, block("decryption front end Content successful", string(origin)))
					if isLogined(username, password) {
						blocks = append(blocks, block("Username and password verification successful", "Congratulations, your login is successful!"))
					} else {
						blocks = append(blocks, block("Username and password verification failed", "origin data: "+string(origin)))
					}
					DefaultRender(BlockContent(blocks...), writer, request)
					return
				}

				writer.WriteHeader(http.StatusMethodNotAllowed)
			},
			RiskDetected: true,
		},
		{
			Path:  "/js/rsa/fromserver",
			Title: "RSA: Encrypt forms server transport key",
			Handler: func(writer http.ResponseWriter, request *http.Request) {
				var params = make(map[string]interface{})

				var data, _ = utils.HttpDumpWithBody(request, true)
				params["packet"] = string(data)

				if request.Method == "GET" {
					results, err := mutate.FuzzTagExec(cryptoRsaKeyFromServerHtml, mutate.Fuzz_WithParams(params))
					if err != nil {
						writer.Write([]byte("<pre>" + string(data) + "<pre> <br/> <br/> <h2>error</h2> <br/>" + err.Error()))
						return
					}
					writer.Write([]byte(results[0]))
					return
				}

				if request.Method == "POST" {
					_, body := lowhttp.SplitHTTPHeadersAndBodyFromPacket(data)
					err := json.Unmarshal(body, &params)
					if err != nil {
						writer.Write([]byte("<pre>" + string(data) + "<pre> <br/> <br/> <h2>error</h2> <br/>" + err.Error()))
						return
					}

					pubkey := pub
					prikey := pri
					_ = pubkey

					println(prikey)

					encrypted := utils.MapGetString(params, "data")
					encryptedBase64Decoded, _ := codec.DecodeBase64(encrypted)

					var origin, decErr = tlsutils.PemPkcsOAEPDecrypt([]byte(prikey), encryptedBase64Decoded)
					spew.Dump(origin, decErr)
					var handled string
					var raw, _ = json.MarshalIndent(map[string]any{
						"publicKey":                pubkey,
						"privateKey":               prikey,
						"encrypted":                encrypted,
						"encrypted_base64_decoded": strconv.Quote(string(encryptedBase64Decoded)),
						"decrypted":                string(origin),
						"decrypted_error":          fmt.Sprint(decErr),
					}, "", "    ")
					handled = string(raw)
					_ = handled

					if !utf8.Valid(origin) {
						origin = []byte(strconv.Quote(string(origin)))
					} else {
						if strings.HasPrefix(string(origin), `"`) && strings.HasSuffix(string(origin), `"`) {
							var after, _ = strconv.Unquote(string(origin))
							if after != "" {
								origin = []byte(after)
							}
						}
					}

					var i any
					json.Unmarshal(origin, &i)
					if i != nil {
						params = utils.InterfaceToGeneralMap(i)
					} else {
						params = utils.InterfaceToGeneralMap(origin)
					}
					username := utils.MapGetString(params, "username")
					password := utils.MapGetString(params, "password")
					var blocks []string
					blocks = append(blocks, block("decryption front end Content successful", string(origin)))
					if isLogined(username, password) {
						blocks = append(blocks, block("Username and password verification successful", "Congratulations, your login is successful!"))
					} else {
						blocks = append(blocks, block("Username and password verification failed", "origin data: "+string(origin)))
					}
					DefaultRender(BlockContent(blocks...), writer, request)
					return
				}

				writer.WriteHeader(http.StatusMethodNotAllowed)
			},
			RiskDetected: true,
		},
		{
			Path:  "/js/rsa/fromserver/response",
			Title: "RSA: Encrypt forms server transport key + response encryption",
			Handler: func(writer http.ResponseWriter, request *http.Request) {
				var params = make(map[string]interface{})

				var data, _ = utils.HttpDumpWithBody(request, true)
				params["packet"] = string(data)

				if request.Method == "GET" {
					results, err := mutate.FuzzTagExec(cryptoRsaKeyFromServerHtmlWithResponse, mutate.Fuzz_WithParams(params))
					if err != nil {
						writer.Write([]byte("<pre>" + string(data) + "<pre> <br/> <br/> <h2>error</h2> <br/>" + err.Error()))
						return
					}
					writer.Write([]byte(results[0]))
					return
				}

				if request.Method == "POST" {
					_, body := lowhttp.SplitHTTPHeadersAndBodyFromPacket(data)
					err := json.Unmarshal(body, &params)
					if err != nil {
						writer.Write([]byte("<pre>" + string(data) + "<pre> <br/> <br/> <h2>error</h2> <br/>" + err.Error()))
						return
					}

					pubkey := pub
					prikey := pri
					_ = pubkey

					println(prikey)

					encrypted := utils.MapGetString(params, "data")
					encryptedBase64Decoded, _ := codec.DecodeBase64(encrypted)

					var origin, decErr = tlsutils.PemPkcsOAEPDecrypt([]byte(prikey), encryptedBase64Decoded)
					spew.Dump(origin, decErr)
					var handled string
					var raw, _ = json.MarshalIndent(map[string]any{
						"publicKey":                pubkey,
						"privateKey":               prikey,
						"encrypted":                encrypted,
						"encrypted_base64_decoded": strconv.Quote(string(encryptedBase64Decoded)),
						"decrypted":                string(origin),
						"decrypted_error":          fmt.Sprint(decErr),
					}, "", "    ")
					handled = string(raw)

					if !utf8.Valid(origin) {
						origin = []byte(strconv.Quote(string(origin)))
					} else {
						if strings.HasPrefix(string(origin), `"`) && strings.HasSuffix(string(origin), `"`) {
							var after, _ = strconv.Unquote(string(origin))
							if after != "" {
								origin = []byte(after)
							}
						}
					}

					var rawResponseBody = `<br>` +
						`<pre>` + string(data) + `</pre> <br><br><br>	` +
						`<pre>` + handled + `</pre> <br><br>	` +
						`<pre>` + string(origin) + `</pre> <br><br>	` +
						`<pre>` + fmt.Sprint(err) + `</pre> <br><br>	`

					var i any
					json.Unmarshal(origin, &i)
					if i != nil {
						params = utils.InterfaceToGeneralMap(i)
					} else {
						params = utils.InterfaceToGeneralMap(origin)
					}
					username := utils.MapGetString(params, "username")
					password := utils.MapGetString(params, "password")
					var results = make(map[string]any)
					results["username"] = username
					results["success"] = isLogined(username, password)
					encryptedData, err := tlsutils.PemPkcsOAEPEncrypt(pub, utils.Jsonify(results))
					if err != nil {
						writer.Write([]byte(rawResponseBody + "<br/> <br/> <h2>error</h2> <br/>" + err.Error()))
						return
					}
					originData, err := tlsutils.PemPkcsOAEPDecrypt(pri, encryptedData)
					println("-------------------")
					println("-------------------")
					println("-------------------")
					spew.Dump(originData, err)
					println("-------------------")
					println("-------------------")
					println("-------------------")
					raw, _ = json.Marshal(map[string]any{
						"data":   codec.EncodeBase64(encryptedData),
						"origin": string(originData),
					})
					renderLoginSuccess(writer, username, password, raw, raw)
					return
				}

				writer.WriteHeader(http.StatusMethodNotAllowed)
			},
			RiskDetected: true,
		},
		{
			Path:  "/js/rsa/fromserver/response/aes-gcm",
			Title: "Frontend RSA encryption AES key, server transport",
			Handler: func(writer http.ResponseWriter, request *http.Request) {
				var params = make(map[string]interface{})

				var data, _ = utils.HttpDumpWithBody(request, true)
				params["packet"] = string(data)

				if request.Method == "GET" {
					results, err := mutate.FuzzTagExec(cryptoRsaKeyAndAesHtml, mutate.Fuzz_WithParams(params))
					if err != nil {
						writer.Write([]byte("<pre>" + string(data) + "<pre> <br/> <br/> <h2>error</h2> <br/>" + err.Error()))
						return
					}
					writer.Write([]byte(results[0]))
					return
				}

				if request.Method == "POST" {
					_, body := lowhttp.SplitHTTPHeadersAndBodyFromPacket(data)
					err := json.Unmarshal(body, &params)
					if err != nil {
						writer.Write([]byte("<pre>" + string(data) + "<pre> <br/> <br/> <h2>error</h2> <br/>" + err.Error()))
						return
					}

					pubkey := pub
					prikey := pri
					_ = pubkey

					println(prikey)

					encrypted := utils.MapGetString(params, "data")
					encryptedBase64Decoded, _ := codec.DecodeBase64(encrypted)

					encryptedKey := utils.MapGetString(params, "encryptedKey")
					encryptedIV := utils.MapGetString(params, "encryptedIV")

					encKeyDec, _ := codec.DecodeBase64(encryptedKey)
					encIVDec, _ := codec.DecodeBase64(encryptedIV)

					var originKey, decErr = tlsutils.PemPkcsOAEPDecrypt([]byte(prikey), encKeyDec)
					spew.Dump(originKey, decErr)
					originIV, decErr := tlsutils.PemPkcsOAEPDecrypt([]byte(prikey), encIVDec)
					spew.Dump(originIV, decErr)

					origin, err := codec.AESGCMDecryptWithNonceSize12(originKey, encryptedBase64Decoded, originIV)
					if err != nil {
						spew.Dump(originKey, originIV, encryptedBase64Decoded)
						log.Warnf("AES-GCM Decrypt failed nonce size 12: %v", err)
						writer.Write([]byte("<pre>" + string(data) + "<pre> <br/> <br/> <h2>error</h2> <br/>" + err.Error()))
						return
					}

					var handled string
					var raw, _ = json.MarshalIndent(map[string]any{
						"publicKey":    pubkey,
						"privateKey":   prikey,
						"aes-gcm":      true,
						"encryptedKey": encryptedKey,
						"encryptedIV":  encryptedIV,

						"encrypted":                encrypted,
						"encrypted_base64_decoded": strconv.Quote(string(encryptedBase64Decoded)),
						"decrypted":                string(origin),
						"decrypted_error":          fmt.Sprint(decErr),
					}, "", "    ")
					handled = string(raw)

					if !utf8.Valid(origin) {
						origin = []byte(strconv.Quote(string(origin)))
					} else {
						if strings.HasPrefix(string(origin), `"`) && strings.HasSuffix(string(origin), `"`) {
							var after, _ = strconv.Unquote(string(origin))
							if after != "" {
								origin = []byte(after)
							}
						}
					}

					var rawResponseBody = `<br>` +
						`<pre>` + string(data) + `</pre> <br><br><br>	` +
						`<pre>` + handled + `</pre> <br><br>	` +
						`<pre>` + string(origin) + `</pre> <br><br>	` +
						`<pre>` + fmt.Sprint(err) + `</pre> <br><br>	`

					var key, iv = utils.RandStringBytes(16), utils.RandStringBytes(12)
					encryptedKeyData, err := tlsutils.PemPkcsOAEPEncrypt(pub, key)
					if err != nil {
						log.Error(err)
						writer.Write([]byte(rawResponseBody + "<br/> <br/> <h2>error</h2> <br/>" + err.Error()))
						return
					}

					encryptedIVData, err := tlsutils.PemPkcsOAEPEncrypt(pub, iv)
					if err != nil {
						writer.Write([]byte(rawResponseBody + "<br/> <br/> <h2>error</h2> <br/>" + err.Error()))
						return
					}

					originData, err := tlsutils.PemPkcsOAEPDecrypt(pri, encryptedKeyData)
					println("-------------------")
					println("-------------------")
					println("-------------------")
					spew.Dump(originData, err)
					println("-------------------")
					println("-------------------")
					println("-------------------")

					var i any
					json.Unmarshal(origin, &i)
					if i != nil {
						params = utils.InterfaceToGeneralMap(i)
					} else {
						params = utils.InterfaceToGeneralMap(origin)
					}
					username := utils.MapGetString(params, "username")
					password := utils.MapGetString(params, "password")
					var results = make(map[string]any)
					results["username"] = username
					results["success"] = isLogined(username, password)
					_ = rawResponseBody
					aesEncrypted, err := codec.AESGCMEncryptWithNonceSize12([]byte(key), utils.Jsonify(results), []byte(iv))
					if err != nil {
						log.Errorf("AES-GCM Encrypt failed nonce size 12: %v", err)
						writer.Write([]byte(rawResponseBody + "<br/> <br/> <h2>error</h2> <br/>" + err.Error()))
						return
					}

					raw, _ = json.Marshal(map[string]any{
						"encryptedKey": codec.EncodeBase64(encryptedKeyData),
						"encryptedIV":  codec.EncodeBase64(encryptedIVData),
						"data":         codec.EncodeBase64(aesEncrypted),
					})
					writer.Write(raw)
					return
				}

				writer.WriteHeader(http.StatusMethodNotAllowed)
			},
			RiskDetected: true,
		},
		{
			Path: "/js/rsa/generator",
			Handler: func(writer http.ResponseWriter, request *http.Request) {
				writer.Write([]byte(`{"ok": true, "publicKey": ` + strconv.Quote(string(pub)) + `, "privateKey": ` + strconv.Quote(string(pri)) + `}`))
			},
		},
		{
			Path: "/js/rsa/public/key",
			Handler: func(writer http.ResponseWriter, request *http.Request) {
				writer.Write(pub)
			},
		},
	}
	for _, v := range cryptoRoutes {
		addRouteWithVulInfo(cryptoGroup, v)
	}

	testHandler := func(writer http.ResponseWriter, request *http.Request) {
		DefaultRender("<h1>This test page will test the front-end verification signature (verification) form for /static/js/cryptojs_4.0.0/*.js file import</h1>\n"+`

<p id='error'>CryptoJS Status</p>	
<p id='404'></p>	

<script>
window.onerror = function(message, source, lineno, colno, error) {
    console.log('An error has occurred: ', message);
	document.getElementById('error').innerHTML = "CryptoJS ERROR: " + message;
    return true;
};

handle404 = function(event) {
	const p = document.createElement("p")
	p.innerText = "CryptoJS 404: " + event.target.src
	document.getElementById('404').appendChild(p)
}
</script>
<script src="/static/js/cryptojs_4.0.0/core.min.js" onerror="handle404(event)"></script>
<script src="/static/js/cryptojs_4.0.0/enc-base64.min.js" onerror="handle404(event)"></script>
<script src="/static/js/cryptojs_4.0.0/md5.min.js" onerror="handle404(event)"></script>
<script src="/static/js/cryptojs_4.0.0/evpkdf.min.js" onerror="handle404(event)"></script>
<script src="/static/js/cryptojs_4.0.0/cipher-core.min.js" onerror="handle404(event)"></script>
<script src="/static/js/cryptojs_4.0.0/aes.min.js" onerror="handle404(event)"></script>
<script src="/static/js/cryptojs_4.0.0/pad-pkcs7.min.js" onerror="handle404(event)"></script>
<script src="/static/js/cryptojs_4.0.0/mode-ecb.min.js" onerror="handle404(event)"></script>
<script src="/static/js/cryptojs_4.0.0/enc-utf8.min.js" onerror="handle404(event)"></script>
<script src="/static/js/cryptojs_4.0.0/enc-hex.min.js" onerror="handle404(event)"></script>

You can observe the Console 
`, writer, request)
	}
	cryptoGroup.HandleFunc("/_test/", testHandler)
	cryptoGroup.HandleFunc("/", testHandler)
}
