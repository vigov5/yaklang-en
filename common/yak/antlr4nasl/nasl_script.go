package antlr4nasl

import (
	"encoding/json"
	"github.com/yaklang/yaklang/common/consts"
	"github.com/yaklang/yaklang/common/log"
	"github.com/yaklang/yaklang/common/utils"
	"github.com/yaklang/yaklang/common/yakgrpc/yakit"
	"sync"
)

type NaslKBs struct {
	data map[string]interface{}
	mux  *sync.RWMutex
}

func NewNaslKBs() *NaslKBs {
	return &NaslKBs{
		data: make(map[string]interface{}),
		mux:  &sync.RWMutex{},
	}
}
func (n *NaslKBs) GetData() map[string]interface{} {
	return n.data
}
func (n *NaslKBs) AddKB(name string, value interface{}) error {
	n.mux.Lock()
	if v, ok := n.data[name]; ok {
		if v1, ok := v.([]interface{}); ok {
			v1 = append(v1, value)
		} else {
			return utils.Errorf("KB %s is not array", name)
		}
	}
	n.mux.Unlock()
	return n.SetKB(name, []interface{}{value})
}
func (n *NaslKBs) SetKB(name string, value interface{}) error {
	n.mux.Lock()
	defer n.mux.Unlock()
	n.data[name] = value
	return nil
}
func (n *NaslKBs) GetKB(name string) interface{} {
	n.mux.Lock()
	defer n.mux.Unlock()
	if v, ok := n.data[name]; ok {
		return v
	}
	return nil
}
func (n *NaslKBs) GetKBByPattern(name string) (res map[string]interface{}) {
	n.mux.Lock()
	defer n.mux.Unlock()
	res = make(map[string]interface{})
	for k, v := range n.data {
		if utils.MatchAllOfGlob(k, name) {
			res[k] = v
		}
	}
	return
}

type NaslVhost struct {
	Hostname string
	Source   string
}
type NaslScriptInfo struct {
	naslScript     *yakit.NaslScript
	OriginFileName string
	Hash           string
	OID            string
	CVE            []string
	ScriptName     string
	Script         string
	Tags           map[string]interface{}
	Version        string
	Category       string
	Family         string
	Copyright      string
	Dependencies   []string // Dependency script

	Xrefs           map[string]string
	Preferences     map[string]interface{}
	RequirePorts    []string // precondition assertion
	RequireKeys     []string // precondition assertion
	ExcludeKeys     []string // precondition assertion
	RequireUdpPorts []string // precondition assertion
	BugtraqId       []int
	MandatoryKeys   []string // precondition assertion
	Timeout         int      // milliseconds
	RecvTimeout     int      // milliseconds
	Vhosts          []*NaslVhost
	Ip              string
}

