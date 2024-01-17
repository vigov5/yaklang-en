package yakgrpc

import (
	"context"
	"github.com/yaklang/yaklang/common/consts"
	"github.com/yaklang/yaklang/common/utils"
	"github.com/yaklang/yaklang/common/utils/pcapfix"
	"github.com/yaklang/yaklang/common/yakgrpc/ypb"
	"os"
	"runtime"
)

func (s *Server) ResetAndInvalidUserData(ctx context.Context, req *ypb.ResetAndInvalidUserDataRequest) (*ypb.Empty, error) {
	os.RemoveAll(consts.GetDefaultYakitBaseTempDir())
	for _, table := range YakitAllTables {
		s.GetProjectDatabase().Unscoped().DropTableIfExists(table)
	}
	for _, table := range YakitProfileTables {
		s.GetProfileDatabase().Unscoped().DropTableIfExists(table)
	}
	os.Exit(1)
	return &ypb.Empty{}, nil
}

func (s *Server) IsPrivilegedForNetRaw(ctx context.Context, req *ypb.Empty) (*ypb.IsPrivilegedForNetRawResponse, error) {
	if runtime.GOOS == "windows" {
		return &ypb.IsPrivilegedForNetRawResponse{
			IsPrivileged:  pcapfix.IsPrivilegedForNetRaw(),
			Advice:        "use administrator privileges for opening yak.exe or yakit",
			AdviceVerbose: "Use administrator rights to open Yakit or yak.exe",
		}, nil
	}
	return &ypb.IsPrivilegedForNetRawResponse{
		IsPrivileged:  pcapfix.IsPrivilegedForNetRaw(),
		Advice:        "use pcapfix.Fix or Yakit FixPcapPermission to fix this;",
		AdviceVerbose: "Use pcapfix.Fix or Yakit to fix the original network card permission operation",
	}, nil
}

func (s *Server) PromotePermissionForUserPcap(ctx context.Context, req *ypb.Empty) (*ypb.Empty, error) {
	err := pcapfix.Fix()
	if err != nil {
		return nil, utils.Errorf("call pcapfix.Fix error: %s", err)
	}
	return &ypb.Empty{}, nil
}
