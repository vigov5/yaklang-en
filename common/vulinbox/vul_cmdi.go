package vulinbox

import (
	"context"
	"fmt"
	"github.com/google/shlex"
	"github.com/yaklang/yaklang/common/utils"
	"net/http"
	"os/exec"
	"time"
)

func (s *VulinServer) registerPingCMDI() {
	r := s.router

	cmdIGroup := r.PathPrefix("/exec").Name("Command injection test case (Unsafe Mode)").Subrouter()
	cmdIRoutes := []*VulInfo{
		{
			DefaultQuery: "ip=127.0.0.1",
			Path:         "/ping/shlex",
			Title:        "Shlex parsed command injection",
			Handler: func(writer http.ResponseWriter, request *http.Request) {
				ip := request.URL.Query().Get("ip")
				if ip == "" {
					writer.Write([]byte(`no ip set`))
					return
				}
				var raw = fmt.Sprintf("ping %v", ip)
				list, err := shlex.Split(raw)
				if err != nil {
					writer.Write([]byte(`shlex parse failed: ` + err.Error()))
					return
				}
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				outputs, err1 := exec.CommandContext(ctx, list[0], list[1:]...).CombinedOutput()
				// Attempt to convert GBK to UTF-8
				utf8Outputs, err2 := utils.GbkToUtf8(outputs)
				if err2 != nil {
					writer.Write(outputs)
				} else {
					writer.Write(utf8Outputs)
				}
				if err1 != nil {
					writer.Write([]byte("exec : " + err1.Error()))
					return
				}
			},
			RiskDetected: false,
		},
		{
			DefaultQuery: "ip=127.0.0.1",
			Path:         "/ping/bash",
			Title:        "Bash parsed command injection",
			Handler: func(writer http.ResponseWriter, request *http.Request) {
				ip := request.URL.Query().Get("ip")
				if ip == "" {
					writer.Write([]byte(`no ip set`))
					return
				}
				var raw = fmt.Sprintf("ping %v", ip)
				ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
				defer cancel()
				outputs, err1 := exec.CommandContext(ctx, `bash`, "-c", raw).CombinedOutput()
				// Attempt to convert GBK to UTF-8
				utf8Outputs, err2 := utils.GbkToUtf8(outputs)
				if err2 != nil {
					writer.Write(outputs)
				} else {
					writer.Write(utf8Outputs)
				}
				if err1 != nil {
					writer.Write([]byte("exec : " + err1.Error()))
					return
				}
			},
			RiskDetected: true,
		},
	}

	for _, v := range cmdIRoutes {
		addRouteWithVulInfo(cmdIGroup, v)
	}
}
