package coreplugin

import (
	"embed"
	"fmt"
	"sync"

	"github.com/yaklang/yaklang/common/log"
)

//go:embed base-yak-plugin/*
var basePlugin embed.FS

type PlugInfo struct {
	PlugName    string
	BinDataPath string
}

var initDB = sync.Once{}

func GetCorePluginData(name string) []byte {
	codeBytes, err := basePlugin.ReadFile(fmt.Sprintf("base-yak-plugin/%v.yak", name))
	if err != nil {
		log.Errorf("%v is not a core plugin", name)
		return nil
	}
	return codeBytes
}
