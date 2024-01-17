package yakgrpc

import (
	"context"
	"github.com/yaklang/yaklang/common/consts"
	"github.com/yaklang/yaklang/common/utils"
	"github.com/yaklang/yaklang/common/yakgrpc/yakit"
	"github.com/yaklang/yaklang/common/yakgrpc/ypb"
	"strconv"
	"sync"
)

func (s *Server) UpdateFromYakitResource(ctx context.Context, req *ypb.UpdateFromYakitResourceRequest) (*ypb.Empty, error) {
	err := yakit.UpdateYakitStore(s.GetProfileDatabase(), req.GetBaseSourceUrl())
	if err != nil {
		return nil, err
	}
	return &ypb.Empty{}, nil
}

func (s *Server) UpdateFromGithub(ctx context.Context, req *ypb.UpdateFromGithubRequest) (*ypb.Empty, error) {
	return nil, utils.Errorf("not implemeted")
}

func (s *Server) GetKey(ctx context.Context, req *ypb.GetKeyRequest) (*ypb.GetKeyResult, error) {
	// patch
	switch req.GetKey() {
	case "YAKIT_MITMDefaultDnsServers":
		patched := req.GetKey() + "_PATCHED"
		result, _ := strconv.ParseBool(yakit.GetKey(s.GetProfileDatabase(), patched))
		if !result {
			yakit.SetKey(s.GetProfileDatabase(), req.GetKey(), "[]")
			yakit.SetKey(s.GetProfileDatabase(), patched, "true")
		}
	}
	result := yakit.GetKey(s.GetProfileDatabase(), req.GetKey())
	return &ypb.GetKeyResult{
		Value: utils.EscapeInvalidUTF8Byte([]byte(result)),
	}, nil
}

func (s *Server) SetKey(ctx context.Context, req *ypb.SetKeyRequest) (*ypb.Empty, error) {
	if req.GetTTL() > 0 {
		err := yakit.SetKeyWithTTL(s.GetProfileDatabase(), req.GetKey(), req.GetValue(), int(req.GetTTL()))
		if err != nil {
			return nil, err
		}
	} else {
		err := yakit.SetKey(s.GetProfileDatabase(), req.GetKey(), req.GetValue())
		if err != nil {
			return nil, err
		}
	}
	return &ypb.Empty{}, nil
}

type envBuildin struct {
	Key     string
	Value   string
	Verbose string
}

var processEnv = []*envBuildin{
	{Key: "YAKIT_DINGTALK_WEBHOOK", Verbose: "Set up the DingTalk robot webhook, which can be used to receive information such as vulnerabilities"},
	{Key: "YAKIT_DINGTALK_SECRET", Verbose: "Set up Ding Password (SecretKey) of Dingbot Webhook"},
	{Key: "YAKIT_WORKWX_WEBHOOK", Verbose: "sets the Feishu Bot Webhook address, which can be used to receive information such as vulnerabilities."},
	{Key: "YAKIT_WORKWX_SECRET", Verbose: "Set the password for the enterprise WeChat robot webhook (SecretKey)"},
	{Key: "YAKIT_FEISHU_WEBHOOK", Verbose: "Set the Feishu Bot Webhook address, which can be used to receive vulnerability and other information"},
	{Key: "YAKIT_FEISHU_SECRET", Verbose: "Set the password for Feishu Bot Webhook address (SecretKey)"},
	{Key: "YAK_PROXY", Verbose: "Set up the proxy configuration of the Yaklang engine."},
	{Key: consts.CONST_YAK_EXTRA_DNS_SERVERS, Verbose: "Set an additional DNS server for the Yaklang engine (comma separated)"},
	{Key: consts.CONST_YAK_OVERRIDE_DNS_SERVERS, Verbose: "Do you want to use user-configured DNS to overwrite the original DNS? (true/false）"},
}
var onceInitProcessEnv = new(sync.Once)

func (s *Server) GetAllProcessEnvKey(ctx context.Context, req *ypb.Empty) (*ypb.GetProcessEnvKeyResult, error) {
	var result []*ypb.GeneralStorage

	onceInitProcessEnv.Do(func() {
		for _, k := range processEnv {
			yakit.InitKey(s.GetProfileDatabase(), k.Key, k.Verbose, true)
		}
	})

	for _, k := range yakit.GetProcessEnvKey(s.GetProfileDatabase()) {
		if k.Key == "" || k.Key == `""` {
			continue
		}
		result = append(result, k.ToGRPCModel())
	}
	return &ypb.GetProcessEnvKeyResult{Results: result}, nil
}

func (s *Server) SetProcessEnvKey(ctx context.Context, req *ypb.SetKeyRequest) (*ypb.Empty, error) {
	if req.GetKey() == "" {
		return nil, utils.Errorf("empty key")
	}
	_, err := s.SetKey(ctx, req)
	if err != nil {
		return nil, err
	}
	yakit.SetKeyProcessEnv(s.GetProfileDatabase(), req.GetKey(), true)
	yakit.RefreshProcessEnv(s.GetProfileDatabase())
	return &ypb.Empty{}, nil
}

func (s *Server) DelKey(ctx context.Context, req *ypb.GetKeyRequest) (*ypb.Empty, error) {
	key, err := yakit.GetKeyModel(s.GetProfileDatabase(), req.GetKey())
	if err != nil {
		return nil, err
	}

	if key.ProcessEnv {
		s.SetProcessEnvKey(ctx, &ypb.SetKeyRequest{
			Key: req.GetKey(), Value: "",
		})
	}
	yakit.DelKey(s.GetProfileDatabase(), req.GetKey())

	return &ypb.Empty{}, nil
}

func (s *Server) GetProjectKey(ctx context.Context, req *ypb.GetKeyRequest) (*ypb.GetKeyResult, error) {
	result := yakit.GetProjectKey(s.GetProjectDatabase(), req.GetKey())
	return &ypb.GetKeyResult{
		Value: utils.EscapeInvalidUTF8Byte([]byte(result)),
	}, nil
}

func (s *Server) SetProjectKey(ctx context.Context, req *ypb.SetKeyRequest) (*ypb.Empty, error) {
	if req.GetTTL() > 0 {
		err := yakit.SetProjectKeyWithTTL(s.GetProjectDatabase(), req.GetKey(), req.GetValue(), int(req.GetTTL()))
		if err != nil {
			return nil, err
		}
	} else {
		err := yakit.SetProjectKey(s.GetProjectDatabase(), req.GetKey(), req.GetValue())
		if err != nil {
			return nil, err
		}
	}
	return &ypb.Empty{}, nil
}
