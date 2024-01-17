package vulinbox

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/yaklang/yaklang/common/netx"
	utils2 "github.com/yaklang/yaklang/common/utils"
	"github.com/yaklang/yaklang/common/yak/yaklib/codec"
	"io"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"
)

//go:embed html/vul_fastjson.html
var fastjson_loginPage []byte

type JsonParser func(data string) (map[string]any, error)

var DnsRecord = sync.Map{}

func generateFastjsonParser(version string) JsonParser {
	if version == "intranet" { // Intranet version, the dnslog in the payload cannot be successfully parsed and the parsing timeout
		return func(data string) (map[string]any, error) {
			return fastjsonParser(data, "qqqqqqqqqq.qqqqqqqqqq.qqqqqqqqqq.qqqqqqqqqq")
		}
	}
	return func(data string) (map[string]any, error) {
		return fastjsonParser(data)
	}
}

// Here simulates the parsing process of fastjson
func fastjsonParser(data string, forceDnslog ...string) (map[string]any, error) {
	// redos
	if strings.Contains(data, "regex") {
		time.Sleep(5 * time.Second)
	}
	var dnslog string
	// Search dnslog
	re, err := regexp.Compile(`(\w+\.)+((dnslog\.cn)|(ceye\.io)|(vcap\.me)|(vcap\.io)|(xip\.io)|(burpcollaborator\.net)|(dgrh3\.cn))`)
	if err != nil {
		return nil, err
	}
	res := re.FindAllStringSubmatch(data, -1)
	if len(res) > 0 {
		dnslog = res[0][0]
		if len(forceDnslog) > 0 {
			dnslog = forceDnslog[0]
		}
	}
	if dnslog != "" {
		ip := netx.LookupFirst(dnslog, netx.WithTimeout(5*time.Second), netx.WithDNSNoCache(true))
		if ip != "" {
			DnsRecord.Store(dnslog, ip)
		}
	}
	var js map[string]any
	err = json.Unmarshal([]byte(data), &js)
	if err != nil {
		return nil, err
	}
	return js, nil
}
func mockController(parser JsonParser, req *http.Request, data string) string {
	newErrorResponse := func(err error) string {
		response, _ := json.Marshal(map[string]any{
			"code": -1,
			"err":  err.Error(),
		})
		return string(response)
	}
	results := codec.AutoDecode(data)
	if len(results) == 0 {
		return newErrorResponse(errors.New("decode error"))
	}
	data = results[len(results)-1].Result
	_, err := parser(data)
	if err != nil {
		return newErrorResponse(err)
	}

	err = errors.New("user or password error")
	return newErrorResponse(err)
}
func jacksonParser(data string) (map[string]any, error) {
	var js map[string]any
	err := json.Unmarshal([]byte(data), &js)
	if err != nil {
		return nil, err
	}
	return js, nil
}
func mockJacksonController(req *http.Request, data string) string {
	newErrorResponse := func(err error) string {
		response, _ := json.Marshal(map[string]any{
			"code": -1,
			"err":  err.Error(),
		})
		return string(response)
	}
	results := codec.AutoDecode(data)
	if len(results) == 0 {
		return newErrorResponse(errors.New("decode error"))
	}
	data = results[len(results)-1].Result
	js, _ := jacksonParser(data)
	unrecognizedFields := map[string]struct{}{}
	allowFields := map[string]struct{}{}
	if req.URL.Path == "/fastjson/json-in-cookie" {
		for k, _ := range js {
			if k != "id" {
				unrecognizedFields[k] = struct{}{}
				allowFields["id"] = struct{}{}
			}
		}
	} else {
		for k, _ := range js {
			if k != "user" && k != "password" {
				unrecognizedFields[k] = struct{}{}
				allowFields["user"] = struct{}{}
				allowFields["password"] = struct{}{}
			}
		}
	}

	// Simulate Jacksons error
	if len(js) > 2 {
		unrecognizedStr := ""
		allowFieldsStr := ""
		for k, _ := range unrecognizedFields {
			unrecognizedStr += k + ","

		}
		for k, _ := range allowFields {
			allowFieldsStr += k + ","
		}
		response, _ := json.Marshal(map[string]any{
			"timestamp": time.Now().Format("2006-01-02T15:04:05.000+0000"),
			"status":    500,
			"error":     "Internal Server Error",
			"message":   fmt.Sprintf("Unrecognized field %s (class com.vulbox.User), not marked as ignorable (2 known properties: %s])\n at [Source: (String)%s] (through reference chain: com.vulbox.User[%s])", unrecognizedStr, allowFieldsStr, data, unrecognizedStr),
			"path":      req.URL.Path,
		})
		return string(response)
	}
	return "ok"
}
func (s *VulinServer) registerFastjson() {
	r := s.router
	var fastjsonGroup = r.PathPrefix("/fastjson").Name("Fastjson case").Subrouter()
	var vuls = []*VulInfo{
		{
			Title:        "GET parameter passing case",
			Path:         "/json-in-query",
			DefaultQuery: `auth={"user":"admin","password":"password"}`,
			Handler: func(writer http.ResponseWriter, request *http.Request) {
				if request.Method == http.MethodGet {
					action := request.URL.Query().Get("action")
					if action == "" {
						writer.Write([]byte(utils2.Format(string(fastjson_loginPage), map[string]string{
							"script": `function load(){
            name=$("#username").val();
            password=$("#password").val();
            auth = {"user":name,"password":password};
            $.ajax({
                type:"get",
                url:"/fastjson/json-in-query",
                data:{"auth":JSON.stringify(auth),"action":"login"},
                success: function (data ,textStatus, jqXHR)
                {
                    $("#response").text(JSON.stringify(data));
					console.log(data);
                },
                error:function (XMLHttpRequest, textStatus, errorThrown) {      
                    alert("Request error");
                },
            })
        }`,
						})))
						return
					}
					auth := request.URL.Query().Get("auth")
					if auth == "" {
						writer.Write([]byte("auth parameter cannot be empty"))
						return
					}
					response := mockController(generateFastjsonParser("1.2.43"), request, auth)
					writer.Write([]byte(response))
				} else {
					writer.WriteHeader(http.StatusMethodNotAllowed)
				}
			},
		},
		{
			Title: "POST Form parameter passing case",
			Path:  "/json-in-form",
			Handler: func(writer http.ResponseWriter, request *http.Request) {
				if request.Method == http.MethodPost {
					body := request.FormValue("auth")
					response := mockController(generateFastjsonParser("1.2.43"), request, body)
					writer.Write([]byte(response))
				} else {
					writer.Write([]byte(utils2.Format(string(fastjson_loginPage), map[string]string{
						"script": `function load(){
            name=$("#username").val();
            password=$("#password").val();
            auth = {"user":name,"password":password};
            $.ajax({
                type:"post",
                url:"/fastjson/json-in-form",
                data:{"auth":JSON.stringify(auth),"action":"login"},
				dataType: "json",
                success: function (data ,textStatus, jqXHR)
                {
                    $("#response").text(JSON.stringify(data));
					console.log(data);
                },
                error:function (XMLHttpRequest, textStatus, errorThrown) {      
                    alert("Request error");
                },
            })
        }`,
					})))
					return
				}
			},
		},
		{
			Title: "POST Body parameter passing case",
			Path:  "/json-in-body",
			Handler: func(writer http.ResponseWriter, request *http.Request) {
				if request.Method == http.MethodPost {
					body, err := io.ReadAll(request.Body)
					if err != nil {
						writer.WriteHeader(http.StatusBadRequest)
						writer.Write([]byte("Invalid request"))
						return
					}
					defer request.Body.Close()
					response := mockController(generateFastjsonParser("1.2.43"), request, string(body))
					writer.Write([]byte(response))
				} else {
					writer.Write([]byte(utils2.Format(string(fastjson_loginPage), map[string]string{
						"script": `function load(){
            name=$("#username").val();
            password=$("#password").val();
            auth = {"user":name,"password":password};
            $.ajax({
                type:"post",
                url:"/fastjson/json-in-body",
                data:JSON.stringify(auth),
				dataType: "json",
                success: function (data ,textStatus, jqXHR)
                {
                    $("#response").text(JSON.stringify(data));
					console.log(data);
                },
                error:function (XMLHttpRequest, textStatus, errorThrown) {      
                    alert("Request error");
                },
            })
        }`,
					})))
					return
				}
			},
		},
		{
			Title: "Cookie Parameter passing case case",
			Path:  "/json-in-cookie",
			Handler: func(writer http.ResponseWriter, request *http.Request) {
				if request.Method == http.MethodGet {
					action := request.URL.Query().Get("action")
					if action == "" {
						writer.Header().Set("Set-Cookie", `auth=`+codec.EncodeBase64Url(`{"id":"-1"}`)) // Fuzz Coookie is temporarily unavailable and can only decode, not encode
						writer.Write([]byte(utils2.Format(string(fastjson_loginPage), map[string]string{
							"script": `function load(){
            name=$("#username").val();
            password=$("#password").val();
            auth = {"user":name,"password":password};
            $.ajax({
                type:"get",
                url:"/fastjson/json-in-cookie",
                data:{"auth":JSON.stringify(auth),"action":"login"},
                success: function (data ,textStatus, jqXHR)
                {
                    $("#response").text(JSON.stringify(data));
					console.log(data);
                },
                error:function (XMLHttpRequest, textStatus, errorThrown) {      
                    alert("Request error");
                },
            })
        }`,
						})))
						return
					}
					cookie, err := request.Cookie("auth")
					if err != nil {
						writer.Write([]byte("auth parameter cannot be empty"))
						return
					}
					response := mockController(generateFastjsonParser("1.2.43"), request, cookie.Value)
					writer.Write([]byte(response))
				} else {
					writer.WriteHeader(http.StatusMethodNotAllowed)
				}
			},
		},
		{
			Title: "Authorization parameter passing case",
			Path:  "/json-in-authorization",
			Handler: func(writer http.ResponseWriter, request *http.Request) {
				if request.Method == http.MethodGet {
					action := request.URL.Query().Get("action")
					if action == "" {
						writer.Write([]byte(utils2.Format(string(fastjson_loginPage), map[string]string{
							"script": `function load(){
            name=$("#username").val();
            password=$("#password").val();
            auth = {"user":name,"password":password};
			authHeaderValue = btoa(JSON.stringify(auth));
            $.ajax({
                type:"get",
                url:"/fastjson/json-in-authorization?action=login",
                headers: {"Authorization": "Basic "+authHeaderValue},
				dataType: "json",
                success: function (data ,textStatus, jqXHR)
                {
                    $("#response").text(JSON.stringify(data));
					console.log(data);
                },
                error:function (XMLHttpRequest, textStatus, errorThrown) {      
                    alert("Request error");
                },
            })
        }`,
						})))
						return
					}
					auth := request.Header.Get("Authorization")
					if len(auth) < 6 {
						writer.Write([]byte("auth parameter cannot be empty"))
						return
					}
					response := mockController(generateFastjsonParser("1.2.43"), request, auth[6:])
					writer.Write([]byte(response))
				} else {
					writer.WriteHeader(http.StatusMethodNotAllowed)
				}
			},
		},
		{
			Title:        "GET parameter passing Jackson backend case",
			Path:         "/jackson-in-query",
			DefaultQuery: `auth={"user":"admin","password":"password"}`,
			Handler: func(writer http.ResponseWriter, request *http.Request) {
				if request.Method == http.MethodGet {
					action := request.URL.Query().Get("action")
					if action == "" {
						writer.Write([]byte(utils2.Format(string(fastjson_loginPage), map[string]string{
							"script": `function load(){
            name=$("#username").val();
            password=$("#password").val();
            auth = {"user":name,"password":password};
            $.ajax({
                type:"get",
                url:"/fastjson/json-in-query",
                data:{"auth":JSON.stringify(auth),"action":"login"},
                success: function (data ,textStatus, jqXHR)
                {
                    $("#response").text(JSON.stringify(data));
					console.log(data);
                },
                error:function (XMLHttpRequest, textStatus, errorThrown) {      
                    alert("Request error");
                },
            })
        }`,
						})))
						return
					}
					auth := request.URL.Query().Get("auth")
					if auth == "" {
						writer.Write([]byte("auth parameter cannot be empty"))
						return
					}
					response := mockJacksonController(request, auth)
					writer.Write([]byte(response))
				} else {
					writer.WriteHeader(http.StatusMethodNotAllowed)
				}
			},
		},
		{
			Title: "GET Parameter passing case and application deployed on the intranet",
			Path:  "/get-in-query-intranet",
			Handler: func(writer http.ResponseWriter, request *http.Request) {
				if request.Method == http.MethodGet {
					action := request.URL.Query().Get("action")
					if action == "" {
						writer.Write([]byte(utils2.Format(string(fastjson_loginPage), map[string]string{
							"script": `function load(){
            name=$("#username").val();
            password=$("#password").val();
            auth = {"user":name,"password":password};
            $.ajax({
                type:"get",
                url:"/fastjson/json-in-query",
                data:{"auth":JSON.stringify(auth),"action":"login"},
                success: function (data ,textStatus, jqXHR)
                {
                    $("#response").text(JSON.stringify(data));
					console.log(data);
                },
                error:function (XMLHttpRequest, textStatus, errorThrown) {      
                    alert("Request error");
                },
            })
        }`,
						})))
						return
					}
					auth := request.URL.Query().Get("auth")
					if auth == "" {
						writer.Write([]byte("auth parameter cannot be empty"))
						return
					}
					response := mockController(generateFastjsonParser("intranet"), request, auth)
					writer.Write([]byte(response))
				} else {
					writer.WriteHeader(http.StatusMethodNotAllowed)
				}
			},
		},
	}
	for _, v := range vuls {
		addRouteWithVulInfo(fastjsonGroup, v)
	}
}