func NewNaslScriptObject() *NaslScriptInfo {
	return &NaslScriptInfo{
		naslScript:  yakit.NewEmptyNaslScript(),
		Tags:        make(map[string]interface{}),
		Xrefs:       make(map[string]string),
		Preferences: make(map[string]interface{}),
	}
}
func NewNaslScriptObjectFromDb(originName string) (*NaslScriptInfo, error) {
	db := consts.GetGormProfileDatabase()
	if db == nil {
		return nil, utils.Errorf("gorm database is nil")
	}
	var scripts []*yakit.NaslScript
	if err := db.Where("origin_file_name = ?", originName).First(&scripts).Error; err != nil {
		log.Error(err)
		return nil, err
	}
	if len(scripts) > 0 {
		return NewNaslScriptObjectFromNaslScript(scripts[0]), nil
	}
	return nil, utils.Errorf("script %s not found", originName)
}
func NewNaslScriptObjectFromFile(path string) (*NaslScriptInfo, error) {
	e := New()
	e.InitBuildInLib()
	e.SetDescription(true)
	err := e.RunFile(path)
	if err != nil {
		return nil, err
	}
	return e.GetScriptObject(), nil
}
func NewNaslScriptObjectFromNaslScript(s *yakit.NaslScript) *NaslScriptInfo {
	info := NewNaslScriptObject()
	n := func() error {
		if err := json.Unmarshal([]byte(s.Tags), &info.Tags); err != nil {
			return err
		}
		if err := json.Unmarshal([]byte(s.Dependencies), &info.Dependencies); err != nil {
			return err
		}
		if err := json.Unmarshal([]byte(s.RequirePorts), &info.RequirePorts); err != nil {
			return err
		}
		if err := json.Unmarshal([]byte(s.RequireKeys), &info.RequireKeys); err != nil {
			return err
		}
		if err := json.Unmarshal([]byte(s.ExcludeKeys), &info.ExcludeKeys); err != nil {
			return err
		}
		if err := json.Unmarshal([]byte(s.Xref), &info.Xrefs); err != nil {
			return err
		}
		if err := json.Unmarshal([]byte(s.Preferences), &info.Preferences); err != nil {
			return err
		}
		if err := json.Unmarshal([]byte(s.RequireUdpPorts), &info.RequireUdpPorts); err != nil {
			return err
		}
		if err := json.Unmarshal([]byte(s.MandatoryKeys), &info.MandatoryKeys); err != nil {
			return err
		}
		if err := json.Unmarshal([]byte(s.BugtraqId), &info.BugtraqId); err != nil {
			return err
		}
		if err := json.Unmarshal([]byte(s.CVE), &info.CVE); err != nil {
			return err
		}
		return nil
	}
	n()
	info.OriginFileName = s.OriginFileName
	info.Hash = s.Hash
	info.OID = s.OID
	info.ScriptName = s.ScriptName
	info.Script = s.Script
	info.Version = s.Version
	info.Category = s.Category
	info.Family = s.Family
	info.Copyright = s.Copyright
	info.Timeout = s.Timeout
	info.naslScript = s
	return info
}
func (n *NaslScriptInfo) Save() error {
	tagMarshal, err := json.Marshal(n.Tags)
	if err != nil {
		return err
	}
	dependenciesMarshal, err := json.Marshal(n.Dependencies)
	if err != nil {
		return err
	}
	portsMarshal, err := json.Marshal(n.RequirePorts)
	if err != nil {
		return err
	}
	excludeKeysMarshal, err := json.Marshal(n.ExcludeKeys)
	if err != nil {
		return err
	}
	xrefsMarshal, err := json.Marshal(n.Xrefs)
	if err != nil {
		return err
	}
	preferencesMarshal, err := json.Marshal(n.Preferences)
	if err != nil {
		return err
	}
	RequireUdpPorts, err := json.Marshal(n.RequireUdpPorts)
	if err != nil {
		return err
	}
	MandatoryKeys, err := json.Marshal(n.MandatoryKeys)
	if err != nil {
		return err
	}
	BugtraqId, err := json.Marshal(n.BugtraqId)
	if err != nil {
		return err
	}
	cves, err := json.Marshal(n.CVE)
	if err != nil {
		return err
	}
	RequireKeys, err := json.Marshal(n.RequireKeys)
	if err != nil {
		return err
	}
	n.naslScript.Script = n.Script
	n.naslScript.ScriptName = n.ScriptName
	n.naslScript.Tags = string(tagMarshal)
	n.naslScript.Version = n.Version
	n.naslScript.Category = n.Category
	n.naslScript.Family = n.Family
	n.naslScript.Copyright = n.Copyright
	n.naslScript.Dependencies = string(dependenciesMarshal)
	n.naslScript.RequirePorts = string(portsMarshal)
	n.naslScript.ExcludeKeys = string(excludeKeysMarshal)
	n.naslScript.Hash = n.Hash
	n.naslScript.OID = n.OID
	n.naslScript.CVE = string(cves)
	n.naslScript.Xref = string(xrefsMarshal)
	n.naslScript.Preferences = string(preferencesMarshal)
	n.naslScript.RequireUdpPorts = string(RequireUdpPorts)
	n.naslScript.BugtraqId = string(BugtraqId)
	n.naslScript.MandatoryKeys = string(MandatoryKeys)
	n.naslScript.Timeout = n.Timeout
	n.naslScript.RequireKeys = string(RequireKeys)
	n.naslScript.OriginFileName = n.OriginFileName
	return n.naslScript.CreateOrUpdateNaslScript(consts.GetGormProfileDatabase())
}
