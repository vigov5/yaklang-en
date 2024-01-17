package wsm

import (
	"github.com/yaklang/yaklang/common/wsm/payloads/behinder"
)

var WebShellExports = map[string]interface{}{
	"NewWebshell": NewWebShell,

	// Set shell information
	"tools":       SetShellType,
	"setProxy":    SetProxy,
	"useBehinder": SetBeinderTool,
	"useGodzilla": SetGodzillaTool,
	"useBase64":   SetBase64Aes,
	"useRaw":      SetRawAes,
	"script":      SetShellScript,
	"secretKey":   SetSecretKey,
	"passParams":  SetPass,

	// set parameters
	"cmdPath": behinder.SetCommandPath,
}
