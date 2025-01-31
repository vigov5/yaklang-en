package yakgrpc

import (
	"context"
	"github.com/yaklang/yaklang/common/utils"
	"github.com/yaklang/yaklang/common/yakgrpc/yakit"
	"github.com/yaklang/yaklang/common/yakgrpc/ypb"
)

const (
	YAK_ENGINE_DEFAULT_SCAN_PROXY = "YAK_ENGINE_DEFAULT_SCAN_PROXY"
)

func (s *Server) GetEngineDefaultProxy(ctx context.Context, e *ypb.Empty) (*ypb.DefaultProxyResult, error) {
	return &ypb.DefaultProxyResult{Proxy: yakit.GetKey(s.GetProfileDatabase(), YAK_ENGINE_DEFAULT_SCAN_PROXY)}, nil
}

func (s *Server) SetEngineDefaultProxy(ctx context.Context, d *ypb.DefaultProxyResult) (*ypb.Empty, error) {
	var err = yakit.SetKey(s.GetProfileDatabase(), YAK_ENGINE_DEFAULT_SCAN_PROXY, d.GetProxy())
	if err != nil {
		return nil, utils.Errorf("Setting engine default scanning proxy failed")
	}
	return &ypb.Empty{}, nil
}
